package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	i := Get()
	assert.NotEmpty(t, i.Version)
}

func TestString(t *testing.T) {
	Version, Commit, Date = "1.2.3", "abc", "2026-01-01"
	assert.Contains(t, Get().String(), "1.2.3")
	assert.Contains(t, Get().String(), "abc")
}
