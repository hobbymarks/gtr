# gtr

Go multi-engine translation CLI, inspired by [translate-shell](https://github.com/soimort/translate-shell). This repository follows a phased roadmap (see **Development plan** below).

## Goals

- Cross-platform CLI (Linux, macOS, Windows) with a small static binary.
- Multiple translation engines (`google`, `bing`, `yandex`, `apertium`, spell backends, and an `auto` router), aligned with translate-shell behavior where practical.
- Testable core: URL builders and response parsers covered by fixtures; default `go test` stays off the network.

## Non-goals

- **Azure Translator** and other Microsoft cloud APIs (explicit product choice).
- Guaranteed stability of scrape-style engines: upstream sites change without notice.

## Engine fragility

Undocumented HTTP endpoints can break, rate-limit, or conflict with provider terms of use. Prefer official APIs for production workloads when available. This project may add optional official API engines later without removing scrape-style backends.

## Upstream reference

Translator logic and HTTP contracts are traced to translate-shell AWK sources. The pinned commit used for parity work is recorded in [docs/UPSTREAM_TRANSLATE_SHELL.md](docs/UPSTREAM_TRANSLATE_SHELL.md).

## Development plan

Phased roadmap (Phase 0–7) lives in translate-shell as `docs/DEVELOPMENT_PLAN.md` in that checkout. If you clone only `gtr`, copy that file into `docs/DEVELOPMENT_PLAN.md` here so the tree stays self-contained.

**Current status:** Phase 0 is implemented in this repository; Phase 1 (Google + minimal translate path) is next.

## Per-phase testing and verification

Each roadmap phase is only **done** when it ships **automated tests** plus a **repeatable verify** step. Treat CI as the source of truth; local runs should mirror it before you merge or tag.

| Requirement | What to do |
|-------------|------------|
| **Automated tests** | New behavior gets table tests, golden fixtures under `testdata/` (parsers, URL builders), or focused unit tests. Default `go test ./...` must stay **off the live network** unless behind `-tags=integration` (see plan). |
| **Static checks** | `go vet ./...` and `go mod verify` (CI runs both). |
| **Concurrency** | On Linux, CI runs `go test -race ./...` for data races. |
| **Binary smoke** | After build: `gtr -V` and `gtr --help` must succeed (CI runs these). |

**One-shot local verify (Linux / macOS):**

```bash
./scripts/verify.sh
```

On **Windows** (PowerShell), run the equivalent:

```powershell
go mod tidy; go mod verify; go vet ./...; go test ./...; go build -o gtr.exe ./cmd/gtr; ./gtr.exe -V
```

When you finish a phase, confirm **GitHub Actions is green** on your branch and attach a short note in the PR or commit (what you tested manually, if anything—e.g. one live translation smoke for an engine).

**Git history:** land each completed phase as **one dedicated commit** on `main` (or your integration branch), using a message like `feat(phaseN): short summary` so history stays easy to bisect and review. WIP work can use multiple commits on a feature branch, then squash merge to one phase commit if you prefer a strictly linear log.

## Build

Requires Go 1.22 or newer.

```bash
go mod tidy
go build -o gtr ./cmd/gtr
./gtr --help
./gtr -V
```

Link a version string at build time:

```bash
go build -ldflags "-X main.version=0.1.0" -o gtr ./cmd/gtr
```

## Environment

| Variable       | Effect                                      |
|----------------|---------------------------------------------|
| `HTTP_PROXY`   | Standard Go proxy support for HTTP clients. |
| `HTTPS_PROXY`  | Same for HTTPS.                             |
| `NO_PROXY`     | Bypass list for proxies.                   |
| `USER_AGENT`   | Default `User-Agent` on outbound requests (same name as translate-shell). |

## License

SPDX: add a `LICENSE` file when you pick one for this project.
