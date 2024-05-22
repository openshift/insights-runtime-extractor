use std::collections::HashMap;

use clap::Parser;
use insights_operator_runtime::{RuntimeInfo, ScannerError};
use log::{info, warn};

#[derive(Parser, Debug)]
#[command(about, long_about = None)]
struct Args {
    #[arg(
        short,
        long,
        help = "Log level (default is warn) [possible values: debug, info, warn, error]"
    )]
    log_level: Option<String>,
}

fn main() -> Result<(), ScannerError> {
    let args = Args::parse();

    let log_level = args.log_level.unwrap_or(String::from("warn"));

    env_logger::Builder::from_env(env_logger::Env::default().default_filter_or(log_level)).init();

    info!("Scanning all containers");
    let containers = insights_operator_runtime::get_containers();

    // keys are pod-namespace, pod-name, container-id
    let mut out: HashMap<String, HashMap<String, HashMap<String, RuntimeInfo>>> = HashMap::new();

    for container in containers {
        info!("Scanning container 🫙 {}", container.id);
        let namespace = out.entry(container.pod_namespace).or_insert(HashMap::new());
        let pod = namespace
            .entry(container.pod_name)
            .or_insert(HashMap::new());
        if let Some(runtime_info) = insights_operator_runtime::scan_container(&container.id) {
            pod.insert(container.id, runtime_info);
        }
    }

    match serde_json::to_string(&out) {
        Err(_err) => warn!("Unable to serialize JSON"),
        Ok(json) => println!("{}", json),
    };

    Ok(())
}
