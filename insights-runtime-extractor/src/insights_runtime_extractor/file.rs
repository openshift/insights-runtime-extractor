use log::debug;
use rc_zip_sync::ReadZip;
use std::collections::HashMap;
use std::env::split_paths;
use std::fs::{self, File};
use std::io::BufRead;
use std::io::{self, Write};
use std::os::unix::fs::PermissionsExt;
use std::path::Path;
use std::path::PathBuf;
use std::process::Command;
use std::process::Stdio;

/// Create a directory and return its File.
/// Remove the directory if it exists before creating it.
///
/// The directory is created with 777 permissions so that anybody can write into it.
pub fn create_dir(name: &str) -> io::Result<File> {
    debug!("📂  (re)creating dir {}", name);

    if let Err(e) = fs::remove_dir_all(&name) {
        if e.kind() != io::ErrorKind::NotFound {
            return Err(e);
        }
    }
    fs::create_dir(name)?;
    fs::set_permissions(&name, fs::Permissions::from_mode(0o777))?;

    File::open(&name)
}

/// Read a key=value file and return its content in a HashMap.
///
/// Key and values are separated by '='.
pub fn read_key_value_file(path: &str) -> io::Result<HashMap<String, String>> {
    debug!("📄  Reading key=value file {:?}", path);

    let content = fs::read_to_string(path)?;
    let mut map = HashMap::new();

    for line in content.lines() {
        let parts: Vec<&str> = line.splitn(2, '=').collect();
        if parts.len() == 2 {
            let key = parts[0].trim().to_string();
            let value = parts[1].trim_matches('"').to_string();
            map.insert(key, value);
        }
    }

    Ok(map)
}

/// Read a Jar manifest and return its entries in a HashMap
///
/// Keys and values are separated by ':'.
pub fn read_jar_manifest(manifest_file: &str) -> io::Result<HashMap<String, String>> {
    debug!("📄  Reading jar manifest {:?}", manifest_file);

    let content = fs::read_to_string(manifest_file)?;

    read_jar_manifest_content(&content)
}

/// Read a Jar manifest and return its entries in a HashMap
///
/// Keys and values are separated by ':'.
fn read_jar_manifest_content(manifest_content: &str) -> io::Result<HashMap<String, String>> {
    // /// 	manifestEntries := make(map[string]string)
    // currentKey := ""
    // for _, entry := range strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n") {
    // 	key, value, keyFound := strings.Cut(entry, ": ")
    // 	if keyFound {
    // 		manifestEntries[key] = value
    // 		currentKey = key
    // 	} else {
    // 		manifestEntries[currentKey] = manifestEntries[currentKey] + strings.TrimLeft(entry, " ")
    // 	}
    // }
    // return manifestEntries, nil

    let mut map = HashMap::new();

    let mut current_key = "";
    for line in manifest_content.lines() {
        let parts: Vec<&str> = line.splitn(2, ':').collect();
        if parts.len() == 2 {
            let key = parts[0].trim();
            let value = parts[1].trim().trim_matches('"').to_string();
            map.insert(key.to_string(), value);
            current_key = key;
        } else if parts.len() == 1 {
            let value = parts[0].trim().trim_matches('"').to_string();
            let existing_value = map.get(current_key);
            map.insert(
                current_key.to_string(),
                format!("{}{}", existing_value.unwrap(), value),
            );
        }
    }

    Ok(map)
}

pub fn get_jar_manifest(jar_path: &str) -> io::Result<HashMap<String, String>> {
    debug!("📄  Reading manifest for jar {:?}", jar_path);
    eprintln!("📄  Reading manifest for jar {:?}", jar_path);

    if let Ok(jar_file) = File::open(jar_path) {
        eprintln!("got jar file {:#?}", jar_file);
        if let Ok(archive) = jar_file.read_zip() {
            eprintln!("got archive");
            if let Some(manifest) = archive.by_name("META-INF/MANIFEST.MF") {
                let lines = io::BufReader::new(manifest.reader()).lines();
                let content: String = lines
                    .map(|line| line.map(|l| l + "\n"))
                    .collect::<Result<_, _>>()?;
                println!("content: {:#?}", content);
                return read_jar_manifest_content(&content);
            }
        }
    }

    Ok(HashMap::new())
}

pub fn jar_contains_class(jar_path: &PathBuf, class_name: &str) -> bool {
    let class_file = class_name.replace(".", "/") + ".class";

    debug!("📄  List classes in the jar {:?}", jar_path);
    eprintln!("📄  Reading manifest for jar {:?}", jar_path);
    if let Ok(jar_file) = File::open(jar_path) {
        if let Ok(archive) = jar_file.read_zip() {
            match archive.entries().find(|file| file.name == class_file) {
                Some(_) => {
                    return true;
                }
                None => {}
            }
        }
    }

    false
}

/// Write fingerprints entries to a file in the `out` directory.
///
/// The file is named `${prefix}-fingerprints.txt`.
pub fn write_fingerprint(
    out: &Path,
    prefix: &str,
    entries: &HashMap<String, String>,
) -> io::Result<()> {
    if entries.len() == 0 || prefix.len() == 0 {
        return Ok(());
    }

    let binding = Path::new(&out).join(format!("{}-fingerprints.txt", prefix));
    let fingerprint_file = binding.as_path();
    let mut file = File::create(fingerprint_file)?;

    let mut keys: Vec<_> = entries.keys().collect();
    keys.sort();

    for key in keys {
        if let Some(value) = entries.get(key) {
            writeln!(&mut file, "{}={}", key, value)?;
        }
    }

    debug!("📄  Wrote fingerprints file {:?}", fingerprint_file);

    Ok(())
}

pub fn get_executable_version_output(exec: &str) -> io::Result<String> {
    let output = Command::new(&exec)
        .arg("--version")
        .stdout(Stdio::piped())
        .output();

    match output {
        Ok(output) => {
            if output.status.success() {
                if let Ok(output_str) = String::from_utf8(output.stdout) {
                    Ok(output_str.trim().to_string())
                } else {
                    let custom_error =
                        io::Error::new(io::ErrorKind::Other, "No output from command");
                    Err(custom_error)
                }
            } else {
                let custom_error = io::Error::new(io::ErrorKind::Other, "command did not succeed");
                Err(custom_error)
            }
        }
        Err(err) => Err(err),
    }
}

pub fn recursively_resolve_symlink(path: PathBuf) -> io::Result<PathBuf> {
    let mut current_path = path;
    loop {
        let metadata = fs::symlink_metadata(&current_path)?;
        if metadata.file_type().is_symlink() {
            current_path = fs::read_link(&current_path)?;
        } else {
            return Ok(current_path);
        }
    }
}

pub fn find_executable_in_path(executable_name: &str, path_var: &str) -> Option<PathBuf> {
    for path in split_paths(&path_var) {
        let exe_path = path.join(executable_name);
        if exe_path.is_file() {
            return Some(exe_path);
        }
    }
    None
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::collections::HashMap;

    use crate::file::create_dir;

    #[test]
    fn it_write_fingerprint() {
        let out = String::from("target/tmp-test");
        create_dir(&out).expect("Create temporary directory");

        let mut entries = HashMap::new();
        entries.insert(String::from("foo"), String::from("value1"));
        entries.insert(String::from("bar"), String::from("value2"));

        let dir_path = Path::new(&out);
        write_fingerprint(dir_path, "test", &entries).expect("wrote fingerprints");

        let content = fs::read_to_string(dir_path.join("test-fingerprints.txt"))
            .expect("Read content from fingerprint file");
        // map is stored in natural order
        assert_eq!(
            content,
            "\
bar=value2
foo=value1
"
            .to_string()
        );

        let _ = fs::remove_dir_all(dir_path);
    }
}
