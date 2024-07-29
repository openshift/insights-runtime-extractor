use std::collections::HashMap;
use std::io;
use std::path::Path;
use std::time::Instant;

use insights_runtime_extractor::file;

fn main() -> io::Result<()> {
    let start = Instant::now();

    let out = std::env::args()
        .nth(1)
        .expect("Missing argument for output directory");

    println!("🔎 Fingerprinting the Operating System to {}", out);

    let mut entries = HashMap::new();

    if let Ok(release_entries) = file::read_key_value_file("/etc/os-release") {
        for (key, value) in release_entries {
            match key.as_str() {
                // "PRETTY_NAME" => entries.insert(String::from("os-release-pretty-name"), value),
                // "NAME" => entries.insert(String::from("os-release-name"), value),
                "VERSION_ID" => entries.insert(String::from("os-release-version-id"), value),
                "ID" => entries.insert(String::from("os-release-id"), value),
                _ => None,
            };
        }
    }

    let duration = start.elapsed().as_micros();
    println!("🕑 OS Fingerprint executed in  {:?}μs", duration);

    file::write_entries(Path::new(&out), "os.txt", &entries)
}
