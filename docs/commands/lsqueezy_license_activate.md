## lsqueezy license activate

Activate a license key for a new instance

```
lsqueezy license activate [flags]
```

### Examples

```
  lsqueezy license activate --key 38b1460a-... --instance-name my-laptop
```

### Options

```
  -h, --help                   help for activate
      --instance-name string   a label for this activation (required)
      --key string             license key (required)
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

* [lsqueezy license](lsqueezy_license.md)	 - Activate, validate, and deactivate license keys (License API)

