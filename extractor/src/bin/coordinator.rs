use clap::Parser;
use log::info;
use std::time::{SystemTime, UNIX_EPOCH};

use insights_runtime_extractor::{config, file, get_containers, perms};

#[derive(Parser, Debug)]
#[command(about, long_about = None)]
struct Args {
    #[arg(
        short,
        long,
        help = "Log level (default is warn) [possible values: debug, info, warn, error]"
    )]
    log_level: Option<String>,

    #[arg(
        help = "Comma-separated list of container IDs to scan. If absent, all containers are scanned"
    )]
    container_ids: Option<String>,
}

fn main() {
    let args = Args::parse();

    let log_level = args.log_level.unwrap_or(String::from("info"));
    env_logger::Builder::from_env(env_logger::Env::default().default_filter_or(log_level)).init();

    info!("Gather runtime information from containers on OpenShift");

    perms::check_privileged_perms().expect("Must have privileged permissions to scan containers");

    let timestamp = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .expect("Get Unix timestamp")
        .subsec_nanos();

    let exec_dir = format!("data/out-{}", timestamp);
    file::create_dir(exec_dir.as_str()).expect("Can not create execution directory");

    let config = config::get_config("/");

    info!(
        "Scanning all containers in execution directory {}",
        &exec_dir
    );

    let container_ids: Vec<String> = args
        .container_ids
        .map(|ids| {
            ids.split(',')
                .map(|id| id.trim().to_string())
                .filter(|id| !id.is_empty())
                .collect()
        })
        .unwrap_or_default();

    let containers = get_containers(container_ids);

    info!("Scanning {} containers", containers.len());

    for container in containers {
        info!(
            "Scanning container 🫙 {} in {}/{}",
            container.id, container.pod_namespace, container.pod_name
        );
        insights_runtime_extractor::scan_container(&config, &exec_dir, &container)
    }

    info!(
        "Scanning DONE. Sending back the path to the execution directory {}",
        &exec_dir
    );

    println!("{}", &exec_dir);
}
