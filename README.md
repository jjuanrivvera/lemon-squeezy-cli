# lsqueezy

> 🏭 Part of the [cliwright](https://cliwright.jjuanrivvera.com) CLI fleet.

A polished, production-grade command-line interface for [Lemon Squeezy](https://lemonsqueezy.com) —
manage stores, products, orders, subscriptions, customers, discounts, license keys, checkouts,
and webhooks, and script it all with table/json/yaml/csv output, a `--jq` filter, named accounts,
retries, and an MCP server for AI agents.

Built with the [cliwright](https://github.com/jjuanrivvera/cliwright) playbook: a generic core
with thin resources, secrets in the OS keyring, idempotent-only retries, and a `--dry-run` that
prints the equivalent `curl`. Lemon Squeezy speaks [JSON:API](https://jsonapi.org); `lsqueezy`
hides that envelope behind flat, table-friendly records and a `--rel name=type:id` flag for
relationships.

## Install

```bash
# Install script (macOS/Linux) — downloads the release binary, verifies its checksum
curl -fsSL https://raw.githubusercontent.com/jjuanrivvera/lemon-squeezy-cli/main/install.sh | sh

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

Global flags: `-o/--output` (table|json|yaml|csv), `--jq <expr>`, `--columns`, `--filter`,
`--sort`, `--all`, `--limit`, `--page`, `--account`, `--base-url`, `--dry-run`, `--show-token`,
`--no-color`, `--quiet`, `-v/--verbose`. Table color is on only for a TTY and honors `NO_COLOR`.

`--jq` runs a built-in [gojq](https://github.com/itchyny/gojq) program over the result before
it renders, in any format:

```bash
lsqueezy orders list -o json --jq '.[].id'
lsqueezy orders list --jq '[.[] | {id, email: .user_email, total: .total_formatted}]'
lsqueezy stores get 1 -o json --jq '.total_sales'
```

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
`--account` and its hidden `--profile` alias, `--base-url`) so an agent can't read the key or
switch accounts.

## Configuration & precedence

`flag > env (LEMONSQUEEZY_*) > config file > default`. Config lives at
`$XDG_CONFIG_HOME/lemon-squeezy-cli/config.yaml` (or `~/.lemon-squeezy-cli/config.yaml`). The API
key is stored in the OS keyring (with an encrypted-file fallback for headless boxes), never in the
config file. A profile is an account, so the selector is `--account`: one machine can hold several
keys (e.g. a live key and a test-mode key) and switch between them —
`lsqueezy --account test orders list` (or `LEMONSQUEEZY_ACCOUNT=test`). `--profile` still works as
a hidden back-compat alias.

## Development

```bash
make build      # build bin/lsqueezy
make check      # fmt + vet + lint + test (local gate)
make verify     # DETERMINISTIC gate: check + spec-check + spec-completeness + cover-check + dod
make judge      # LLM-scored subjective gate (needs claude/codex; not part of verify)
make accept     # full acceptance: verify + judge
make docs-build # regenerate the CLI reference and build the docs site (mkdocs, strict)
make setup-hooks
```

`make verify` is the token-free gate CI runs; `make judge` adds the one non-deterministic LLM
check, and `make accept` is both together (the build-acceptance gate).

## Docs site

The documentation is [MkDocs](https://www.mkdocs.org/) (Material theme). `make docs-gen`
regenerates the per-command reference from the live CLI into `docs/commands/`, and
`make docs-build` builds the whole site (getting-started + user-guide + command reference).

## Roadmap

Everything below is now shipped:

- [x] **Docs site** — `mkdocs.yml` + getting-started and user-guide pages; `make docs-build` passes.
- [x] **`--jq` output filter** — a global `--jq <expr>` (gojq) filters the result before rendering.
- [x] **API completeness** — the manifest wraps the full Lemon Squeezy API (Store + License API),
  enumerated against the official SDK (59/59 methods; enforced by `make spec-completeness`).
- [x] **Deeper tests** — broadened the `httptest` mock coverage across the resource set (≥80% gate).
- [x] **Per-account selector** — `--account` (with `--profile` kept as a hidden alias).

## License

MIT — see [LICENSE](LICENSE).
