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

- **Coverage is measured with `-coverpkg=./...`.** The suite is integration-heavy
  (`commands_test` drives the resources + api code paths end-to-end via httptest), so total
  coverage credits that exercise. The ≥80% gate is genuinely met by real tests, not loosened.

## Binary identity

- **Command name → `lsqueezy`** (not `ls`, which shadows coreutils, nor `lsq`, which clashes
  with the Logseq CLI `jrswab/lsq`). Repo/module → `lemon-squeezy-cli`.
