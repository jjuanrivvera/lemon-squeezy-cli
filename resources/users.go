package resources

import (
	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/lemon-squeezy-cli/commands"
)

// usersCommand builds the `users` group. The API exposes only the authenticated user
// (GET /users/me), so this is a singleton read, not a generic collection.
func usersCommand() *cobra.Command {
	parent := &cobra.Command{
		Use:     "users",
		Aliases: []string{"user"},
		Short:   "The authenticated user (read-only)",
	}
	me := &cobra.Command{
		Use:     "me",
		Short:   "Show the authenticated user",
		Example: "  lsqueezy users me\n  lsqueezy users me -o json",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, render, err := commands.ClientRender(true)
			if err != nil {
				return err
			}
			user, err := c.Me(cmd.Context())
			if err != nil {
				return err
			}
			if c.DryRun {
				return nil
			}
			return render(user, []string{"id", "name", "email"})
		},
	}
	me.Annotations = map[string]string{"mcp.readOnly": "true"}
	parent.AddCommand(me)
	return parent
}
