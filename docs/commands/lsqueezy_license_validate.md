## lsqueezy license validate

Validate a license key (optionally scoped to an instance)

```
lsqueezy license validate [flags]
```

### Examples

```
  lsqueezy license validate --key 38b1460a-...
  lsqueezy license validate --key 38b1460a-... --instance-id 1c0c...
```

### Options

```
  -h, --help                 help for validate
      --instance-id string   instance id to scope the check (optional)
      --key string           license key (required)
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

* [lsqueezy license](lsqueezy_license.md)	 - Activate, validate, and deactivate license keys (License API)

