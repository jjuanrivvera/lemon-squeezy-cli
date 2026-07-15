package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/lemon-squeezy-cli/internal/update"
	"github.com/jjuanrivvera/lemon-squeezy-cli/internal/version"
)

func init() {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update lsqueezy to the latest GitHub release",
		Long: `Download the latest lsqueezy release, verify it against checksums.txt, and replace
the running binary in place. Use 'lsqueezy update check' to see what's available without
installing.`,
		Example: "  lsqueezy update\n  lsqueezy update check",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
			defer cancel()

			res := update.NewUpdater(version.Version).CheckAndUpdate(ctx)
			if res.Error != nil {
				return res.Error
			}
			if res.Updated {
				fmt.Printf("Updated %s → %s. Restart to use the new version.\n", res.FromVersion, res.ToVersion)
			} else {
				fmt.Println("Already on the latest version.")
			}
			return nil
		},
	}
	annotate(cmd, annWrite)

	check := &cobra.Command{
		Use:   "check",
		Short: "Check for a newer release without installing it",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
			defer cancel()

			rel, err := update.NewUpdater(version.Version).GetLatestRelease(ctx)
			if err != nil {
				return err
			}
			fmt.Printf("Current: %s\n", version.Version)
			fmt.Printf("Latest:  %s\n", rel.TagName)
			switch {
			case version.Version == "dev" || version.Version == "":
				fmt.Println("This is a development build; self-update is disabled.")
			case strings.TrimPrefix(rel.TagName, "v") == strings.TrimPrefix(version.Version, "v"):
				fmt.Println("You are on the latest version.")
			default:
				fmt.Println("A newer version is available. Run `lsqueezy update` to install it.")
			}
			return nil
		},
	}
	annotate(check, annReadOnly)
	cmd.AddCommand(check)

	rootCmd.AddCommand(cmd)
}
