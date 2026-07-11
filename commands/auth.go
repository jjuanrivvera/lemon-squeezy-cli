package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/lemon-squeezy-cli/internal/api"
	"github.com/jjuanrivvera/lemon-squeezy-cli/internal/config"
)

func init() {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage API credentials",
	}
	authCmd.AddCommand(authLoginCmd(), authLogoutCmd(), authStatusCmd())
	rootCmd.AddCommand(authCmd)
}

func authLoginCmd() *cobra.Command {
	var apiKey string
	cmd := &cobra.Command{
		Use:     "login",
		Short:   "Store an API key in the OS keyring and verify it",
		Example: "  lsqueezy auth login --api-key live_xxx",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			profile := activeProfileName(cfg)
			if apiKey == "" {
				apiKey, err = promptSecret("Enter your Lemon Squeezy key (hidden): ")
				if err != nil {
					return fmt.Errorf("read key: %w", err)
				}
			}
			if apiKey == "" {
				return fmt.Errorf("an API key is required")
			}
			// Verify against /users/me before persisting — never store a bad key.
			user, err := verifyKey(cmd.Context(), cfg, profile, apiKey)
			if err != nil {
				return fmt.Errorf("key verification failed: %w", err)
			}
			if err := tokenStore().Set(profile, apiKey); err != nil {
				return fmt.Errorf("store key: %w", err)
			}
			if !gf.quiet {
				fmt.Printf("✓ key stored for profile %q and verified as %s <%s>\n",
					profile, user.Name, user.Email)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&apiKey, "api-key", "", "API key (omit to be prompted)")
	return cmd
}

// authUser is the minimal shape of GET /v1/users/me used to verify a key and show identity.
type authUser struct {
	ID    api.ID `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// verifyKey confirms a key works by fetching the authenticated user (GET /users/me) — the
// canonical "whoami" endpoint, cheaper and clearer than listing a resource. Returns the
// user so callers can display identity.
func verifyKey(ctx context.Context, cfg *config.Config, profile, key string) (*authUser, error) {
	prof := cfg.Resolve(profile)
	baseURL := prof.BaseURL
	if gf.baseURL != "" {
		baseURL = gf.baseURL
	}
	c := api.New(baseURL, key)
	return api.GetOne[authUser](ctx, c, "users/me", nil)
}

func authLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "logout",
		Short:   "Remove the stored API key for the active profile",
		Example: "  lsqueezy auth logout",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			profile := activeProfileName(cfg)
			if err := tokenStore().Delete(profile); err != nil {
				return fmt.Errorf("remove key: %w", err)
			}
			if !gf.quiet {
				fmt.Printf("✓ key removed for profile %q\n", profile)
			}
			return nil
		},
	}
}

func authStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status",
		Aliases: []string{"whoami"},
		Short:   "Show the active profile, base URL, and whether auth works",
		Example: "  lsqueezy auth status\n  lsqueezy whoami",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}
			profile := activeProfileName(cfg)
			prof := cfg.Resolve(profile)
			baseURL := prof.BaseURL
			if baseURL == "" {
				baseURL = api.DefaultBaseURL
			}
			fmt.Printf("Profile:  %s\n", profile)
			fmt.Printf("Base URL: %s\n", baseURL)
			key, kerr := resolveAPIKey(profile)
			if kerr != nil {
				fmt.Println("Auth:     no key stored (run `lsqueezy auth login`)")
				return nil
			}
			user, err := verifyKey(cmd.Context(), cfg, profile, key)
			if err != nil {
				fmt.Printf("Auth:     key present but verification FAILED: %v\n", err)
				return err
			}
			fmt.Printf("Auth:     ✓ valid\n")
			fmt.Printf("Identity: %s <%s> (user id %s)\n", user.Name, user.Email, user.ID)
			return nil
		},
	}
	annotate(cmd, annReadOnly)
	return cmd
}
