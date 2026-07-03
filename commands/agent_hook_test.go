package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// guardTestClassification mirrors the real tree's money-touching commands so the hook and
// renderer tests exercise the exact payloads an agent would emit, without depending on the
// resources package (import cycle).
func guardTestClassification() classification {
	return classification{
		Read:  []string{"orders list", "orders get", "stores list", "license validate"},
		Write: []string{"customers create", "checkouts create", "license activate"},
		Irreversible: []string{
			"discounts delete",
			"license deactivate",
			"orders refund",
			"subscription-invoices refund",
			"subscriptions cancel",
			"webhooks delete",
		},
	}
}

// TestClaudeCodeIncludesHookAndAPIDenies asserts the claude-code output ships the
// PreToolUse hook (wired for both Bash and the MCP namespace) and hard-blocks the raw-api
// escape hatch for destructive HTTP methods.
func TestClaudeCodeIncludesHookAndAPIDenies(t *testing.T) {
	out, err := renderHostConfig("claude-code", "lsqueezy", guardTestClassification(), hostOptions{})
	require.NoError(t, err)

	// Hook script present and path-prefix hardened.
	assert.Contains(t, out, ".claude/hooks/lsqueezy-guard.sh")
	assert.Contains(t, out, `([^[:space:]]*/)?lsqueezy`)
	assert.Contains(t, out, "bash_is_blocked")
	assert.Contains(t, out, "api_is_blocked")

	// Settings wire the hook on both surfaces.
	assert.Contains(t, out, `"PreToolUse"`)
	assert.Contains(t, out, `"matcher": "Bash"`)
	assert.Contains(t, out, `"matcher": "mcp__lsqueezy__"`)

	// Raw-api destructive methods are denied (method is the first positional arg).
	for _, m := range []string{"DELETE", "PUT", "POST", "PATCH"} {
		assert.Contains(t, out, "Bash(lsqueezy api "+m+":*)")
	}
}

// TestOpenCodeSchemaAndAPIDenies asserts opencode rules live under permission.bash (the
// schema OpenCode actually reads) and include the raw-api method denies.
func TestOpenCodeSchemaAndAPIDenies(t *testing.T) {
	out, err := renderHostConfig("opencode", "lsqueezy", guardTestClassification(), hostOptions{})
	require.NoError(t, err)

	var cfg map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &cfg))
	perm, ok := cfg["permission"].(map[string]any)
	require.True(t, ok, "permission (singular) map must exist")
	bash, ok := perm["bash"].(map[string]any)
	require.True(t, ok, "permission.bash map must exist")

	assert.Equal(t, "deny", bash["lsqueezy orders refund"])
	assert.Equal(t, "deny", bash["lsqueezy api DELETE"])
	assert.Equal(t, "ask", bash["lsqueezy checkouts create"])
	assert.Equal(t, "allow", bash["lsqueezy orders list"])
	assert.Equal(t, "deny", perm["lsqueezy_orders_refund"])
}

// TestCodexTopLevelKeys asserts the codex output uses Codex's real top-level config keys
// (sandbox_mode / approval_policy), not an invented [sandbox] table.
func TestCodexTopLevelKeys(t *testing.T) {
	out, err := renderHostConfig("codex", "lsqueezy", guardTestClassification(), hostOptions{})
	require.NoError(t, err)
	assert.Contains(t, out, `sandbox_mode    = "read-only"`)
	assert.Contains(t, out, `approval_policy = "untrusted"`)
	assert.NotContains(t, out, "[sandbox]")
}

// TestClaudeCodeWriteInstallsFiles asserts --write lands the hook and settings under the
// project root's .claude/ directory and refuses to overwrite existing files.
func TestClaudeCodeWriteInstallsFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, ".git"), 0o750))
	t.Chdir(dir)

	out, err := renderHostConfig("claude-code", "lsqueezy", guardTestClassification(), hostOptions{Write: true})
	require.NoError(t, err)
	assert.Contains(t, out, "wrote ")

	hook, err := os.ReadFile(filepath.Join(dir, ".claude", "hooks", "lsqueezy-guard.sh"))
	require.NoError(t, err)
	assert.Contains(t, string(hook), "blocked_cmds=(")

	settings, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	require.NoError(t, err)
	var cfg map[string]any
	require.NoError(t, json.Unmarshal(settings, &cfg))
	require.Contains(t, cfg, "hooks")

	// Second write must not clobber.
	out2, err := renderHostConfig("claude-code", "lsqueezy", guardTestClassification(), hostOptions{Write: true})
	require.NoError(t, err)
	assert.Contains(t, out2, "already exists")
}

// TestHookScriptBashExecution exercises the generated hook with real bash against the
// adversarial payload battery: obfuscation, command chaining, path-invoked binaries, the
// raw-api escape hatch, and the MCP branch.
func TestHookScriptBashExecution(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("bash hook tests require a POSIX shell; skipping on windows")
	}
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not found in PATH; skipping hook execution tests")
	}

	blocked, _ := blockedPaths(guardTestClassification(), hostOptions{})
	hookContent := hookScript("lsqueezy", blocked)
	hookFile := filepath.Join(t.TempDir(), "lsqueezy-guard.sh")
	require.NoError(t, os.WriteFile(hookFile, []byte(hookContent), 0o755)) // #nosec G306 -- executable hook

	bashPayload := func(command string) string {
		b, _ := json.Marshal(map[string]any{
			"tool_name":  "Bash",
			"tool_input": map[string]any{"command": command},
		})
		return string(b)
	}
	mcpPayload := func(toolName string) string {
		b, _ := json.Marshal(map[string]any{
			"tool_name":  toolName,
			"tool_input": map[string]any{},
		})
		return string(b)
	}
	runHook := func(t *testing.T, payload string) string {
		t.Helper()
		cmd := exec.Command(bash, hookFile) // #nosec G204 -- test fixture
		cmd.Stdin = strings.NewReader(payload)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		// The hook always exits 0; the decision is in the JSON output.
		require.NoError(t, cmd.Run(), "hook output: %s", out.String())
		return out.String()
	}
	isDenied := func(output string) bool {
		return strings.Contains(output, `"permissionDecision":"deny"`)
	}

	cases := []struct {
		name       string
		payload    string
		wantDenied bool
	}{
		// SHOULD DENY
		{"plain_refund", bashPayload("lsqueezy orders refund 123"), true},
		{"plain_cancel", bashPayload("lsqueezy subscriptions cancel 9"), true},
		{"plain_deactivate", bashPayload("lsqueezy license deactivate --key k --instance-id i"), true},
		{"quote_obfuscation", bashPayload(`lsqueezy orders re""fund 123`), true},
		{"backslash_obfuscation", bashPayload(`lsqueezy orders re\fund 123`), true},
		{"after_semicolon", bashPayload("echo hi; lsqueezy orders refund 123"), true},
		{"after_pipe", bashPayload("echo 123 | lsqueezy orders refund 123"), true},
		{"after_and", bashPayload("true && lsqueezy discounts delete 5"), true},
		{"newline_continuation", bashPayload("lsqueezy orders \\\nrefund 123"), true},
		{"api_delete", bashPayload("lsqueezy api DELETE /discounts/5"), true},
		{"api_delete_lowercase", bashPayload("lsqueezy api delete /discounts/5"), true},
		{"api_post", bashPayload("lsqueezy api POST /checkouts -d @x.json"), true},
		{"env_prefix", bashPayload("env X=1 lsqueezy orders refund 123"), true},
		// Path-invoked binaries (the ([^[:space:]]*/)? hardening; would bypass a
		// bare-binary-name anchor).
		{"relative_path_binary", bashPayload("./bin/lsqueezy orders refund 123"), true},
		{"absolute_path_binary", bashPayload("/usr/local/bin/lsqueezy orders refund 123"), true},
		{"absolute_path_api_delete", bashPayload("/usr/local/bin/lsqueezy api DELETE /webhooks/1"), true},

		// SHOULD ALLOW
		{"read_command", bashPayload("lsqueezy orders list --all"), false},
		{"blocked_verb_in_argument", bashPayload(`lsqueezy customers create --set name="refund dept"`), false},
		{"unrelated_file", bashPayload("cat orders_refund.go"), false},
		{"api_get_refund_in_path", bashPayload("lsqueezy api GET /orders/1/refund-status"), false},
		// A different binary whose name merely ends in "lsqueezy" must NOT be blocked.
		{"other_binary_suffix", bashPayload("mylsqueezy orders refund 123"), false},
		// Shell separator glued directly to the verb must still match.
		{"glued_separator_denied", bashPayload("lsqueezy webhooks delete;true"), true},
		{"glued_pipe_denied", bashPayload("lsqueezy subscriptions cancel|cat"), true},

		// MCP branch: exact set membership.
		{"mcp_blocked_tool", mcpPayload("mcp__lsqueezy__lsqueezy_orders_refund"), true},
		{"mcp_read_tool", mcpPayload("mcp__lsqueezy__lsqueezy_orders_list"), false},
		{"mcp_near_miss_suffix", mcpPayload("mcp__lsqueezy__lsqueezy_orders_refunded"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := runHook(t, tc.payload)
			assert.Equal(t, tc.wantDenied, isDenied(output), "output: %s", output)
		})
	}
}

// TestHookScriptBashExecutionNoJq exercises the fail-safe no-jq fallback path.
func TestHookScriptBashExecutionNoJq(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("bash hook tests require a POSIX shell; skipping on windows")
	}
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not found in PATH; skipping hook execution tests")
	}

	blocked, _ := blockedPaths(guardTestClassification(), hostOptions{})
	hookContent := hookScript("lsqueezy", blocked)
	tmpDir := t.TempDir()
	hookFile := filepath.Join(tmpDir, "lsqueezy-guard.sh")
	require.NoError(t, os.WriteFile(hookFile, []byte(hookContent), 0o755)) // #nosec G306 -- executable hook

	// Build a strict PATH containing ONLY the tools the hook needs, minus jq.
	// (Merely prepending an empty dir does NOT hide jq — it stays reachable
	// later in PATH and the no-jq branch never runs; that flaw previously
	// masked a fail-open bug in this branch.)
	strictPath := filepath.Join(tmpDir, "strictbin")
	require.NoError(t, os.Mkdir(strictPath, 0o750))
	for _, tool := range []string{"cat", "tr", "grep", "sed", "printf"} {
		p, err := exec.LookPath(tool)
		if err != nil {
			t.Skipf("required tool %s not found: %v", tool, err)
		}
		require.NoError(t, os.Symlink(p, filepath.Join(strictPath, tool)))
	}
	noJqPath := strictPath

	bashPayload := func(command string) string {
		b, _ := json.Marshal(map[string]any{
			"tool_name":  "Bash",
			"tool_input": map[string]any{"command": command},
		})
		return string(b)
	}
	runHook := func(t *testing.T, payload string) string {
		t.Helper()
		cmd := exec.Command(bash, hookFile) // #nosec G204 -- test fixture
		cmd.Stdin = strings.NewReader(payload)
		env := []string{}
		for _, e := range os.Environ() {
			if !strings.HasPrefix(e, "PATH=") {
				env = append(env, e)
			}
		}
		cmd.Env = append(env, "PATH="+noJqPath)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		require.NoError(t, cmd.Run(), "hook output: %s", out.String())
		return out.String()
	}

	cases := []struct {
		name       string
		payload    string
		wantDenied bool
	}{
		{"nojq_refund_denied", bashPayload("lsqueezy orders refund 123"), true},
		{"nojq_path_binary_denied", bashPayload("./bin/lsqueezy orders refund 123"), true},
		{"nojq_api_delete_denied", bashPayload("lsqueezy api DELETE /discounts/5"), true},
		{"nojq_read_allowed", bashPayload("lsqueezy orders list"), false},
		{"nojq_unrelated_allowed", bashPayload("cat orders_refund.go"), false},
		// Regression: the compact JSON payload glues the command to its key
		// ("command":"lsqueezy ...). Without JSON-punctuation flattening the
		// anchor can never match and the branch is silently fail-open.
		{"nojq_glued_json_key_denied", bashPayload("lsqueezy subscriptions cancel 7"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := runHook(t, tc.payload)
			assert.Equal(t, tc.wantDenied, strings.Contains(out, `"permissionDecision":"deny"`), "output: %s", out)
		})
	}
}

// TestClassifyTreeMoneyVerbsNotRead is the bug-class-1 regression guard: no money-touching
// verb may ever land in the Read (auto-allowed) bucket, and the raw-api escape hatch must
// stay gated (Write) rather than allowed.
func TestClassifyTreeMoneyVerbsNotRead(t *testing.T) {
	registerSynthetic()
	cls := classifyTree(rootCmd)

	for _, p := range cls.Read {
		leaf := p[strings.LastIndex(p, " ")+1:]
		for _, verb := range []string{"refund", "cancel", "delete", "deactivate", "create", "update", "activate", "archive", "api"} {
			assert.NotEqual(t, verb, leaf, "mutating verb %q must not be in the Read bucket (path %q)", verb, p)
		}
	}
	assert.Contains(t, cls.Write, "api", "raw-api escape hatch must require approval")
	assert.NotContains(t, cls.Read, "api")
}
