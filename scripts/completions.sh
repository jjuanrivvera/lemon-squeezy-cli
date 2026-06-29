#!/usr/bin/env bash
# completions.sh — generate shell completions into completions/ (used by GoReleaser before-hook).
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p completions
go build -o bin/lsqueezy .
for sh in bash zsh fish; do
  ./bin/lsqueezy completion "$sh" > "completions/lsqueezy.$sh"
done
echo "completions generated"
