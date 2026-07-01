package commands

import (
	"strings"
	"sync"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jjuanrivvera/lemon-squeezy-cli/internal/api"
)

var syntheticOnce sync.Once

// registerSynthetic adds a CRUD resource + custom read-only and destructive verbs directly,
// without importing the real resources package (which would create an import cycle). It
// exercises the generic builder and the MCP/agent classification. Idempotent across tests.
func registerSynthetic() {
	syntheticOnce.Do(func() {
		RegisterResource(ResourceSpec[api.Store]{
			Use:       "widgets",
			Short:     "synthetic",
			New:       func(c *api.Client) *api.Resource[api.Store] { return c.Stores() },
			Columns:   []string{"id", "name"},
			CanCreate: true,
			CanUpdate: true,
			CanDelete: true,
			Extra: []ExtraCommand{
				{ReadOnly: true, Build: func(func(bool) (*api.Client, renderFunc, error)) *cobra.Command {
					return &cobra.Command{Use: "scan", RunE: func(*cobra.Command, []string) error { return nil }}
				}},
				{Destructive: true, Build: func(func(bool) (*api.Client, renderFunc, error)) *cobra.Command {
					return &cobra.Command{Use: "purge", RunE: func(*cobra.Command, []string) error { return nil }}
				}},
			},
		})
		Setup() // drain the registrar onto rootCmd
	})
}

func TestMCPExcluded(t *testing.T) {
	registerSynthetic()
	for _, name := range excludedFromMCP {
		if cmd, _, err := rootCmd.Find([]string{name}); err == nil {
			assert.True(t, mcpExcluded(cmd), "%q must be excluded", name)
		}
	}
	wl, _, err := rootCmd.Find([]string{"widgets", "list"})
	require.NoError(t, err)
	assert.False(t, mcpExcluded(wl))
}

func TestMCPFlagAllowed(t *testing.T) {
	pf := rootCmd.PersistentFlags()
	assert.False(t, mcpFlagAllowed(pf.Lookup("show-token")))
	assert.False(t, mcpFlagAllowed(pf.Lookup("base-url")))
	// The account selector is excluded under BOTH its natural name and the hidden alias.
	assert.False(t, mcpFlagAllowed(pf.Lookup("account")))
	assert.False(t, mcpFlagAllowed(pf.Lookup("profile")))
	assert.True(t, mcpFlagAllowed(pf.Lookup("output")))
	assert.True(t, mcpFlagAllowed(pf.Lookup("jq")))
}

func TestMCPConfigSelectors(t *testing.T) {
	cfg := mcpConfig()
	require.Len(t, cfg.Selectors, 1)
	sel := cfg.Selectors[0]
	if authCmd, _, _ := rootCmd.Find([]string{"auth"}); authCmd != nil {
		assert.False(t, sel.CmdSelector(authCmd))
	}
}

func TestClassifyTree(t *testing.T) {
	registerSynthetic()
	cls := classifyTree(rootCmd)
	assert.Contains(t, cls.Irreversible, "widgets delete")
	assert.Contains(t, cls.Irreversible, "widgets purge")
	assert.Contains(t, cls.Write, "widgets create")
	assert.Contains(t, cls.Write, "widgets update")
	assert.Contains(t, cls.Read, "widgets list")
	assert.Contains(t, cls.Read, "widgets get")
	assert.Contains(t, cls.Read, "widgets scan")
	for _, p := range cls.Read {
		assert.False(t, strings.HasPrefix(p, "mcp"))
		assert.False(t, strings.HasPrefix(p, "auth"))
	}
}

func TestAgentHostsRender(t *testing.T) {
	cls := classification{
		Read:         []string{"orders list"},
		Write:        []string{"customers create"},
		Irreversible: []string{"orders refund"},
	}
	cc, err := renderHostConfig("claude-code", "lsqueezy", cls, hostOptions{})
	require.NoError(t, err)
	assert.Contains(t, cc, "Bash(lsqueezy orders refund:*)")
	assert.Contains(t, cc, "mcp__lsqueezy__lsqueezy_orders_refund")

	codex, err := renderHostConfig("codex", "lsqueezy", cls, hostOptions{})
	require.NoError(t, err)
	assert.Contains(t, codex, "read-only")

	oc, err := renderHostConfig("opencode", "lsqueezy", cls, hostOptions{})
	require.NoError(t, err)
	assert.Contains(t, oc, `"lsqueezy orders refund": "deny"`)
	assert.Contains(t, oc, `"lsqueezy orders list": "allow"`)

	_, err = renderHostConfig("bogus", "lsqueezy", cls, hostOptions{})
	assert.Error(t, err)
}

func TestAgentAllWritesDeny(t *testing.T) {
	cls := classification{Write: []string{"customers create"}}
	cc, err := renderHostConfig("claude-code", "lsqueezy", cls, hostOptions{AllWrites: true})
	require.NoError(t, err)
	assert.Contains(t, cc, `"deny"`)
	oc, err := renderHostConfig("opencode", "lsqueezy", cls, hostOptions{AllWrites: true})
	require.NoError(t, err)
	assert.Contains(t, oc, `"deny"`)
}

func TestApplyClientFilters(t *testing.T) {
	type rec struct {
		Name string `json:"name"`
		N    int    `json:"n"`
	}
	items := []rec{{"a", 1}, {"b", 2}}
	got, err := applyClientFilters(items, []string{"name=a"})
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "a", got[0].Name)

	got, err = applyClientFilters(items, nil)
	require.NoError(t, err)
	assert.Len(t, got, 2)

	_, err = applyClientFilters(items, []string{"bad"})
	assert.Error(t, err)
}

func TestSingular(t *testing.T) {
	assert.Equal(t, "order", singular("orders"))
	assert.Equal(t, "discount-redemption", singular("discount-redemptions"))
	assert.Equal(t, "x", singular("x"))
}

func TestCommandPath(t *testing.T) {
	registerSynthetic()
	cmd, _, err := rootCmd.Find([]string{"widgets", "delete"})
	require.NoError(t, err)
	assert.Equal(t, "widgets delete", commandPath(cmd))
}

func TestNormalizeColumns(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, normalizeColumns([]string{" a ", "", "b"}))
}

func TestCoerceValue(t *testing.T) {
	assert.Equal(t, float64(5), coerceValue("5"))
	assert.Equal(t, true, coerceValue("true"))
	assert.Equal(t, "hello", coerceValue("hello"))
	assert.Equal(t, nil, coerceValue("null"))
}
