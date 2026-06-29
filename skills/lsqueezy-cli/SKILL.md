---
name: lsqueezy-cli
description: >
  Operate the Lemon Squeezy e-commerce API from the terminal with the `lsqueezy` CLI —
  browse stores, products, variants, prices, and files; read and act on orders
  (refund, generate invoice); manage customers (create/update/archive),
  subscriptions (update/cancel), subscription items and usage, invoices, discounts,
  license keys, checkouts, and webhooks; and activate/validate/deactivate software
  licenses via the License API. Use whenever the user wants to read, report on, or
  change Lemon Squeezy data from the shell, scripts, agent workflows, or CI. Load
  this skill before running `lsqueezy` — it carries the noun/verb grammar, the
  JSON:API write contract (--data/--set/--rel), the output-format rules, and the
  safety rules that prevent irreversible mistakes (refunds, cancellations, deletes).
version: 0.1.0
homepage: https://github.com/jjuanrivvera/lemon-squeezy-cli
license: MIT
allowed-tools: Bash(lsqueezy:*)
metadata: {"openclaw":{"category":"ecommerce","emoji":"🍋","requires":{"bins":["lsqueezy"],"env":["LEMONSQUEEZY_API_KEY"]},"install":[{"kind":"brew","formula":"jjuanrivvera/lemon-squeezy-cli/lemon-squeezy-cli","bins":["lsqueezy"]},{"kind":"go","package":"github.com/jjuanrivvera/lemon-squeezy-cli@latest","bins":["lsqueezy"]}]}}
---

# Lemon Squeezy CLI (`lsqueezy`)

## Prerequisites

- Install: `brew install jjuanrivvera/lemon-squeezy-cli/lemon-squeezy-cli`, or
  `go install github.com/jjuanrivvera/lemon-squeezy-cli@latest`.
- Authenticate once: `lsqueezy auth login --api-key <key>` (stored in the OS keyring and
  verified against `/users/me`), or set `LEMONSQUEEZY_API_KEY` for CI/headless use.

## Prefer the CLI over raw curl

Lemon Squeezy speaks JSON:API (`application/vnd.api+json`, `data/attributes/relationships`,
`page[size]`/`page[number]`). `lsqueezy` hides that envelope, normalizes ids/money/booleans,
walks pagination with `--all`, retries idempotent calls, and prints actionable errors. Reach
for `lsqueezy api <METHOD> <PATH>` only for an endpoint the typed commands don't cover.

## Golden rules

1. **Agents must opt into a parseable format**: pass `-o json` (or `--jq`) — table is the
   default for humans.
2. **Irreversible verbs**: `orders refund`, `subscription-invoices refund`,
   `subscriptions cancel`, `discounts delete`, `webhooks delete`, and
   `license deactivate` cannot be undone. Confirm intent and prefer `--dry-run` first.
3. **`--dry-run` before any write** prints the exact `curl` (token redacted) and sends nothing.
4. **Never reveal the key**: it stays in the keyring; `--show-token` is for local debugging only.
5. **Writes use JSON:API flags**: `--data '<json>'` (or `@file`) for attributes, repeated
   `--set key=value` for scalars, repeated `--rel name=type:id` for relationships.

## Workflow: auth → discover → act → verify

```bash
lsqueezy auth status                          # who am I, is the key valid
lsqueezy stores list -o json                  # discover store ids
lsqueezy products list --store-id 1 --all -o json
lsqueezy orders get 12345 -o json             # inspect before acting
lsqueezy orders refund 12345 --amount 500 --dry-run   # preview
lsqueezy orders refund 12345 --amount 500     # act (irreversible)
```

## Command cheatsheet

```bash
# Catalog (read-only)
lsqueezy stores list ; lsqueezy products list --store-id 1 ; lsqueezy variants list --product-id 9
lsqueezy prices list --variant-id 3 ; lsqueezy files list --variant-id 3

# Customers
lsqueezy customers list --email a@b.co -o json
lsqueezy customers create --set name=Acme --set email=a@b.co --rel store=stores:1
lsqueezy customers update 7 --set city=Bogota
lsqueezy customers archive 7

# Orders & subscriptions
lsqueezy orders list --store-id 1 --all
lsqueezy orders refund 123 --amount 500
lsqueezy orders generate-invoice 123 --name 'Acme' --country US
lsqueezy subscriptions list --status active
lsqueezy subscriptions update 99 --set pause=null
lsqueezy subscriptions cancel 99
lsqueezy subscription-items current-usage 4 -o json

# Discounts, license keys, checkouts, webhooks
lsqueezy discounts create --data '{"name":"SAVE","code":"SAVE10","amount":10,"amount_type":"percent"}' --rel store=stores:1
lsqueezy license-keys update 7 --set activation_limit=10
lsqueezy checkouts create --rel store=stores:1 --rel variant=variants:2 -o json
lsqueezy webhooks create --data '{"url":"https://x/h","events":["order_created"]}' --rel store=stores:1

# License API (just a license key; store key optional)
lsqueezy license validate   --key <license-key>
lsqueezy license activate   --key <license-key> --instance-name my-laptop
lsqueezy license deactivate --key <license-key> --instance-id <instance-id>
```

## Output & filtering

`-o table|json|yaml|csv`, `--columns id,name`, `--filter status=active` (client-side),
server filters via per-resource flags (e.g. `--store-id`), `--sort -createdAt`, `--all`,
`--limit`, `--page`, `--jq '<expr>'`. CSV cells are sanitized against formula injection.

## Troubleshooting

- `401` → `lsqueezy auth login` (or check `LEMONSQUEEZY_API_KEY`).
- `403` → the key lacks permission / wrong test-vs-live mode.
- `422` → validation; check required attributes/relationships against the docs (`--dry-run`
  to see the exact body you're sending).
- `429` → rate limited (300/min); the client backs off automatically.
- `lsqueezy doctor` checks config, credentials, connectivity, and clock.

See `references/` for deeper guides: `auth-and-config.md`, `lemonsqueezy-commands.md`,
`output-and-filtering.md`.
