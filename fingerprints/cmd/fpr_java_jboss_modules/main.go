package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fingerprints/pkg/utils"
)

const (
	versionDelimiter           = "- Version"
	productManifestPath        = "modules/system/layers/base/org/jboss/as/product/main/dir/META-INF/MANIFEST.MF"
	jbossProductReleaseName    = "JBoss-Product-Release-Name"
	jbossProductReleaseVersion = "JBoss-Product-Release-Version"
)

func main() {
	// The program has parameters:
	// - 1 - the subdirectory to write the manifest to
	// - 2 - the directory corresponding to the jboss.home.dir system property
	outputDir := os.Args[1]
	jbossHomeDir := os.Args[2]

	startTime := time.Now()
	log.Printf("🔎 Fingerprinting the JBoss modules from %s\n", jbossHomeDir)

	entries := make(map[string]string)

	foundRuntime := false
	if found, content := readFile(filepath.Join(jbossHomeDir, "version.txt")); found {
		parts := strings.Split(content, versionDelimiter)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			version := strings.TrimSpace(parts[1])
			entries[name] = version
			foundRuntime = true
		}
	}
	if !foundRuntime {
		// version.txt does not exist, let's look at $JBOSS_HOME/modules/system/layers/base/org/jboss/as/product/main/dir/META-INF/MANIFEST.MF
		manifestPath := filepath.Join(jbossHomeDir, productManifestPath)
		if found, content := readFile(manifestPath); found {
			manifestEntries := utils.ReadManifest(content)
			entries[manifestEntries[jbossProductReleaseName]] = manifestEntries[jbossProductReleaseVersion]
			foundRuntime = true
		}
	}

	utils.WriteEntries(outputDir, "java-jboss-modules-fingerprints.txt", entries)
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	log.Printf("🕑 Java JBoss Modules fingerprint executed in time: %s\n", duration)
}

func readFile(path string) (bool, string) {
	content, error := os.ReadFile(path)

	// Check whether the 'error' is nil or not. If it
	//is not nil, then print the error and exit.
	if error != nil {
		return false, ""
	}
	return true, string(content)
}
