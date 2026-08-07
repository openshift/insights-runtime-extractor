use caps::errors::CapsError;
use caps::{has_cap, CapSet, Capability};
use log::{debug, info};
use nix::unistd::{setgroups, setresgid, setresuid, Gid, Uid};

use crate::ScannerError;

pub fn check_privileged_perms() -> Result<(), ScannerError> {
    info!("🔏  Checking permissions required to scan the containers and their processes");

    let required = [
        Capability::CAP_SYS_ADMIN,
        Capability::CAP_SYS_PTRACE,
        Capability::CAP_SETUID,
        Capability::CAP_SETGID,
        Capability::CAP_SETPCAP,
    ];

    let all_present = required
        .iter()
        .all(|cap| has_cap(None, CapSet::Effective, *cap).unwrap_or(false));

    match all_present {
        true => Ok(()),

        false => Err(ScannerError::Caps(CapsError::from(
            "Must have CAP_SYS_ADMIN, CAP_SYS_PTRACE, CAP_SETUID, CAP_SETGID, and CAP_SETPCAP permissions",
        ))),
    }
}

pub fn check_no_privileged_perms() -> Result<(), ScannerError> {
    debug!("🔏  Checking permissions");

    for cap_set in [CapSet::Effective, CapSet::Permitted, CapSet::Inheritable, CapSet::Ambient] {
        let caps = caps::read(None, cap_set)
            .map_err(|e| ScannerError::Caps(e))?;
        if !caps.is_empty() {
            return Err(ScannerError::Caps(CapsError::from(
                "All capability sets must be empty after privilege drop",
            )));
        }
    }

    Ok(())
}

pub fn drop_privileges(uid: u32, gid: u32) -> Result<(), ScannerError> {
    debug!("🧑‍💻 Dropping privileges to uid={uid} gid={gid}");
    let uid = Uid::from_raw(uid);
    let gid = Gid::from_raw(gid);

    // Order matters: groups/gid before uid — once uid is non-zero we lose
    // the capability to change groups.
    setgroups(&[gid]).map_err(|e| ScannerError::Errno(e))?;
    setresgid(gid, gid, gid).map_err(|e| ScannerError::Errno(e))?;
    // Clear Bounding set before setresuid — requires CAP_SETPCAP in Effective.
    caps::clear(None, caps::CapSet::Bounding)
        .map_err(|e| ScannerError::Caps(e))?;
    setresuid(uid, uid, uid).map_err(|e| ScannerError::Errno(e))?;

    // For non-zero uid, the kernel clears Permitted/Effective automatically.
    // For uid 0 (root workloads), setresuid is a no-op and caps survive.
    // Clear all cap sets explicitly so it works regardless of target uid.
    caps::clear(None, caps::CapSet::Permitted)
        .map_err(|e| ScannerError::Caps(e))?;
    caps::clear(None, caps::CapSet::Effective)
        .map_err(|e| ScannerError::Caps(e))?;
    caps::clear(None, caps::CapSet::Inheritable)
        .map_err(|e| ScannerError::Caps(e))?;
    caps::clear(None, caps::CapSet::Ambient)
        .map_err(|e| ScannerError::Caps(e))?;

    nix::sys::prctl::set_no_new_privs().map_err(|e| ScannerError::Errno(e))?;

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use nix::sys::wait::{waitpid, WaitStatus};
    use nix::unistd::{fork, geteuid, getgid, getgroups, getuid, seteuid, ForkResult, Uid};

    fn run_in_fork(f: fn() -> Vec<&'static str>) {
        assert!(getuid().is_root(), "test must run as root — use make privileged-tests");

        match unsafe { fork() }.expect("fork failed") {
            ForkResult::Parent { child, .. } => {
                match waitpid(child, None).expect("waitpid failed") {
                    WaitStatus::Exited(_, 0) => {}
                    WaitStatus::Exited(_, code) => {
                        panic!("privilege drop is incomplete — {code} check(s) failed, see stderr");
                    }
                    status => panic!("unexpected child status: {status:?}"),
                }
            }
            ForkResult::Child => {
                let failures = f();
                if failures.is_empty() {
                    std::process::exit(0);
                }
                for msg in &failures {
                    eprintln!("FAIL: {msg}");
                }
                std::process::exit(failures.len() as i32);
            }
        }
    }

    /// Verifies privilege drop to a non-root uid (65534/nobody).
    /// Must run as root on Linux.
    #[test]
    #[ignore]
    fn privilege_drop_to_non_root() {
        run_in_fork(check_drop_to_non_root);
    }

    fn check_drop_to_non_root() -> Vec<&'static str> {
        let mut failures = Vec::new();

        if drop_privileges(65534, 65534).is_err() {
            failures.push("drop_privileges(65534, 65534) failed");
            return failures;
        }

        if getuid().is_root() {
            failures.push("real UID is still 0");
        }

        if getgid().as_raw() != 65534 {
            failures.push("real GID is not 65534");
        }

        if let Ok(groups) = getgroups() {
            if groups != vec![Gid::from_raw(65534)] {
                failures.push("supplementary groups not set to [65534]");
            }
        }

        if check_no_privileged_perms().is_err() {
            failures.push("capabilities not fully cleared");
        }

        if !nix::sys::prctl::get_no_new_privs().unwrap_or(false) {
            failures.push("PR_SET_NO_NEW_PRIVS not set");
        }

        if seteuid(Uid::from_raw(0)).is_ok() && geteuid().is_root() {
            failures.push("recovered root via seteuid(0)");
        }

        failures
    }

    /// Verifies privilege drop when the target workload runs as root (uid 0).
    /// setresuid(0,0,0) is a no-op so caps must be cleared explicitly.
    /// Must run as root on Linux.
    #[test]
    #[ignore]
    fn privilege_drop_to_root() {
        run_in_fork(check_drop_to_root);
    }

    fn check_drop_to_root() -> Vec<&'static str> {
        let mut failures = Vec::new();

        if drop_privileges(0, 0).is_err() {
            failures.push("drop_privileges(0, 0) failed");
            return failures;
        }

        if getgid().as_raw() != 0 {
            failures.push("real GID is not 0");
        }

        if let Ok(groups) = getgroups() {
            if groups != vec![Gid::from_raw(0)] {
                failures.push("supplementary groups not set to [0]");
            }
        }

        if check_no_privileged_perms().is_err() {
            failures.push("capabilities not fully cleared after drop to uid 0");
        }

        if !nix::sys::prctl::get_no_new_privs().unwrap_or(false) {
            failures.push("PR_SET_NO_NEW_PRIVS not set after drop to uid 0");
        }

        failures
    }
}
