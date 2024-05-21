use log::info;
use serde_json::Value;
use std::process::Command;

#[derive(Clone, Debug)]
pub struct Container {
    // Container ID
    pub id: String,
    // Image pulled to create the container,
    pub image_ref: String,
    // Name of the container (from Kubernetes descriptors)
    pub name: String,
    // Name of the pod owning the container
    pub pod_name: String,
    // Namespace of the container's pod
    pub pod_namespace: String,
    // Root pid of the container
    pub pid: u32,
}

pub fn get_container(container_id: &String) -> Option<Container> {
    info!("🔎  Reading container information with crictl...");

    let output = Command::new("/crictl")
        .args(["ps", "-o", "json", "-s", "RUNNING"])
        .output()
        .expect("List containers with crictl");
    let json = String::from_utf8(output.stdout).unwrap();

    let deserialized_containers: Value = serde_json::from_str(&json).unwrap();

    let container_id = match container_id.strip_prefix("cri-o://") {
        Some(container_id) => container_id.to_string(),
        None => container_id.to_string(),
    };

    for c in deserialized_containers["containers"].as_array().unwrap() {
        let id = c["id"].as_str().unwrap().to_string();

        match id == container_id {
            false => {}
            true => {
                let pod_namespace = c["labels"]["io.kubernetes.pod.namespace"]
                    .as_str()
                    .unwrap()
                    .to_string();
                let image_ref = c["imageRef"].as_str().unwrap().to_string();
                let name = c["labels"]["io.kubernetes.container.name"]
                    .as_str()
                    .unwrap()
                    .to_string();
                let pod_name = c["labels"]["io.kubernetes.pod.name"]
                    .as_str()
                    .unwrap()
                    .to_string();
                let pid: u32 = get_root_pid(&id);
                return Some(Container {
                    id: "cri-o://".to_owned() + &id,
                    image_ref,
                    name,
                    pod_name,
                    pod_namespace,
                    pid,
                });
            }
        }
    }

    None
}

pub fn get_containers() -> Vec<Container> {
    info!("🔎  Reading container information with crictl...");

    let output = Command::new("/crictl")
        .args(["ps", "-o", "json", "-s", "RUNNING"])
        .output()
        .expect("List containers with crictl");
    let json = String::from_utf8(output.stdout).unwrap();

    let deserialized_containers: Value = serde_json::from_str(&json).unwrap();

    let mut containers: Vec<Container> = Vec::new();

    for c in deserialized_containers["containers"].as_array().unwrap() {
        let pod_namespace = c["labels"]["io.kubernetes.pod.namespace"]
            .as_str()
            .unwrap()
            .to_string();

        let id = c["id"].as_str().unwrap().to_string();
        let image_ref = c["imageRef"].as_str().unwrap().to_string();
        let name = c["labels"]["io.kubernetes.container.name"]
            .as_str()
            .unwrap()
            .to_string();
        let pod_name = c["labels"]["io.kubernetes.pod.name"]
            .as_str()
            .unwrap()
            .to_string();
        let pid: u32 = get_root_pid(&id);

        let container = Container {
            id: "cri-o://".to_owned() + &id,
            image_ref,
            name,
            pod_name,
            pod_namespace,
            pid,
        };
        containers.push(container);
    }

    return containers;
}

pub fn get_root_pid(container_id: &String) -> u32 {
    let output = Command::new("/crictl")
        .args([
            "inspect",
            "-o",
            "go-template",
            "--template",
            "{{.info.pid}}",
            container_id,
        ])
        .output()
        .expect("Inspect container with crictl");

    let pid = String::from_utf8(output.stdout)
        .unwrap()
        .trim()
        .parse::<u32>()
        .unwrap();
    return pid;
}
