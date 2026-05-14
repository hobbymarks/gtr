#!/usr/bin/env bash
# Optional quarterly check: compare pinned translate-shell translator sources.
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
pin_file="$root/docs/UPSTREAM_TRANSLATE_SHELL.md"
if [[ ! -f "$pin_file" ]]; then
  echo "missing $pin_file" >&2
  exit 1
fi
echo "See $pin_file for the pinned translate-shell commit."
echo "Manually diff include/Translators/*.awk against your translate-shell checkout."
echo "When parsers or URLs change, update fixtures and bump the pin in that file."
