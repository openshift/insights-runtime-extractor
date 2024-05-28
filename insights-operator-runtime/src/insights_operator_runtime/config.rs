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
    #[serde(rename = "process-names")]
    pub process_names: Vec<String>,
    #[serde(rename = "runtime-kind-name")]
    pub runtime_kind_name: String,
}

pub fn get_config(dir: &str) -> Config {
    let config_content =
        fs::read_to_string(dir.to_owned() + "/config.toml").expect("Configuration file is missing");
    toml::from_str(&config_content).expect("unable to read configuration file")
}
