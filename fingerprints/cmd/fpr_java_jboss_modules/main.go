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
	versionDelimiter                    = "- Version"
	productManifestPath                 = "modules/system/layers/base/org/jboss/as/product/main/dir/META-INF/MANIFEST.MF"
	productMainPath                     = "modules/system/layers/base/org/jboss/as/product/main/"
	wildflyFeaturePackProductConfPrefix = "wildfly-ee-feature-pack-product-conf"
	jbossProductReleaseName             = "JBoss-Product-Release-Name"
	jbossProductReleaseVersion          = "JBoss-Product-Release-Version"
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
	if found, content := utils.ReadFile(filepath.Join(jbossHomeDir, "version.txt")); found {
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
		if found, content := utils.ReadFile(manifestPath); found {
			manifestEntries := utils.ReadManifest(content)
			entries[manifestEntries[jbossProductReleaseName]] = manifestEntries[jbossProductReleaseVersion]
			foundRuntime = true
		}
	}

	if !foundRuntime {
		// more recent versions of WildFly store this info in the Manifest of wildfly-feature-pack-product-conf jar
		productMainDir := filepath.Join(jbossHomeDir, productMainPath)
		_ = filepath.Walk(productMainDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasPrefix(info.Name(), wildflyFeaturePackProductConfPrefix) && strings.HasSuffix(info.Name(), ".jar") {
				manifestEntries, err := utils.GetJarManifest(path)
				if err != nil {
					return err
				}
				entries[manifestEntries[jbossProductReleaseName]] = manifestEntries[jbossProductReleaseVersion]
				foundRuntime = true
			}
			return nil
		})
	}

	if len(entries) > 0 {
		utils.WriteEntries(outputDir, "java-jboss-modules-fingerprints.txt", entries)
	}
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	log.Printf("🕑 Java JBoss Modules fingerprint executed in time: %s\n", duration)
}
