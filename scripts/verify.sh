#!/usr/bin/env bash
# Local verification gate: run before merging each phase (matches CI on Linux).
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

go mod tidy
go mod verify
go vet ./...
go test -race ./...
go build -o gtr ./cmd/gtr
./gtr -V
./gtr --help >/dev/null
echo "verify: OK"
