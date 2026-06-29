## lsqueezy init

First-run wizard: capture base URL + key, write config, smoke-test

```
lsqueezy init [flags]
```

### Examples

```
  lsqueezy init
  lsqueezy init --api-key live_xxx
```

### Options

```
      --api-key string    API key (omit to be prompted)
      --base-url string   base URL (default LEMONSQUEEZY default)
  -h, --help              help for init
```

### Options inherited from parent commands

```
      --all               fetch all pages (list commands)
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

* [lsqueezy](lsqueezy.md)	 - A polished CLI for the Lemon Squeezy e-commerce API

