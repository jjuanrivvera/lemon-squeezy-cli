package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/lemon-squeezy-cli/internal/api"
	"github.com/jjuanrivvera/lemon-squeezy-cli/internal/config"
)

func init() {
	var apiKey, baseURL string
	cmd := &cobra.Command{
		Use:     "init",
		Aliases: []string{"setup"},
		Short:   "First-run wizard: capture base URL + key, write config, smoke-test",
		Example: "  lsqueezy init\n  lsqueezy init --api-key live_xxx",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			profile := activeProfileName(cfg)

			if baseURL == "" {
				baseURL = api.DefaultBaseURL
			}
			if apiKey == "" {
				apiKey, err = promptSecret("Enter your Lemon Squeezy key (hidden): ")
				if err != nil {
					return fmt.Errorf("read key: %w", err)
				}
			}
			if apiKey == "" {
				return fmt.Errorf("an API key is required")
			}

			// Persist non-secret bits, store the key in the keyring, then smoke-test.
			p := cfg.Profiles[profile]
			p.BaseURL = baseURL
			cfg.SetProfile(profile, p)
			if err := cfg.Save(); err != nil {
				return err
			}
			user, err := verifyKey(cmd.Context(), cfg, profile, apiKey)
			if err != nil {
				return fmt.Errorf("smoke test failed: %w", err)
			}
			if err := tokenStore().Set(profile, apiKey); err != nil {
				return err
			}
			if !gf.quiet {
				fmt.Printf("✓ configured profile %q as %s <%s> — try `lsqueezy stores list`\n",
					profile, user.Name, user.Email)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&apiKey, "api-key", "", "API key (omit to be prompted)")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "base URL (default "+config.EnvPrefix+" default)")
	rootCmd.AddCommand(cmd)
}
