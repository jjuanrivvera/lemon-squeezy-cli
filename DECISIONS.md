# DECISIONS

Pinned assumptions and design rules so the same API in → the same CLI out (cliwright §11).
One line each: question → decision → why. Read these back before changing the surface.

## Architecture

- **Resource pattern → Pattern A (generic-core).** Lemon Squeezy is uniform JSON:API CRUD, so
  one generic `Resource[T]` + a per-resource struct/accessor covers every collection. The few
  non-CRUD endpoints (refund, generate-invoice, cancel, current-usage, archive) are added via
  the generic `Extra`/`Action` mechanism — not a per-resource service layer. No Pattern-B
  trigger (per-resource includes are handled by one generic `--include` flag; there is no
  masquerade) was present.

- **JSON:API envelope handled once in the core.** `internal/api/jsonapi.go` flattens
  `{type,id,attributes,relationships}` into a flat struct (id merged into attributes) on read,
  and wraps `{data:{type,id,attributes,relationships}}` on write. Resource structs stay flat
  and table-friendly; the envelope never leaks into resource files.

- **Writes are universal, not per-resource.** Create/update use `--data` (raw attributes JSON
  or `@file`), `--set key=value`, and `--rel name=type:id`. Adding a writable resource needs
  zero flag code — and we never fabricate or hardcode attribute names; the user supplies the
  exact documented attributes. Write verbs are opt-in via `CanCreate/CanUpdate/CanDelete`
  because most Lemon Squeezy resources are read-only.

## API facts (from docs + the official SDK)

- **Auth → bearer token.** `Authorization: Bearer <key>`, plus `Accept`/`Content-Type:
  application/vnd.api+json` on every request (Lemon Squeezy rejects requests without it).
- **Key verification → `GET /users/me`.** The canonical whoami; used by `auth login/status`,
  `doctor`, and `init`.
- **Pagination → page-based.** `page[size]` (1–100, default 10) + `page[number]` (1-based);
  `--all` walks via `meta.page.lastPage`, falling back to `links.next`.
- **Rate limit → 300 req/min.** No quota headers documented, so a fixed 5 req/s base with
  halve-on-429 + gradual restore.
- **Money → integer cents.** Stored as exact decimal text (`Money`), never float64.
- **License API is NOT JSON:API.** `POST /v1/licenses/{activate,validate,deactivate}` take
  form-encoded params, return a flat object, and the store API key is optional (license checks
  run on customer machines). Implemented as a dedicated `license` group, not a `Resource[T]`.
- **Customer "archive" is a PATCH, not a DELETE.** Customers cannot be deleted; `archive` sets
  `status=archived` (reversible), so it is classified as a write, not destructive.
- **Subscription "cancel" is a DELETE** on `/subscriptions/{id}` (returns the cancelled
  subscription); classified destructive.

## Coverage methodology

- **Coverage is measured with `-coverpkg=./...` and `-count=1`.** The suite is integration-heavy
  (`commands_test` drives the resources + api code paths end-to-end via httptest), so total
  coverage credits that exercise. `-count=1` is required: with `-coverpkg`, a cached test result
  merges into the profile wrong and under-reports the total, which would flake `cover-check`. The
  ≥80% gate is genuinely met by real tests, not loosened.

## API completeness (cliwright §0/§11)

- **The manifest wraps the FULL Lemon Squeezy API.** `api_method_total`/`api_method_source`
  record the enumerated surface — the official SDK `@lemonsqueezy/lemonsqueezy.js`
  (github.com/lmsqueezy/lemonsqueezy.js), whose exported API functions are the canonical,
  exhaustive method list for the Store API + License API: **59 operations** (excluding the
  client-only `lemonSqueezySetup`). The manifest's verbs map 1:1 to those functions, so
  `spec-completeness.sh` reports **59/59 = 100%** — no coverage-waiver needed.

## Multi-account selector (cliwright v0.3.0)

- **The profile flag is `--account`** (`profile_flag`/`profile_noun` = `account`). A Lemon
  Squeezy profile IS an account — one API key; a machine may hold a live + a test-mode key.
  `--profile` (and `LEMONSQUEEZY_PROFILE`) stay as HIDDEN back-compat aliases, and the account
  selector is excluded from the MCP surface under BOTH names so an agent can't switch accounts.

## Gate structure (cliwright v0.3.0)

- **`verify` is deterministic; `judge` is separate.** `make verify` = check + spec-check +
  spec-completeness + cover-check + dod-check (no LLM, token-free, CI-safe). `make judge` is the
  ONE non-deterministic gate (an LLM scores subjective DoD items). `make accept` = verify + judge
  is the build-acceptance gate. CI is anchored on `make verify`.

## Binary identity

- **Command name → `lsqueezy`** (not `ls`, which shadows coreutils, nor `lsq`, which clashes
  with the Logseq CLI `jrswab/lsq`). Repo/module → `lemon-squeezy-cli`.
