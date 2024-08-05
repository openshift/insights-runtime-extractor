package main

import (
	"log"
	"os"
	"time"

	"fingerprints/pkg/utils"
)

func main() {
	// The program has parameters:
	// - 1 - the subdirectory to write the manifest to
	// - 2 - the name of the executable
	// - 3 - the name of the runtime-kind corresponding to the executable
	outputDir := os.Args[1]
	executable := os.Args[2]
	runtimeKindName := os.Args[3]

	startTime := time.Now()
	log.Printf("🔎 Fingerprinting the version-able executable %s to %s\n", executable, outputDir)

	if versionOutput, err := utils.GetExecutableVersionOutput(executable); err == nil {
		entries := make(map[string]string)
		entries["runtime-kind"] = runtimeKindName
		entries["runtime-kind-version"] = versionOutput
		utils.WriteEntries(outputDir, "runtime-kind.txt", entries)
	}
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	log.Printf("🕑 version-able executable fingerprint executed in time: %s\n", duration)
}
