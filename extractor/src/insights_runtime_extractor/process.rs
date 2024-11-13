use log::{debug, trace};
use std::collections::HashMap;
use std::fs;
use std::time::Instant;
use sysinfo::{Pid, System, Uid};

#[derive(Clone, Debug)]
pub struct ContainerProcess {
    pub pid: u32,
    pub uid: Uid,
    pub name: String,
    pub command_line: Vec<String>,
    pub cwd: Option<String>,
    pub environ: HashMap<String, String>,
}

pub fn get_process_leaves(pid: &u32) -> Vec<ContainerProcess> {
    let start = Instant::now();

    let s = System::new_all();

    let leaves = get_leaves(&pid);
    debug!("got leaves: {:#?}", leaves);

    let res: Vec<ContainerProcess>= leaves
        .iter()
        .map(|pid| s.process(Pid::from_u32(*pid)))
        .flatten()
        .map(|process| self::ContainerProcess {
            pid: process.pid().as_u32(),
            uid: process.user_id().unwrap().clone(),
            name: process.name().to_string(),
            command_line: process.cmd().iter().map(String::from).collect(),
            cwd: process
                .cwd()
                .and_then(|p| Some(p.to_string_lossy().into_owned())),
            environ: get_environ_hashmap(process.environ()),
        })
        .collect();

    let duration = start.elapsed().as_millis();
    trace!("Processed {} processes in {:?}ms", res.len(), duration);

    res
}

fn get_leaves(root_pid: &u32) -> Vec<u32> {
    debug!("Detecting leaves: {:#?}", root_pid);

    let mut leaves = Vec::new();
    collect_leaves(&root_pid, &mut leaves);

    leaves
}

fn collect_leaves(pid: &u32, leaves: &mut Vec<u32>) {
    debug!("collect_leaves for {:#?}: {:#?}", pid, leaves);

    let path = format!("/proc/{}/task/{}/children", pid, pid);
    if let Ok(content) = fs::read_to_string(path) {
        if content.len() == 0 {
            // no child, add the pid itself and returns
            leaves.push(*pid);
            return;
        }

        for line in content.lines() {
            for child_pid in line.split(" ") {
                if let Ok(child_pid) = child_pid.parse() {
                    debug!("child_pid >>> {:#?}", child_pid);
                    collect_leaves(&child_pid, leaves)
                }
            }
        }
    }
}

fn get_environ_hashmap(envs: &[String]) -> HashMap<String, String> {
    let mut map = HashMap::new();

    for env in envs {
        if let Some(pos) = env.find('=') {
            let key = env[..pos].to_string();
            let value = env[pos + 1..].to_string();
            map.insert(key, value);
        }
    }
    map
}
