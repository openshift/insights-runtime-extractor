use clap::Parser;
use log::{debug, error, info};
use std::fs;
use std::time::{SystemTime, UNIX_EPOCH};
use tokio::io::AsyncWriteExt;
use tokio::net::TcpListener;

use insights_runtime_extractor::{config, file, perms};

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

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = Args::parse();

    let log_level = args.log_level.unwrap_or(String::from("info"));

    env_logger::Builder::from_env(env_logger::Env::default().default_filter_or(log_level)).init();

    info!("Gather runtime information from containers on OpenShift");

    perms::check_privileged_perms().expect("Must have privileged permissions to scan containers");

    // verify that the configuration is properly setup
    let config_content = fs::read_to_string("/config.toml").expect("Configuration file is missing");
    info!("Configuration:\n----\n{}\n----", config_content);
    config::get_config("/");

    // bound to the loopback address so that it can only be contacted
    // by containers in the same pod
    let addr = "127.0.0.1:3000".to_string();

    // Create a TCP listener
    let listener = TcpListener::bind(&addr).await?;

    info!("Listening on {}", addr);

    loop {
        // Accept an incoming connection
        let (mut socket, _) = listener.accept().await?;

        // Spawn a new task to handle the connection
        tokio::spawn(async move {
            debug!("Trigger new runtime info extraction");
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

            let config = config::get_config("/");
            let containers = insights_runtime_extractor::get_containers();

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

            let response = format!("{}\n", &exec_dir);
            // Write the response message to the socket
            if let Err(e) = socket.write_all(response.as_bytes()).await {
                error!("Failed to write to socket; err = {:?}", e);
            }
            // Close the socket
            if let Err(e) = socket.shutdown().await {
                error!("Failed to shutdown socket; err = {:?}", e);
            }

            info!("DONE");
        });
    }
}
