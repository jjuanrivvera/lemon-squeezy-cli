## lsqueezy subscriptions

Manage subscriptions

### Options

```
  -h, --help[=true|false]   help for subscriptions
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
* [lsqueezy subscriptions cancel](lsqueezy_subscriptions_cancel.md)	 - Cancel a subscription (sets it to cancelled; IRREVERSIBLE)
* [lsqueezy subscriptions get](lsqueezy_subscriptions_get.md)	 - Get a single subscription by id
* [lsqueezy subscriptions list](lsqueezy_subscriptions_list.md)	 - List subscriptions
* [lsqueezy subscriptions update](lsqueezy_subscriptions_update.md)	 - Update a subscription by id

