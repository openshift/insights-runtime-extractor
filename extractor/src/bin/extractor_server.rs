use clap::Parser;
use log::{debug, error, info};
use std::fs;
use std::io::Write;
use std::net::{Shutdown, TcpListener, TcpStream};
use std::process::Command;
use std::thread;

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

fn main() {
    let args = Args::parse();

    let log_level = args.log_level.unwrap_or(String::from("info"));

    env_logger::Builder::from_env(env_logger::Env::default().default_filter_or(log_level)).init();

    info!("Gather runtime information from containers on OpenShift");

    perms::check_privileged_perms().expect("Must have privileged permissions to scan containers");

    // verify that the configuration is properly setup
    let config_content = fs::read_to_string("/config.toml").expect("Configuration file is missing");
    info!("Configuration:\n----\n{}\n----", config_content);
    config::get_config("/");

    // Create a TCP listener
    // bound to the loopback address so that it can only be contacted
    // by containers in the same pod
    let addr = "127.0.0.1:3000";
    let listener = TcpListener::bind(addr).expect("Failed to bind to address");

    info!("Listening on {}", addr);

    for stream in listener.incoming() {
        match stream {
            Ok(stream) => {
                thread::spawn(|| handle_trigger_extraction(stream));
            }
            Err(err) => error!("Error during TCP connection: {}", err),
        }
    }
}

fn handle_trigger_extraction(mut stream: TcpStream) {
    debug!("Trigger new runtime info extraction");

    // Execute the "extractor_coordinator" program
    let output = Command::new("/coordinator").output();
    match output {
        Ok(output) => {
            let stdout = String::from_utf8_lossy(&output.stdout);
            let response = format!("{}\n", stdout);
            if let Err(e) = stream.write_all(response.as_bytes()) {
                error!("Failed to write to socket; err = {:?}", e);
            }
        }
        Err(err) => {
            error!("Error during the extraction of the runtime info: {}", err)
        }
    }
    if let Err(e) = stream.shutdown(Shutdown::Both) {
        error!("Failed to shutdown socket; err = {:?}", e);
    }
    info!("DONE");
}
