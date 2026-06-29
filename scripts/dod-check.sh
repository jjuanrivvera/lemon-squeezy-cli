#!/usr/bin/env bash
# dod-check.sh — deterministic Definition-of-Done checks (cliwright GOAL.md §9/§12).
# One concrete check per atomic criterion. Copied into a generated CLI under scripts/.
# Usage: ./scripts/dod-check.sh <binary-name>
set -uo pipefail
BIN="${1:-__BINARY__}"
fail=0

ok()   { printf "  ✓ %s\n" "$1"; }
bad()  { printf "  ✗ %s\n" "$1"; fail=1; }
# have <description> <test-command...>
have() { if eval "${@:2}" >/dev/null 2>&1; then ok "$1"; else bad "$1"; fi; }

echo "Definition-of-Done checks for '$BIN':"

# Agent surface
have "mcp server command present"        "rg -lq 'ophis|mcp' commands/mcp.go"
have "agent guard command present"       "test -f commands/agent.go"

# Output formats (atomic — one per format)
for f in json yaml csv table; do
  have "output format: $f"               "rg -liq '\"$f\"|format$f|$f *format' internal/output"
done

# Resilience & safety
have "--dry-run prints equivalent curl"  "rg -lq 'dry-run' . && rg -lq 'curl' internal/api"
# NOTE: this project keeps main.go at the repo root (not cmd/), so check there.
have "Ctrl-C: signal.NotifyContext"      "rg -lq 'signal.NotifyContext' main.go"
# The rule targets stray call sites that should thread cmd.Context(). Exclude _test.go
# (tests legitimately use Background) and main.go (the ONE legitimate root context, passed to
# signal.NotifyContext — the template scoped this to commands/internal precisely to allow it).
have "no stray context.Background()"     "! rg -lq --glob '!*_test.go' 'context.Background()' commands internal/api resources"
have "secrets in OS keyring"             "rg -q 'zalando/go-keyring' go.mod"
have "idempotent-only retry"             "rg -lq 'idempotent|MethodGet|MethodPut|MethodDelete' internal/api"

# Meta commands (atomic — one per command)
for c in auth config init doctor completion alias api version; do
  have "meta command: $c"                "test -f commands/$c.go || rg -lq '\"$c\"' commands"
done

# Distribution & CI
have "GoReleaser config present"         "test -f .goreleaser.yaml || test -f .goreleaser.yml"
have "goreleaser check clean"            "command -v goreleaser >/dev/null && goreleaser check || true"
have "CI workflow present"               "test -f .github/workflows/ci.yml"
have "release workflow present"          "test -f .github/workflows/release.yml"

# Hygiene
for doc in README.md LICENSE CHANGELOG.md SECURITY.md AGENTS.md; do
  have "doc: $doc"                       "test -f $doc"
done
have "no committed token"                "! rg -lq '(api[_-]?key|token)\\s*[:=]\\s*[A-Za-z0-9_-]{16,}' --glob '!*.sh' --glob '!scripts/**' ."

if [[ $fail -ne 0 ]]; then
  echo "✗ Definition-of-Done incomplete"; exit 1
fi
echo "✓ Definition-of-Done satisfied"
