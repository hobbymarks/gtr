# gtr

Multi-engine translation CLI, inspired by [translate-shell](https://github.com/soimort/translate-shell).

## Install

### Go install (requires Go 1.22+)

```bash
go install github.com/hobbymarks/gtr/cmd/gtr@latest
```

### Pre-built binaries

Download the latest release for your platform from the [Releases page](https://github.com/hobbymarks/gtr/releases/latest).

### Homebrew (macOS / Linux)

```bash
brew tap hobbymarks/release https://github.com/hobbymarks/release
brew install gtr
```

### Scoop (Windows)

```powershell
scoop bucket add release https://github.com/hobbymarks/release
scoop install gtr
```

## Quick start

```bash
gtr -t de hello                       # translate to German (default engine: auto)
gtr :en 'bonjour'                     # auto -> en (SRC:TL token shorthand)
gtr 'auto:en+de' 'hello'              # multi-target parallel translation
gtr -b -t fr hello                    # brief output (translation text only)
gtr --host-lang ja -t de hello        # set host/UI language
gtr --no-autocorrect -t fr hello      # disable Google autocorrect
gtr -i in.txt -o out.txt -t fr        # file I/O
gtr -j -t ja a b c                    # joined args as input text (never stdin)
gtr --identify 'hola'                 # detect language code
gtr --dump -t de 'test'               # raw HTTP response body (debug)
gtr -t fr --json hello                # structured JSON output
gtr -v -t de hello                    # verbose (print engine, source, target)
gtr --timeout 10 -t de hello          # custom HTTP timeout (seconds)
gtr --no-color -t fr hello            # disable ANSI colors
gtr --speak -t de hello               # translate and play TTS (needs mpv/ffplay)
gtr -B -t fr "hello"                  # open in browser instead of translating
```

`--identify` requires `google`, `bing`, `yandex`, or `auto`. `apertium` and spell engines do not support identification.

## Subcommands

### repl

Interactive REPL with line editing, persistent history (`~/.gtr_history`), and tab completion for commands, engine names, and language codes.

```bash
gtr repl -e google -t fr             # start interactive session
```

Inside the REPL, type `:help` for all meta-commands:

| Command | Effect |
|---------|--------|
| `:engine <name>` | Switch translation engine |
| `:target <code>` | Set target language |
| `:source <code>` | Set source language (`auto` for detect) |
| `:host <code>` | Set host/UI language |
| `:browser` | Open last translation in browser |
| `:narrator <code>` | Set TTS voice language |
| `:brief` / `:nobrief` | Toggle brief output |
| `:dict` / `:nodict` | Toggle dictionary payload |
| `:speak` / `:nospeak` | Toggle TTS after translation |
| `:dump` / `:nodump` | Toggle raw HTTP dump |
| `:autocorrect` / `:noautocorrect` | Toggle autocorrect (Google) |
| `:debug` / `:nodebug` | Toggle debug logging |
| `:info` | Show current settings |
| `:help` | Show help text |

### config

Manage settings stored in `~/.gtrrc` (simple `KEY=VALUE` format).

```bash
gtr config                           # show all settings
gtr config set GTR_DEFAULT_TARGET de # set default target language
gtr config get GTR_DEFAULT_TARGET    # get a value
gtr config unset GTR_TIMEOUT         # remove a value
gtr config path                      # show config file path
```

Supported keys: `GTR_DEFAULT_ENGINE`, `GTR_DEFAULT_TARGET`, `GTR_TIMEOUT`.

### update

Self-update to the latest GitHub release.

```bash
gtr update                           # update to latest
gtr update --dry-run                 # check without installing
```

On Linux/macOS the binary is replaced in-place. On Windows the new binary is written as `gtr.exe.new` and must be moved manually.

## Engines

| Engine | Description | TTS | Dictionary | Identify |
|--------|-------------|-----|------------|----------|
| `auto` | Default. Routes to Google or Bing based on language support. | delegated | delegated | delegated |
| `google` | Google Translate public endpoint. | yes | yes | yes |
| `bing` | Bing Web Translator. Caches setup tokens. | yes | yes | yes |
| `yandex` | Yandex Translate mobile API. | no | no | yes |
| `apertium` | Apertium APy translate. Source `auto` falls back to `en`. | no | no | no |
| `spell` | Auto-selects `aspell` or `hunspell` (requires local binary). | no | no | no |
| `aspell` | GNU Aspell spell checker. | no | no | no |
| `hunspell` | Hunspell spell checker. | no | no | no |

Engine names are case-insensitive with fuzzy prefix matching (e.g. `-e goo` matches `google`, `-e ap` matches `aspell` before `apertium`). Prefer full names when in doubt.

```bash
gtr --list-engines                   # table of all engines
gtr --list-languages                 # language codes with engine coverage
gtr --list-codes                     # plain list of language codes
gtr -L en                            # language details (name, family, script, ISO)
gtr -L ja+zh-CN                      # multiple codes at once
gtr -e google -t de 'Wanderlust'     # translation
gtr -e google -d -t de 'Wanderlust'  # with dictionary segments
gtr -e google -t zh-CN hello         # pinyin shown when available
gtr -e google --speak -t de hello    # translate and play TTS (needs mpv/ffplay)
gtr -e google --speak -t de --download-audio /tmp/out.mp3 hello  # save audio
gtr -e google --speak -n ja -t fr hello  # TTS with Japanese voice
gtr -B -t fr "hello"                 # open in browser instead of translating
gtr -e apertium -s en -t es "hello"  # Apertium (only valid pairs return text)
gtr -e spell -s en 'some text'       # spell check in source language
```

Engine language support metadata is embedded from translate-shell `LanguageData.awk`. Regenerate:

```bash
python3 scripts/gen_language_support.py /path/to/translate-shell/include/LanguageData.awk
```

## Configuration

### Environment variables

| Variable | Effect |
|----------|--------|
| `GTR_DEFAULT_ENGINE` | Default translation engine (also settable in `~/.gtrrc`) |
| `GTR_DEFAULT_TARGET` | Default target language when `-t` is omitted |
| `GTR_TIMEOUT` | HTTP request timeout in seconds (default 30; overridden by `--timeout`) |
| `NO_COLOR` | Disable ANSI color output (same as `--no-color`) |
| `HTTP_PROXY` / `HTTPS_PROXY` / `NO_PROXY` | Standard Go proxy support |
| `USER_AGENT` | Custom `User-Agent` header on outbound requests |

### Config file (`~/.gtrrc`)

```ini
GTR_DEFAULT_ENGINE=google
GTR_DEFAULT_TARGET=de
GTR_TIMEOUT=15
```

Environment variables take precedence over config file values. Use `gtr config` to manage settings.

## Build from source

Requires Go 1.22+.

```bash
go build -o gtr ./cmd/gtr
```

With version embedding:

```bash
go build -ldflags "-X main.version=0.1.0 -X main.commit=$(git rev-parse --short HEAD)" -o gtr ./cmd/gtr
```

## Development

Run tests:

```bash
go test ./...                        # unit/parser tests (no network)
go test -race ./...                  # with race detector
go test -tags=integration -count=1 ./internal/cli/   # integration tests (live HTTP)
./scripts/verify.sh                  # full verify
```

Static checks: `go vet ./...`, `go mod verify`, and golangci-lint (see [.golangci.yml](.golangci.yml)).

## License

[MIT](LICENSE)
