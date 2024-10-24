package utils

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func WriteEntries(outputDir string, fileName string, entries map[string]string) {
	file, err := os.Create(filepath.Join(outputDir, fileName))
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

// / Read a key=value file and return its content in a map.
// /
// / Key and values are separated by '='.
func ReadPropertiesFile(filePath string) (map[string]string, bool) {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		// File does not exist
		return nil, false
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	properties := make(map[string]string)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		// Split the line into key and value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.Trim(strings.TrimSpace(parts[0]), "\"")
			value := strings.Trim(strings.TrimSpace(parts[1]), "\"")
			properties[key] = value
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return nil, false
	}

	return properties, true
}

func GetExecutableVersionOutput(executable string) (string, error) {
	out, err := exec.Command(executable, "--version").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// FindExecutableInPath returns the directory in the $PATH env var that contains the given executable
func FindExecutableInPath(executable string, pathEnvVar string) (string, error) {
	paths := strings.Split(pathEnvVar, string(os.PathListSeparator))
	for _, dir := range paths {
		fullPath := filepath.Join(dir, executable)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}
	return "", fmt.Errorf("executable %s not found in PATH", executable)
}

func ReadManifest(content string) map[string]string {
	manifestEntries := make(map[string]string)
	currentKey := ""
	for _, entry := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		key, value, keyFound := strings.Cut(entry, ": ")
		if keyFound {
			manifestEntries[key] = value
			currentKey = key
		} else {
			manifestEntries[currentKey] = manifestEntries[currentKey] + strings.TrimLeft(entry, " ")
		}
	}
	return manifestEntries
}

func GetJarManifest(jarPath string) (map[string]string, error) {
	r, err := zip.OpenReader(jarPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open JAR file: %w", err)
	}
	defer r.Close()

	for _, file := range r.File {
		if file.Name == "META-INF/MANIFEST.MF" {
			manifest, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open manifest file: %w", err)
			}
			defer manifest.Close()

			var buffer bytes.Buffer
			if _, err := io.Copy(&buffer, manifest); err != nil {
				return nil, fmt.Errorf("failed to read manifest file: %w", err)
			}

			manifestContent := buffer.String()

			return ReadManifest(manifestContent), nil
		}
	}

	return nil, fmt.Errorf("manifest file not found in jar %s", jarPath)
}

func JarFileContainsClass(jarPath string, className string) bool {
	classFile := strings.ReplaceAll(className, ".", "/") + ".class"

	r, err := zip.OpenReader(jarPath)
	if err != nil {
		return false
	}
	defer r.Close()

	for _, file := range r.File {
		if file.Name == classFile {
			return true
		}
	}
	return false
}
