package utils

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Fingerprints Fingerprints `toml:"fingerprints"`
}

type Fingerprints struct {
	VersionExecutables []VersionExecutable      `toml:"version-executables"`
	Java               []JavaRuntimeExecutables `toml:"java"`
}

type VersionExecutable struct {
	ProcessNames    []string `toml:"process-names"`
	RuntimeKindName string   `toml:"runtime-kind-name"`
}

type JavaRuntimeExecutables struct {
	RuntimeName                 string `toml:"runtime-name"`
	MainClass                   string `toml:"main-class"`
	MainJar                     string `toml:"main-jar,omitempty"`
	ReadManifestOfExecutableJar bool   `toml:"read-manifest-of-executable-jar"`
	JarVersionManifestEntry     string `toml:"jar-version-manifest-entry"`
}

func GetConfig(filepath string) (Config, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return Config{}, err
	}

	var config Config
	_, err = toml.Decode(string(content), &config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}
