mod insights_operator_runtime;

pub use crate::insights_operator_runtime::file;
pub use crate::insights_operator_runtime::perms;

#[derive(Debug)]
pub enum ScannerError {
    Caps(caps::errors::CapsError),
    Errno(nix::errno::Errno),
    String,
}
