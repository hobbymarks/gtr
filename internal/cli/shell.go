package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ueki/gtr/internal/engine"
)

// RunShell reads lines from stdin, translates each non-empty line, and prints results until EOF.
func RunShell(cmd *cobra.Command, eng engine.Engine, base engine.TranslateInput) error {
	sc := bufio.NewScanner(os.Stdin)
	const max = 512 << 10
	sc.Buffer(make([]byte, 0, 64*1024), max)
	for {
		fmt.Fprint(cmd.OutOrStdout(), "gtr> ")
		if !sc.Scan() {
			break
		}
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if strings.EqualFold(line, "exit") || strings.EqualFold(line, "quit") {
			break
		}
		in := base
		in.Text = line
		out, err := eng.Translate(cmd.Context(), in)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "error: %v\n", err)
			continue
		}
		fmt.Fprintln(cmd.OutOrStdout(), out.Text)
		if out.Dictionary != "" {
			fmt.Fprintln(cmd.OutOrStdout(), "--")
			fmt.Fprintln(cmd.OutOrStdout(), out.Dictionary)
		}
	}
	if err := sc.Err(); err != nil {
		return fmt.Errorf("stdin: %w", err)
	}
	return nil
}
