package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeYAML(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o600)
}

func TestLoad_MissingReturnsDefault(t *testing.T) {
	c, err := Load(filepath.Join(t.TempDir(), "nope.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "default", c.ActiveProfile)
	assert.Equal(t, "table", c.ResolveOutput())
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	c, err := Load(path)
	require.NoError(t, err)
	c.Output = "json"
	c.SetProfile("work", Profile{BaseURL: "https://x"})
	require.NoError(t, c.Use("work"))
	require.NoError(t, c.Save())

	got, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, "work", got.ActiveProfile)
	assert.Equal(t, "json", got.ResolveOutput())
	assert.Equal(t, "https://x", got.Profiles["work"].BaseURL)
}

func TestResolve_EnvOverride(t *testing.T) {
	c, _ := Load(filepath.Join(t.TempDir(), "c.yaml"))
	c.SetProfile("default", Profile{BaseURL: "https://file"})
	t.Setenv("LEMONSQUEEZY_BASE_URL", "https://env")
	p := c.Resolve("default")
	assert.Equal(t, "https://env", p.BaseURL)
}

func TestResolveOutput_EnvOverride(t *testing.T) {
	c, _ := Load(filepath.Join(t.TempDir(), "c.yaml"))
	c.Output = "csv"
	t.Setenv("LEMONSQUEEZY_OUTPUT", "yaml")
	assert.Equal(t, "yaml", c.ResolveOutput())
}

func TestResolveProfileName_EnvOverride(t *testing.T) {
	c, _ := Load(filepath.Join(t.TempDir(), "c.yaml"))
	t.Setenv("LEMONSQUEEZY_PROFILE", "prod")
	assert.Equal(t, "prod", c.ResolveProfileName())
}

func TestResolveProfileName_AccountEnvWins(t *testing.T) {
	c, _ := Load(filepath.Join(t.TempDir(), "c.yaml"))
	// LEMONSQUEEZY_ACCOUNT is the natural env for the --account selector and takes precedence
	// over the legacy LEMONSQUEEZY_PROFILE alias.
	t.Setenv("LEMONSQUEEZY_PROFILE", "legacy")
	t.Setenv("LEMONSQUEEZY_ACCOUNT", "primary")
	assert.Equal(t, "primary", c.ResolveProfileName())
}

func TestResolveProfileName_DefaultsToActive(t *testing.T) {
	c, _ := Load(filepath.Join(t.TempDir(), "c.yaml"))
	c.ActiveProfile = "work"
	assert.Equal(t, "work", c.ResolveProfileName())
}

func TestUse_UnknownProfileErrors(t *testing.T) {
	c, _ := Load(filepath.Join(t.TempDir(), "c.yaml"))
	assert.Error(t, c.Use("ghost"))
}

func TestDefaultDir_XDG(t *testing.T) {
	// Build expected paths with filepath.Join so the test is portable: on Windows the
	// separator is "\", and a hardcoded "/tmp/xdg/..." string would never match.
	xdg := filepath.Join(t.TempDir(), "xdg")
	t.Setenv("XDG_CONFIG_HOME", xdg)
	wantDir := filepath.Join(xdg, "lemon-squeezy-cli")
	assert.Equal(t, wantDir, DefaultDir())
	assert.Equal(t, filepath.Join(wantDir, "config.yaml"), DefaultPath())
}

func TestDefaultDir_HomeFallback(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	dir := DefaultDir()
	assert.Contains(t, dir, "lemon-squeezy-cli")
}

func TestPath(t *testing.T) {
	p := filepath.Join(t.TempDir(), "c.yaml")
	c, err := Load(p)
	require.NoError(t, err)
	assert.Equal(t, p, c.Path())
}

func TestLoad_InvalidYAML(t *testing.T) {
	p := filepath.Join(t.TempDir(), "c.yaml")
	require.NoError(t, writeYAML(p, "::: not yaml :::"))
	_, err := Load(p)
	assert.Error(t, err)
}
