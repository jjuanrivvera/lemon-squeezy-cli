package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/jjuanrivvera/lemon-squeezy-cli/internal/config"
)

// aliasFile stores user-defined aliases separately from config so they can be loaded with
// zero cobra dependencies during pre-parse expansion.
func aliasFile() string { return filepath.Join(config.DefaultDir(), "aliases.yaml") }

func loadAliases() (map[string]string, error) {
	raw, err := os.ReadFile(aliasFile()) // #nosec G304 -- fixed config path
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	m := map[string]string{}
	if err := yaml.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func saveAliases(m map[string]string) error {
	if err := os.MkdirAll(config.DefaultDir(), 0o700); err != nil {
		return err
	}
	out, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	// #nosec G306 -- non-secret aliases, 0600 by policy.
	return os.WriteFile(aliasFile(), out, 0o600)
}

// builtins are command names an alias must never shadow.
func builtins(root *cobra.Command) map[string]bool {
	b := map[string]bool{}
	for _, c := range root.Commands() {
		b[c.Name()] = true
		for _, a := range c.Aliases {
			b[a] = true
		}
	}
	return b
}

// ExpandAlias rewrites args[0] if it names a user alias and does not shadow a builtin. It is
// called from main.go BEFORE cobra parses, so an alias can expand into multiple tokens.
func ExpandAlias(root *cobra.Command, args []string) []string {
	if len(args) == 0 {
		return args
	}
	aliases, err := loadAliases()
	if err != nil {
		return args
	}
	expansion, ok := aliases[args[0]]
	if !ok || builtins(root)[args[0]] {
		return args
	}
	return append(strings.Fields(expansion), args[1:]...)
}

func init() {
	aliasCmd := &cobra.Command{Use: "alias", Short: "Manage user-defined command aliases"}

	setCmd := &cobra.Command{
		Use:     "set <name> <expansion>",
		Short:   "Define an alias (expansion may be multiple words)",
		Example: "  lsqueezy alias set ords 'orders list'",
		Args:    cobra.MinimumNArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			if builtins(rootCmd)[args[0]] {
				return fmt.Errorf("%q shadows a built-in command", args[0])
			}
			m, err := loadAliases()
			if err != nil {
				return err
			}
			m[args[0]] = strings.Join(args[1:], " ")
			if err := saveAliases(m); err != nil {
				return err
			}
			if !gf.quiet {
				fmt.Printf("✓ alias %q -> %q\n", args[0], m[args[0]])
			}
			return nil
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List defined aliases",
		RunE: func(_ *cobra.Command, _ []string) error {
			m, err := loadAliases()
			if err != nil {
				return err
			}
			names := make([]string, 0, len(m))
			for n := range m {
				names = append(names, n)
			}
			sort.Strings(names)
			for _, n := range names {
				fmt.Printf("%s = %s\n", n, m[n])
			}
			return nil
		},
	}

	removeCmd := &cobra.Command{
		Use:     "remove <name>",
		Aliases: []string{"rm"},
		Short:   "Remove an alias",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			m, err := loadAliases()
			if err != nil {
				return err
			}
			delete(m, args[0])
			if err := saveAliases(m); err != nil {
				return err
			}
			if !gf.quiet {
				fmt.Printf("✓ removed alias %q\n", args[0])
			}
			return nil
		},
	}

	aliasCmd.AddCommand(setCmd, listCmd, removeCmd)
	rootCmd.AddCommand(aliasCmd)
}
