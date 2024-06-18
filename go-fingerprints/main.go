package main

import (
	"bufio"
	"debug/buildinfo"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/saferwall/elf"
)

func main() {
	// The program has parameters:
	// - 1 - the subdirectory to write the manifest to
	// - 2 - the executable current working directory
	// - 2 - the name of the executable
	outputDir := os.Args[1]
	cwd := os.Args[2]
	executable := os.Args[3]

	startTime := time.Now()

	path := executable
	if !strings.HasPrefix(executable, "/") {
		path = cwd + executable
	}

	isElf := isElfExecutable(path)
	if !isElf {
		return
	}

	entries := make(map[string]string)

	goVersion, err := getGoVersion(path)
	if err == nil && goVersion != "" {

		entries["runtime-kind"] = "Golang"
		entries["runtime-kind-version"] = goVersion

		writeFingerprints(outputDir, "runtime-kind", entries)

		endTime := time.Now()
		duration := endTime.Sub(startTime)
		fmt.Printf("Golang fingerprint executed in time: %s\n", duration)
		return
	}

	// check whether the executable is a GraalVM executable
	graalVMExec, err := checkGraalVMExecutable(path)
	if err != nil {
		return
	}

	if graalVMExec {
		entries["runtime-kind"] = "GraalVM"
		writeFingerprints(outputDir, "runtime-kind", entries)

		containsQuarkusStrings, err := checkQuarkusStrings(path)
		if err != nil {
			return
		}
		if containsQuarkusStrings {

			runtimeEntries := make(map[string]string)
			runtimeEntries["Quarkus"] = ""
			writeFingerprints(outputDir, "quarkus", runtimeEntries)

		}
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		fmt.Printf("GraalVM/Quarkus fingerprint executed in time: %s\n", duration)
	}
}

func checkQuarkusStrings(executable string) (bool, error) {
	file, err := os.Open(executable)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	found := get_strings(file, 14, 14, true)
	for _, str := range found {
		if strings.Contains(str, "quarkus.native") {
			return true, nil
		}
	}
	return false, nil
}

func checkGraalVMExecutable(executable string) (bool, error) {
	p, err := elf.New(executable)
	defer p.CloseFile()
	if err != nil {
		return false, err
	}
	err = p.Parse()
	if err != nil {
		return false, err
	}
	for _, section := range p.F.ELFBin64.Sections64 {
		if section.SectionName == ".svm_heap" {
			return true, nil
		}
	}
	return false, nil
}

func getGoVersion(executable string) (string, error) {
	bi, err := buildinfo.ReadFile(executable)
	if err != nil {
		return "", err
	}
	return bi.GoVersion, nil
}

func writeFingerprints(outputDir string, fingerPrintKey string, entries map[string]string) {
	file, err := os.Create(filepath.Join(outputDir, fingerPrintKey+"-fingerprints.txt"))
	if err != nil {
		panic(err)
	}
	defer file.Close()

	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		file.WriteString(k + "=" + entries[k] + "\n")
	}
}

// copied from https://github.com/robpike/strings/blob/master/strings.go
func get_strings(file *os.File, min int, max int, ascii bool) []string {
	in := bufio.NewReader(file)
	str := make([]rune, 0, max)
	found := make([]string, 1)
	filePos := int64(0)
	add := func() {
		if len(str) >= min {
			s := string(str)
			found = append(found, s)
		}
		str = str[0:0]
	}
	for {
		var (
			r   rune
			wid int
			err error
		)
		// One string per loop.
		for ; ; filePos += int64(wid) {
			r, wid, err = in.ReadRune()
			if err != nil {
				if err != io.EOF {
					panic(err)
				}
				return found
			}
			if !strconv.IsPrint(r) || ascii && r >= 0xFF {
				add()
				continue
			}
			// It's printable. Keep it.
			if len(str) >= max {
				add()
			}
			str = append(str, r)
		}
	}
}

const (
	ELFMAG = "\177ELF"
)

func isElfExecutable(executable string) bool {

	r, err := os.Open(executable)
	if err != nil {
		return false
	}
	defer r.Close()

	header := make([]byte, 4)
	_, err = io.ReadFull(r, header[:])
	if err != nil {
		return false
	}

	return string(header) == ELFMAG
}
