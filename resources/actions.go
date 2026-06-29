// Package resources registers every Lemon Squeezy resource (and the custom action verbs and
// singleton command groups) against the generic core. main.go blank-imports this package so
// its init() self-registers everything — no edits to shared code when a resource is added.
package resources

import (
	"net/http"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/lemon-squeezy-cli/commands"
	"github.com/jjuanrivvera/lemon-squeezy-cli/internal/api"
)

// clientFn is the bound client+render factory handed to Extra command builders.
type clientFn = func(bool) (*api.Client, func(any, []string) error, error)

// orderRefundExtra: POST /orders/{id}/refund. Irreversible (real money), so destructive.
func orderRefundExtra() commands.ExtraCommand {
	return commands.ExtraCommand{
		Destructive: true,
		Build: func(getClient clientFn) *cobra.Command {
			var amount int
			cmd := &cobra.Command{
				Use:     "refund <id>",
				Short:   "Issue a refund for an order (IRREVERSIBLE)",
				Example: "  lsqueezy orders refund 12345 --amount 500\n  lsqueezy orders refund 12345 --dry-run",
				Args:    cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					c, render, err := getClient(true)
					if err != nil {
						return err
					}
					id := args[0]
					attrs := map[string]any{}
					if amount > 0 {
						attrs["amount"] = amount
					}
					body := map[string]any{"data": map[string]any{"type": "orders", "id": id, "attributes": attrs}}
					order, err := c.Orders().ActionOne(cmd.Context(), http.MethodPost, id+"/refund", nil, body)
					if err != nil {
						return err
					}
					if c.DryRun {
						return nil
					}
					return render(order, []string{"id", "order_number", "status", "refunded", "total_formatted"})
				},
			}
			cmd.Flags().IntVar(&amount, "amount", 0, "amount in cents to refund (0 = full refund)")
			return cmd
		},
	}
}

// orderInvoiceExtra: POST /orders/{id}/generate-invoice. Generates a downloadable invoice;
// reversible-ish (creates a document), classified as a write.
func orderInvoiceExtra() commands.ExtraCommand {
	return commands.ExtraCommand{
		Write: true,
		Build: func(getClient clientFn) *cobra.Command {
			var name, address, city, state, zip, country, notes, locale string
			cmd := &cobra.Command{
				Use:     "generate-invoice <id>",
				Short:   "Generate a downloadable invoice for an order",
				Example: "  lsqueezy orders generate-invoice 12345 --name 'Acme' --country US",
				Args:    cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					c, render, err := getClient(true)
					if err != nil {
						return err
					}
					q := invoiceQuery(name, address, city, state, zip, country, notes, locale)
					var out map[string]any
					if err := c.Orders().Action(cmd.Context(), http.MethodPost, args[0]+"/generate-invoice", q, nil, &out); err != nil {
						return err
					}
					if c.DryRun {
						return nil
					}
					return render(out, nil)
				},
			}
			invoiceFlags(cmd, &name, &address, &city, &state, &zip, &country, &notes, &locale)
			return cmd
		},
	}
}

// subscriptionCancelExtra: DELETE /subscriptions/{id}. Irreversible, so destructive.
func subscriptionCancelExtra() commands.ExtraCommand {
	return commands.ExtraCommand{
		Destructive: true,
		Build: func(getClient clientFn) *cobra.Command {
			cmd := &cobra.Command{
				Use:     "cancel <id>",
				Short:   "Cancel a subscription (sets it to cancelled; IRREVERSIBLE)",
				Example: "  lsqueezy subscriptions cancel 9999\n  lsqueezy subscriptions cancel 9999 --dry-run",
				Args:    cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					c, render, err := getClient(true)
					if err != nil {
						return err
					}
					sub, err := c.Subscriptions().ActionOne(cmd.Context(), http.MethodDelete, args[0], nil, nil)
					if err != nil {
						return err
					}
					if c.DryRun {
						return nil
					}
					return render(sub, []string{"id", "product_name", "user_email", "status", "ends_at"})
				},
			}
			return cmd
		},
	}
}

// subItemUsageExtra: GET /subscription-items/{id}/current-usage. Read-only.
func subItemUsageExtra() commands.ExtraCommand {
	return commands.ExtraCommand{
		ReadOnly: true,
		Build: func(getClient clientFn) *cobra.Command {
			return &cobra.Command{
				Use:     "current-usage <id>",
				Short:   "Show current usage for a usage-based subscription item",
				Example: "  lsqueezy subscription-items current-usage 1 -o json",
				Args:    cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					c, render, err := getClient(true)
					if err != nil {
						return err
					}
					var out map[string]any
					if err := c.SubscriptionItems().Action(cmd.Context(), http.MethodGet, args[0]+"/current-usage", nil, nil, &out); err != nil {
						return err
					}
					if c.DryRun {
						return nil
					}
					return render(out, nil)
				},
			}
		},
	}
}

// subInvoiceRefundExtra: POST /subscription-invoices/{id}/refund. Irreversible, destructive.
func subInvoiceRefundExtra() commands.ExtraCommand {
	return commands.ExtraCommand{
		Destructive: true,
		Build: func(getClient clientFn) *cobra.Command {
			var amount int
			cmd := &cobra.Command{
				Use:     "refund <id>",
				Short:   "Issue a refund for a subscription invoice (IRREVERSIBLE)",
				Example: "  lsqueezy subscription-invoices refund 55 --amount 500",
				Args:    cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					c, render, err := getClient(true)
					if err != nil {
						return err
					}
					id := args[0]
					attrs := map[string]any{}
					if amount > 0 {
						attrs["amount"] = amount
					}
					body := map[string]any{"data": map[string]any{"type": "subscription-invoices", "id": id, "attributes": attrs}}
					inv, err := c.SubscriptionInvoices().ActionOne(cmd.Context(), http.MethodPost, id+"/refund", nil, body)
					if err != nil {
						return err
					}
					if c.DryRun {
						return nil
					}
					return render(inv, []string{"id", "subscription_id", "status", "refunded", "total_formatted"})
				},
			}
			cmd.Flags().IntVar(&amount, "amount", 0, "amount in cents to refund (0 = full refund)")
			return cmd
		},
	}
}

// subInvoiceGenerateExtra: POST /subscription-invoices/{id}/generate-invoice.
func subInvoiceGenerateExtra() commands.ExtraCommand {
	return commands.ExtraCommand{
		Write: true,
		Build: func(getClient clientFn) *cobra.Command {
			var name, address, city, state, zip, country, notes, locale string
			cmd := &cobra.Command{
				Use:     "generate-invoice <id>",
				Short:   "Generate a downloadable invoice for a subscription invoice",
				Example: "  lsqueezy subscription-invoices generate-invoice 55 --name 'Acme' --country US",
				Args:    cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					c, render, err := getClient(true)
					if err != nil {
						return err
					}
					q := invoiceQuery(name, address, city, state, zip, country, notes, locale)
					var out map[string]any
					if err := c.SubscriptionInvoices().Action(cmd.Context(), http.MethodPost, args[0]+"/generate-invoice", q, nil, &out); err != nil {
						return err
					}
					if c.DryRun {
						return nil
					}
					return render(out, nil)
				},
			}
			invoiceFlags(cmd, &name, &address, &city, &state, &zip, &country, &notes, &locale)
			return cmd
		},
	}
}

// customerArchiveExtra: PATCH /customers/{id} with status=archived (customers can't be
// deleted, only archived). Reversible (un-archive by updating status), so a write.
func customerArchiveExtra() commands.ExtraCommand {
	return commands.ExtraCommand{
		Write: true,
		Build: func(getClient clientFn) *cobra.Command {
			return &cobra.Command{
				Use:     "archive <id>",
				Short:   "Archive a customer (sets status=archived)",
				Example: "  lsqueezy customers archive 123\n  lsqueezy customers archive 123 --dry-run",
				Args:    cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					c, render, err := getClient(true)
					if err != nil {
						return err
					}
					var out api.Customer
					body := api.WriteBody{Attributes: map[string]any{"status": "archived"}}
					if err := c.Customers().Update(cmd.Context(), args[0], body, &out); err != nil {
						return err
					}
					if c.DryRun {
						return nil
					}
					return render(&out, []string{"id", "name", "email", "status"})
				},
			}
		},
	}
}

// invoiceFlags registers the shared generate-invoice billing flags.
func invoiceFlags(cmd *cobra.Command, name, address, city, state, zip, country, notes, locale *string) {
	f := cmd.Flags()
	f.StringVar(name, "name", "", "billing name")
	f.StringVar(address, "address", "", "billing address")
	f.StringVar(city, "city", "", "billing city")
	f.StringVar(state, "state", "", "billing state")
	f.StringVar(zip, "zip-code", "", "billing zip/postal code")
	f.StringVar(country, "country", "", "billing country (ISO code)")
	f.StringVar(notes, "notes", "", "invoice notes")
	f.StringVar(locale, "locale", "", "invoice locale")
}

// invoiceQuery builds the generate-invoice query params (snake_case per the API), omitting
// empty values so the request stays minimal.
func invoiceQuery(name, address, city, state, zip, country, notes, locale string) url.Values {
	q := url.Values{}
	set := func(k, v string) {
		if v != "" {
			q.Set(k, v)
		}
	}
	set("name", name)
	set("address", address)
	set("city", city)
	set("state", state)
	set("zip_code", zip)
	set("country", country)
	set("notes", notes)
	set("locale", locale)
	return q
}
