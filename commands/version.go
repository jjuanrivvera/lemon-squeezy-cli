package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/lemon-squeezy-cli/internal/version"
)

func init() {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Print version, commit, and build date",
		Example: "  lsqueezy version\n  lsqueezy version --json",
		RunE: func(_ *cobra.Command, _ []string) error {
			info := version.Get()
			if asJSON {
				b, err := json.MarshalIndent(info, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(b))
				return nil
			}
			fmt.Println(info.String())
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "output as JSON")
	annotate(cmd, annReadOnly)
	rootCmd.AddCommand(cmd)
}
