package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/lemon-squeezy-cli/internal/config"
)

func init() {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect and edit configuration",
	}
	configCmd.AddCommand(
		configPathCmd(), configViewCmd(), configSetCmd(),
		configUseCmd(), configListProfilesCmd(),
	)
	rootCmd.AddCommand(configCmd)
}

func configPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the config file path",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Println(config.DefaultPath())
			return nil
		},
	}
}

func configViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "view",
		Short:   "Show the current config (secrets redacted)",
		Example: "  lsqueezy config view",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			fmt.Printf("active_profile: %s\n", cfg.ActiveProfile)
			fmt.Printf("output: %s\n", cfg.ResolveOutput())
			fmt.Println("profiles:")
			for name, p := range cfg.Profiles {
				base := p.BaseURL
				if base == "" {
					base = "(default)"
				}
				// Secrets live in the keyring, never in config; nothing to redact here, but
				// we still never echo a key even if one were present.
				fmt.Printf("  %s: base_url=%s api_key=<keyring>\n", name, base)
			}
			return nil
		},
	}
}

func configSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set <key> <value>",
		Short:   "Set a config value (base_url|output) for the active profile",
		Example: "  lsqueezy config set base_url https://api.lemonsqueezy.com/v1\n  lsqueezy config set output json",
		Args:    cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			profile := activeProfileName(cfg)
			p := cfg.Profiles[profile]
			switch args[0] {
			case "base_url":
				p.BaseURL = args[1]
				cfg.SetProfile(profile, p)
			case "output":
				cfg.Output = args[1]
			default:
				return fmt.Errorf("unknown key %q (want base_url|output)", args[0])
			}
			if err := cfg.Save(); err != nil {
				return err
			}
			if !gf.quiet {
				fmt.Printf("✓ set %s for profile %q\n", args[0], profile)
			}
			return nil
		},
	}
	return cmd
}

func configUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "use <profile>",
		Short:   "Switch the active profile",
		Example: "  lsqueezy config use work",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			// Auto-create the profile if it's new, so `use` doubles as "create + switch".
			if _, ok := cfg.Profiles[args[0]]; !ok {
				cfg.SetProfile(args[0], config.Profile{})
			}
			if err := cfg.Use(args[0]); err != nil {
				return err
			}
			if err := cfg.Save(); err != nil {
				return err
			}
			if !gf.quiet {
				fmt.Printf("✓ active profile is now %q\n", args[0])
			}
			return nil
		},
	}
}

func configListProfilesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list-profiles",
		Short:   "List configured profiles",
		Example: "  lsqueezy config list-profiles",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			for name := range cfg.Profiles {
				marker := "  "
				if name == cfg.ActiveProfile {
					marker = "* "
				}
				fmt.Printf("%s%s\n", marker, name)
			}
			return nil
		},
	}
	annotate(cmd, annReadOnly)
	return cmd
}
