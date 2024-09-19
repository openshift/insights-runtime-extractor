package utils

import (
	"crypto/sha256"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashEmptryString(t *testing.T) {
	h := sha256.New()
	actual := HashString(true, h, "")
	assert.Equal(t, "", actual)
}

func TestDoNotHash(t *testing.T) {
	h := sha256.New()
	str := "this must not be hashed"
	assert.Equal(t, str, HashString(false, h, str))
	assert.Equal(t, "", HashString(false, h, ""))
}

func TestHashValues(t *testing.T) {
	h := sha256.New()

	assert.Equal(t, "yqx1WBefoIAq", HashString(true, h, "rhel"))
	assert.Equal(t, "UOVueXxKyJ95", HashString(true, h, "Golang"))
	assert.Equal(t, "wbpgzhNYZQOi", HashString(true, h, "Java"))
	assert.Equal(t, "H8DHhTMugPuT", HashString(true, h, "Node.js"))
	assert.Equal(t, "zBVuVhrrC_vC", HashString(true, h, "Quarkus"))
	assert.Equal(t, "mjPGfOhcD0JI", HashString(true, h, "GraalVM"))
}
