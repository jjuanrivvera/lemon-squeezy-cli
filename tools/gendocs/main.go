// Command gendocs generates docs/commands/*.md from the cobra tree so the reference docs
// stay in lockstep with the CLI (the CI drift check fails if they diverge). It also writes
// a commands/index.md landing page that the MkDocs nav points at.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/jjuanrivvera/lemon-squeezy-cli/commands"

	_ "github.com/jjuanrivvera/lemon-squeezy-cli/resources"
)

func main() {
	root := commands.Setup()
	root.DisableAutoGenTag = true // stable output: no timestamp line that would churn the drift check

	out := filepath.Join("docs", "commands")
	if err := os.MkdirAll(out, 0o750); err != nil {
		log.Fatal(err)
	}
	if err := doc.GenMarkdownTree(root, out); err != nil {
		log.Fatal(err)
	}
	if err := writeIndex(root, out); err != nil {
		log.Fatal(err)
	}
}

// writeIndex renders a stable commands/index.md listing every top-level command, so the docs
// site has a single Command Reference entry point. Deterministic (alphabetical) so the CI
// drift check never flaps.
func writeIndex(root *cobra.Command, dir string) error {
	cmds := append([]*cobra.Command(nil), root.Commands()...)
	sort.Slice(cmds, func(i, j int) bool { return cmds[i].Name() < cmds[j].Name() })

	var b strings.Builder
	b.WriteString("# Command Reference\n\n")
	b.WriteString("Auto-generated from the CLI (`make docs-gen`). Start at the root command, ")
	fmt.Fprintf(&b, "[`%s`](%s.md), or jump to a group below.\n\n", root.Name(), root.Name())
	for _, c := range cmds {
		if c.Hidden || c.Name() == "help" {
			continue
		}
		file := strings.ReplaceAll(c.CommandPath(), " ", "_") + ".md"
		fmt.Fprintf(&b, "- [`%s`](%s) — %s\n", c.Name(), file, c.Short)
	}
	// #nosec G306 -- generated documentation, world-readable is intended.
	return os.WriteFile(filepath.Join(dir, "index.md"), []byte(b.String()), 0o644)
}
