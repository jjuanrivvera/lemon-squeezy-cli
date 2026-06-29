## lsqueezy discounts create

Create a discount

### Synopsis

Create a discount.

Supply attributes with --data (raw JSON object or @file) and/or repeated --set
key=value, and relationships with repeated --rel name=type:id. The JSON:API
envelope (type/attributes/relationships) is added for you.

```
lsqueezy discounts create [flags]
```

### Examples

```
  lsqueezy discounts create --data '{"name":"Acme"}' --rel store=stores:1
  lsqueezy discounts create --set name=Acme --set email=a@b.co --rel store=stores:1 --dry-run
```

### Options

```
  -d, --data string       attributes as a JSON object, or @file
  -h, --help              help for create
      --rel stringArray   relationship name=type:id (repeatable), e.g. store=stores:1
      --set stringArray   attribute key=value (repeatable)
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

* [lsqueezy discounts](lsqueezy_discounts.md)	 - Manage discounts

