# lsqueezy

A command-line interface for [Lemon Squeezy](https://lemonsqueezy.com). Manage stores,
products, orders, subscriptions, customers, discounts, license keys, checkouts, and webhooks
from your terminal, and script it all with `table`/`json`/`yaml`/`csv` output, a `--jq`
filter, named accounts, retries, and an MCP server for AI agents.

Built with the [cliwright](https://github.com/jjuanrivvera/cliwright) playbook: a generic core
with thin resources, secrets in the OS keyring, idempotent-only retries, and a `--dry-run`
that prints the equivalent `curl`. Lemon Squeezy speaks [JSON:API](https://jsonapi.org);
`lsqueezy` hides that envelope behind flat, table-friendly records and a `--rel name=type:id`
flag for relationships.

```bash
lsqueezy auth login --api-key eyJ0eXAiOi...
lsqueezy stores list
lsqueezy orders list --store-id 1 -o json --jq '.[].total_formatted'
lsqueezy subscriptions cancel 9999 --dry-run
```

## Where to next

- [Installation](getting-started/installation.md) — Homebrew, `go install`, or a release binary.
- [Authentication](getting-started/authentication.md) — get a key, store it in the keyring.
- [First call](getting-started/first-call.md) — verify auth and read your first resource.
- [Output & filtering](user-guide/output-and-filtering.md) — formats, columns, `--filter`, `--jq`.
- [Accounts & configuration](user-guide/accounts-and-config.md) — profiles, config precedence.
- [Creating & updating](user-guide/writing-data.md) — the JSON:API write flags.
- [AI agents](user-guide/agents.md) — the MCP server and the agent guard.
- [Command reference](commands/index.md) — every command, auto-generated from the CLI.

## The full API, wrapped

`lsqueezy` covers the entire Lemon Squeezy API surface — the Store API (stores, products,
variants, prices, files, customers, orders, order items, subscriptions, subscription items and
invoices, usage records, discounts and redemptions, license keys and instances, checkouts,
webhooks, users) and the separate License API (`activate`/`validate`/`deactivate`).
