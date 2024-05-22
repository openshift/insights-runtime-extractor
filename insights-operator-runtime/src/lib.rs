mod insights_operator_runtime;

pub use crate::insights_operator_runtime::container::get_container;
pub use crate::insights_operator_runtime::container::get_containers;
pub use crate::insights_operator_runtime::file;
pub use crate::insights_operator_runtime::perms;
pub use crate::insights_operator_runtime::scan_container;
pub use crate::insights_operator_runtime::RuntimeInfo;

#[derive(Debug)]
pub enum ScannerError {
    Caps(caps::errors::CapsError),
    Errno(nix::errno::Errno),
    String,
}
