# gtr

Go multi-engine translation CLI, inspired by [translate-shell](https://github.com/soimort/translate-shell).

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

Phased roadmap (Phase 0--7) is tracked in [docs/DEVELOPMENT_PLAN.md](docs/DEVELOPMENT_PLAN.md) (synced from translate-shell), including **strengths, gaps, security notes, risks, and enhancements**.

**Current status:** Phases **0--7** from [docs/DEVELOPMENT_PLAN.md](docs/DEVELOPMENT_PLAN.md) are implemented, plus six enhancement phases (A--F) covering bug fixes, code consolidation, test coverage, defense/resilience, UX features, and performance.

## Quick start

```bash
./gtr -t de hello                     # translate to German (default engine: auto)
./gtr :en 'bonjour'                   # auto -> en (no -t/-s when using SRC:TL token)
./gtr 'auto:en+de' 'hello'            # translate to en then de (multi-target, parallel)
./gtr -i ./in.txt -o ./out.txt -t fr  # file I/O (paths or file:// URLs)
./gtr -j -t ja a b c                  # input text "a b c" (never stdin)
./gtr --identify 'hola'               # print detected language code
./gtr --dump -t de 'test'             # raw HTTP body (debug; engine-specific)
```

`apertium` does not implement language identification; use `google`, `bing`, `yandex`, or `auto` for **`--identify`**.

### New features (beyond upstream plan)

```bash
./gtr -t fr --json hello              # structured JSON output
./gtr --timeout 10 -t de hello        # custom HTTP timeout (also GTR_TIMEOUT env)
./gtr --no-color -t fr hello          # disable ANSI color output (also NO_COLOR env)
./gtr --shell -e google -t fr         # line REPL with :engine, :target, :source etc.
```

### Shell meta-commands

Inside `--shell` mode, type `:help` for a list. Supported commands:

```
:engine google     switch translation engine
:target de         set target language
:source en         set source language (auto for detect)
:host en           set host/UI language
:brief / :nobrief  toggle brief output
:info              show current settings
exit / quit        leave REPL
```

### Engine name matching (`-e` / `--engine`)

Names are **case-insensitive**. If the name is not exact, **fuzzy match** picks the **shortest registered engine name** that has your input as a **prefix** (ties broken lexicographically). For example, **`-e ap`** can match **`aspell`** before **`apertium`**. Prefer **full names** when in doubt (`-e apertium`, `-e aspell`).

### Pager (`--view` and `$PAGER`)

The pager command is built by **splitting `$PAGER` on spaces** (no shell-style quoting). If the pager binary lives under a path **with spaces**, put a wrapper script on `PATH` or point `PAGER` at a single-token executable name.

## Engines

| Engine | Role | TTS | Dictionary payload |
|--------|------|-----|----------------------|
| `auto` | **Default.** Picks `google` or `bing` from translate-shell language tables; else Google. | yes* | yes* |
| `google` | `translate.googleapis.com` `translate_a/single`. | yes | yes |
| `bing` | Bing Web Translator (`/translator` + `ttranslatev3`). Setup tokens cached 5 min. | no | yes |
| `yandex` | `translate.yandex.net` `api/v1/tr.json/translate` (mobile-style; `ucid` per process). | no | no (upstream path disabled in translate-shell) |
| `apertium` | `www.apertium.org/apy/translate` GET; `auto` source -> `en` like translate-shell. | no | no |
| `spell` / `aspell` / `hunspell` | Local ispell-protocol checkers (requires binaries on `PATH`). Lazily resolved on first use. | no | no |

\*`auto` delegates capabilities to the chosen backend; **`--speak`** / **`-play`** is implemented only when routing matches Google TTS.

### Engine features

```bash
./gtr --list-engines              # table: ENGINE / TTS / DICT
./gtr -t de hello                 # default -e auto -> google or bing by pair
./gtr -e yandex -t ru "hello"     # may fail if API changes
./gtr -e apertium -s en -t es "hello"   # only valid Apertium pairs return text
./gtr -e goo -t fr hi             # fuzzy prefix -> google
./gtr -e spell -s en 'some text'          # aspell or hunspell (target defaults to source)
./gtr -e google -d -t de 'Wanderlust'     # translation + dictionary JSON segments when present
./gtr -e google --speak -t de 'hello'     # translate then play TTS (mpv / ffplay / afplay)
./gtr -e google -play -t de 'hello'      # same as --speak (translate-shell-style flag)
./gtr --shell -e auto -t fr               # line REPL; :help for commands, exit/quit to stop
./scripts/check_upstream.sh               # reminder to diff against pinned translate-shell
```

Language support metadata is embedded from translate-shell `LanguageData.awk`. Regenerate after updating the upstream pin:

```bash
# go:generate directive is in internal/lang/lang.go
python3 scripts/gen_language_support.py /path/to/translate-shell/include/LanguageData.awk
```

## Testing

21 test files covering CLI arguments, input parsing, language spec parsing, all engine parsers (golden fixtures), auto routing, registry, fuzzy lookup, TTS URL building, Bing setup scraping, shell REPL, Yandex UCID, Google identify HTTP flow, pager, and audio player.

| Type | Command |
|------|---------|
| Unit / parser tests (no network) | `go test ./...` |
| Race detector | `go test -race ./...` |
| Integration (live HTTP) | `go test -tags=integration -count=1 ./internal/cli/` |
| Full verify | `./scripts/verify.sh` |

### Static checks

`go vet ./...`, `go mod verify`, and **golangci-lint** in CI (see [.golangci.yml](.golangci.yml); locally: `just lint` or `scripts/verify.sh` if `golangci-lint` is on `PATH`).

## Build

Requires Go 1.22 or newer.

```bash
go mod tidy
go build -o gtr ./cmd/gtr
./gtr --help
./gtr -V
./gtr -e google -t fr hello
echo hello | ./gtr -t fr -b
./gtr -e google -t fr+de --json hello    # multi-target JSON output
```

Link a version string at build time:

```bash
go build -ldflags "-X main.version=0.1.0" -o gtr ./cmd/gtr
```

## Environment

| Variable | Effect |
|----------|--------|
| `HTTP_PROXY` | Standard Go proxy support for HTTP clients. |
| `HTTPS_PROXY` | Same for HTTPS. |
| `NO_PROXY` | Bypass list for proxies. |
| `USER_AGENT` | Default `User-Agent` on outbound requests (same name as translate-shell). |
| `PAGER` | Used by **`--view`** (default `less -R`, or `more` on Windows). |
| `GTR_TIMEOUT` | HTTP request timeout in seconds (default 30; overridden by `--timeout`). |
| `GTR_DEFAULT_TARGET` | Default target language when `-t` is omitted (also settable in `~/.gtrrc`). |
| `NO_COLOR` | Disable ANSI color output (same as `--no-color`). |

### Config file (`~/.gtrrc`)

Simple `KEY=VALUE` format. Supported keys:

```ini
# ~/.gtrrc
GTR_DEFAULT_TARGET=de
GTR_TIMEOUT=15
```

Environment variables take precedence over config file values.

## Releases

- Version is injected with **`-ldflags "-X main.version=VERSION"`** (see **Build** above).
- [GoReleaser](https://goreleaser.com/) config: [.goreleaser.yaml](.goreleaser.yaml). Example: tag `v0.1.0`, set `GITHUB_TOKEN`, then run **`goreleaser release`** so **GitHub Releases** publishes archives for that tag (first-time setup: confirm repo permissions and `.goreleaser.yaml` targets).

## License

[LICENSE](LICENSE) (MIT).
