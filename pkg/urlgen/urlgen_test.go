package urlgen

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateURL(t *testing.T) {
	baseURL := "https://short.url"
	result := GenerateURL(baseURL)

	if len(result) != len(baseURL)+1+6 {
		t.Errorf("Expected URL length %d, got %d", len(baseURL)+7, len(result))
	}

	if !strings.HasPrefix(result, baseURL) {
		t.Errorf("Expected URL to start with %s, got %s", baseURL, result)
	}
	_, err := url.Parse(result)
	assert.NoError(t, err)
}
