use log::debug;

use super::FingerPrint;
use crate::insights_operator_runtime::Config;
use crate::insights_operator_runtime::ContainerProcess;

pub struct Java {}

impl Java {
    fn jar_executable(process: &ContainerProcess, jar: &str) -> Option<Vec<String>> {
        let jar = match jar {
            jar if jar.starts_with("/") => jar.to_owned(),
            jar => format!("{:?}/{}", process.cwd, jar),
        };

        return Some(vec![
            String::from("./fpr_java"),
            format!("out/{}", process.pid),
            jar.to_string(),
        ]);
    }
}

impl FingerPrint for Java {
    fn can_apply_to(&self, _config: &Config, process: &ContainerProcess) -> Option<Vec<String>> {
        if !process.name.ends_with("java") {
            return None;
        }

        debug!("Fingerprint Java application from process: {:#?}", process);

        // check "java -jar" process
        let exec = process
            .command_line
            .iter()
            .position(|s| s == "-jar")
            .and_then(|i| process.command_line.get(i + 1))
            .and_then(|jar| {
                debug!("Executable jar is {:?}", jar);
                return Java::jar_executable(process, jar);
            });

        if exec.is_some() {
            return exec;
        }

        // check for Java classpath-based process
        process
            .command_line
            .iter()
            .position(|s| s == "-classpath" || s == "-cp")
            .and_then(|i| process.command_line.get(i + 1))
            .and_then(|classpath| {
                debug!("java process is using classpath {}", classpath);
                let mut jars = classpath.split(":");
                match process.command_line.join("") {
                    cmd if cmd.contains("org.apache.catalina.startup.Bootstrap") => {
                        debug!("Detected Apache Tomcat application");
                        jars.find(|jar| jar.ends_with("bootstrap.jar"))
                            .and_then(|jar| Java::jar_executable(process, &jar))
                    }

                    cmd if cmd.contains("io.quarkus.bootstrap.runner.QuarkusEntryPoint") => {
                        debug!("Detected Quarkus Application");
                        // Quarkus application, let's extract the version from the io.quarkus.quarkus-core jar
                        jars.find(|jar| jar.contains("io.quarkus.quarkus-core"))
                            .and_then(|jar| Java::jar_executable(process, &jar))
                    }

                    cmd if cmd.contains("kafka.Kafka") => {
                        debug!("Detected Kafka Broker");
                        // Kafka application, let's extract the version from the kafka_* jar
                        jars.find(|jar| jar.contains("kafka_"))
                            .and_then(|jar| Java::jar_executable(process, &jar))
                    }

                    _ => None,
                }
            })
    }
}
