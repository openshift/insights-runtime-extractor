package main

import (
	"fingerprints/pkg/utils"
	"log"
	"os"
	"time"
)

func main() {
	// The program has parameters:
	// - 1 - the subdirectory to write the manifest to
	outputDir := os.Args[1]

	startTime := time.Now()
	log.Printf("🔎 Fingerprinting the Operating System to %s\n", outputDir)

	entries := make(map[string]string)

	if properties, exists := utils.ReadPropertiesFile("/etc/os-release"); exists {
		for k, v := range properties {
			switch k {
			case "ID":
				entries["os-release-id"] = v
			case "VERSION_ID":
				entries["os-release-version-id"] = v
			}
		}
		utils.WriteEntries(outputDir, "os.txt", entries)
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	log.Printf("🕑 OS fingerprint executed in time: %s\n", duration)
}
