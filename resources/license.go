package resources

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/lemon-squeezy-cli/commands"
)

// licenseCommand builds the `license` group wrapping Lemon Squeezy's License API
// (activate/validate/deactivate). These endpoints are not JSON:API and the store API key is
// optional — they are meant to run on a customer's machine with just the license key — so
// they use ClientRender with requireAuth=false.
func licenseCommand() *cobra.Command {
	parent := &cobra.Command{
		Use:   "license",
		Short: "Activate, validate, and deactivate license keys (License API)",
	}
	parent.AddCommand(licenseActivateCmd(), licenseValidateCmd(), licenseDeactivateCmd())
	return parent
}

func licenseActivateCmd() *cobra.Command {
	var key, instanceName string
	cmd := &cobra.Command{
		Use:     "activate",
		Short:   "Activate a license key for a new instance",
		Example: "  lsqueezy license activate --key 38b1460a-... --instance-name my-laptop",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if key == "" || instanceName == "" {
				return fmt.Errorf("--key and --instance-name are required")
			}
			c, render, err := commands.ClientRender(false)
			if err != nil {
				return err
			}
			res, err := c.License().Activate(cmd.Context(), key, instanceName)
			if err != nil {
				return err
			}
			if c.DryRun {
				return nil
			}
			return render(res, nil)
		},
	}
	cmd.Flags().StringVar(&key, "key", "", "license key (required)")
	cmd.Flags().StringVar(&instanceName, "instance-name", "", "a label for this activation (required)")
	cmd.Annotations = map[string]string{"mcp.write": "true"}
	return cmd
}

func licenseValidateCmd() *cobra.Command {
	var key, instanceID string
	cmd := &cobra.Command{
		Use:     "validate",
		Short:   "Validate a license key (optionally scoped to an instance)",
		Example: "  lsqueezy license validate --key 38b1460a-...\n  lsqueezy license validate --key 38b1460a-... --instance-id 1c0c...",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if key == "" {
				return fmt.Errorf("--key is required")
			}
			c, render, err := commands.ClientRender(false)
			if err != nil {
				return err
			}
			res, err := c.License().Validate(cmd.Context(), key, instanceID)
			if err != nil {
				return err
			}
			if c.DryRun {
				return nil
			}
			return render(res, nil)
		},
	}
	cmd.Flags().StringVar(&key, "key", "", "license key (required)")
	cmd.Flags().StringVar(&instanceID, "instance-id", "", "instance id to scope the check (optional)")
	cmd.Annotations = map[string]string{"mcp.readOnly": "true"}
	return cmd
}

func licenseDeactivateCmd() *cobra.Command {
	var key, instanceID string
	cmd := &cobra.Command{
		Use:     "deactivate",
		Short:   "Deactivate a license key instance (IRREVERSIBLE)",
		Example: "  lsqueezy license deactivate --key 38b1460a-... --instance-id 1c0c...",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if key == "" || instanceID == "" {
				return fmt.Errorf("--key and --instance-id are required")
			}
			c, render, err := commands.ClientRender(false)
			if err != nil {
				return err
			}
			res, err := c.License().Deactivate(cmd.Context(), key, instanceID)
			if err != nil {
				return err
			}
			if c.DryRun {
				return nil
			}
			return render(res, nil)
		},
	}
	cmd.Flags().StringVar(&key, "key", "", "license key (required)")
	cmd.Flags().StringVar(&instanceID, "instance-id", "", "instance id to deactivate (required)")
	cmd.Annotations = map[string]string{"mcp.destructive": "true"}
	return cmd
}
