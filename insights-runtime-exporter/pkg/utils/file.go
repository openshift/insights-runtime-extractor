package utils

import (
	"bufio"
	"os"
	"strings"
)

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
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			properties[key] = value
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return nil, false
	}

	return properties, true
}
