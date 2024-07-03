use std::collections::HashMap;
use std::io;
use std::path::{Path, PathBuf};

use insights_runtime_extractor::config;
use insights_runtime_extractor::file;

fn main() -> io::Result<()> {
    let out = std::env::args()
        .nth(1)
        .expect("Missing argument for output directory");
    let scanned_jar = std::env::args()
        .nth(2)
        .expect("Missing argument for jar file");

    println!("🔎 Fingerprinting Java jar {} to {}", scanned_jar, out);

    let config = config::get_config(&out);
    let java_fp_config = config.fingerprints.java;
    println!("Java fingerpint configs are: {:#?}\n", &java_fp_config);

    let mut entries = HashMap::new();

    if let Ok(manifest_entries) = file::get_jar_manifest(&scanned_jar) {
        println!("Entries: {:#?}", manifest_entries);
        for (key, value) in &manifest_entries {
            match key.as_str() {
                "Implementation-Title" => match value.as_str() {
                    impl_title if impl_title.starts_with("kafka_") => {
                        let kafka_version = impl_title.rsplit('_').next().unwrap_or("");
                        if let Some(version) = manifest_entries.get("Implementation-Version") {
                            entries.insert(
                                String::from("Kafka") + " " + kafka_version,
                                version.to_string(),
                            );
                        }
                    }
                    _ => {}
                },

                "Version" => {
                    println!("🔎 Java jar is {:?}", Path::new(&scanned_jar).file_name());
                    // Upstream Kafka specifies only a Version entry in its manifest
                    if Path::new(&scanned_jar)
                        .file_name()
                        .is_some_and(|f| f.to_str().is_some_and(|f| f.starts_with("kafka_")))
                    {
                        let kafka_runtime_version = value.to_string();
                        let kafka_scala_version = &scanned_jar.rsplit('_').next().unwrap_or("");
                        let kafka_scala_version =
                            kafka_scala_version.split('-').next().unwrap_or("");
                        if kafka_scala_version != "" {
                            entries.insert(
                                String::from("Kafka") + " " + kafka_scala_version,
                                kafka_runtime_version,
                            );
                        }
                    }
                }

                "Main-Class" => {
                    let main_class = value.as_str();
                    let config = java_fp_config
                        .iter()
                        .find(|&java_fp| java_fp.main_class == main_class);
                    println!(
                        "Found config for main-class {}:\n{:#?}",
                        &main_class, &config
                    );

                    match config {
                        Some(config) => {
                            if config.read_manifest_of_executable_jar {
                                if let Some(version) =
                                    manifest_entries.get(&config.jar_version_manifest_entry)
                                {
                                    entries
                                        .insert(config.runtime_name.clone(), version.to_string());
                                }
                            } else {
                                // find the jars that contains the main class
                                if let Some(classpath) = manifest_entries.get("Class-Path") {
                                    for jar in classpath.split(" ") {
                                        let jar = match jar {
                                            jar if jar.starts_with("/") => PathBuf::from(jar),
                                            jar => {
                                                let mut dir = Path::new(&scanned_jar)
                                                    .parent()
                                                    .unwrap()
                                                    .to_path_buf();
                                                dir.push(jar);
                                                dir
                                            }
                                        };
                                        if file::jar_contains_class(&jar, main_class) {
                                            println!(
                                                "found jar container main-class {}: {:#?}",
                                                main_class, jar
                                            );
                                            if let Ok(manifest_entries) = file::get_jar_manifest(
                                                jar.display().to_string().as_str(),
                                            ) {
                                                if let Some(version) = manifest_entries
                                                    .get(&config.jar_version_manifest_entry)
                                                {
                                                    entries.insert(
                                                        config.runtime_name.clone(),
                                                        version.to_string(),
                                                    );
                                                }
                                            }
                                        }
                                    }
                                }
                            }
                        }
                        None => {}
                    }
                }
                _ => {}
            }
        }
    }

    file::write_fingerprint(Path::new(&out), "java", &entries)
}
