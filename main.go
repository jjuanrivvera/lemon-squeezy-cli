// Command lsqueezy is a production-grade CLI for Lemon Squeezy.
//
// main wires a cancellable context (signal.NotifyContext) so Ctrl-C cancels in-flight
// pagination and retry backoff, expands user aliases before cobra parses, and blank-imports
// the resources package so its init() self-registers every resource against the generic core.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jjuanrivvera/lemon-squeezy-cli/commands"

	// The resources package self-registers every resource via init(). Importing it here (and
	// only here) keeps the dependency direction one-way: resources -> commands -> internal/api.
	_ "github.com/jjuanrivvera/lemon-squeezy-cli/resources"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	root := commands.Setup()

	// Expand a user alias in args[0] before cobra parses (aliases may be multi-token).
	root.SetArgs(commands.ExpandAlias(root, os.Args[1:]))

	if err := root.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
