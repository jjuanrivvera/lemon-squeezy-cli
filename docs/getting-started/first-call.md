# Your first call

Once you've [authenticated](authentication.md), read something.

## List your stores

```bash
lsqueezy stores list
```

```text
ID  NAME       SLUG       CURRENCY  TOTAL_SALES
1   My Store   my-store   USD       42
```

## Get one record

```bash
lsqueezy stores get 1
lsqueezy products get 55 -o json
```

## Page through everything

List commands return one page by default. Use `--all` to walk every page, or `--limit` /
`--page` to control paging yourself:

```bash
lsqueezy products list --store-id 1 --all
lsqueezy orders list --limit 50 --page 2
```

## Look before you leap

Any write can be previewed with `--dry-run`, which prints the equivalent `curl` (with the token
redacted) and makes no request:

```bash
lsqueezy subscriptions cancel 9999 --dry-run
#  -> curl -X DELETE 'https://api.lemonsqueezy.com/v1/subscriptions/9999' \
#       -H 'Authorization: Bearer REDACTED' -H 'Accept: application/vnd.api+json'
```

Next: shape the output with [formats, columns, `--filter`, and `--jq`](../user-guide/output-and-filtering.md).
