use std::thread;
use std::time::Duration;

use insights_operator_runtime::{file, perms};

fn main() {
    println!("Gather runtime information from containers on OpenShift");

    perms::check_privileged_perms().expect("Must have privileged permissions to scan containers");

    // create a out directory to store all fingerprints data
    file::create_dir("out").expect("Can not create output directory");

    println!("\nRun the /gather-containers or /gather-container executables to extract runtime information from containers.");
    println!("Sleeping 💤");
    thread::sleep(Duration::MAX);
}
