package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/peterh/liner"
	"github.com/spf13/cobra"
	"github.com/hobbymarks/gtr/internal/engine"
	"github.com/hobbymarks/gtr/internal/lang"
)

// shellStdinFn returns the reader for shell input. Tests may replace it.
var shellStdinFn = func() io.Reader { return os.Stdin }

// RunShell reads lines from stdin, translates each non-empty line, and prints results until EOF.
// Supports meta-commands prefixed with : (:engine, :target, :source, :host,
// :brief, :nobrief, :dict, :nodict, :speak, :nospeak, :dump, :nodump,
// :noautocorrect, :autocorrect, :debug, :nodebug, :info, :help).
// When stdin is a terminal, uses line editing with history (~/.gtr_history).
func RunShell(cmd *cobra.Command, eng engine.Engine, base engine.TranslateInput, engineName string) error {
	if stdinIsTTYFn() {
		return runShellLiner(cmd, eng, base, engineName)
	}
	return runShellScanner(cmd, eng, base, engineName)
}

func shellPrompt(engineName, target string) string {
	if target == "" {
		return fmt.Sprintf("[%s]> ", engineName)
	}
	return fmt.Sprintf("[%s:%s]> ", engineName, target)
}

func runShellScanner(cmd *cobra.Command, eng engine.Engine, base engine.TranslateInput, engineName string) error {
	sc := bufio.NewScanner(shellStdinFn())
	const max = 512 << 10
	sc.Buffer(make([]byte, 0, 64*1024), max)

	var speak bool

	fmt.Fprintln(cmd.OutOrStdout(), "Type :help for meta-commands, exit/quit to leave.")
	if base.Target == "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Set a target language with \":target <code>\" (e.g. \":target de\") or use -t when launching.\n")
	}

	for {
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
		if err := processLine(cmd, &eng, &base, &engineName, &speak, line); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "error: %v\n", err)
		}
	}
	if err := sc.Err(); err != nil {
		return fmt.Errorf("stdin: %w", err)
	}
	return nil
}

func runShellLiner(cmd *cobra.Command, eng engine.Engine, base engine.TranslateInput, engineName string) error {
	l := liner.NewLiner()
	defer l.Close()

	l.SetCtrlCAborts(true)

	histPath := historyPath()
	if f, err := os.Open(histPath); err == nil {
		l.ReadHistory(f)
		f.Close()
	}

	l.SetCompleter(func(line string) []string {
		return shellComplete(line, engineName)
	})

	var speak bool

	fmt.Fprintf(cmd.OutOrStdout(), "gtr shell — Type :help for commands, exit/quit to leave, Ctrl+C to cancel input, Ctrl+D to exit.\n")
	if base.Target == "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Set a target language with \":target <code>\" (e.g. \":target de\") or use -t when launching.\n")
	}

	for {
		prompt := shellPrompt(engineName, base.Target)
		line, err := l.Prompt(prompt)
		if err != nil {
			if err == liner.ErrPromptAborted {
				fmt.Fprintln(cmd.OutOrStdout(), "^C")
				continue
			}
			if err == io.EOF {
				fmt.Fprintln(cmd.OutOrStdout())
				break
			}
			return fmt.Errorf("readline: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.EqualFold(line, "exit") || strings.EqualFold(line, "quit") {
			break
		}

		l.AppendHistory(line)

		if err := processLine(cmd, &eng, &base, &engineName, &speak, line); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "error: %v\n", err)
		}
	}

	if err := saveHistory(l, histPath); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "warning: saving history: %v\n", err)
	}
	return nil
}

func processLine(cmd *cobra.Command, eng *engine.Engine, base *engine.TranslateInput, engName *string, speak *bool, line string) error {
	if strings.HasPrefix(line, ":") {
		return handleMetaCommand(cmd, eng, base, engName, speak, line)
	}
	in := *base
	in.Text = line
	out, err := (*eng).Translate(cmd.Context(), in)
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), out.Text)
	if out.Dictionary != "" {
		fmt.Fprintln(cmd.OutOrStdout(), "--")
		fmt.Fprintln(cmd.OutOrStdout(), out.Dictionary)
	}
	if *speak {
		u, werr := ttsURLForEngine(*eng, in, out.Text)
		if werr != nil {
			return fmt.Errorf("tts: %w", werr)
		}
		if werr := playGoogleTTS(cmd.Context(), u); werr != nil {
			return fmt.Errorf("play: %w", werr)
		}
	}
	return nil
}

func handleMetaCommand(cmd *cobra.Command, eng *engine.Engine, base *engine.TranslateInput, engName *string, speak *bool, line string) error {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}
	switch strings.ToLower(parts[0]) {
	case ":help":
		fmt.Fprintln(cmd.OutOrStdout(), `Commands:
  :engine <name>   switch translation engine
  :target <code>   set target language
  :source <code>   set source language (auto for detect)
  :host <code>     set host language
  :brief           enable brief output
  :nobrief         disable brief output
  :dict            enable dictionary payload
  :nodict          disable dictionary payload
  :speak           enable TTS after translation
  :nospeak         disable TTS after translation
  :dump            enable raw HTTP dump output
  :nodump          disable raw HTTP dump output
  :noautocorrect   disable autocorrect
  :autocorrect     enable autocorrect
  :debug           enable debug logging
  :nodebug         disable debug logging
  :info            show current settings
  exit / quit      leave shell
Ctrl+C            cancel current input line
Ctrl+D            exit shell`)
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
	case ":dict":
		base.Dictionary = true
		base.Dump = false
		fmt.Fprintln(cmd.OutOrStdout(), "dict: on")
	case ":nodict":
		base.Dictionary = false
		fmt.Fprintln(cmd.OutOrStdout(), "dict: off")
	case ":speak":
		*speak = true
		base.Dump = false
		fmt.Fprintln(cmd.OutOrStdout(), "speak: on")
	case ":nospeak":
		*speak = false
		fmt.Fprintln(cmd.OutOrStdout(), "speak: off")
	case ":dump":
		base.Dump = true
		base.Dictionary = false
		fmt.Fprintln(cmd.OutOrStdout(), "dump: on")
	case ":nodump":
		base.Dump = false
		fmt.Fprintln(cmd.OutOrStdout(), "dump: off")
	case ":noautocorrect":
		base.NoAutocorrect = true
		fmt.Fprintln(cmd.OutOrStdout(), "noautocorrect: on")
	case ":autocorrect":
		base.NoAutocorrect = false
		fmt.Fprintln(cmd.OutOrStdout(), "autocorrect: on (default)")
	case ":debug":
		base.Debug = true
		fmt.Fprintln(cmd.OutOrStdout(), "debug: on")
	case ":nodebug":
		base.Debug = false
		fmt.Fprintln(cmd.OutOrStdout(), "debug: off")
	case ":info":
		fmt.Fprintf(cmd.OutOrStdout(),
			"engine: %s  source: %s  target: %s  host: %s  brief: %v  dict: %v  speak: %v  dump: %v  noautocorrect: %v  debug: %v\n",
			*engName, base.Source, base.Target, base.HostLang, base.Brief,
			base.Dictionary, *speak, base.Dump, base.NoAutocorrect, base.Debug)
	default:
		return fmt.Errorf("unknown command %q (type :help)", parts[0])
	}
	return nil
}

func historyPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".gtr_history")
}

func saveHistory(l *liner.State, path string) error {
	if path == "" {
		return nil
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = l.WriteHistory(f)
	if closeErr := f.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	return err
}

func shellComplete(line string, currentEngine string) []string {
	var completions []string

	// Meta-command names (when line starts with : and no space yet)
	metaCommands := []string{
		":engine", ":target", ":source", ":host",
		":brief", ":nobrief", ":dict", ":nodict",
		":speak", ":nospeak", ":dump", ":nodump",
		":noautocorrect", ":autocorrect", ":debug", ":nodebug",
		":info", ":help",
	}

	if !strings.Contains(line, " ") {
		for _, mc := range metaCommands {
			if strings.HasPrefix(mc, line) {
				completions = append(completions, mc+" ")
			}
		}
		if strings.EqualFold(line, "exit") || strings.EqualFold(line, "quit") {
			completions = append(completions, "exit ", "quit ")
		}
		return completions
	}

	// After a meta-command, complete arguments
	parts := strings.Fields(line)
	if len(parts) >= 1 && strings.HasPrefix(parts[0], ":") {
		cmd := strings.ToLower(parts[0])
		switch cmd {
		case ":engine":
			prefix := ""
			if len(parts) > 1 {
				prefix = strings.ToLower(parts[1])
			}
			for _, name := range engine.Names() {
				if strings.HasPrefix(name, prefix) {
					completions = append(completions, parts[0]+" "+name+" ")
				}
			}
		case ":target", ":source", ":host":
			prefix := ""
			if len(parts) > 1 {
				prefix = strings.ToLower(parts[1])
			}
			for code := range lang.AllCodes() {
				if strings.HasPrefix(code, prefix) {
					completions = append(completions, parts[0]+" "+code+" ")
				}
			}
		}
	}

	return completions
}
