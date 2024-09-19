package main

import (
	"fingerprints/pkg/utils"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	// The program has parameters:
	// - 1 - the subdirectory to write the manifest to
	// - 2 - the PATH env var of the process
	// - 3 - the JAVA_HOME env var of the process
	outputDir := os.Args[1]
	pathEnvVar := os.Args[2]
	javaHomeEnvVar := os.Args[3]

	startTime := time.Now()
	log.Printf("🔎 Fingerprinting the Java version to %s\n", outputDir)

	javaHomeDir := javaHomeEnvVar
	if javaHomeDir == "" {
		// find the java home directory based on the location of the java executable
		// ($JAVA_HOME/bin/java)
		dir, err := utils.FindExecutableInPath("java", pathEnvVar)
		if err != nil {
			log.Panicf("Unable to find java home directory: %s", err)
		}
		javaHomeDir = filepath.Dir(dir)
	}
	log.Printf("🔎 Fingerprinting the Java version from %s\n", javaHomeDir)

	entries := make(map[string]string)
	// read the release file from the $JAVA_HOME directory
	entries["runtime-kind"] = "Java"
	if properties, exists := utils.ReadPropertiesFile(filepath.Join(javaHomeDir, "release")); exists {
		for k, v := range properties {
			switch k {
			case "JAVA_VERSION":
				entries["runtime-kind-version"] = v
			case "IMPLEMENTOR":
				entries["runtime-kind-implementer"] = v
			}
		}
	}
	utils.WriteEntries(outputDir, "runtime-kind.txt", entries)

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	log.Printf("🕑 Java version fingerprint executed in time: %s\n", duration)
}
