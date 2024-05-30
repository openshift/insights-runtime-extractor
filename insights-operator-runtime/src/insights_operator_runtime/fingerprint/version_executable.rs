use log::debug;

use super::FingerPrint;
use crate::config::Config;
use crate::insights_operator_runtime::ContainerProcess;

pub struct VersionExecutable {}

impl FingerPrint for VersionExecutable {
    fn can_apply_to(&self, config: &Config, process: &ContainerProcess) -> Option<Vec<String>> {
        debug!(
            "Checking if {:#?} is an executable with that has a `--version`",
            { &process.name }
        );

        let fpr_kind_executable = String::from("./fpr_kind_executable");
        let outdir = format!("out/{}", process.pid);

        if let Some(version_executable) = config
            .fingerprints
            .versioned_executables
            .iter()
            .find(|c| c.process_names.contains(&process.name))
        {
            return Some(vec![
                fpr_kind_executable,
                String::from(&process.command_line[0]),
                String::from(&version_executable.runtime_kind_name),
                outdir,
            ]);
        } else if process.command_line[0].contains("java") {
            // JAVA_HOME env var can not be set
            let no_java_home = "".to_string();
            let java_home = process.environ.get("JAVA_HOME").unwrap_or(&no_java_home);
            return Some(vec![
                String::from("./fpr_java_version"),
                outdir,
                process.environ.get("PATH").unwrap().to_string(),
                java_home.to_string(),
            ]);
        }

        None
    }
}

pub fn is_version_executable(process: &ContainerProcess) -> bool {
    let proc_name = &process.name;
    let proc_cmd = &process.command_line[0];

    proc_name.ends_with("node") || proc_cmd.contains("python") || proc_cmd.contains("java")
}
