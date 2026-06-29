package resources

import (
	"github.com/jjuanrivvera/lemon-squeezy-cli/commands"
	"github.com/jjuanrivvera/lemon-squeezy-cli/internal/api"
)

// init self-registers every resource. Adding a resource is exactly this: a type + a Client
// accessor (in internal/api) + one RegisterResource call here. No shared code changes.
func init() {
	// --- Read-only catalog & platform ---
	commands.RegisterResource(commands.ResourceSpec[api.Store]{
		Use: "stores", Aliases: []string{"store"}, Short: "Browse stores (read-only)",
		New:     func(c *api.Client) *api.Resource[api.Store] { return c.Stores() },
		Columns: []string{"id", "name", "slug", "currency", "total_sales"},
	})

	commands.RegisterResource(commands.ResourceSpec[api.Product]{
		Use: "products", Aliases: []string{"product"}, Short: "Browse products (read-only)",
		New:         func(c *api.Client) *api.Resource[api.Product] { return c.Products() },
		Columns:     []string{"id", "store_id", "name", "status", "price_formatted"},
		ListFilters: []commands.ListFilter{{Flag: "store-id", Query: "store_id", Usage: "store id"}},
		Includes:    []string{"store", "variants"},
	})

	commands.RegisterResource(commands.ResourceSpec[api.Variant]{
		Use: "variants", Aliases: []string{"variant"}, Short: "Browse product variants (read-only)",
		New:         func(c *api.Client) *api.Resource[api.Variant] { return c.Variants() },
		Columns:     []string{"id", "product_id", "name", "price", "status"},
		ListFilters: []commands.ListFilter{{Flag: "product-id", Query: "product_id", Usage: "product id"}},
		Includes:    []string{"product"},
	})

	commands.RegisterResource(commands.ResourceSpec[api.Price]{
		Use: "prices", Aliases: []string{"price"}, Short: "Browse variant prices (read-only)",
		New:         func(c *api.Client) *api.Resource[api.Price] { return c.Prices() },
		Columns:     []string{"id", "variant_id", "category", "scheme", "unit_price"},
		ListFilters: []commands.ListFilter{{Flag: "variant-id", Query: "variant_id", Usage: "variant id"}},
		Includes:    []string{"variant"},
	})

	commands.RegisterResource(commands.ResourceSpec[api.File]{
		Use: "files", Aliases: []string{"file"}, Short: "Browse downloadable files (read-only)",
		New:         func(c *api.Client) *api.Resource[api.File] { return c.Files() },
		Columns:     []string{"id", "variant_id", "name", "size_formatted", "status"},
		ListFilters: []commands.ListFilter{{Flag: "variant-id", Query: "variant_id", Usage: "variant id"}},
		Includes:    []string{"variant"},
	})

	// --- Customers (create/update + archive) ---
	commands.RegisterResource(commands.ResourceSpec[api.Customer]{
		Use: "customers", Aliases: []string{"customer"}, Short: "Manage customers",
		New:       func(c *api.Client) *api.Resource[api.Customer] { return c.Customers() },
		Columns:   []string{"id", "name", "email", "status", "country"},
		CanCreate: true, CanUpdate: true,
		ListFilters: []commands.ListFilter{
			{Flag: "store-id", Query: "store_id", Usage: "store id"},
			{Flag: "email", Query: "email", Usage: "email address"},
		},
		Includes: []string{"store", "orders", "subscriptions", "license-keys"},
		Extra:    []commands.ExtraCommand{customerArchiveExtra()},
	})

	// --- Orders (read + refund + generate-invoice) ---
	commands.RegisterResource(commands.ResourceSpec[api.Order]{
		Use: "orders", Aliases: []string{"order"}, Short: "Browse orders; refund and invoice",
		New:     func(c *api.Client) *api.Resource[api.Order] { return c.Orders() },
		Columns: []string{"id", "order_number", "user_email", "total_formatted", "status"},
		ListFilters: []commands.ListFilter{
			{Flag: "store-id", Query: "store_id", Usage: "store id"},
			{Flag: "user-email", Query: "user_email", Usage: "buyer email"},
		},
		Includes: []string{"store", "customer", "order-items"},
		Extra:    []commands.ExtraCommand{orderRefundExtra(), orderInvoiceExtra()},
	})

	commands.RegisterResource(commands.ResourceSpec[api.OrderItem]{
		Use: "order-items", Aliases: []string{"order-item"}, Short: "Browse order line items (read-only)",
		New:     func(c *api.Client) *api.Resource[api.OrderItem] { return c.OrderItems() },
		Columns: []string{"id", "order_id", "product_name", "variant_name", "quantity"},
		ListFilters: []commands.ListFilter{
			{Flag: "order-id", Query: "order_id", Usage: "order id"},
			{Flag: "product-id", Query: "product_id", Usage: "product id"},
		},
		Includes: []string{"order", "product", "variant"},
	})

	// --- Subscriptions (update + cancel) ---
	commands.RegisterResource(commands.ResourceSpec[api.Subscription]{
		Use: "subscriptions", Aliases: []string{"subscription", "subs"}, Short: "Manage subscriptions",
		New:       func(c *api.Client) *api.Resource[api.Subscription] { return c.Subscriptions() },
		Columns:   []string{"id", "product_name", "user_email", "status", "renews_at"},
		CanUpdate: true,
		ListFilters: []commands.ListFilter{
			{Flag: "store-id", Query: "store_id", Usage: "store id"},
			{Flag: "status", Query: "status", Usage: "status (active, cancelled, …)"},
			{Flag: "user-email", Query: "user_email", Usage: "customer email"},
		},
		Includes: []string{"store", "customer", "order", "product", "variant"},
		Extra:    []commands.ExtraCommand{subscriptionCancelExtra()},
	})

	commands.RegisterResource(commands.ResourceSpec[api.SubscriptionItem]{
		Use: "subscription-items", Aliases: []string{"subscription-item", "sub-items"}, Short: "Manage subscription items",
		New:       func(c *api.Client) *api.Resource[api.SubscriptionItem] { return c.SubscriptionItems() },
		Columns:   []string{"id", "subscription_id", "price_id", "quantity", "is_usage_based"},
		CanUpdate: true,
		ListFilters: []commands.ListFilter{
			{Flag: "subscription-id", Query: "subscription_id", Usage: "subscription id"},
		},
		Includes: []string{"subscription", "price"},
		Extra:    []commands.ExtraCommand{subItemUsageExtra()},
	})

	commands.RegisterResource(commands.ResourceSpec[api.SubscriptionInvoice]{
		Use: "subscription-invoices", Aliases: []string{"subscription-invoice", "sub-invoices"}, Short: "Browse subscription invoices; refund and invoice",
		New:     func(c *api.Client) *api.Resource[api.SubscriptionInvoice] { return c.SubscriptionInvoices() },
		Columns: []string{"id", "subscription_id", "user_email", "total_formatted", "status"},
		ListFilters: []commands.ListFilter{
			{Flag: "store-id", Query: "store_id", Usage: "store id"},
			{Flag: "status", Query: "status", Usage: "status"},
			{Flag: "subscription-id", Query: "subscription_id", Usage: "subscription id"},
		},
		Includes: []string{"store", "subscription", "customer"},
		Extra:    []commands.ExtraCommand{subInvoiceRefundExtra(), subInvoiceGenerateExtra()},
	})

	commands.RegisterResource(commands.ResourceSpec[api.UsageRecord]{
		Use: "usage-records", Aliases: []string{"usage-record"}, Short: "Browse and create usage records",
		New:       func(c *api.Client) *api.Resource[api.UsageRecord] { return c.UsageRecords() },
		Columns:   []string{"id", "subscription_item_id", "quantity", "action", "created_at"},
		CanCreate: true,
		ListFilters: []commands.ListFilter{
			{Flag: "subscription-item-id", Query: "subscription_item_id", Usage: "subscription item id"},
		},
		Includes: []string{"subscription-item"},
	})

	// --- Discounts (create + delete) ---
	commands.RegisterResource(commands.ResourceSpec[api.Discount]{
		Use: "discounts", Aliases: []string{"discount"}, Short: "Manage discounts",
		New:       func(c *api.Client) *api.Resource[api.Discount] { return c.Discounts() },
		Columns:   []string{"id", "name", "code", "amount", "amount_type", "status"},
		CanCreate: true, CanDelete: true,
		ListFilters: []commands.ListFilter{{Flag: "store-id", Query: "store_id", Usage: "store id"}},
		Includes:    []string{"store", "variants"},
	})

	commands.RegisterResource(commands.ResourceSpec[api.DiscountRedemption]{
		Use: "discount-redemptions", Aliases: []string{"discount-redemption"}, Short: "Browse discount redemptions (read-only)",
		New:     func(c *api.Client) *api.Resource[api.DiscountRedemption] { return c.DiscountRedemptions() },
		Columns: []string{"id", "discount_id", "order_id", "discount_code", "amount"},
		ListFilters: []commands.ListFilter{
			{Flag: "discount-id", Query: "discount_id", Usage: "discount id"},
			{Flag: "order-id", Query: "order_id", Usage: "order id"},
		},
		Includes: []string{"discount", "order"},
	})

	// --- Licensing ---
	commands.RegisterResource(commands.ResourceSpec[api.LicenseKey]{
		Use: "license-keys", Aliases: []string{"license-key"}, Short: "Manage license keys",
		New:       func(c *api.Client) *api.Resource[api.LicenseKey] { return c.LicenseKeys() },
		Columns:   []string{"id", "key_short", "user_email", "status", "activation_limit", "instances_count"},
		CanUpdate: true,
		ListFilters: []commands.ListFilter{
			{Flag: "store-id", Query: "store_id", Usage: "store id"},
			{Flag: "order-id", Query: "order_id", Usage: "order id"},
			{Flag: "status", Query: "status", Usage: "status"},
		},
		Includes: []string{"store", "customer", "order", "product"},
	})

	commands.RegisterResource(commands.ResourceSpec[api.LicenseKeyInstance]{
		Use: "license-key-instances", Aliases: []string{"license-key-instance"}, Short: "Browse license key instances (read-only)",
		New:     func(c *api.Client) *api.Resource[api.LicenseKeyInstance] { return c.LicenseKeyInstances() },
		Columns: []string{"id", "license_key_id", "name", "identifier", "created_at"},
		ListFilters: []commands.ListFilter{
			{Flag: "license-key-id", Query: "license_key_id", Usage: "license key id"},
		},
		Includes: []string{"license-key"},
	})

	// --- Checkouts (create) ---
	commands.RegisterResource(commands.ResourceSpec[api.Checkout]{
		Use: "checkouts", Aliases: []string{"checkout"}, Short: "Browse and create checkouts",
		New:         func(c *api.Client) *api.Resource[api.Checkout] { return c.Checkouts() },
		Columns:     []string{"id", "store_id", "variant_id", "url", "expires_at"},
		CanCreate:   true,
		ListFilters: []commands.ListFilter{{Flag: "store-id", Query: "store_id", Usage: "store id"}},
		Includes:    []string{"store", "variant"},
	})

	// --- Webhooks (full CRUD) ---
	commands.RegisterResource(commands.ResourceSpec[api.Webhook]{
		Use: "webhooks", Aliases: []string{"webhook"}, Short: "Manage webhooks (full CRUD)",
		New:       func(c *api.Client) *api.Resource[api.Webhook] { return c.Webhooks() },
		Columns:   []string{"id", "store_id", "url", "last_sent_at"},
		CanCreate: true, CanUpdate: true, CanDelete: true,
		ListFilters: []commands.ListFilter{{Flag: "store-id", Query: "store_id", Usage: "store id"}},
		Includes:    []string{"store"},
	})

	// --- Singletons / non-JSON:API groups ---
	commands.RegisterCommand(usersCommand)
	commands.RegisterCommand(licenseCommand)
}
