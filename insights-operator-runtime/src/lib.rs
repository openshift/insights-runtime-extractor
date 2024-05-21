mod insights_operator_runtime;

pub use crate::insights_operator_runtime::file;

#[derive(Debug)]
pub enum ScannerError {
    Caps(caps::errors::CapsError),
    Errno(nix::errno::Errno),
    String,
}
