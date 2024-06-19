use log::debug;

use super::{version_executable, FingerPrint};
use crate::config::Config;
use crate::insights_operator_runtime::ContainerProcess;

pub struct NativeExecutable {}

impl FingerPrint for NativeExecutable {
    fn can_apply_to(
        &self,
        _: &Config,
        out_dir: &String,
        process: &ContainerProcess,
    ) -> Option<Vec<String>> {
        debug!("Checking if {:#?} is a native executable", {
            &process.name
        });

        let fpr_kind_executable = String::from("./fpr_native_executable");

        match !version_executable::is_version_executable(process) {
            false => None,
            true => Some(vec![
                fpr_kind_executable,
                out_dir.to_string(),
                process.cwd.as_ref().unwrap().clone(),
                process.command_line.get(0)?.clone(),
            ]),
        }
    }
}
