pub mod config;
pub mod container;
pub mod file;
mod fingerprint;
pub mod perms;
mod process;

use log::{debug, error, info, trace, warn};
use nix::sched::{setns, CloneFlags};
use nix::sys::wait::{waitpid, WaitStatus};
use nix::unistd::{fork, ForkResult};
use process::ContainerProcess;
use serde::Serialize;
use std::collections::HashMap;
use std::fs::{self, File};
use std::os::fd::{AsFd, AsRawFd, BorrowedFd};
use std::path::Path;
use std::time::Instant;

use crate::config::Config;
use crate::insights_runtime_extractor::container::{get_root_pid, Container};
use crate::ScannerError;

#[derive(Serialize, Debug)]
pub struct RuntimeComponentInfo {
    name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    version: Option<String>,
}

#[derive(Serialize, Debug)]
pub struct RuntimeInfo {
    #[serde(rename = "os-release-id", skip_serializing_if = "Option::is_none")]
    os_id: Option<String>,
    #[serde(
        rename = "os-release-version-id",
        skip_serializing_if = "Option::is_none"
    )]
    os_version_id: Option<String>,
    #[serde(rename = "runtime-kind", skip_serializing_if = "Option::is_none")]
    kind: Option<String>,
    #[serde(
        rename = "runtime-kind-version",
        skip_serializing_if = "Option::is_none"
    )]
    kind_version: Option<String>,
    #[serde(
        rename = "runtime-kind-implementer",
        skip_serializing_if = "Option::is_none"
    )]
    kind_implementer: Option<String>,
    runtimes: Vec<RuntimeComponentInfo>,
}

pub fn scan_container(config: &Config, out: &String, container: &Container) {
    let container_id = match container.id.strip_prefix("cri-o://") {
        Some(id) => id.to_string(),
        None => container.id.to_string(),
    };

    info!(
        "⚙️  Running Container Scanner on container {} ...",
        &container.id
    );

    let current_dir = File::open(".").unwrap();
    debug!("Using {:?} as the orphaned directory", current_dir);

    let root_pid = get_root_pid(&container_id);

    let leaves = process::get_process_leaves(&root_pid);

    // fingerprint only the first process
    info!("🔎  Fingerprinting {} processes...", leaves.len());

    if let Some(process) = leaves.get(0) {
        // create a directory to store this process' fingerprints
        // that is put it under a directory from the executing process so that concurrent
        // execution are stored in separate directories.
        let container_output = format!("{}/{}", out, &process.pid);
        file::create_dir(&container_output).expect(&format!(
            "Can not create output directory for container {}",
            &container.id
        ));

        let mut container_info = HashMap::new();
        container_info.insert(String::from("pod-name"), container.pod_name.clone());
        container_info.insert(
            String::from("pod-namespace"),
            container.pod_namespace.clone(),
        );
        container_info.insert(String::from("container-id"), container.id.clone());
        let _ = file::write_entries(
            Path::new(&container_output),
            "container-info.txt",
            &container_info,
        );

        // copy the config.toml to the pid_output so that it can be read by fingerprints executables
        fs::copy("/config.toml", container_output.clone() + "/config.toml")
            .ok()
            .expect("Copy configuration for fingerprints execution");

        let start = Instant::now();

        if let Err(e) = fork_and_exec(&config, &process, &current_dir, &container_output) {
            error!("Failed to fork and exec fingerprints for container {}: {e:?}", &container.id);
        }

        let duration = start.elapsed().as_millis();
        trace!("🕑 Executed fingerprints in {:?}ms", duration);
    }
}

fn fork_and_exec(
    config: &Config,
    process: &ContainerProcess,
    current_dir: &File,
    out_dir: &String,
) -> Result<(), ScannerError> {
    debug!("Storing fingerprints content in {}", out_dir);

    match unsafe { fork() } {
        Ok(ForkResult::Parent { child, .. }) => {
            // Retry on EINTR to ensure the child is reaped and does not become a zombie.
            loop {
                match waitpid(child, None) {
                    Err(nix::errno::Errno::EINTR) => continue,
                    Err(e) => {
                        warn!("Error: problem waiting for child: {e}");
                        break;
                    }
                    Ok(w) => {
                        match w {
                            WaitStatus::Exited(_, code) if code == 0 => {}
                            WaitStatus::Exited(_, code) if code != 0 => {
                                warn!("Error: problem with child: returned {code}")
                            }
                            _ => warn!("Error: problem with child: {:?}", w),
                        }
                        break;
                    }
                }
            }
            return Ok(());
        }

        Ok(ForkResult::Child) => {
            let start = Instant::now();

            if let Err(e) = join_process_namespaces(process.pid) {
                error!("Failed to join namespaces: {e:?}");
                std::process::exit(1);
            }
            change_dir(current_dir);
            if let Err(e) = perms::drop_privileges(*process.uid, *process.gid) {
                error!("Failed to drop privileges: {e:?}");
                std::process::exit(1);
            }
            if let Err(e) = perms::check_no_privileged_perms() {
                error!("Privilege check failed after drop: {e:?}");
                std::process::exit(1);
            }
            fingerprint::run_fingerprints(&config, out_dir, &process);

            let duration = start.elapsed().as_millis();
            trace!("Child process executed in {:?}ms", duration);

            std::process::exit(0);
        }

        Err(e) => Err(ScannerError::Errno(e)),
    }
}

fn change_dir(dir: &File) {
    nix::unistd::fchdir(dir.as_raw_fd()).unwrap_or_else(|err| {
        error!("Could not change to orphaned directory: {err}");
        std::process::exit(1);
    });
    debug!("Changed directory to {:?}", dir);
}

fn join_process_namespaces(pid: u32) -> Result<(), ScannerError> {
    debug!("📝 Joining pid, mnt, net and cgroup namespaces of {pid}");
    // Not joining the user namespace: sysinfo reads UIDs from the host
    // perspective (/proc/<pid>/status before setns). With K8s user namespaces,
    // the host UID (e.g. 65536) may not be mapped inside the container's user
    // namespace, causing setresuid to fail with EINVAL. The privilege drop
    // (cap clearing + no-new-privs) is sufficient without joining user ns.
    let namespaces: Vec<(String, CloneFlags)> = vec![
        (format!("/proc/{}/ns/pid", pid), CloneFlags::CLONE_NEWPID),
        (format!("/proc/{}/ns/mnt", pid), CloneFlags::CLONE_NEWNS),
        (format!("/proc/{}/ns/net", pid), CloneFlags::CLONE_NEWNET),
        (format!("/proc/{}/ns/cgroup", pid), CloneFlags::CLONE_NEWCGROUP),
    ];

    // Open all namespace files before joining any. Once we join the mnt
    // namespace, /proc reflects the container's view and the host-pid paths
    // above would no longer exist.
    let files: Vec<(String, CloneFlags, File)> = namespaces
        .into_iter()
        .map(|(ns, flags)| {
            let f = File::open(&ns).map_err(|e| {
                error!("Failed to open namespace file {ns}: {e}");
                ScannerError::String
            })?;
            Ok((ns, flags, f))
        })
        .collect::<Result<Vec<_>, ScannerError>>()?;

    for (ns, flags, f) in &files {
        let fd: BorrowedFd<'_> = f.as_fd();
        setns(fd, *flags).map_err(|e| {
            error!("Failed to join namespace {ns}: {e}");
            ScannerError::Errno(e)
        })?;
    }

    Ok(())
}

