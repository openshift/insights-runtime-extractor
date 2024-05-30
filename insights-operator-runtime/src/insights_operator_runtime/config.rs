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

    pub java: Vec<JavaFingerprint>,
}

#[derive(Deserialize, Debug)]
pub struct VersionExecutables {
    #[serde(rename = "process-names")]
    pub process_names: Vec<String>,
    #[serde(rename = "runtime-kind-name")]
    pub runtime_kind_name: String,
}

#[derive(Deserialize, Debug)]
pub struct JavaFingerprint {
    #[serde(rename = "runtime-name")]
    pub runtime_name: String,
    #[serde(rename = "main-class")]
    pub main_class: String,
    #[serde(rename = "read-manifest-of-executable-jar")]
    pub read_manifest_of_executable_jar: bool,
    #[serde(rename = "jar-version-manifest-entry")]
    pub jar_version_manifest_entry: String,
}

pub fn get_config(dir: &str) -> Config {
    let config_content =
        fs::read_to_string(dir.to_owned() + "/config.toml").expect("Configuration file is missing");
    toml::from_str(&config_content).expect("unable to read configuration file")
}
