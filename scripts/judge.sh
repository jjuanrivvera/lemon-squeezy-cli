#!/usr/bin/env bash
# judge.sh — the ONE non-deterministic part of the gate: an LLM scores the subjective
# Definition-of-Done items a grep can't prove. Copied into a generated CLI under scripts/.
# CI without an agent: set CLIWRIGHT_SKIP_JUDGE=1 to bypass *intentionally* (logs a warning;
# it never silently passes).
set -uo pipefail
THRESHOLD="${CLIWRIGHT_JUDGE_MIN:-3}"

if [[ "${CLIWRIGHT_SKIP_JUDGE:-0}" == "1" ]]; then
  echo "⚠ judge skipped (CLIWRIGHT_SKIP_JUDGE=1) — subjective DoD items NOT verified" >&2
  exit 0
fi

read -r -d '' PROMPT <<'EOF' || true
You are a STRICT senior reviewer. Inspect this CLI repo: read internal/api/errors.go,
two command files, and the root --help/-h output. Score each 0-5 (FAIL if any < 3):
  1. Errors carry actionable hints keyed by status (401/403/404/429/5xx), not "request failed".
  2. Comments explain WHY, not WHAT.
  3. --help text includes runnable examples.
  4. Output/UX reads like a first-party tool (gh-quality).
End with exactly one line: "VERDICT: PASS" or "VERDICT: FAIL".
EOF

if command -v claude >/dev/null 2>&1; then out=$(claude -p "$PROMPT" 2>/dev/null)
elif command -v codex >/dev/null 2>&1; then out=$(codex exec "$PROMPT" 2>/dev/null)
else
  echo "⚠ no agent (claude/codex) found to run the judge — set CLIWRIGHT_SKIP_JUDGE=1 to bypass intentionally" >&2
  exit 1
fi

echo "$out"
grep -q 'VERDICT: PASS' <<<"$out" || { echo "✗ judge verdict: FAIL (subjective DoD items)"; exit 1; }
echo "✓ judge verdict: PASS"
