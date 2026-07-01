// Package commands wires the cobra command tree. root.go owns the global flags, the shared
// API client factory, and the single render() path used by every resource command.
package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/lemon-squeezy-cli/internal/api"
	"github.com/jjuanrivvera/lemon-squeezy-cli/internal/auth"
	"github.com/jjuanrivvera/lemon-squeezy-cli/internal/config"
	"github.com/jjuanrivvera/lemon-squeezy-cli/internal/output"
)

// profileFlag/profileNoun name the multi-profile selector (cliwright manifest fields). A
// Lemon Squeezy "profile" IS an account (one API key per account; a machine may hold a live
// and a test-mode key), so the user-facing flag reads --account; --profile stays as a hidden
// back-compat alias. Keep these in sync with api-manifest.json profile_flag/profile_noun.
const (
	profileFlag = "account"
	profileNoun = "account"
)

// globalFlags holds the persistent flag values, resolved once in PersistentPreRunE.
type globalFlags struct {
	outputFormat string
	profile      string // the selected account/profile (bound to both --account and --profile)
	baseURL      string
	dryRun       bool
	showToken    bool
	verbose      bool
	noColor      bool
	columns      []string
	quiet        bool
	jq           string

	// list flags (registered globally so the generic builder can read them)
	all    bool
	limit  int
	page   int
	sort   string
	filter []string
}

var gf globalFlags

// rootCmd is the application root. It is package-global so resource files can self-register
// via init().
var rootCmd = &cobra.Command{
	Use:   "lsqueezy",
	Short: "A polished CLI for the Lemon Squeezy e-commerce API",
	Long: `lsqueezy is a production-grade command-line interface for Lemon Squeezy.

Manage stores, products, orders, subscriptions, customers, discounts, license keys,
checkouts, and webhooks. Script it all with table/json/yaml/csv output, a --jq filter,
named accounts for multiple keys, and a --dry-run that prints the equivalent curl.

Examples:
  lsqueezy auth login --api-key eyJ0eX...
  lsqueezy stores list
  lsqueezy products list --filter store_id=1 --all
  lsqueezy orders get 12345 -o json
  lsqueezy orders list -o json --jq '.[].total_formatted'
  lsqueezy --account test subscriptions list
  lsqueezy subscriptions cancel 9999 --dry-run
  lsqueezy license validate --key 38b1460a-5104-4067-a91d-77b872934d51`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		// Validate the output format early so a typo fails fast and uniformly.
		if gf.outputFormat != "" {
			if _, err := output.ParseFormat(gf.outputFormat); err != nil {
				return err
			}
		}
		return nil
	},
}

// Root returns the configured root command. main.go calls Root().ExecuteContext(ctx) with a
// context from signal.NotifyContext so Ctrl-C cancels in-flight work.
func Root() *cobra.Command { return rootCmd }

// Setup applies all deferred resource registrations onto the root command. main.go calls it
// after the resource packages' init() functions have run (via blank imports), so the tree is
// complete before Execute. Idempotent.
func Setup() *cobra.Command {
	for _, reg := range registrars {
		rootCmd.AddCommand(reg())
	}
	registrars = nil
	return rootCmd
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.StringVarP(&gf.outputFormat, "output", "o", "", "output format: table|json|yaml|csv")
	// The multi-profile selector is named per the API (profileFlag). Both --account and the
	// legacy --profile target the same var, so existing scripts using --profile keep working;
	// --profile is hidden so help shows only the natural name.
	pf.StringVar(&gf.profile, profileFlag, "", "named "+profileNoun+" to use (env LEMONSQUEEZY_"+strings.ToUpper(profileNoun)+")")
	if profileFlag != "profile" {
		pf.StringVar(&gf.profile, "profile", "", "alias for --"+profileFlag)
		_ = pf.MarkHidden("profile")
	}
	pf.StringVar(&gf.baseURL, "base-url", "", "override the API base URL")
	pf.BoolVar(&gf.dryRun, "dry-run", false, "print the equivalent curl and make no request")
	pf.BoolVar(&gf.showToken, "show-token", false, "reveal the API key in dry-run output")
	pf.BoolVarP(&gf.verbose, "verbose", "v", false, "verbose request logging")
	pf.BoolVar(&gf.noColor, "no-color", false, "disable colored output")
	pf.StringSliceVar(&gf.columns, "columns", nil, "comma-separated columns to show")
	pf.BoolVar(&gf.quiet, "quiet", false, "suppress non-essential chatter")
	pf.StringVar(&gf.jq, "jq", "", "gojq expression applied to the result before rendering")

	// List flags (read by the generic builder's list command).
	pf.BoolVar(&gf.all, "all", false, "fetch all pages (list commands)")
	pf.IntVar(&gf.limit, "limit", 0, "page size, 1-100 (list commands)")
	pf.IntVar(&gf.page, "page", 0, "page number, 1-based (list commands)")
	pf.StringVar(&gf.sort, "sort", "", "JSON:API sort field, prefix with - for desc (list commands)")
	pf.StringSliceVar(&gf.filter, "filter", nil, "client-side field=value filters (list commands)")
}

// loadConfig loads the config file (honoring LEMONSQUEEZY_ACCOUNT/env) once per invocation.
func loadConfig() (*config.Config, error) {
	return config.Load("")
}

// activeProfileName resolves the profile from flag > env > config active profile.
func activeProfileName(cfg *config.Config) string {
	if gf.profile != "" {
		return gf.profile
	}
	return cfg.ResolveProfileName()
}

// tokenStore builds the auth store backed by the config directory (for the file fallback).
func tokenStore() *auth.Store {
	return auth.NewStore(config.DefaultDir())
}

// resolveAPIKey applies precedence: LEMONSQUEEZY_API_KEY env > keyring/file for the profile.
func resolveAPIKey(profile string) (string, error) {
	if v := os.Getenv(config.APIKeyEnv); v != "" {
		return v, nil
	}
	return tokenStore().Get(profile)
}

// getAPIClient builds an authenticated client honoring all precedence and global flags.
// requireAuth=false lets commands like `doctor`/`api --dry-run` proceed without a key.
func getAPIClient(requireAuth bool) (*api.Client, *config.Config, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, nil, err
	}
	profile := activeProfileName(cfg)
	prof := cfg.Resolve(profile)

	baseURL := prof.BaseURL
	if gf.baseURL != "" {
		baseURL = gf.baseURL
	}

	key, keyErr := resolveAPIKey(profile)
	// --dry-run never makes a request (and redacts the token anyway), so a missing key must
	// not block it — dry-run is a teaching/debugging tool that has to work before login.
	if requireAuth && !gf.dryRun && keyErr != nil {
		return nil, cfg, fmt.Errorf("no API key for profile %q: run `lsqueezy auth login` or set %s", profile, config.APIKeyEnv)
	}

	opts := []api.Option{
		api.WithDryRun(gf.dryRun, os.Stdout),
	}
	c := api.New(baseURL, key, opts...)
	c.ShowToken = gf.showToken
	c.Verbose = gf.verbose
	c.VerboseOut = os.Stderr
	return c, cfg, nil
}

// outputOptions builds renderer options from the resolved global flags + config default.
func outputOptions(cfg *config.Config) output.Options {
	format := gf.outputFormat
	if format == "" {
		format = cfg.ResolveOutput()
	}
	f, err := output.ParseFormat(format)
	if err != nil {
		f = output.FormatTable
	}
	return output.Options{
		Format:  f,
		Columns: normalizeColumns(gf.columns),
		NoColor: gf.noColor,
		JQ:      gf.jq,
		Writer:  os.Stdout,
	}
}

func normalizeColumns(cols []string) []string {
	var out []string
	for _, c := range cols {
		c = strings.TrimSpace(c)
		if c != "" {
			out = append(out, c)
		}
	}
	return out
}

// render is the single output path for every command.
func render(cfg *config.Config, v any, defaultColumns []string) error {
	return output.Render(v, defaultColumns, outputOptions(cfg))
}
