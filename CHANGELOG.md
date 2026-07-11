# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.1] - 2026-07-11

### Fixed
- `auth login` and `init` now read the API key with a **hidden** prompt (`term.ReadPassword`)
  instead of `fmt.Scanln`, which echoed the key in plaintext and could hang on long terminal
  pastes (Lemon Squeezy keys are long JWTs). The key is no longer shown on screen, and the
  read no longer stalls.

### Added
- One-line install script (`install.sh`) for macOS/Linux, checksum-verified:
  `curl -fsSL https://raw.githubusercontent.com/jjuanrivvera/lemon-squeezy-cli/main/install.sh | sh`.

## [0.2.0] - 2026-07-02

### Added
- **`agent guard` now generates a real PreToolUse enforcement hook** (Bash + MCP
  matchers) instead of only permission deny/ask rules. The hook anchors blocked
  subcommands at the command position, matches path-invoked binaries
  (`./bin/lsqueezy`, `/usr/local/bin/lsqueezy`) while ignoring a different binary
  that merely ends in `lsqueezy`, defeats quote/backslash/newline obfuscation and
  separators glued to a verb, and blocks `api DELETE/PUT/POST/PATCH` at the
  method position (a `GET` whose path contains `delete` stays allowed).
- Global `--jq <expr>` filter: runs a built-in [gojq](https://github.com/itchyny/gojq) program
  over the result before rendering, in any output format.
- MkDocs (Material) documentation site: getting-started + user-guide pages and an
  auto-generated command reference; `make docs-build` builds it (strict).
- `spec-completeness` gate: `api_method_total`/`api_method_source` in `api-manifest.json`
  enumerate the full Lemon Squeezy API (59 methods, from the official SDK); the manifest wraps
  100% of it.

### Changed
- The multi-account selector is now `--account` (env `LEMONSQUEEZY_ACCOUNT`); `--profile`
  (and `LEMONSQUEEZY_PROFILE`) remain as hidden back-compat aliases. Both are excluded from
  the MCP tool surface.
- Split the acceptance gate: `make verify` is deterministic (check + spec-check +
  spec-completeness + cover-check + dod, no LLM), `make judge` is the LLM gate, and
  `make accept` = verify + judge. CI is anchored on `make verify`.
- Deepened the `httptest` mock coverage across the resource set.

### Fixed
- `agent guard` Codex and OpenCode output now emit the config schemas those hosts
  actually read (previously an invented `[sandbox]` table and a plural
  `permissions` key were silently ignored), and `--write` now writes the files
  instead of being a silent no-op. The no-jq hook fallback no longer fails open.

### Added — initial release surface
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
