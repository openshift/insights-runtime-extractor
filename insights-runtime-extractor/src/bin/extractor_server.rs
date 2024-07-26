use clap::Parser;
use log::{error, info};
use std::fs;
use tempfile::Builder;
use tokio::io::AsyncWriteExt;
use tokio::net::TcpListener;

use insights_runtime_extractor::{config, perms};

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

    let log_level = args.log_level.unwrap_or(String::from("warn"));

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

    println!("Listening on {}", addr);

    loop {
        // Accept an incoming connection
        let (mut socket, _) = listener.accept().await?;

        // Spawn a new task to handle the connection
        tokio::spawn(async move {
            let exec_dir = Builder::new()
                .prefix("out")
                .rand_bytes(6)
                .tempdir_in("/data/")
                .expect("Can not create execution directory");

            // use a relative path for the execution directory as the
            // fingerprints will add files from the orphaned directory
            let relative_exec_dir = exec_dir
                .path()
                .strip_prefix("/")
                .unwrap()
                .to_str()
                .unwrap()
                .to_string();

            info!(
                "Scanning all containers in execution directory {}",
                &relative_exec_dir
            );

            let config = config::get_config("/");
            let containers = insights_runtime_extractor::get_containers();

            for container in containers {
                info!("Scanning container 🫙 {}", container.id);
                insights_runtime_extractor::scan_container(&config, &relative_exec_dir, &container)
            }

            let response = format!("{}", &exec_dir.path().display());

            // Write the response message to the socket
            if let Err(e) = socket.write_all(response.as_bytes()).await {
                error!("Failed to write to socket; err = {:?}", e);
            }
            // Close the socket
            if let Err(e) = socket.shutdown().await {
                error!("Failed to shutdown socket; err = {:?}", e);
            }
        });
    }
}
