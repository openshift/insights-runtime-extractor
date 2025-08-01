use core::option::Option::{self, None};

use log::debug;

use super::FingerPrint;
use crate::config::Config;
use crate::insights_runtime_extractor::fingerprint::version_executable::VersionExecutable;
use crate::insights_runtime_extractor::ContainerProcess;

pub struct NativeExecutable {}

impl FingerPrint for NativeExecutable {
    fn can_apply_to(
        &self,
        config: &Config,
        out_dir: &String,
        process: &ContainerProcess,
    ) -> Option<Vec<String>> {
        debug!("Checking if {} is a native executable", &process.name);

        let version_exec = VersionExecutable {
        };

        // do not check for native executables if they have a `--version` way
        // to get their versions
        match version_exec.can_apply_to(&config, &out_dir, &process) {
            Some(_) => None,
            None => Some(vec![
                String::from("./fpr_native_executable"),
                out_dir.to_string(),
                process.cwd.as_ref().unwrap().clone(),
                process.command_line.get(0)?.clone(),
            ]),
        }
    }
}
