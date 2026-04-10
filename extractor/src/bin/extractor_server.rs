use clap::Parser;
use log::{error, info, trace, warn};
use openssl::ssl::{Ssl, SslAcceptor, SslContext, SslMethod, SslVersion};
use serde::Deserialize;
use std::collections::HashMap;
use std::fs;
use std::io::{Read, Write};
use std::net::TcpListener;
use std::process::Command;
use std::thread;
use std::time::Instant;

use kube::{
    api::{Api, DynamicObject},
    discovery::ApiResource,
    Client,
};

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

    #[arg(long, required = true, help = "Path to TLS certificate file (PEM format)")]
    tls_cert: String,

    #[arg(long, required = true, help = "Path to TLS private key file (PEM format)")]
    tls_key: String,
}

// TLS profile types matching the OpenShift API Server configuration

#[derive(Debug, Clone, Deserialize)]
#[serde(rename_all = "camelCase")]
struct TLSSecurityProfile {
    #[serde(rename = "type")]
    profile_type: Option<String>,
    custom: Option<CustomTLSProfile>,
}

#[derive(Debug, Clone, Deserialize)]
#[serde(rename_all = "camelCase")]
struct CustomTLSProfile {
    min_tls_version: Option<String>,
    ciphers: Option<Vec<String>>,
}

#[derive(Debug, Clone)]
struct TLSProfileSpec {
    min_tls_version: String,
    ciphers: Vec<String>,
}

fn builtin_profiles() -> HashMap<&'static str, TLSProfileSpec> {
    HashMap::from([
        ("Old", TLSProfileSpec {
            min_tls_version: "VersionTLS10".into(),
            ciphers: vec![],
        }),
        ("Intermediate", TLSProfileSpec {
            min_tls_version: "VersionTLS12".into(),
            ciphers: vec![
                "ECDHE-RSA-AES128-GCM-SHA256",
                "ECDHE-ECDSA-AES128-GCM-SHA256",
                "ECDHE-RSA-AES256-GCM-SHA384",
                "ECDHE-ECDSA-AES256-GCM-SHA384",
            ].into_iter().map(String::from).collect(),
        }),
        ("Modern", TLSProfileSpec {
            min_tls_version: "VersionTLS13".into(),
            ciphers: vec![],
        }),
    ])
}

fn resolve_profile(body: &serde_json::Value) -> TLSProfileSpec {
    let profiles = builtin_profiles();
    let default = profiles["Intermediate"].clone();

    let profile = body
        .pointer("/spec/tlsSecurityProfile")
        .and_then(|v| serde_json::from_value::<TLSSecurityProfile>(v.clone()).ok());

    let profile = match profile {
        Some(p) => p,
        None => return default,
    };

    let profile_type = profile.profile_type.as_deref().unwrap_or("Intermediate");

    if profile_type == "Custom" {
        if let Some(custom) = profile.custom {
            return TLSProfileSpec {
                min_tls_version: custom.min_tls_version.unwrap_or("VersionTLS12".into()),
                ciphers: custom.ciphers.unwrap_or_default(),
            };
        }
    }

    profiles.get(profile_type).cloned().unwrap_or(default)
}

fn to_ssl_version(version: &str) -> SslVersion {
    match version {
        "VersionTLS13" => SslVersion::TLS1_3,
        "VersionTLS12" => SslVersion::TLS1_2,
        "VersionTLS11" => SslVersion::TLS1_1,
        "VersionTLS10" => SslVersion::TLS1,
        _ => SslVersion::TLS1_2,
    }
}

fn build_ssl_context(
    spec: &TLSProfileSpec,
    certfile: &str,
    keyfile: &str,
) -> Result<SslContext, openssl::error::ErrorStack> {
    let mut builder = SslAcceptor::mozilla_intermediate_v5(SslMethod::tls_server())?;

    builder.set_min_proto_version(Some(to_ssl_version(&spec.min_tls_version)))?;

    if !spec.ciphers.is_empty() {
        builder.set_cipher_list(&spec.ciphers.join(":"))?;
    }

    builder.set_certificate_chain_file(certfile)?;
    builder.set_private_key_file(keyfile, openssl::ssl::SslFiletype::PEM)?;


    Ok(builder.build().into_context())
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

    // Fetch TLS profile from API Server
    let rt = tokio::runtime::Builder::new_current_thread()
        .enable_all()
        .build()
        .expect("Failed to create tokio runtime");
    let profile_spec = rt.block_on(async {
        let client = Client::try_default().await.expect("Failed to create kube client");

        let ar = ApiResource::from_gvk(&kube::api::GroupVersionKind {
            group: "config.openshift.io".into(),
            version: "v1".into(),
            kind: "APIServer".into(),
        });

        let api: Api<DynamicObject> = Api::all_with(client, &ar);
        let obj = api.get("cluster").await.expect("Failed to get APIServer config");
        let value = serde_json::to_value(&obj).expect("Failed to serialize APIServer");

        resolve_profile(&value)
    });

    info!("TLS profile: min_version={}, ciphers={:?}", profile_spec.min_tls_version, profile_spec.ciphers);

    // Build SSL context with the API Server TLS profile
    let ssl_context = std::sync::Arc::new(
        build_ssl_context(&profile_spec, &args.tls_cert, &args.tls_key)
            .expect("Failed to build SSL context"),
    );

    // Create a TCP listener
    // bound to the loopback address so that it can only be contacted
    // by containers in the same pod
    let addr = "127.0.0.1:3000";
    let listener = TcpListener::bind(addr).expect("Failed to bind to address");

    info!("Listening on {}", addr);

    for stream in listener.incoming() {
        match stream {
            Ok(stream) => {
                let ctx = ssl_context.clone();
                let log_level = log_level.clone();
                thread::spawn(move || {
                    let ssl = match Ssl::new(&ctx) {
                        Ok(ssl) => ssl,
                        Err(err) => {
                            error!("Failed to create SSL session: {}", err);
                            return;
                        }
                    };
                    match ssl.accept(stream) {
                        Ok(tls_stream) => handle_trigger_extraction(tls_stream, log_level),
                        Err(err) => error!("TLS handshake failed: {}", err),
                    }
                });
            }
            Err(err) => error!("Error during TCP connection: {}", err),
        }
    }
}

fn handle_trigger_extraction(mut stream: impl Read + Write, log_level: String) {
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
