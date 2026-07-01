#!/usr/bin/env bash
# spec-completeness.sh — the COMPLETENESS half of the spec gate (cliwright GOAL.md §0/§11).
#
# spec-check.sh proves the CLI surface ⊆ the manifest (consistency): every command maps to a
# declared resource/verb. It can NOT see what the manifest left out. This script proves the
# other direction — manifest ≈ the FULL API (completeness): the manifest must derive from an
# ENUMERATED method/endpoint list (OpenAPI/Postman/llms.txt, else the docs' full method index
# or a community machine spec — §0), not model recall, and cover ≥ THRESHOLD% of it.
#
# Under-capture is otherwise invisible: a hand-curated manifest can wrap a third of an API and
# still pass spec-check, because every command it does ship is consistent. That is exactly how
# tgctl wrapped only ~⅓ of the Telegram Bot API. This gate makes the gap fail loudly.
#
# Copied into a generated CLI under scripts/.
# Usage: ./scripts/spec-completeness.sh [api-manifest.json] [min-coverage-% (default 90)]
set -uo pipefail
MANIFEST="${1:-api-manifest.json}"
THRESHOLD="${2:-${CLIWRIGHT_COVERAGE_MIN:-90}}"

[[ -f "$MANIFEST" ]] || { echo "✗ $MANIFEST missing — §11 requires a checked-in spec-derived manifest"; exit 1; }

TOTAL="$(jq -r '.api_method_total // 0' "$MANIFEST")"
SOURCE="$(jq -r '.api_method_source // ""' "$MANIFEST")"

# Enumeration is mandatory (§0). No recorded total/source ⇒ the manifest was authored from
# recall — the precise failure this gate exists to catch. Fail with the fix inline.
if ! [[ "$TOTAL" =~ ^[0-9]+$ ]] || [[ "$TOTAL" -le 0 ]] || [[ -z "$SOURCE" ]]; then
  cat >&2 <<'EOF'
✗ completeness: api_method_total / api_method_source missing from the manifest.
  §0 REQUIRES enumerating the COMPLETE method/endpoint set from a source (OpenAPI / Postman /
  llms.txt; else the docs' full method index or a community machine spec) BEFORE authoring the
  manifest. Record the enumerated total and where it came from:
      "api_method_total":  <enumerated total methods/endpoints in the full API>,
      "api_method_source": "<OpenAPI / Postman / llms.txt / doc-index / community-spec URL>"
EOF
  exit 1
fi

# Covered = operations the manifest actually declares. A resource contributes one per verb;
# a flat RPC-style `methods` array (Telegram-shaped APIs) contributes one each. spec-check.sh
# independently proves every declared resource/verb is a REAL reachable command, so this
# numerator is grounded in the built surface and cannot be inflated by an optimistic manifest.
COVERED="$(jq -r '((.resources // [] | map(.verbs // [] | length) | add) // 0) + (.methods // [] | length)' "$MANIFEST")"
[[ "$COVERED" =~ ^[0-9]+$ ]] || COVERED=0

PCT=$(( COVERED * 100 / TOTAL ))

# A materially-below-threshold manifest is allowed ONLY with an explicit, recorded waiver in
# DECISIONS.md (§11 "pin every assumption") — so a deliberate "read surface first, writes in v2"
# is a recorded decision the loop sees every pass, never a silent shrug.
WAIVER_NOTE=""
if [[ -f DECISIONS.md ]] && grep -Eiq 'coverage[ -]?waiver' DECISIONS.md; then
  WAIVER_NOTE="$(grep -Ei 'coverage[ -]?waiver' DECISIONS.md | head -1 | sed 's/^[[:space:]]*//')"
fi

printf "completeness: %s/%s methods covered (%d%%), threshold %d%%, source: %s\n" \
  "$COVERED" "$TOTAL" "$PCT" "$THRESHOLD" "$SOURCE"

if (( PCT < THRESHOLD )); then
  if [[ -n "$WAIVER_NOTE" ]]; then
    printf "⚠ coverage %d%% < %d%% — ACCEPTED via recorded waiver in DECISIONS.md: %s\n" \
      "$PCT" "$THRESHOLD" "$WAIVER_NOTE" >&2
    exit 0
  fi
  cat >&2 <<EOF
✗ completeness: coverage ${PCT}% < ${THRESHOLD}% and no waiver recorded.
  The manifest captures ${COVERED} of ${TOTAL} enumerated methods. Either wrap the missing
  methods (re-derive resources/verbs from the ENUMERATED list — not recall — then ship them),
  OR record an explicit, justified waiver line in DECISIONS.md containing "coverage-waiver",
  e.g.:  coverage-waiver: shipping the read surface first; ${COVERED}/${TOTAL} now, write
         methods deferred to v2 — tracked in #123.
EOF
  exit 1
fi
echo "✓ completeness: coverage ${PCT}% ≥ ${THRESHOLD}%"
