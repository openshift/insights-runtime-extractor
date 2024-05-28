use crate::config::Config;
use crate::insights_operator_runtime::ContainerProcess;

use super::FingerPrint;

pub struct Os {}

impl FingerPrint for Os {
    fn can_apply_to(&self, _config: &Config, process: &ContainerProcess) -> Option<Vec<String>> {
        Some(vec![
            String::from("./fpr_os"),
            format!("out/{}", process.pid),
        ])
    }
}
