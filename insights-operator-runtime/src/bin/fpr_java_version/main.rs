use std::collections::HashMap;
use std::io;
use std::path::Path;

use insights_operator_runtime::file;

fn main() -> io::Result<()> {
    let out = std::env::args()
        .nth(1)
        .expect("Missing argument for output directory");
    let path_var = std::env::args()
        .nth(2)
        .expect("PATH env var of the java process");
    let java_home_arg = std::env::args().nth(3);

    let java_home: String = match java_home_arg {
        Some(val) if val.is_empty() => {
            let java_exe = file::find_executable_in_path("java", &path_var);
            let actual_file = file::recursively_resolve_symlink(java_exe.unwrap()).unwrap();
            println!("java is called with path: {:#?}", actual_file);
            actual_file
                .parent()
                .and_then(|parent| parent.parent())
                .unwrap()
                .to_string_lossy()
                .to_string()
        }
        Some(val) => val,
        _ => "".to_owned(),
    };

    println!("🔎 Fingerprinting Java from {:#?}", java_home);

    let mut entries = HashMap::new();
    entries.insert(String::from("runtime-kind"), String::from("Java"));

    if !java_home.is_empty() {
        if let Ok(release_entries) = file::read_key_value_file(&(java_home.to_owned() + "/release"))
        {
            for (key, value) in release_entries {
                match key.as_str() {
                    "JAVA_VERSION" => entries.insert(String::from("runtime-kind-version"), value),
                    "IMPLEMENTOR" => {
                        entries.insert(String::from("runtime-kind-implementer"), value)
                    }
                    _ => None,
                };
            }
        }
    }
    file::write_fingerprint(Path::new(&out), "runtime-kind", &entries)
}
