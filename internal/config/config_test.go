package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetConfig(t *testing.T) {
	cfg := GetConfig()
	assert.NotNil(t, cfg)
	copy := cfg
	cfg = GetConfig()
	assert.Equal(t, cfg, copy)
}
