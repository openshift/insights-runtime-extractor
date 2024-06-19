use log::{debug, warn};
use std::process::Command;

use crate::config::Config;
use crate::insights_operator_runtime::ContainerProcess;

mod java;
mod native_executable;
mod os;
mod version_executable;

trait FingerPrint {
    fn can_apply_to(
        &self,
        config: &Config,
        out_dir: &String,
        process: &ContainerProcess,
    ) -> Option<Vec<String>>;
}

fn fingerprints() -> Vec<Box<dyn FingerPrint>> {
    vec![
        Box::new(os::Os {}),
        Box::new(version_executable::VersionExecutable {}),
        Box::new(java::Java {}),
        Box::new(native_executable::NativeExecutable {}),
    ]
}

pub fn run_fingerprints(config: &Config, out_dir: &String, process: &ContainerProcess) {
    debug!("👆 Fingerprinting process {}", &process.pid);

    for fingerprint in fingerprints() {
        if let Some(exec) = fingerprint.can_apply_to(config, &out_dir, &process) {
            debug!("Executing {:?}", &exec);
            if let Some((command, args)) = exec.split_first() {
                let command = Command::new(&command).args(args).output();

                match command {
                    Ok(output) => match output.status.success() {
                        true => {
                            let output = String::from_utf8_lossy(&output.stdout);
                            debug!("{}", output);
                        }
                        false => {
                            let error = String::from_utf8_lossy(&output.stderr);
                            warn!("Command {:#?} failed with error:\n{:#?}", exec, error);
                        }
                    },
                    Err(e) => {
                        // Print the error if command execution fails
                        warn!("Error executing command {:#?}: {:#?}", exec, e);
                    }
                }
            }
        }
    }
}
