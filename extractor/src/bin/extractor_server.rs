use clap::Parser;
use log::{error, info, trace, warn};
use std::fs;
use std::io::{BufReader, Read, Write};
use std::net::TcpListener;
use std::process::Command;
use std::sync::Arc;
use std::thread;
use std::time::Instant;

use rustls::ServerConfig;

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

    #[arg(long, help = "Path to TLS certificate file (PEM format)")]
    tls_cert: Option<String>,

    #[arg(long, help = "Path to TLS private key file (PEM format)")]
    tls_key: Option<String>,
}

fn main() {
    let args = Args::parse();

    let log_level = args.log_level.unwrap_or(String::from("info"));

    env_logger::Builder::from_env(env_logger::Env::default().default_filter_or(log_level.clone())).init();

    info!("Gather runtime information from containers on OpenShift");

    perms::check_privileged_perms().expect("Must have privileged permissions to scan containers");

    // verify that the configuration is properly setup
    let config_content = fs::read_to_string("/config.toml").expect("Configuration file is missing");
    info!("Configuration:\n----\n{}\n----", config_content);
    config::get_config("/");

    // Configure TLS if both cert and key are provided
    let tls_config: Option<Arc<ServerConfig>> = match (&args.tls_cert, &args.tls_key) {
        (Some(cert_path), Some(key_path)) => {
            let cert_file =
                fs::File::open(cert_path).expect("Failed to open TLS certificate file");
            let key_file =
                fs::File::open(key_path).expect("Failed to open TLS private key file");

            let certs: Vec<_> = rustls_pemfile::certs(&mut BufReader::new(cert_file))
                .collect::<Result<Vec<_>, _>>()
                .expect("Failed to parse TLS certificate PEM");
            let key = rustls_pemfile::private_key(&mut BufReader::new(key_file))
                .expect("Failed to read TLS private key PEM")
                .expect("No private key found in PEM file");

            let config = ServerConfig::builder()
                .with_no_client_auth()
                .with_single_cert(certs, key)
                .expect("Failed to build TLS configuration");

            info!("TLS enabled with cert={} key={}", cert_path, key_path);
            Some(Arc::new(config))
        }
        (None, None) => {
            info!("TLS not configured, running plain TCP");
            None
        }
        _ => {
            panic!("Both --tls-cert and --tls-key must be provided together");
        }
    };

    // Create a TCP listener
    // bound to the loopback address so that it can only be contacted
    // by containers in the same pod
    let addr = "127.0.0.1:3000";
    let listener = TcpListener::bind(addr).expect("Failed to bind to address");

    info!("Listening on {}", addr);

    for stream in listener.incoming() {
        match stream {
            Ok(mut tcp_stream) => {
                let log_level = log_level.clone();
                let tls_config = tls_config.clone();
                thread::spawn(move || {
                    if let Some(config) = tls_config {
                        let conn = rustls::ServerConnection::new(config)
                            .expect("Failed to create TLS connection");
                        let mut tls_stream = rustls::StreamOwned::new(conn, tcp_stream);
                        handle_trigger_extraction(&mut tls_stream, log_level);
                    } else {
                        handle_trigger_extraction(&mut tcp_stream, log_level);
                    }
                });
            }
            Err(err) => error!("Error during TCP connection: {}", err),
        }
    }
}

fn handle_trigger_extraction(stream: &mut (impl Read + Write), log_level: String) {
    info!("Triggering new runtime info extraction");

    let start = Instant::now();

    // Read container IDs from TCP stream until EOF
    let mut buffer = String::new();
    if let Err(e) = stream.read_to_string(&mut buffer) {
        warn!("Failed to read from socket; err = {:?}", e);
        return;
    }
    let container_ids = buffer.trim().to_string();

    // Execute the "extractor_coordinator" program
    let output = Command::new("/coordinator")
        .arg("--log-level")
        .arg(log_level)
        .arg(container_ids)
        .output();
    match output {
        Ok(output) => {
            let stderr = String::from_utf8_lossy(&output.stderr);
            trace!("{}\n", stderr);

            let stdout = String::from_utf8_lossy(&output.stdout);

            let response = format!("{}\n", stdout);

            let duration = start.elapsed().as_secs();
            info!("Info extracted in {}s, stored at {}", duration, response);

            if let Err(e) = stream.write_all(response.as_bytes()) {
                error!("Failed to write to socket; err = {:?}", e);
            }
        }
        Err(err) => {
            error!("Error during the extraction of the runtime info: {}", err)
        }
    }
}
