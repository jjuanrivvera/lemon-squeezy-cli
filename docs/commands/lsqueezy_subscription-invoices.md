## lsqueezy subscription-invoices

Browse subscription invoices; refund and invoice

### Options

```
  -h, --help[=true|false]   help for subscription-invoices
```

### Options inherited from parent commands

```
      --account string            named account to use (env LEMONSQUEEZY_ACCOUNT)
      --all[=true|false]          fetch all pages (list commands)
      --base-url string           override the API base URL
      --columns strings           comma-separated columns to show
      --dry-run[=true|false]      print the equivalent curl and make no request
      --filter strings            client-side field=value filters (list commands)
      --jq string                 gojq expression applied to the result before rendering
      --limit int                 page size, 1-100 (list commands)
      --no-color[=true|false]     disable colored output
  -o, --output string             output format: table|json|yaml|csv
      --page int                  page number, 1-based (list commands)
      --quiet[=true|false]        suppress non-essential chatter
      --show-token[=true|false]   reveal the API key in dry-run output
      --sort string               JSON:API sort field, prefix with - for desc (list commands)
  -v, --verbose[=true|false]      verbose request logging
```

### SEE ALSO

* [lsqueezy](lsqueezy.md)	 - A polished CLI for the Lemon Squeezy e-commerce API
* [lsqueezy subscription-invoices generate-invoice](lsqueezy_subscription-invoices_generate-invoice.md)	 - Generate a downloadable invoice for a subscription invoice
* [lsqueezy subscription-invoices get](lsqueezy_subscription-invoices_get.md)	 - Get a single subscription-invoice by id
* [lsqueezy subscription-invoices list](lsqueezy_subscription-invoices_list.md)	 - List subscription-invoices
* [lsqueezy subscription-invoices refund](lsqueezy_subscription-invoices_refund.md)	 - Issue a refund for a subscription invoice (IRREVERSIBLE)

