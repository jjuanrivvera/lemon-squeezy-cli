package commands

import (
	"strings"

	"github.com/njayp/ophis"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// excludedFromMCP are command subtrees an agent must never reach through the MCP server:
// setup/meta commands that switch instances or touch credentials. Excluding them keeps the
// agent on whatever profile is active at server startup — it can't read the key or disable
// its own rails (the `agent` guard).
var excludedFromMCP = []string{"agent", "auth", "config", "alias", "init", "doctor", "completion"}

// secretFlags must never be exposed as MCP tool inputs: they'd let an agent switch
// instances or exfiltrate the key.
var secretFlags = map[string]bool{
	"show-token": true,
	"profile":    true,
	"base-url":   true,
	"api-key":    true,
}

// mcpExcluded reports whether a command is in an excluded subtree (checked by name along the
// command's path so subcommands of an excluded parent are excluded too).
func mcpExcluded(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		for _, name := range excludedFromMCP {
			if c.Name() == name {
				return true
			}
		}
	}
	return false
}

// mcpConfig is the ophis configuration; exported via a function so tests can assert the
// surface (TestMCPExcludesSetupCommands).
func mcpConfig() *ophis.Config {
	return &ophis.Config{
		Selectors: []ophis.Selector{{
			// Expose every runnable command EXCEPT the setup/meta subtrees.
			CmdSelector: func(cmd *cobra.Command) bool {
				return !mcpExcluded(cmd)
			},
			// Drop secret/instance flags from both local and inherited flag sets.
			LocalFlagSelector:     mcpFlagAllowed,
			InheritedFlagSelector: mcpFlagAllowed,
		}},
	}
}

func mcpFlagAllowed(f *pflag.Flag) bool {
	return !secretFlags[strings.ToLower(f.Name)]
}

func init() {
	// ophis.Command registers the whole `mcp` subtree (start/tools/claude/vscode/...). It
	// walks the tree and replays the matching cobra command on tool invocation, so every
	// tool reuses the same client, keyring, profiles, and --dry-run — no separate handlers.
	rootCmd.AddCommand(ophis.Command(mcpConfig()))
}
