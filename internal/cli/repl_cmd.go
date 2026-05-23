package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/hobbymarks/gtr/internal/config"
	"github.com/hobbymarks/gtr/internal/engine"
	"github.com/hobbymarks/gtr/internal/httpx"
	"github.com/hobbymarks/gtr/internal/lang"
)

func newReplCmd() *cobra.Command {
	var (
		engineName    string
		target        string
		source        string
		hostLang      string
		brief         bool
		noAutocorrect bool
		debug         bool
		dump          bool
		dictionary    bool
		timeoutSec    int
		noColor       bool
	)

	cmd := &cobra.Command{
		Use:   "repl",
		Short: "Start an interactive translation REPL",
		Long: strings.TrimSpace(`
Start an interactive Read-Eval-Print-Loop for translation.

Provides line editing, persistent history (~/.gtr_history), tab completion for
commands/engine names/language codes, and meta-commands (:engine, :target, etc.).
Type :help inside the REPL for full command list.`),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			engineName = strings.TrimSpace(strings.ToLower(engineName))
			if engineName == "" {
				return fmt.Errorf("engine name must not be empty")
			}

			target = strings.TrimSpace(target)
			if target == "" {
				target = config.DefaultTarget()
			}
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

			canon, factory, ok := engine.LookupFuzzy(engineName)
			if !ok {
				names := engine.Names()
				if len(names) == 0 {
					return fmt.Errorf("unknown engine %q (no engines registered)", engineName)
				}
				return fmt.Errorf("unknown engine %q (registered: %s)", engineName, strings.Join(names, ", "))
			}
			eng, err := factory()
			if err != nil {
				return fmt.Errorf("engine %q: %w", canon, err)
			}

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
		},
	}

	defEngine := config.DefaultEngine()
	cmd.Flags().StringVarP(&engineName, "engine", "e", defEngine, "translation engine")
	cmd.Flags().StringVarP(&target, "target", "t", "", "target language code")
	cmd.Flags().StringVarP(&source, "source", "s", "auto", "source language code")
	cmd.Flags().StringVar(&hostLang, "host-lang", "en", "host / UI language code")
	cmd.Flags().BoolVarP(&brief, "brief", "b", false, "Brief output (translation text only, trimmed)")
	cmd.Flags().BoolVar(&noAutocorrect, "no-autocorrect", false, "Disable autocorrect (Google: qc instead of qca)")
	cmd.Flags().BoolVar(&debug, "debug", false, "Log request URL to stderr")
	cmd.Flags().BoolVar(&dump, "dump", false, "Print raw HTTP response body instead of parsed translation")
	cmd.Flags().BoolVarP(&dictionary, "dictionary", "d", false, "Include dictionary / auxiliary JSON segments (Google)")
	cmd.Flags().IntVar(&timeoutSec, "timeout", 0, "HTTP request timeout in seconds (also GTR_TIMEOUT env)")
	cmd.Flags().BoolVar(&noColor, "no-color", false, "Disable ANSI color output")

	return cmd
}
