package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadConfig(t *testing.T) {
	config, err := GetConfig("../../../extractor/config/config.toml")
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, 2, len(config.Fingerprints.VersionExecutables))
	assert.Equal(t, 4, len(config.Fingerprints.Java))
}
