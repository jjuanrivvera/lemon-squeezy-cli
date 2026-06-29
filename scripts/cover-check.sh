#!/usr/bin/env bash
# cover-check.sh — fail if total test coverage is below the threshold.
# Copied into a generated CLI under scripts/. Usage: ./scripts/cover-check.sh [min%]
set -euo pipefail
THRESHOLD="${1:-80}"

[[ -f coverage.out ]] || go test ./... -coverprofile=coverage.out >/dev/null
pct=$(go tool cover -func=coverage.out | awk '/^total:/ { gsub(/%/,"",$3); print $3 }')

awk -v p="$pct" -v t="$THRESHOLD" 'BEGIN {
  if (p+0 < t+0) { printf "✗ coverage %.1f%% < %d%%\n", p, t; exit 1 }
  printf "✓ coverage %.1f%% ≥ %d%%\n", p, t
}'
