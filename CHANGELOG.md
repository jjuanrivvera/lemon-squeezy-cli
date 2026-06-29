# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial CLI for the Lemon Squeezy API covering 18 JSON:API resources (stores, products,
  variants, prices, files, customers, orders, order-items, subscriptions, subscription-items,
  subscription-invoices, usage-records, discounts, discount-redemptions, license-keys,
  license-key-instances, checkouts, webhooks), the `users me` singleton, and the License API
  (`license activate|validate|deactivate`).
- Custom action verbs: orders/subscription-invoices `refund` + `generate-invoice`,
  subscriptions `cancel`, subscription-items `current-usage`, customers `archive`.
- Generic-core architecture: one generic `Resource[T]` decodes the JSON:API envelope
  (data/attributes/relationships, page-based pagination) and powers list/get/create/update/delete.
- Universal write flags `--data`, `--set`, and `--rel name=type:id` for JSON:API writes.
- Output formats: table, json, yaml, csv with `--columns`, `--filter`, `--sort`, `--all`, `--page`.
- Bearer auth verified against `/users/me`; key in the OS keyring with encrypted-file fallback;
  `LEMONSQUEEZY_API_KEY` env override; named profiles for multiple accounts.
- Resilient client: idempotent-only retry with backoff/jitter, adaptive rate limiting (300/min).
- `--dry-run` prints a redacted, copy-pasteable curl.
- Meta commands: auth, config, init, doctor, completion, alias, api, version.
- MCP server and `agent guard` (refund/cancel/delete/deactivate denied, writes ask, reads allowed).
