package commands_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jjuanrivvera/lemon-squeezy-cli/commands"
)

func TestE2E_AuthLoginLogoutStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/users/me", r.URL.Path)
		jsonAPI(w, `{"data":{"type":"users","id":"1","attributes":{"name":"Jo","email":"jo@x.co"}}}`)
	}))
	defer srv.Close()
	// No env API key: login must store it. Use a temp config dir + base URL override.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("LEMONSQUEEZY_BASE_URL", srv.URL)

	out := captureStdout(t, func() { require.NoError(t, run(t, "auth", "login", "--api-key", "k123")) })
	assert.Contains(t, out, "verified as Jo")

	out = captureStdout(t, func() { require.NoError(t, run(t, "auth", "status")) })
	assert.Contains(t, out, "valid")
	assert.Contains(t, out, "jo@x.co")

	out = captureStdout(t, func() { require.NoError(t, run(t, "auth", "logout")) })
	assert.Contains(t, out, "key removed")
}

func TestE2E_ConfigUseProfile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("LEMONSQUEEZY_API_KEY", "k")
	require.NoError(t, run(t, "config", "set", "base_url", "https://api.lemonsqueezy.com/v1"))
	// create a second profile via --profile then use it
	require.NoError(t, run(t, "--profile", "staging", "config", "set", "base_url", "https://staging"))
	err := run(t, "config", "use", "staging")
	require.NoError(t, err)
}

func TestE2E_AliasExpansion(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	require.NoError(t, run(t, "alias", "set", "myls", "stores list"))
	root := commands.Root()
	expanded := commands.ExpandAlias(root, []string{"myls", "-o", "json"})
	assert.Equal(t, []string{"stores", "list", "-o", "json"}, expanded)
}

func TestE2E_DryRunGet(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	out := captureStdout(t, func() {
		require.NoError(t, run(t, "orders", "get", "5", "--dry-run", "--base-url", "https://api.lemonsqueezy.com/v1"))
	})
	assert.Contains(t, out, "curl -X GET")
	assert.Contains(t, out, "/orders/5")
}

func TestE2E_Quiet(t *testing.T) {
	assert.IsType(t, false, commands.Quiet())
}
