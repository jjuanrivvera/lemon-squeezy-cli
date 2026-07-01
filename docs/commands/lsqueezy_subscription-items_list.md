## lsqueezy subscription-items list

List subscription-items

```
lsqueezy subscription-items list [flags]
```

### Examples

```
  lsqueezy subscription-items list --limit 25
  lsqueezy subscription-items list -o json --all
```

### Options

```
  -h, --help                     help for list
      --include strings          embed related resources: subscription,price
      --subscription-id string   filter: subscription id
```

### Options inherited from parent commands

```
      --account string    named account to use (env LEMONSQUEEZY_ACCOUNT)
      --all               fetch all pages (list commands)
      --base-url string   override the API base URL
      --columns strings   comma-separated columns to show
      --dry-run           print the equivalent curl and make no request
      --filter strings    client-side field=value filters (list commands)
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

* [lsqueezy subscription-items](lsqueezy_subscription-items.md)	 - Manage subscription items

