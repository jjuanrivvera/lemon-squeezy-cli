# AGENTS.md — lsqueezy

`lsqueezy` is a production-grade CLI for [Lemon Squeezy](https://lemonsqueezy.com), built with the
cliwright playbook. This file orients any AI agent working in the repo.

## Architecture

- **Generic core, thin resources.** The CRUD/pagination/retry/output **and the JSON:API
  envelope** logic lives once in `internal/api/`. Adding a resource is a type + a `Client`
  accessor + one `commands.RegisterResource(...)` call — no edits to shared code.
- **JSON:API.** Lemon Squeezy wraps records as `{data:{type,id,attributes,relationships}}`.
  `internal/api/jsonapi.go` flattens id+attributes into the flat resource struct on read and
  builds the write envelope on write, so resource structs stay flat and table-friendly.
- **Pattern A (generic-core).** `internal/api/resource.go` exposes `Resource[T]` with
  `List/ListAll/Get/Create/Update/Delete/Action/ActionOne`. `commands/generic.go` builds the
  subcommands; write verbs are opt-in (`CanCreate/CanUpdate/CanDelete`). Writes use universal
  flags `--data`/`--set`/`--rel` (see `commands/write.go`).
- **Layout:** entry point is `main.go` at the repo root (not `cmd/lsqueezy/`). Every resource
  is registered from the single `resources/` package (`register.go` + action verbs in
  `actions.go`, singletons in `users.go`/`license.go`), blank-imported by `main.go`.
- **License API** (`license activate|validate|deactivate`) is NOT JSON:API — see
  `internal/api/license_api.go` (form-encoded, flat response, key optional).
- Pinned design rules and API facts live in `DECISIONS.md` (cliwright §11).

## Build & verify

- `make build` — build `bin/lsqueezy`.
- `make check` — fmt + vet + lint + test (local gate).
- `make verify` — full acceptance gate (check + spec-check + cover-check + dod-check + judge).
  Set `CLIWRIGHT_SKIP_JUDGE=1` to skip the LLM judge in CI environments without an agent.

## Rules

- Comments explain WHY, not WHAT.
- Never print or commit a real token; redact by default.
- Wire `cmd.Context()` from `ExecuteContext` everywhere — no stray `context.Background()`.
- Annotate resource commands (read-only/write/destructive) in the generic builder.
- Keep the CLI surface in sync with `api-manifest.json` (enforced by `make spec-check`).
