pub mod config;
pub mod container;
pub mod file;
mod fingerprint;
pub mod perms;
mod process;

use log::{debug, error, info, trace, warn};
use nix::sched::{setns, CloneFlags};
use nix::sys::wait::{waitpid, WaitStatus};
use nix::unistd::{fork, seteuid, ForkResult};
use process::ContainerProcess;
use serde::Serialize;
use std::fs::{self, File};
use std::os::fd::{AsFd, AsRawFd, BorrowedFd};
use std::path::PathBuf;
use std::time::Instant;

use crate::config::Config;
use crate::file::read_key_value_file;
use crate::insights_operator_runtime::container::get_root_pid;
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

pub fn scan_container(config: &Config, container_id: &String) -> Option<RuntimeInfo> {
    info!(
        "⚙️  Running Container Scanner on container {}...",
        container_id
    );

    let container_id = match container_id.strip_prefix("cri-o://") {
        Some(container_id) => container_id.to_string(),
        None => container_id.to_string(),
    };

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
        let pid_output = format!("{}/{}", std::process::id(), &process.pid);
        file::create_dir(&pid_output).expect(&format!(
            "Can not create output directory for pid {}",
            &process.pid
        ));

        // copy the config.toml to the pid_output so that it can be read by fingerprints executables
        fs::copy("/config.toml", pid_output.clone() + "/config.toml").ok()?;

        let start = Instant::now();

        let _ = fork_and_exec(&config, &process, &current_dir, &pid_output);

        let duration = start.elapsed().as_millis();
        trace!("🕑 Executed fingerprints in {:?}ms", duration);

        let start = Instant::now();

        let mut os_id: Option<String> = None;
        let mut os_version_id: Option<String> = None;
        let mut kind: Option<String> = None;
        let mut kind_version: Option<String> = None;
        let mut kind_implementer: Option<String> = None;
        let mut runtimes: Vec<RuntimeComponentInfo> = Vec::new();

        // read the files written to pid_output by the fingerprint and returns a JSON
        if let Ok(fp_files) = fs::read_dir(PathBuf::from(&pid_output)) {
            for fingerprint in fp_files {
                match fingerprint {
                    Err(_err) => warn!("Unable to read fingerprints"),
                    Ok(fingerprint) if fingerprint.path().is_file() => {
                        if let Some(path) = fingerprint.path().to_str() {
                            if let Ok(fp_entries) = read_key_value_file(path) {
                                match fingerprint.file_name().to_str() {
                                    Some(filename) if filename == "os-fingerprints.txt" => {
                                        os_id = fp_entries.get("os-release-id").cloned();
                                        os_version_id =
                                            fp_entries.get("os-release-version-id").cloned();
                                    }
                                    Some(filename)
                                        if filename == "runtime-kind-fingerprints.txt" =>
                                    {
                                        kind = fp_entries.get("runtime-kind").cloned();
                                        kind_version =
                                            fp_entries.get("runtime-kind-version").cloned();
                                        kind_implementer =
                                            fp_entries.get("runtime-kind-implementer").cloned();
                                    }
                                    Some(_filename) if _filename.ends_with("-fingerprints.txt") => {
                                        for (name, version) in fp_entries.iter() {
                                            let version: Option<String> = if version == "" {
                                                None
                                            } else {
                                                Some(version.to_string())
                                            };
                                            runtimes.push(RuntimeComponentInfo {
                                                name: name.to_string(),
                                                version,
                                            });
                                        }
                                    }
                                    Some(_) => (),
                                    None => (),
                                }
                            }
                        }
                    }
                    Ok(_) => (),
                }
            }
        }

        let duration = start.elapsed().as_millis();
        trace!("🕑 Collected fingerprints files in {:?}ms", duration);

        let info = RuntimeInfo {
            os_id,
            os_version_id,
            kind,
            kind_version,
            kind_implementer,
            runtimes,
        };

        return Some(info);
    }

    None
}

fn fork_and_exec(
    config: &Config,
    process: &ContainerProcess,
    current_dir: &File,
    out_dir: &String,
) -> Result<(), ScannerError> {
    match unsafe { fork() } {
        Ok(ForkResult::Parent { child, .. }) => {
            match waitpid(child, None) {
                Err(e) => warn!("Error: problem waiting for child: {e}"),
                Ok(w) => match w {
                    WaitStatus::Exited(_, code) if code == 0 => {}
                    WaitStatus::Exited(_, code) if code != 0 => {
                        warn!("Error: problem with child: returned {code}")
                    }
                    _ => warn!("Error: problem with child: {:?}", w),
                },
            }
            return Ok(());
        }

        Ok(ForkResult::Child) => {
            let start = Instant::now();

            join_process_namespaces(process.pid)?;
            change_dir(current_dir);
            switch_user(*process.uid)?;
            if *process.uid != 0 {
                perms::check_no_privileged_perms()
                    .expect("Must not have privileged permissions to run fingerprints");
            }
            fingerprint::run_fingerprints(&config, &out_dir, &process);

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
    debug!("📝 Joining pid and mnt namespaces of {pid}");
    let namespaces = vec![
        format!("/proc/{}/ns/pid", pid),
        format!("/proc/{}/ns/mnt", pid),
    ];

    for ns in namespaces {
        let f = File::open(&ns).expect(&format!("Open namespace file {:?}", &ns));
        let fd: BorrowedFd<'_> = f.as_fd();
        setns(fd, CloneFlags::empty()).expect(&format!("join namespace {:?}", &ns));
    }

    Ok(())
}

fn switch_user(uid: u32) -> Result<(), ScannerError> {
    debug!("🧑‍💻 Becoming user: {uid}");
    seteuid(uid.into()).map_err(|e| ScannerError::Errno(e))
}
