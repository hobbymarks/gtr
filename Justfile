# Development tasks for gtr (https://github.com/casey/just)
# Install: https://just.systems/   then: just --list
#
# Editor: if you see bogus "Makefile" errors here, open gtr.code-workspace in
# VS Code / Cursor (File → Open Workspace from File…) and install the suggested
# "vscode-just" extension; or set files.associations for "Justfile" → "just" or
# "plaintext" in your user settings.

set shell := ["bash", "-eu", "-o", "pipefail", "-c"]

version := `git describe --tags --always --dirty 2>/dev/null || echo dev`
commit  := `git rev-parse --short HEAD 2>/dev/null || echo unknown`

# Show available recipes
default:
    @just --list

# Full local gate (matches scripts/verify.sh; Linux-friendly race tests)
verify: tidy mod-verify vet test-race build
    ./gtr -V
    ./gtr --help >/dev/null
    @echo "verify: OK"

# Quick smoke check (build + version + help)
smoke: build
    ./gtr -V
    ./gtr --help >/dev/null

# Build binary for the current platform
build binary="gtr":
    go build -ldflags "-X main.version={{ version }} -X main.commit={{ commit }}" -o "{{ binary }}" ./cmd/gtr

# Run gtr with arguments (e.g. just run -t fr "Hello")
run *args:
    go run ./cmd/gtr {{ args }}

# Install gtr to $GOPATH/bin (or $GOBIN)
install:
    go install ./cmd/gtr

# Run unit tests (no network)
test:
    go test ./...

# Run unit tests with race detector
test-race:
    go test -race ./...

# Tidy go.mod and go.sum
tidy:
    go mod tidy

# Verify module dependencies are unmodified
mod-verify:
    go mod verify

# Run go vet static analysis
vet:
    go vet ./...

# Run golangci-lint (requires golangci-lint installed)
lint:
    golangci-lint run ./...

# Format Go source files
fmt:
    go fmt ./...

# Regenerate embedded language tables (default path: sibling translate-shell checkout)
gen-lang path="../translate-shell/include/LanguageData.awk":
    python3 scripts/gen_language_support.py "{{ path }}"

# Remove build artifacts
clean:
    rm -f gtr gtr.exe

# Windows-friendly verify (no race detector; run from Git Bash or WSL if `just` is available)
verify-win: tidy mod-verify vet test build smoke
    @echo "verify-win: OK"
