## lsqueezy files list

List files

```
lsqueezy files list [flags]
```

### Examples

```
  lsqueezy files list --limit 25
  lsqueezy files list -o json --all
```

### Options

```
  -h, --help                help for list
      --include strings     embed related resources: variant
      --variant-id string   filter: variant id
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

* [lsqueezy files](lsqueezy_files.md)	 - Browse downloadable files (read-only)

