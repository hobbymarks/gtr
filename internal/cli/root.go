// Package cli implements the cobra command tree, I/O paths, shell REPL,
// TTS playback, pager integration, and color output for gtr.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/hobbymarks/gtr/internal/config"
	"github.com/hobbymarks/gtr/internal/engine"
	"github.com/hobbymarks/gtr/internal/httpx"
	"github.com/hobbymarks/gtr/internal/lang"
	"golang.org/x/term"
)

// Version is set by main via -ldflags for releases.
var Version = "dev"

// Execute runs the root command.
func Execute() error {
	return newRoot().ExecuteContext(context.Background())
}

// stdinIsTTYFn reports whether os.Stdin is a character device. Tests may replace it.
var stdinIsTTYFn = func() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func newRoot() *cobra.Command {
	var (
		engineName    string
		printVersion  bool
		listEngines   bool
		target        string
		source        string
		hostLang      string
		brief         bool
		noAutocorrect bool
		debug         bool
		dump          bool
		identify      bool
		joinArgv      bool
		inputPath     string
		outputPath    string
		dictionary    bool
		speak         bool
		view          bool
		shell         bool
		timeoutSec    int
		jsonOut       bool
		noColor       bool
	)

	cmd := &cobra.Command{
		Use:   "gtr [SRC:TL] [text ...]",
		Short: "Multi-engine translation CLI (translate-shell-inspired)",
		Long: strings.TrimSpace(`
gtr is a Go rewrite-in-progress of the translate-shell idea: one CLI, multiple
translation backends. Remote engines rely on undocumented HTTP endpoints and
may break without notice; use responsibly and see the README for scope.

Provide text as arguments, or pipe stdin when there are no arguments. Target
language (-t / --target) is required for translation unless you use an optional
leading SRC:TL token (e.g. :en or ja:en) without setting -s/-t.

Phase 4 I/O: -i / -o (paths or file:// URLs), -j to force argv as input,
--identify for language detection, --dump for raw HTTP response bodies.
Phase 5+: -d dictionary payload (Google), spell engines; --speak / -play (Google TTS);
--view (pager); --shell (line REPL).`),
		SilenceUsage:     true,
		TraverseChildren: true,
		Args:             cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if printVersion {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), Version)
				return err
			}
			if listEngines {
				w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
				if _, err := fmt.Fprintln(w, "ENGINE\tTTS\tDICT"); err != nil {
					return err
				}
				for _, n := range engine.Names() {
					c := engine.CapabilitiesOf(n)
					if _, err := fmt.Fprintf(w, "%s\t%v\t%v\n", n, c.SupportsTTS, c.SupportsDictionary); err != nil {
						return err
					}
				}
				return w.Flush()
			}

			engineName = strings.TrimSpace(strings.ToLower(engineName))
			if engineName == "" {
				return errors.New("engine name must not be empty")
			}

			initColorOut(noColor)

			if timeoutSec == 0 {
				if s := strings.TrimSpace(config.EnvOverride("GTR_TIMEOUT")); s != "" {
					if n, err := strconv.Atoi(s); err == nil && n > 0 {
						timeoutSec = n
					}
				}
			}
			if timeoutSec > 0 {
				httpx.SharedClientTimeout = time.Duration(timeoutSec) * time.Second
			}

			target = strings.TrimSpace(target)
			source = strings.TrimSpace(source)
			if source == "" {
				source = "auto"
			}
			hostLang = strings.TrimSpace(hostLang)
			if hostLang == "" {
				hostLang = "en"
			}

			if !lang.IsKnownLanguage(source) {
				return fmt.Errorf("unknown source language code %q", source)
			}
			if !lang.IsKnownLanguage(hostLang) {
				return fmt.Errorf("unknown host language code %q", hostLang)
			}

			sourceChanged := cmd.Flags().Lookup("source").Changed
			targetChanged := cmd.Flags().Lookup("target").Changed

			stdinTTY := stdinIsTTYFn()

			var extraTargets []string
			args, pairSrc, pairTgts, stripped := stripLeadingLangSpec(args, sourceChanged, targetChanged)
			if stripped {
				if !sourceChanged {
					source = pairSrc
				}
				if !targetChanged {
					target = pairTgts[0]
					extraTargets = append(extraTargets, pairTgts[1:]...)
				}
			}

			if identify && dump {
				return errors.New("cannot combine --identify and --dump")
			}
			if shell && identify {
				return errors.New("cannot combine --shell and --identify")
			}
			if shell && view {
				return errors.New("cannot combine --shell and --view")
			}
			if shell && speak {
				return errors.New("cannot combine --shell and --speak / --play")
			}
			if view && strings.TrimSpace(outputPath) != "" {
				return errors.New("cannot combine --view and -o")
			}
			if dictionary && dump {
				return errors.New("cannot combine --dictionary and --dump")
			}
			if speak && (identify || dump) {
				return errors.New("cannot combine --speak / --play with --identify or --dump")
			}
			if joinArgv && strings.TrimSpace(inputPath) != "" {
				return errors.New("cannot combine -j and -i")
			}
			if strings.TrimSpace(inputPath) != "" && len(args) > 0 {
				return errors.New("cannot combine -i and positional text arguments")
			}
			if joinArgv && len(args) == 0 {
				return errors.New("-j requires at least one text argument")
			}

			canon, factory, ok := engine.LookupFuzzy(engineName)
			if !ok {
				names := engine.Names()
				if len(names) == 0 {
					return fmt.Errorf("unknown engine %q (no engines registered)", engineName)
				}
				return fmt.Errorf("unknown engine %q (registered: %s)", engineName, strings.Join(names, ", "))
			}
			eng, engErr := factory()
			if engErr != nil {
				return fmt.Errorf("engine %q: %w", canon, engErr)
			}

			if !shell && target == "" && !targetChanged {
				target = config.DefaultTarget()
			}

			if isSpellEngine(canon) && !identify && !shell && strings.TrimSpace(target) == "" {
				target = strings.TrimSpace(source)
			}

			if !identify && !shell && target == "" && len(args) == 0 && stdinTTY && !joinArgv && strings.TrimSpace(inputPath) == "" {
				return cmd.Help()
			}
			if !identify && !shell && target == "" {
				return errors.New("target language is required (-t / --target, or a leading SRC:TL token)")
			}

			var text string
			var err error
			if !shell {
				switch {
				case joinArgv:
					text = strings.Join(args, " ")
				case strings.TrimSpace(inputPath) != "":
					text, err = readTextFile(inputPath)
					if err != nil {
						return err
					}
				default:
					text, err = textFromArgsOrStdin(args, os.Stdin, stdinTTY)
					if err != nil {
						return err
					}
				}
			}

			if dictionary && !identify && !engine.CapabilitiesOf(canon).SupportsDictionary {
				return fmt.Errorf("engine %q does not support dictionary mode (-d)", canon)
			}
			if speak && !identify {
				switch canon {
				case "google", "auto", "bing":
				default:
					return fmt.Errorf("engine %q does not support --speak / --play (only google, bing, and auto)", canon)
				}
			}

			out := cmd.OutOrStdout()
			var closeOut func()
			if op := strings.TrimSpace(outputPath); op != "" {
				p := stripFileURLPrefix(op)
				f, err := os.Create(p)
				if err != nil {
					return fmt.Errorf("create output file: %w", err)
				}
				out = f
				closeOut = func() { _ = f.Close() }
			}
			if closeOut != nil {
				defer closeOut()
			}
			if view {
				pw, cleanup, err := openPagerWriter(cmd.Context())
				if err != nil {
					return err
				}
				out = pw
				defer cleanup()
			}

			if identify {
				li, ok := eng.(engine.LanguageIdentifier)
				if !ok {
					return fmt.Errorf("engine %q does not support language identification", canon)
				}
				lang, err := li.IdentifyLanguage(cmd.Context(), text, hostLang)
				if err != nil {
					return err
				}
				_, err = fmt.Fprintln(out, lang)
				return err
			}

			if shell {
				base := engine.TranslateInput{
					Source:        source,
					Target:        target,
					HostLang:      hostLang,
					Brief:         brief,
					NoAutocorrect: noAutocorrect,
					Debug:         debug,
					Dump:          dump,
					Dictionary:    dictionary,
				}
				return RunShell(cmd, eng, base, canon)
			}

			allTargets := append([]string{target}, extraTargets...)
			for _, tl := range allTargets {
				if !lang.IsKnownLanguage(tl) {
					return fmt.Errorf("unknown target language code %q", tl)
				}
			}
			type jsonSingle struct {
				Source     string `json:"source"`
				Target     string `json:"target"`
				Text       string `json:"text"`
				Phonetic   string `json:"phonetic,omitempty"`
				Dictionary string `json:"dictionary,omitempty"`
			}
			type result struct {
				idx int
				out engine.TranslateOutput
				err error
			}
			ch := make(chan result, len(allTargets))
			for i, tl := range allTargets {
				go func(idx int, tl string) {
					outi, err := eng.Translate(cmd.Context(), engine.TranslateInput{
						Text:          text,
						Source:        source,
						Target:        tl,
						HostLang:      hostLang,
						Brief:         brief,
						NoAutocorrect: noAutocorrect,
						Debug:         debug,
						Dump:          dump,
						Dictionary:    dictionary,
					})
					ch <- result{idx: idx, out: outi, err: err}
				}(i, tl)
			}
			results := make([]result, len(allTargets))
			for i := 0; i < len(allTargets); i++ {
				r := <-ch
				if r.err != nil {
					return r.err
				}
				results[r.idx] = r
			}
			var jsonResults []jsonSingle
			for i, r := range results {
				tl := allTargets[i]
				outi := r.out
				if jsonOut {
					js := jsonSingle{
						Source:   source,
						Target:   tl,
						Text:     outi.Text,
						Phonetic: outi.Phonetic,
					}
					if outi.Dictionary != "" {
						js.Dictionary = outi.Dictionary
					}
					jsonResults = append(jsonResults, js)
				}
				inForTTS := engine.TranslateInput{
					Text: text, Source: source, Target: tl, HostLang: hostLang,
				}
				if jsonOut {
					continue
				}
				if len(allTargets) == 1 {
					if _, werr := fmt.Fprintf(out, "%s\n", Green(outi.Text)); werr != nil {
						return werr
					}
				} else {
					if i > 0 {
						if _, werr := fmt.Fprintln(out); werr != nil {
							return werr
						}
					}
					if _, werr := fmt.Fprintf(out, "[%s]\n%s", Cyan(tl), Green(outi.Text)); werr != nil {
						return werr
					}
				}
				if outi.Dictionary != "" {
					if _, werr := fmt.Fprintf(out, "\n%s\n%s\n", Yellow("--"), outi.Dictionary); werr != nil {
						return werr
					}
				}
				if outi.Phonetic != "" {
					if _, werr := fmt.Fprintf(out, "(%s)\n", Cyan(outi.Phonetic)); werr != nil {
						return werr
					}
				}
				if speak {
					u, werr := ttsURLForEngine(eng, inForTTS, outi.Text)
					if werr != nil {
						return werr
					}
					if werr := playGoogleTTS(cmd.Context(), u); werr != nil {
						return werr
					}
				}
			}
			if jsonOut {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(jsonResults)
			}
			if len(allTargets) > 1 {
				_, _ = fmt.Fprintln(out)
			}
			return nil
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().BoolVarP(&printVersion, "version", "V", false, "Print version and exit")
	cmd.Flags().BoolVar(&listEngines, "list-engines", false, "Print registered engine names and exit")
	defEngine := config.DefaultEngine()
	cmd.Flags().StringVarP(&engineName, "engine", "e", defEngine, "translation engine (default "+defEngine+")")
	cmd.Flags().StringVarP(&target, "target", "t", "", "target language code (required)")
	cmd.Flags().StringVarP(&source, "source", "s", "auto", "source language code (default auto)")
	cmd.Flags().StringVar(&hostLang, "host-lang", "en", "host / UI language code sent to the engine (default en)")
	cmd.Flags().BoolVarP(&brief, "brief", "b", false, "Brief output (translation text only, trimmed)")
	cmd.Flags().BoolVar(&noAutocorrect, "no-autocorrect", false, "Disable autocorrect (Google: qc instead of qca)")
	cmd.Flags().BoolVar(&debug, "debug", false, "Log request URL to stderr (no credentials; includes query text)")
	cmd.Flags().BoolVar(&dump, "dump", false, "Print raw HTTP response body instead of parsed translation")
	cmd.Flags().BoolVar(&identify, "identify", false, "Detect source language and print its code (no translation)")
	cmd.Flags().BoolVarP(&joinArgv, "join", "j", false, "Use joined arguments as input text (never read stdin)")
	cmd.Flags().StringVarP(&inputPath, "input", "i", "", "Read input text from this file path or file:// URL")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Write output to this file path or file:// URL (truncates)")
	cmd.Flags().BoolVarP(&dictionary, "dictionary", "d", false, "Include dictionary / auxiliary JSON segments when the engine supports it (Google)")
	cmd.Flags().BoolVar(&speak, "speak", false, "After translation, play Google TTS for the translated text (requires local player: mpv, ffplay, …)")
	cmd.Flags().BoolVar(&speak, "play", false, "Same as --speak: play translated text via Google TTS (translate-shell-style)")
	cmd.Flags().BoolVar(&view, "view", false, "Send output through $PAGER (default less -R, or more on Windows)")
	cmd.Flags().BoolVar(&shell, "shell", false, "Interactive line-at-a-time translation on stdin (exit/quit to leave)")
	cmd.Flags().IntVar(&timeoutSec, "timeout", 0, "HTTP request timeout in seconds (default 30; also GTR_TIMEOUT env)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Output structured JSON instead of plain text")
	cmd.Flags().BoolVar(&noColor, "no-color", false, "Disable ANSI color output")

	return cmd
}

func isSpellEngine(name string) bool {
	switch name {
	case "spell", "aspell", "hunspell":
		return true
	default:
		return false
	}
}

// Main is a tiny entrypoint helper so cmd/gtr can stay minimal.
func Main() {
	if err := Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "gtr:", err)
		os.Exit(1)
	}
}
