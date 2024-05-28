use serde::Deserialize;
use std::fs;

#[derive(Deserialize, Debug)]
pub struct Config {
    #[serde(rename = "fingerprints")]
    pub fingerprints: Fingerprints,
}

#[derive(Deserialize, Debug)]
pub struct Fingerprints {
    #[serde(rename = "version-executables")]
    pub versioned_executables: Vec<VersionExecutables>,
}

#[derive(Deserialize, Debug)]
pub struct VersionExecutables {
    #[serde(rename = "process-command")]
    pub process_command: String,
    #[serde(rename = "runtime-kind-name")]
    pub runtime_kind_name: String,
}

pub fn get_config(dir: &str) -> Config {
    let config_content =
        fs::read_to_string(dir.to_owned() + "/config.toml").expect("Configuration file is missing");
    toml::from_str(&config_content).expect("unable to read configuration file")
}

#[cfg(test)]
mod tests {
    use super::*;
    use regex::Regex;

    #[test]
    fn it_can_match_node_process() {
        let config = get_config("./config");

        let node_config = config
            .fingerprints
            .versioned_executables
            .iter()
            .find(|c| c.runtime_kind_name == "Node.js")
            .unwrap();
        let re = Regex::new(&node_config.process_command).unwrap();

        assert!(re.is_match("/usr/local/bin/node"));
        assert!(re.is_match("node"));
        assert!(!re.is_match("nodexx"));
        assert!(!re.is_match("/usr/local/node/bin/foo"));
    }

    #[test]
    fn it_can_match_python_process() {
        let config = get_config("./config");

        let node_config = config
            .fingerprints
            .versioned_executables
            .iter()
            .find(|c| c.runtime_kind_name == "Python")
            .unwrap();
        let re = Regex::new(&node_config.process_command).unwrap();

        assert!(re.is_match("/usr/local/bin/python"));
        assert!(re.is_match("/usr/local/bin/python3"));
        assert!(re.is_match("python"));
        assert!(re.is_match("python3"));
    }
}
