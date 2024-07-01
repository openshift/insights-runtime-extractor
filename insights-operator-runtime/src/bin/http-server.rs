use std::collections::HashMap;
use std::fs;
use std::time::Instant;

use insights_operator_runtime::config::{self, Config};
use insights_operator_runtime::{perms, RuntimeInfo};

use clap::Parser;
use log::{debug, info, trace};
use rouille::{router, Request, Response};
use tempfile::Builder;

fn gather_runtime_info(request: &Request, config: &Config) -> Response {
    let start = Instant::now();

    let container_id = request.get_param("cid");

    let containers = match container_id {
        Some(container_id) => match insights_operator_runtime::get_container(&container_id) {
            None => return Response::empty_404(),
            Some(container) => vec![container],
        },
        None => insights_operator_runtime::get_containers(),
    };

    let exec_dir = Builder::new()
        .prefix("out")
        .rand_bytes(6)
        .tempdir_in("./")
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

    debug!("Relative execution directory: {}", &relative_exec_dir);

    // keys are pod-namespace, pod-name, container-id
    let mut out: HashMap<String, HashMap<String, HashMap<String, RuntimeInfo>>> = HashMap::new();

    for container in containers {
        debug!("Scanning container 🫙 {}", container.id);
        let namespace = out.entry(container.pod_namespace).or_insert(HashMap::new());
        let pod = namespace
            .entry(container.pod_name)
            .or_insert(HashMap::new());
        if let Some(runtime_info) =
            insights_operator_runtime::scan_container(&config, &relative_exec_dir, &container.id)
        {
            pod.insert(container.id, runtime_info);
        }
    }

    exec_dir.close().ok();

    let duration = start.elapsed().as_millis();
    trace!("🕑 Scanned container(s) in {:?}ms", duration);

    Response::json(&out)
}

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

    let log_level = args.log_level.unwrap_or(String::from("warn"));

    env_logger::Builder::from_env(env_logger::Env::default().default_filter_or(log_level)).init();

    info!("Container Scanning on OpenShift 🔎 ☁️");

    perms::check_privileged_perms().expect("Must have privileged permissions to scan containers");

    // verify that the configuration is properly setup
    let config_content = fs::read_to_string("/config.toml").expect("Configuration file is missing");
    println!("Configuration:\n----\n{}\n----", config_content);
    let config = config::get_config("/");

    rouille::start_server("0.0.0.0:8000", move |request| {
        router!(request,
            (GET) (/gather_runtime_info) =>
            {
                gather_runtime_info(request, &config)
            },
            (GET) (/health/live) => {
                Response::empty_204()
            },
            (GET) (/health/ready) => {
                Response::empty_204()
            },
            _ => Response::empty_404()
        )
    });
}
