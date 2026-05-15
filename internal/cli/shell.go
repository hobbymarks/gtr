package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/hobbymarks/gtr/internal/engine"
)

// shellStdinFn returns the reader for shell input. Tests may replace it.
var shellStdinFn = func() io.Reader { return os.Stdin }

// RunShell reads lines from stdin, translates each non-empty line, and prints results until EOF.
// Supports meta-commands prefixed with : (:engine, :target, :source, :host, :brief, :nobrief, :help).
func RunShell(cmd *cobra.Command, eng engine.Engine, base engine.TranslateInput, engineName string) error {
	sc := bufio.NewScanner(shellStdinFn())
	const max = 512 << 10
	sc.Buffer(make([]byte, 0, 64*1024), max)

	fmt.Fprintln(cmd.OutOrStdout(), "Type :help for meta-commands, exit/quit to leave.")

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
		if strings.HasPrefix(line, ":") {
			if err := handleMetaCommand(cmd, &eng, &base, &engineName, line); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error: %v\n", err)
			}
			continue
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

func handleMetaCommand(cmd *cobra.Command, eng *engine.Engine, base *engine.TranslateInput, engName *string, line string) error {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}
	switch strings.ToLower(parts[0]) {
	case ":help":
		fmt.Fprintln(cmd.OutOrStdout(), `Meta-commands:
  :engine <name>   switch translation engine
  :target <code>   set target language
  :source <code>   set source language (auto for detect)
  :host <code>     set host language
  :brief           enable brief output
  :nobrief         disable brief output
  :info            show current settings
  exit / quit      leave`)
	case ":engine":
		if len(parts) < 2 {
			return fmt.Errorf("usage: :engine <name>")
		}
		name := strings.TrimSpace(strings.ToLower(parts[1]))
		_, factory, ok := engine.LookupFuzzy(name)
		if !ok {
			return fmt.Errorf("unknown engine %q", name)
		}
		newEng, err := factory()
		if err != nil {
			return fmt.Errorf("engine %q: %w", name, err)
		}
		*eng = newEng
		*engName = name
		fmt.Fprintf(cmd.OutOrStdout(), "engine: %s\n", (*eng).Name())
	case ":target":
		if len(parts) < 2 {
			return fmt.Errorf("usage: :target <code>")
		}
		base.Target = strings.TrimSpace(parts[1])
		fmt.Fprintf(cmd.OutOrStdout(), "target: %s\n", base.Target)
	case ":source":
		if len(parts) < 2 {
			return fmt.Errorf("usage: :source <code>")
		}
		base.Source = strings.TrimSpace(parts[1])
		fmt.Fprintf(cmd.OutOrStdout(), "source: %s\n", base.Source)
	case ":host":
		if len(parts) < 2 {
			return fmt.Errorf("usage: :host <code>")
		}
		base.HostLang = strings.TrimSpace(parts[1])
		fmt.Fprintf(cmd.OutOrStdout(), "host: %s\n", base.HostLang)
	case ":brief":
		base.Brief = true
		fmt.Fprintln(cmd.OutOrStdout(), "brief: on")
	case ":nobrief":
		base.Brief = false
		fmt.Fprintln(cmd.OutOrStdout(), "brief: off")
	case ":info":
		fmt.Fprintf(cmd.OutOrStdout(), "engine: %s  source: %s  target: %s  host: %s  brief: %v\n",
			*engName, base.Source, base.Target, base.HostLang, base.Brief)
	default:
		return fmt.Errorf("unknown command %q (type :help)", parts[0])
	}
	return nil
}
