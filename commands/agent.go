package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// classification buckets a command into read / write / irreversible using the same MCP
// annotations the generic builder stamps. The agent guard derives host safety config from
// the LIVE tree, so it stays correct as resources are added.
type classification struct {
	Read         []string // safe; no approval needed
	Write        []string // mutating but reversible; require approval
	Irreversible []string // destructive (delete); hard-block by default
}

// classifyTree walks the command tree and buckets every runnable leaf by its annotation.
func classifyTree(root *cobra.Command) classification {
	var c classification
	var walk func(cmd *cobra.Command)
	walk = func(cmd *cobra.Command) {
		for _, sub := range cmd.Commands() {
			walk(sub)
		}
		if !cmd.Runnable() || mcpExcluded(cmd) || agentSkip(cmd) {
			return
		}
		path := commandPath(cmd)
		switch {
		case cmd.Annotations[annDestructive] == "true":
			c.Irreversible = append(c.Irreversible, path)
		case cmd.Annotations[annWrite] == "true":
			c.Write = append(c.Write, path)
		case cmd.Annotations[annReadOnly] == "true":
			c.Read = append(c.Read, path)
		default:
			// Unannotated runnable leaves are treated as writes (fail safe, not open).
			c.Write = append(c.Write, path)
		}
	}
	walk(root)
	sort.Strings(c.Read)
	sort.Strings(c.Write)
	sort.Strings(c.Irreversible)
	return c
}

// agentSkip excludes commands that aren't API operations from the guard: cobra's `help`
// and the ophis-managed `mcp` subtree (which administers the server, not the API).
func agentSkip(cmd *cobra.Command) bool {
	for c := cmd; c != nil; c = c.Parent() {
		if c.Name() == "help" || c.Name() == "mcp" {
			return true
		}
	}
	return false
}

// commandPath returns the space-joined path below the root ("orders refund").
func commandPath(cmd *cobra.Command) string {
	var parts []string
	for c := cmd; c != nil && c.Parent() != nil; c = c.Parent() {
		parts = append([]string{c.Name()}, parts...)
	}
	return strings.Join(parts, " ")
}

func init() {
	var host string
	var allWrites, write bool
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Generate AI-agent safety config from the live command tree",
	}
	guard := &cobra.Command{
		Use:   "guard",
		Short: "Emit host safety config that blocks destructive ops for an agent driving lsqueezy",
		Long: `Classify every lsqueezy command (read / write / irreversible) using the same
annotations the MCP server uses, then emit safety config for the chosen agent host.

Reads are left free, writes require approval, and irreversible verbs (refund, cancel,
delete, deactivate) are blocked. Because it derives from the live tree, it stays correct
as resources are added.

For claude-code the output includes a PreToolUse hook script alongside the permission
rules: permission rules are literal prefix patterns, so on their own they can be
sidestepped by path-invoking the binary (./bin/lsqueezy ...) or quote obfuscation. The
hook re-checks every Bash command (anchored command-position matching, de-obfuscation)
and hard-blocks blocked lsqueezy MCP tools by exact name. The "lsqueezy api" escape
hatch is blocked for DELETE/PUT/POST/PATCH at the method position; variable indirection
and shell aliases are NOT defeated — run MCP-only or in a read-only sandbox for a hard
guarantee.`,
		Example: "  lsqueezy agent guard --host claude-code\n  lsqueezy agent guard --host codex\n  lsqueezy agent guard --host opencode",
		RunE: func(_ *cobra.Command, _ []string) error {
			cls := classifyTree(rootCmd)
			out, err := renderHostConfig(host, "lsqueezy", cls, hostOptions{AllWrites: allWrites, Write: write})
			if err != nil {
				return err
			}
			fmt.Print(out)
			return nil
		},
	}
	guard.Flags().StringVar(&host, "host", "claude-code", "agent host: claude-code|codex|opencode")
	guard.Flags().BoolVar(&allWrites, "all-writes", false, "block all writes, not just irreversible ones")
	guard.Flags().BoolVar(&write, "write", false, "write the config to the host's default path instead of stdout")
	annotate(guard, annReadOnly)
	cmd.AddCommand(guard)
	rootCmd.AddCommand(cmd)
}
