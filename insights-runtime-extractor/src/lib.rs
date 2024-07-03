mod insights_runtime_extractor;

pub use crate::insights_runtime_extractor::config;
pub use crate::insights_runtime_extractor::container::get_container;
pub use crate::insights_runtime_extractor::container::get_containers;
pub use crate::insights_runtime_extractor::file;
pub use crate::insights_runtime_extractor::perms;
pub use crate::insights_runtime_extractor::scan_container;
pub use crate::insights_runtime_extractor::RuntimeInfo;

#[derive(Debug)]
pub enum ScannerError {
    Caps(caps::errors::CapsError),
    Errno(nix::errno::Errno),
    String,
}
