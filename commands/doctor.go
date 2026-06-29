package commands

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/lemon-squeezy-cli/internal/config"
	"github.com/jjuanrivvera/lemon-squeezy-cli/internal/version"
)

type checkResult struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail"`
}

func init() {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "doctor",
		Short:   "Diagnose config, credentials, and connectivity",
		Example: "  lsqueezy doctor\n  lsqueezy doctor --json",
		RunE: func(cmd *cobra.Command, _ []string) error {
			results := runDoctor(cmd)
			allOK := true
			for _, r := range results {
				if !r.OK {
					allOK = false
				}
			}
			if asJSON {
				b, _ := json.MarshalIndent(results, "", "  ")
				fmt.Println(string(b))
			} else {
				for _, r := range results {
					mark := "✓"
					if !r.OK {
						mark = "✗"
					}
					fmt.Printf("%s %s: %s\n", mark, r.Name, r.Detail)
				}
			}
			if !allOK {
				return fmt.Errorf("doctor found problems")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "output as JSON")
	annotate(cmd, annReadOnly)
	rootCmd.AddCommand(cmd)
}

func runDoctor(cmd *cobra.Command) []checkResult {
	var out []checkResult

	cfg, err := loadConfig()
	if err != nil {
		out = append(out, checkResult{"config", false, err.Error()})
		return out
	}
	out = append(out, checkResult{"config", true, "loaded " + config.DefaultPath()})

	profile := activeProfileName(cfg)
	out = append(out, checkResult{"profile", true, profile})

	key, kerr := resolveAPIKey(profile)
	if kerr != nil {
		out = append(out, checkResult{"credentials", false, "no key (run `lsqueezy auth login`)"})
	} else {
		out = append(out, checkResult{"credentials", true, "key resolvable"})
		if user, err := verifyKey(cmd.Context(), cfg, profile, key); err != nil {
			out = append(out, checkResult{"connectivity", false, err.Error()})
		} else {
			out = append(out, checkResult{"connectivity", true, "auth valid: " + user.Email})
		}
	}

	// Clock sanity: if the system year is implausibly old, TLS/cert checks may fail.
	if time.Now().Year() < 2020 {
		out = append(out, checkResult{"clock", false, "system clock looks wrong"})
	} else {
		out = append(out, checkResult{"clock", true, time.Now().Format(time.RFC3339)})
	}

	out = append(out, checkResult{"version", true, version.Get().String()})
	return out
}
