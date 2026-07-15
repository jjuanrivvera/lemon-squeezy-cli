## lsqueezy update

Update lsqueezy to the latest GitHub release

### Synopsis

Download the latest lsqueezy release, verify it against checksums.txt, and replace
the running binary in place. Use 'lsqueezy update check' to see what's available without
installing.

```
lsqueezy update [flags]
```

### Examples

```
  lsqueezy update
  lsqueezy update check
```

### Options

```
  -h, --help   help for update
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

* [lsqueezy](lsqueezy.md)	 - A polished CLI for the Lemon Squeezy e-commerce API
* [lsqueezy update check](lsqueezy_update_check.md)	 - Check for a newer release without installing it

