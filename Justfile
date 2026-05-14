# Development tasks for gtr (https://github.com/casey/just)
# Install: https://just.systems/   then: just --list

set shell := ["bash", "-eu", "-o", "pipefail", "-c"]

# Show available recipes
default:
    @just --list

# Full local gate (matches scripts/verify.sh; Linux-friendly race tests)
verify: tidy mod-verify vet test-race build
    ./gtr -V
    ./gtr --help >/dev/null
    @echo "verify: OK"

smoke: build
    ./gtr -V
    ./gtr --help >/dev/null

build binary="gtr":
    go build -o "{{ binary }}" ./cmd/gtr

run *args:
    go run ./cmd/gtr {{ args }}

install:
    go install ./cmd/gtr

test:
    go test ./...

test-race:
    go test -race ./...

tidy:
    go mod tidy

mod-verify:
    go mod verify

vet:
    go vet ./...

fmt:
    go fmt ./...

# Regenerate embedded language tables (default path: sibling translate-shell checkout)
gen-lang path="../translate-shell/include/LanguageData.awk":
    python3 scripts/gen_language_support.py "{{ path }}"

clean:
    rm -f gtr gtr.exe

# Windows-friendly verify (no race detector; run from Git Bash or WSL if `just` is available)
verify-win: tidy mod-verify vet test build smoke
    @echo "verify-win: OK"
