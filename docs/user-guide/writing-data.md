# Creating & updating (JSON:API)

Lemon Squeezy wraps records in a JSON:API envelope (`{data:{type,id,attributes,
relationships}}`). `lsqueezy` builds that envelope for you from three universal write flags, so
there's no per-resource flag code and nothing to memorize per command.

## The three write flags

| Flag                 | Purpose                                                        |
|----------------------|----------------------------------------------------------------|
| `--data <json>`      | Raw attributes as a JSON object, or `@file` to read from disk. |
| `--set key=value`    | One attribute (repeatable). Values are coerced (numbers, bools, `null`). |
| `--rel name=type:id` | One relationship (repeatable), e.g. `--rel store=stores:1`.    |

You supply the exact attribute names the API documents — `lsqueezy` never invents or hardcodes
them.

```bash
# --set + --rel
lsqueezy customers create --set name=Acme --set email=ops@acme.co --rel store=stores:1

# --data (raw JSON) + --rel
lsqueezy webhooks create \
  --data '{"url":"https://x.co/hook","events":["order_created","subscription_updated"]}' \
  --rel store=stores:1

# --data from a file
lsqueezy discounts create --data @discount.json --rel store=stores:1

# Update by id
lsqueezy subscriptions update 9999 --set pause=null --dry-run
```

Add `--dry-run` to any write to see the exact request as `curl` without sending it.

## Which resources are writable

Most Lemon Squeezy resources are read-only; write verbs are opt-in per resource:

| Resource                | Writable verbs                              |
|-------------------------|---------------------------------------------|
| `customers`             | create, update, **archive** (sets status=archived) |
| `orders`                | **refund** (irreversible), **generate-invoice** |
| `subscriptions`         | update, **cancel** (DELETE)                 |
| `subscription-items`    | update, current-usage (read)                |
| `subscription-invoices` | **refund**, **generate-invoice**            |
| `usage-records`         | create                                      |
| `discounts`             | create, delete                              |
| `license-keys`          | update                                      |
| `checkouts`             | create                                      |
| `webhooks`              | create, update, delete                      |

## The License API

The License API is not JSON:API — it takes form-encoded params, returns a flat object, and the
store API key is optional (so it's safe in customer-facing tooling):

```bash
lsqueezy license validate   --key 38b1460a-5104-4067-a91d-77b872934d51
lsqueezy license activate   --key 38b1460a-... --instance-name my-laptop
lsqueezy license deactivate --key 38b1460a-... --instance-id 1c0c...
```
