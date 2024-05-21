use caps::errors::CapsError;
use caps::{has_cap, CapSet, Capability};
use log::{debug, info};

use crate::ScannerError;

pub fn check_privileged_perms() -> Result<(), ScannerError> {
    info!("🔏  Checking permissions required to scan the containers and their processes");

    let has_suid = has_cap(None, CapSet::Effective, Capability::CAP_SETUID).is_ok_and(|suid| suid);
    let has_sysadmin =
        has_cap(None, CapSet::Effective, Capability::CAP_SYS_ADMIN).is_ok_and(|sysadmin| sysadmin);

    match has_suid && has_sysadmin {
        true => Ok(()),

        false => Err(ScannerError::Caps(CapsError::from(
            "Must have CAP_SETUID and CAP_SYS_ADMIN permissions",
        ))),
    }
}

pub fn check_no_privileged_perms() -> Result<(), ScannerError> {
    debug!("🔏  Checking permissions");

    let has_not_suid =
        has_cap(None, CapSet::Effective, Capability::CAP_SETUID).is_ok_and(|suid| !suid);
    let has_not_sysadmin =
        has_cap(None, CapSet::Effective, Capability::CAP_SYS_ADMIN).is_ok_and(|sysadmin| !sysadmin);

    match has_not_suid && has_not_sysadmin {
        true => Ok(()),

        false => Err(ScannerError::Caps(CapsError::from(
            "Must not have CAP_SETUID and CAP_SYS_ADMIN permissions",
        ))),
    }
}
