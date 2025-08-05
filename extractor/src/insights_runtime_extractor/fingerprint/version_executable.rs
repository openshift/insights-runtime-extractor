use log::debug;

use super::FingerPrint;
use crate::config::Config;
use crate::insights_runtime_extractor::ContainerProcess;

pub struct VersionExecutable {}

impl FingerPrint for VersionExecutable {
    fn can_apply_to(
        &self,
        config: &Config,
        out_dir: &String,
        process: &ContainerProcess,
    ) -> Option<Vec<String>> {
        debug!(
            "Checking if {} is an executable with that has a `--version`",
            &process.name
        );

        let fpr_kind_executable = String::from("./fpr_kind_executable");

        if let Some(version_executable) = config
            .fingerprints
            .versioned_executables
            .iter()
            .find(|c| c.process_names.contains(&process.name))
        {
            return Some(vec![
                fpr_kind_executable,
                out_dir.to_string(),
                String::from(&process.command_line[0]),
                String::from(&version_executable.runtime_kind_name),
            ]);
        } else if process.command_line[0].contains("java") {
            // JAVA_HOME env var can not be set
            let no_java_home = "".to_string();
            let java_home = process.environ.get("JAVA_HOME").unwrap_or(&no_java_home);
            return Some(vec![
                String::from("./fpr_java_version"),
                out_dir.to_string(),
                process.environ.get("PATH").unwrap().to_string(),
                java_home.to_string(),
            ]);
        }

        None
    }
}
