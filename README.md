# lsqueezy

A polished, production-grade command-line interface for [Lemon Squeezy](https://lemonsqueezy.com) —
manage stores, products, orders, subscriptions, customers, discounts, license keys, checkouts,
and webhooks, and script it all with table/json/yaml/csv output, named profiles, retries, and an
MCP server for AI agents.

Built with the [cliwright](https://github.com/jjuanrivvera/cliwright) playbook: a generic core
with thin resources, secrets in the OS keyring, idempotent-only retries, and a `--dry-run` that
prints the equivalent `curl`. Lemon Squeezy speaks [JSON:API](https://jsonapi.org); `lsqueezy`
hides that envelope behind flat, table-friendly records and a `--rel name=type:id` flag for
relationships.

## Install

```bash
# Homebrew (tap)
brew install jjuanrivvera/lemon-squeezy-cli/lemon-squeezy-cli

# Go
go install github.com/jjuanrivvera/lemon-squeezy-cli@latest
```

## Quickstart

```bash
# 1. Authenticate (key is stored in the OS keyring, verified against /users/me)
lsqueezy auth login --api-key eyJ0eXAiOi...
#    ...or, for CI / headless use:
export LEMONSQUEEZY_API_KEY=eyJ0eXAiOi...

# 2. Browse
lsqueezy stores list
lsqueezy products list --store-id 1 --all
lsqueezy orders get 12345 -o json

# 3. Act
lsqueezy customers create --set name=Acme --set email=ops@acme.co --rel store=stores:1
lsqueezy subscriptions cancel 9999
lsqueezy orders refund 12345 --amount 500

# 4. Inspect anything without sending it
lsqueezy subscriptions cancel 9999 --dry-run
#  -> curl -X DELETE 'https://api.lemonsqueezy.com/v1/subscriptions/9999' \
#       -H 'Authorization: Bearer REDACTED' -H 'Accept: application/vnd.api+json'
```

## Resources

| Resource                | Verbs                                          | Notes |
|-------------------------|------------------------------------------------|-------|
| `stores`                | list, get                                      | read-only |
| `products`              | list, get                                      | read-only |
| `variants`              | list, get                                      | read-only |
| `prices`                | list, get                                      | read-only |
| `files`                 | list, get                                      | read-only |
| `customers`             | list, get, create, update, **archive**         | archive sets status=archived |
| `orders`                | list, get, **refund**, **generate-invoice**    | refund is irreversible |
| `order-items`           | list, get                                      | read-only |
| `subscriptions`         | list, get, update, **cancel**                  | cancel = DELETE |
| `subscription-items`    | list, get, update, **current-usage**           | usage-based billing |
| `subscription-invoices` | list, get, **refund**, **generate-invoice**    | |
| `usage-records`         | list, get, create                              | |
| `discounts`             | list, get, create, delete                      | |
| `discount-redemptions`  | list, get                                      | read-only |
| `license-keys`          | list, get, update                              | |
| `license-key-instances` | list, get                                      | read-only |
| `checkouts`             | list, get, create                              | custom checkouts |
| `webhooks`              | list, get, create, update, delete              | full CRUD |
| `users`                 | **me**                                         | the authenticated user |
| `license`               | **activate**, **validate**, **deactivate**     | License API (not JSON:API) |

## Creating & updating (JSON:API)

Writes take attributes via `--data` (raw JSON or `@file`) and/or repeated `--set key=value`,
and relationships via repeated `--rel name=type:id`. The JSON:API envelope is added for you.

```bash
lsqueezy customers create --set name=Acme --set email=a@b.co --rel store=stores:1
lsqueezy webhooks create \
  --data '{"url":"https://x.co/hook","events":["order_created","subscription_updated"]}' \
  --rel store=stores:1
lsqueezy subscriptions update 9999 --set pause=null --dry-run
```

## Output & filtering

```bash
lsqueezy orders list -o csv
lsqueezy products list -o yaml --columns id,name,status
lsqueezy subscriptions list --filter status=active --all
lsqueezy orders list --store-id 1 --limit 50 --sort -createdAt
```

Global flags: `-o/--output` (table|json|yaml|csv), `--columns`, `--filter`, `--sort`,
`--all`, `--limit`, `--page`, `--profile`, `--base-url`, `--dry-run`, `--show-token`,
`--no-color`, `--quiet`, `-v/--verbose`. Table color is on only for a TTY and honors `NO_COLOR`.

## License keys (License API)

The License API works with just a license key (the store API key is optional), so it's safe
to ship in customer-facing tooling:

```bash
lsqueezy license validate   --key 38b1460a-5104-4067-a91d-77b872934d51
lsqueezy license activate   --key 38b1460a-... --instance-name my-laptop
lsqueezy license deactivate --key 38b1460a-... --instance-id 1c0c...
```

## Meta commands

`auth login|logout|status`, `config path|view|set|use|list-profiles`, `init`, `doctor`,
`completion`, `alias set|list|remove`, `api <METHOD> <PATH>`, `version`.

## AI agents (MCP + guard)

```bash
# Run lsqueezy as an MCP server so an agent can drive the API
lsqueezy mcp start

# Generate agent safety rules from the live command tree
lsqueezy agent guard --host claude-code   # refund/cancel/delete/deactivate denied, writes ask
lsqueezy agent guard --host codex
lsqueezy agent guard --host opencode
```

The MCP surface excludes setup/secret commands (`auth`, `config`, `--api-key`, `--show-token`,
`--profile`, `--base-url`) so an agent can't read the key or switch stores.

## Configuration & precedence

`flag > env (LEMONSQUEEZY_*) > config file > default`. Config lives at
`$XDG_CONFIG_HOME/lemon-squeezy-cli/config.yaml` (or `~/.lemon-squeezy-cli/config.yaml`). The API
key is stored in the OS keyring (with an encrypted-file fallback for headless boxes), never in the
config file. Named profiles let one machine hold several accounts (e.g. a live key and a
test-mode key): `lsqueezy --profile test orders list`.

## Development

```bash
make build      # build bin/lsqueezy
make check      # fmt + vet + lint + test (local gate)
make verify     # full acceptance gate (check + spec-check + cover-check + dod + judge)
make setup-hooks
```

## License

MIT — see [LICENSE](LICENSE).
