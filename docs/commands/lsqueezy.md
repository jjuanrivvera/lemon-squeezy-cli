## lsqueezy

A polished CLI for the Lemon Squeezy e-commerce API

### Synopsis

lsqueezy is a production-grade command-line interface for Lemon Squeezy.

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
  lsqueezy license validate --key 38b1460a-5104-4067-a91d-77b872934d51

### Options

```
      --account string    named account to use (env LEMONSQUEEZY_ACCOUNT)
      --all               fetch all pages (list commands)
      --base-url string   override the API base URL
      --columns strings   comma-separated columns to show
      --dry-run           print the equivalent curl and make no request
      --filter strings    client-side field=value filters (list commands)
  -h, --help              help for lsqueezy
      --jq string         gojq expression applied to the result before rendering
      --limit int         page size, 1-100 (list commands)
      --no-color          disable colored output
  -o, --output string     output format: table|json|yaml|csv
      --page int          page number, 1-based (list commands)
      --quiet             suppress non-essential chatter
      --show-token        reveal the API key in dry-run output
      --sort string       JSON:API sort field, prefix with - for desc (list commands)
  -v, --verbose           verbose request logging
```

### SEE ALSO

* [lsqueezy agent](lsqueezy_agent.md)	 - Generate AI-agent safety config from the live command tree
* [lsqueezy alias](lsqueezy_alias.md)	 - Manage user-defined command aliases
* [lsqueezy api](lsqueezy_api.md)	 - Make a raw authenticated request (escape hatch)
* [lsqueezy auth](lsqueezy_auth.md)	 - Manage API credentials
* [lsqueezy checkouts](lsqueezy_checkouts.md)	 - Browse and create checkouts
* [lsqueezy completion](lsqueezy_completion.md)	 - Generate shell completion script
* [lsqueezy config](lsqueezy_config.md)	 - Inspect and edit configuration
* [lsqueezy customers](lsqueezy_customers.md)	 - Manage customers
* [lsqueezy discount-redemptions](lsqueezy_discount-redemptions.md)	 - Browse discount redemptions (read-only)
* [lsqueezy discounts](lsqueezy_discounts.md)	 - Manage discounts
* [lsqueezy doctor](lsqueezy_doctor.md)	 - Diagnose config, credentials, and connectivity
* [lsqueezy files](lsqueezy_files.md)	 - Browse downloadable files (read-only)
* [lsqueezy init](lsqueezy_init.md)	 - First-run wizard: capture base URL + key, write config, smoke-test
* [lsqueezy license](lsqueezy_license.md)	 - Activate, validate, and deactivate license keys (License API)
* [lsqueezy license-key-instances](lsqueezy_license-key-instances.md)	 - Browse license key instances (read-only)
* [lsqueezy license-keys](lsqueezy_license-keys.md)	 - Manage license keys
* [lsqueezy mcp](lsqueezy_mcp.md)	 - MCP server management
* [lsqueezy order-items](lsqueezy_order-items.md)	 - Browse order line items (read-only)
* [lsqueezy orders](lsqueezy_orders.md)	 - Browse orders; refund and invoice
* [lsqueezy prices](lsqueezy_prices.md)	 - Browse variant prices (read-only)
* [lsqueezy products](lsqueezy_products.md)	 - Browse products (read-only)
* [lsqueezy stores](lsqueezy_stores.md)	 - Browse stores (read-only)
* [lsqueezy subscription-invoices](lsqueezy_subscription-invoices.md)	 - Browse subscription invoices; refund and invoice
* [lsqueezy subscription-items](lsqueezy_subscription-items.md)	 - Manage subscription items
* [lsqueezy subscriptions](lsqueezy_subscriptions.md)	 - Manage subscriptions
* [lsqueezy update](lsqueezy_update.md)	 - Update lsqueezy to the latest GitHub release
* [lsqueezy usage-records](lsqueezy_usage-records.md)	 - Browse and create usage records
* [lsqueezy users](lsqueezy_users.md)	 - The authenticated user (read-only)
* [lsqueezy variants](lsqueezy_variants.md)	 - Browse product variants (read-only)
* [lsqueezy version](lsqueezy_version.md)	 - Print version, commit, and build date
* [lsqueezy webhooks](lsqueezy_webhooks.md)	 - Manage webhooks (full CRUD)

