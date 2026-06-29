// Command gendocs generates docs/commands/*.md from the cobra tree so the reference docs
// stay in lockstep with the CLI (the CI drift check fails if they diverge).
package main

import (
	"log"
	"os"
	"path/filepath"

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
}
