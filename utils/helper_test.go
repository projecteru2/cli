package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvParser(t *testing.T) {
	assert.NoError(t, os.Setenv("ABC", "123"))
	s := []byte(`{{ .ABC }} eru`)
	b, err := EnvParser(s)
	assert.NoError(t, err)
	assert.Contains(t, string(b), "123")
}
