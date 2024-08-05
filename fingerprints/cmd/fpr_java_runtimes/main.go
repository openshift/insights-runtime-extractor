package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"fingerprints/pkg/utils"
)

func main() {
	// The program has parameters:
	// - 1 - the subdirectory to write the manifest to
	// - 2 - the jar to inspect
	outputDir := os.Args[1]
	inspectedJar := os.Args[2]

	startTime := time.Now()
	log.Printf("🔎 Fingerprinting the Java runtimes from %s\n", inspectedJar)

	config, err := utils.GetConfig(filepath.Join(outputDir, "config.toml"))
	if err != nil {
		log.Fatalf("Unable to read configuration in %s\n", outputDir)
	}
	javaConfigs := config.Fingerprints.Java

	entries := make(map[string]string)

	manifestEntries, err := utils.GetJarManifest(inspectedJar)
	if err != nil {
		log.Fatalf("Unable to read manifest entries from %s\n", inspectedJar)
	}
	for k, v := range manifestEntries {
		switch k {
		case "Main-Class":
			{
				mainClass := v
				// find the config matching this main class
				idx := slices.IndexFunc(javaConfigs, func(jc utils.JavaRuntimeExecutables) bool { return jc.MainClass == mainClass })
				if idx == -1 {
					break
				}
				javaConfig := javaConfigs[idx]
				log.Printf("Found fingerprint configuration for java main-class %s: %+v\n", mainClass, javaConfig)
				if javaConfig.ReadManifestOfExecutableJar {
					entries[javaConfig.RuntimeName] = manifestEntries[javaConfig.JarVersionManifestEntry]
				} else {
					fmt.Printf("Read version for another class\n")
					// find the jars that contains the main class
					classPath := manifestEntries["Class-Path"]
					fmt.Printf("Classpath = %s\n", classPath)
					if classPath != "" {
						for _, otherJar := range strings.Split(classPath, " ") {
							if !filepath.IsAbs(otherJar) {
								// Get the absolute path of the parent directory
								parentDir := filepath.Dir(inspectedJar)
								otherJar = filepath.Join(parentDir, otherJar)
							}
							//  if file::jar_contains_class(&jar, main_class) {
							if utils.JarFileContainsClass(otherJar, mainClass) {
								otherManifestEntries, err := utils.GetJarManifest(otherJar)
								if err != nil {
									log.Fatalf("Unable to read manifest entries from %s\n", inspectedJar)
								}
								entries[javaConfig.RuntimeName] = otherManifestEntries[javaConfig.JarVersionManifestEntry]
							}
						}
					}
				}
			}
		}
	}

	utils.WriteEntries(outputDir, "java-runtimes-fingerprints.txt", entries)
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	log.Printf("🕑 Java runtimes fingerprint executed in time: %s\n", duration)
}
