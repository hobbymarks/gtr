package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ueki/gtr/internal/config"
	"github.com/ueki/gtr/internal/engine"
)

// Version is set by main via -ldflags for releases.
var Version = "dev"

// Execute runs the root command.
func Execute() error {
	return newRoot().Execute()
}

func newRoot() *cobra.Command {
	var engineName string
	var printVersion bool

	cmd := &cobra.Command{
		Use:   "gtr [text ...]",
		Short: "Multi-engine translation CLI (translate-shell-inspired)",
		Long: strings.TrimSpace(`
gtr is a Go rewrite-in-progress of the translate-shell idea: one CLI, multiple
translation backends. Remote engines rely on undocumented HTTP endpoints and
may break without notice; use responsibly and see the README for scope.`),
		SilenceUsage:     true,
		TraverseChildren: true,
		Args:             cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if printVersion {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), Version)
				return err
			}
			engineName = strings.TrimSpace(strings.ToLower(engineName))
			if engineName == "" {
				return errors.New("engine name must not be empty")
			}
			if len(args) == 0 {
				return cmd.Help()
			}
			if _, ok := engine.Lookup(engineName); !ok {
				names := engine.Names()
				if len(names) == 0 {
					return fmt.Errorf("unknown engine %q (no engines registered yet; see Phase 1)", engineName)
				}
				return fmt.Errorf("unknown engine %q (registered: %s)", engineName, strings.Join(names, ", "))
			}
			return errors.New("translation is not implemented yet (Phase 1)")
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().BoolVarP(&printVersion, "version", "V", false, "Print version and exit")
	cmd.Flags().StringVarP(&engineName, "engine", "e", config.DefaultEngine, "translation engine (temporary default: "+config.DefaultEngine+")")

	return cmd
}

// Main is a tiny entrypoint helper so cmd/gtr can stay minimal.
func Main() {
	if err := Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "gtr:", err)
		os.Exit(1)
	}
}
