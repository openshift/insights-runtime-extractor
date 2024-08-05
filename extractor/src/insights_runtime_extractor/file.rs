use log::debug;
use std::collections::HashMap;
use std::fs::{self, File};
use std::io::{self, Write};
use std::os::unix::fs::PermissionsExt;
use std::path::Path;

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

/// Write entries to a file in the `out` directory.
pub fn write_entries(
    out: &Path,
    file_name: &str,
    entries: &HashMap<String, String>,
) -> io::Result<()> {
    if entries.len() == 0 {
        return Ok(());
    }

    let binding = Path::new(&out).join(file_name);
    let file_path = binding.as_path();
    let mut file = File::create(file_path)?;

    let mut keys: Vec<_> = entries.keys().collect();
    keys.sort();

    for key in keys {
        if let Some(value) = entries.get(key) {
            writeln!(&mut file, "{}={}", key, value)?;
        }
    }

    debug!("📄  Wrote fingerprints file {:?}", file_path);

    Ok(())
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
        write_entries(dir_path, "test.txt", &entries).expect("wrote fingerprints");

        let content = fs::read_to_string(dir_path.join("test.txt"))
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
