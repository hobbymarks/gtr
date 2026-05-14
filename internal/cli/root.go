package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ueki/gtr/internal/config"
	"github.com/ueki/gtr/internal/engine"
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
	)

	cmd := &cobra.Command{
		Use:   "gtr [text ...]",
		Short: "Multi-engine translation CLI (translate-shell-inspired)",
		Long: strings.TrimSpace(`
gtr is a Go rewrite-in-progress of the translate-shell idea: one CLI, multiple
translation backends. Remote engines rely on undocumented HTTP endpoints and
may break without notice; use responsibly and see the README for scope.

Provide text as arguments, or pipe stdin when there are no arguments. Target
language (-t / --target) is required for translation.`),
		SilenceUsage:     true,
		TraverseChildren: true,
		Args:             cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if printVersion {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), Version)
				return err
			}
			if listEngines {
				for _, n := range engine.Names() {
					if _, err := fmt.Fprintln(cmd.OutOrStdout(), n); err != nil {
						return err
					}
				}
				return nil
			}

			engineName = strings.TrimSpace(strings.ToLower(engineName))
			if engineName == "" {
				return errors.New("engine name must not be empty")
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

			stdinTTY := stdinIsTTYFn()
			if target == "" && len(args) == 0 && stdinTTY {
				return cmd.Help()
			}
			if target == "" {
				return errors.New("target language is required (-t / --target)")
			}

			text, err := textFromArgsOrStdin(args, os.Stdin, stdinTTY)
			if err != nil {
				return err
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

			out, err := eng.Translate(cmd.Context(), engine.TranslateInput{
				Text:          text,
				Source:        source,
				Target:        target,
				HostLang:      hostLang,
				Brief:         brief,
				NoAutocorrect: noAutocorrect,
				Debug:         debug,
			})
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", out.Text)
			return err
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().BoolVarP(&printVersion, "version", "V", false, "Print version and exit")
	cmd.Flags().BoolVar(&listEngines, "list-engines", false, "Print registered engine names and exit")
	cmd.Flags().StringVarP(&engineName, "engine", "e", config.DefaultEngine, "translation engine (default "+config.DefaultEngine+")")
	cmd.Flags().StringVarP(&target, "target", "t", "", "target language code (required)")
	cmd.Flags().StringVarP(&source, "source", "s", "auto", "source language code (default auto)")
	cmd.Flags().StringVar(&hostLang, "host-lang", "en", "host / UI language code sent to the engine (default en)")
	cmd.Flags().BoolVarP(&brief, "brief", "b", false, "Brief output (translation text only, trimmed)")
	cmd.Flags().BoolVar(&noAutocorrect, "no-autocorrect", false, "Disable autocorrect (Google: qc instead of qca)")
	cmd.Flags().BoolVar(&debug, "debug", false, "Log request URL to stderr (no credentials; includes query text)")

	return cmd
}

// Main is a tiny entrypoint helper so cmd/gtr can stay minimal.
func Main() {
	if err := Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "gtr:", err)
		os.Exit(1)
	}
}
