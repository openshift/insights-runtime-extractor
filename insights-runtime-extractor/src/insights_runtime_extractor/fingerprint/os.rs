use crate::config::Config;
use crate::insights_runtime_extractor::ContainerProcess;

use super::FingerPrint;

pub struct Os {}

impl FingerPrint for Os {
    fn can_apply_to(
        &self,
        _config: &Config,
        out_dir: &String,
        _process: &ContainerProcess,
    ) -> Option<Vec<String>> {
        Some(vec![String::from("./fpr_os"), out_dir.to_string()])
    }
}
