package utils

import (
	"bufio"
	"fmt"
	"os"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadPropertiesFile(t *testing.T) {
	filePath := "test.properties"
	expectedProps := map[string]string{
		"os-release-id":         "rhel",
		"os-release-version-id": "8.9",
	}
	// Create properties file
	err := writePropertiesFile(filePath, expectedProps)
	if err != nil {
		t.Fatalf("Failed to write properties file: %v", err)
	}
	defer os.Remove(filePath)

	actualProps, exists := ReadPropertiesFile(filePath)
	assert.True(t, exists)

	assert.Equal(t, expectedProps, actualProps)
}

func TestReadPropertiesFileWithComment(t *testing.T) {
	filePath := "test.properties"
	expectedProps := map[string]string{
		"os-release-id":          "rhel",
		"#os-release-version-id": "8.9",
	}
	// Create properties file
	err := writePropertiesFile(filePath, expectedProps)
	if err != nil {
		t.Fatalf("Failed to write properties file: %v", err)
	}
	defer os.Remove(filePath)

	actualProps, exists := ReadPropertiesFile(filePath)
	assert.True(t, exists)

	assert.Equal(t, 1, len(actualProps))
	assert.Equal(t, expectedProps["os-release-id"], actualProps["os-release-id"])
}

func TestReadPropertiesFileFromNonExistingFile(t *testing.T) {
	filePath := "that-file-does-not-exist"

	actualProps, exists := ReadPropertiesFile(filePath)
	assert.False(t, exists)

	assert.Nil(t, actualProps)
}

func writePropertiesFile(filePath string, properties map[string]string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for key, value := range properties {
		_, err := writer.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}
