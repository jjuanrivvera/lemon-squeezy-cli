## lsqueezy subscription-invoices generate-invoice

Generate a downloadable invoice for a subscription invoice

```
lsqueezy subscription-invoices generate-invoice <id> [flags]
```

### Examples

```
  lsqueezy subscription-invoices generate-invoice 55 --name 'Acme' --country US
```

### Options

```
      --address string    billing address
      --city string       billing city
      --country string    billing country (ISO code)
  -h, --help              help for generate-invoice
      --locale string     invoice locale
      --name string       billing name
      --notes string      invoice notes
      --state string      billing state
      --zip-code string   billing zip/postal code
```

### Options inherited from parent commands

```
      --all               fetch all pages (list commands)
      --base-url string   override the API base URL
      --columns strings   comma-separated columns to show
      --dry-run           print the equivalent curl and make no request
      --filter strings    client-side field=value filters (list commands)
      --limit int         page size, 1-100 (list commands)
      --no-color          disable colored output
  -o, --output string     output format: table|json|yaml|csv
      --page int          page number, 1-based (list commands)
      --profile string    config profile to use
      --quiet             suppress non-essential chatter
      --show-token        reveal the API key in dry-run output
      --sort string       JSON:API sort field, prefix with - for desc (list commands)
  -v, --verbose           verbose request logging
```

### SEE ALSO

* [lsqueezy subscription-invoices](lsqueezy_subscription-invoices.md)	 - Browse subscription invoices; refund and invoice

