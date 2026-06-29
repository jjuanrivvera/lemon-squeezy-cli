# Output & filtering

## Formats

`-o/--output table|json|yaml|csv` (default `table`). Agents should pass `-o json`.

```bash
lsqueezy orders list -o json
lsqueezy orders list -o csv > orders.csv
lsqueezy products get 1 -o yaml
```

- **table** is colored only on a TTY; honors `NO_COLOR` and `--no-color`. Columns are
  deterministic; wide cells are truncated with `…` (use `-o json` for full values).
- **csv** cells are sanitized against spreadsheet formula injection (a leading `= + @ -` is
  neutralized).
- Notes/warnings go to **stderr**, so stdout stays pipe-clean.

## Columns

```bash
lsqueezy orders list --columns id,order_number,total_formatted,status
```

## Filtering

Two layers:

- **Server filters** (per resource flag → JSON:API `filter[...]`), e.g.
  `lsqueezy orders list --store-id 1 --user-email a@b.co`.
- **Client filters** (post-fetch, any field): `--filter status=active` (repeatable). Combine
  with `--all` to filter across every page.

```bash
lsqueezy subscriptions list --status active --all
lsqueezy products list --store-id 1 --filter status=published -o json
```

## Sorting & pagination

```bash
lsqueezy orders list --sort -createdAt     # leading - = descending (JSON:API sort)
lsqueezy orders list --limit 100 --page 2  # page[size]/page[number]
lsqueezy orders list --all                 # walk every page (uses meta.page.lastPage)
```

## Embedding related resources

```bash
lsqueezy orders get 123 --include customer,order-items -o json
lsqueezy products list --include variants -o json
```

## gojq escape hatch

```bash
lsqueezy orders list -o json --jq '.[] | {id, total_formatted, status}'
lsqueezy subscriptions list -o json --jq '[.[] | select(.status=="active")] | length'
```

## Piping ids

```bash
lsqueezy orders list -o json --jq '.[].id' | xargs -I{} lsqueezy orders get {} -o json
```
