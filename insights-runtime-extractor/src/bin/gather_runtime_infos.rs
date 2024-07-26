use clap::Parser;
use log::info;
use std::time::{SystemTime, UNIX_EPOCH};

use insights_runtime_extractor::ScannerError;
use insights_runtime_extractor::{config, file};

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

    let config = config::get_config("/");

    let timestamp = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .expect("Get Unix timestamp")
        .subsec_nanos();

    let exec_dir = format!("data/out-{}", timestamp);
    file::create_dir(exec_dir.as_str()).expect("Can not create execution directory");

    info!(
        "Scanning all containers in execution directory {}",
        &exec_dir
    );
    let containers = insights_runtime_extractor::get_containers();

    for container in containers {
        info!("Scanning container 🫙 {}", container.id);
        insights_runtime_extractor::scan_container(&config, &exec_dir, &container);
    }

    println!("{}", &exec_dir);

    Ok(())
}
