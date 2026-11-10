## lsqueezy api

Make a raw authenticated request (escape hatch)

### Synopsis

Make a raw authenticated request against the API.

Honors --dry-run (prints the equivalent curl) and --output. Use it for endpoints the
typed resource commands don't cover yet.

```
lsqueezy api <METHOD> <PATH> [flags]
```

### Examples

```
  lsqueezy api GET /products -q 'page[size]=1'
  lsqueezy api POST /checkouts -d @checkout.json
```

### Options

```
  -d, --data string         request body (JSON)
  -h, --help[=true|false]   help for api
  -q, --query stringArray   query param key=value (repeatable)
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

