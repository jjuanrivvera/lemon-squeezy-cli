## lsqueezy orders list

List orders

```
lsqueezy orders list [flags]
```

### Examples

```
  lsqueezy orders list --limit 25
  lsqueezy orders list -o json --all
```

### Options

```
  -h, --help[=true|false]   help for list
      --include strings     embed related resources: store,customer,order-items
      --store-id string     filter: store id
      --user-email string   filter: buyer email
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

* [lsqueezy orders](lsqueezy_orders.md)	 - Browse orders; refund and invoice

