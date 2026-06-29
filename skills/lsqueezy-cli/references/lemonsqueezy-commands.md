# Command reference

Every resource has `list` and `get <id>`. Write verbs are present only where the API allows
them. `list` supports per-resource server filters (flags below), plus the global
`--all/--limit/--page/--sort/--filter/--include`.

## Reads (read-only)

| Resource | list filters | include |
|---|---|---|
| `stores` | — | — |
| `products` | `--store-id` | store, variants |
| `variants` | `--product-id` | product |
| `prices` | `--variant-id` | variant |
| `files` | `--variant-id` | variant |
| `order-items` | `--order-id`, `--product-id` | order, product, variant |
| `discount-redemptions` | `--discount-id`, `--order-id` | discount, order |
| `license-key-instances` | `--license-key-id` | license-key |

## Reads + actions

| Resource | extra verbs |
|---|---|
| `orders` (`--store-id`, `--user-email`) | `refund <id> [--amount cents]`, `generate-invoice <id> --name … --country …` |
| `subscription-invoices` (`--store-id`, `--status`, `--subscription-id`) | `refund <id> [--amount]`, `generate-invoice <id> …` |
| `subscription-items` (`--subscription-id`) | `update <id>`, `current-usage <id>` |

## Writes

| Resource | verbs |
|---|---|
| `customers` (`--store-id`, `--email`) | `create`, `update <id>`, `archive <id>` |
| `subscriptions` (`--store-id`, `--status`, `--user-email`) | `update <id>`, `cancel <id>` |
| `usage-records` (`--subscription-item-id`) | `create` |
| `discounts` (`--store-id`) | `create`, `delete <id>` |
| `license-keys` (`--store-id`, `--order-id`, `--status`) | `update <id>` |
| `checkouts` (`--store-id`) | `create` |
| `webhooks` (`--store-id`) | `create`, `update <id>`, `delete <id>` |

## Singletons & License API

```bash
lsqueezy users me                       # GET /users/me
lsqueezy license validate   --key <k> [--instance-id <i>]
lsqueezy license activate   --key <k> --instance-name <name>
lsqueezy license deactivate --key <k> --instance-id <i>
```

## Writing records (JSON:API)

Attributes come from `--data` and/or `--set`; relationships from `--rel name=type:id`.

```bash
# attributes via raw JSON
lsqueezy webhooks create --data '{"url":"https://x/h","events":["order_created"]}' --rel store=stores:1
# attributes via repeated --set (values are JSON-parsed, else treated as strings)
lsqueezy customers create --set name=Acme --set email=a@b.co --rel store=stores:1
# update only sends the fields you set
lsqueezy subscriptions update 99 --set cancelled=true
```

`--set quantity=5` sends a number; `--set name=Acme` sends a string; `--set pause=null` sends
JSON null. Use `--data @file.json` for large bodies.

## Raw escape hatch

```bash
lsqueezy api GET /orders -q 'page[size]=1'
lsqueezy api POST /checkouts -d @checkout.json
```
