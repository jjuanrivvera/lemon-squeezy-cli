// Package config implements config precedence flag > env > config file > default, with
// named profiles for multi-account use. Secrets are never stored here — only non-secret
// bits (base URL, default output, active profile).
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// EnvPrefix namespaces env overrides, e.g. LEMONSQUEEZY_OUTPUT, LEMONSQUEEZY_PROFILE, LEMONSQUEEZY_BASE_URL.
	EnvPrefix = "LEMONSQUEEZY"
	// APIKeyEnv is the documented override for the API key (read by callers, not stored here).
	APIKeyEnv = "LEMONSQUEEZY_API_KEY"
)

// Profile holds the non-secret settings for one account/instance. Lemon Squeezy is a single
// cloud API, but profiles let one machine hold several accounts (e.g. a live key and a
// test-mode key) and switch between them.
type Profile struct {
	BaseURL string `yaml:"base_url,omitempty"`
}

// Config is the on-disk config file shape.
type Config struct {
	ActiveProfile string             `yaml:"active_profile"`
	Output        string             `yaml:"output,omitempty"`
	Profiles      map[string]Profile `yaml:"profiles"`

	path string `yaml:"-"`
}

// DefaultDir returns the config directory: $XDG_CONFIG_HOME/lemon-squeezy-cli or ~/.lemon-squeezy-cli.
func DefaultDir() string {
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "lemon-squeezy-cli")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".lemon-squeezy-cli"
	}
	return filepath.Join(home, ".lemon-squeezy-cli")
}

// DefaultPath is the config file path.
func DefaultPath() string { return filepath.Join(DefaultDir(), "config.yaml") }

// Load reads the config from path (DefaultPath if empty), returning a populated default
// when the file is absent.
func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultPath()
	}
	c := &Config{
		ActiveProfile: "default",
		Output:        "table",
		Profiles:      map[string]Profile{"default": {BaseURL: ""}},
		path:          path,
	}
	raw, err := os.ReadFile(path) // #nosec G304 -- path is a config location, not user data
	if err != nil {
		if os.IsNotExist(err) {
			return c, nil
		}
		return nil, err
	}
	if err := yaml.Unmarshal(raw, c); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	if c.Profiles == nil {
		c.Profiles = map[string]Profile{}
	}
	if c.ActiveProfile == "" {
		c.ActiveProfile = "default"
	}
	c.path = path
	return c, nil
}

// Path returns the file path this config is bound to.
func (c *Config) Path() string { return c.path }

// Save writes the config to disk with owner-only permissions.
func (c *Config) Save() error {
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	out, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	// #nosec G306 -- config is 0600 even though it holds no secrets, by policy.
	return os.WriteFile(c.path, out, 0o600)
}

// Resolve applies precedence for a profile's settings. The flag layer is applied by the
// caller (cobra); here we merge env > file > default.
func (c *Config) Resolve(profileName string) Profile {
	if profileName == "" {
		profileName = c.ActiveProfile
	}
	p := c.Profiles[profileName] // zero value if absent

	if v := os.Getenv(EnvPrefix + "_BASE_URL"); v != "" {
		p.BaseURL = v
	}
	return p
}

// ResolveOutput returns the default output format honoring env > file > "table".
func (c *Config) ResolveOutput() string {
	if v := os.Getenv(EnvPrefix + "_OUTPUT"); v != "" {
		return v
	}
	if c.Output != "" {
		return c.Output
	}
	return "table"
}

// ResolveProfileName honors LEMONSQUEEZY_PROFILE over the file's active profile.
func (c *Config) ResolveProfileName() string {
	if v := os.Getenv(EnvPrefix + "_PROFILE"); v != "" {
		return v
	}
	return c.ActiveProfile
}

// SetProfile creates or updates a profile.
func (c *Config) SetProfile(name string, p Profile) {
	if c.Profiles == nil {
		c.Profiles = map[string]Profile{}
	}
	c.Profiles[name] = p
}

// Use switches the active profile, erroring if it doesn't exist.
func (c *Config) Use(name string) error {
	if _, ok := c.Profiles[name]; !ok {
		return fmt.Errorf("profile %q not found", name)
	}
	c.ActiveProfile = name
	return nil
}
