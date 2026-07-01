# Output & filtering

Every command renders through one path, so these flags work everywhere.

## Formats

`-o` / `--output` picks the format: `table` (default), `json`, `yaml`, or `csv`.

```bash
lsqueezy orders list -o json
lsqueezy products list -o csv
lsqueezy stores get 1 -o yaml
```

Table color is on only for a TTY and honors `NO_COLOR`. Use `--no-color` to force it off.

## Columns

`--columns` selects and orders the fields shown in `table` and `csv`:

```bash
lsqueezy products list --columns id,name,status
lsqueezy orders list -o csv --columns id,order_number,total_formatted,status
```

## Server-side filters

Resources expose typed filter flags that map to JSON:API `filter[...]` query params:

```bash
lsqueezy products list --store-id 1
lsqueezy orders list --store-id 1 --user-email buyer@acme.co
lsqueezy subscriptions list --status active
```

## Client-side `--filter`

`--filter field=value` (repeatable) trims the fetched rows by any field, on top of any
server-side filter:

```bash
lsqueezy subscriptions list --store-id 1 --filter status=active --all
```

## Sorting

```bash
lsqueezy orders list --sort -createdAt   # prefix with - for descending
```

## `--jq`: filter the result before it renders

`--jq <expr>` runs a [gojq](https://github.com/itchyny/gojq) program over the result before it
is rendered — the same query language as `jq`, built in, no external binary. It applies to
whatever format you pick, so you can extract a field as JSON, reshape into a table, or pull a
single scalar.

```bash
# Pull just the ids
lsqueezy orders list -o json --jq '.[].id'

# One field per row
lsqueezy products list -o json --jq '.[].name'

# Reshape into objects, then render as a table
lsqueezy orders list --jq '[.[] | {id, email: .user_email, total: .total_formatted}]'

# A single scalar
lsqueezy stores get 1 -o json --jq '.total_sales'
```

A malformed expression fails fast with a clear error and makes no change to the data. For
raw endpoints the typed commands don't cover, reach for `lsqueezy api <METHOD> <PATH>` and pipe
into your own tools.
