package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

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
	var lastText string
	var narrator string

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
		if err := processLine(cmd, &eng, &base, &engineName, &speak, &lastText, &narrator, line); err != nil {
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
	var lastText string
	var narrator string

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)
	defer signal.Stop(sigCh)

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

		select {
		case <-sigCh:
			_ = saveHistory(l, histPath)
			fmt.Fprintln(cmd.OutOrStdout())
			os.Exit(0)
		default:
		}

		if err := processLine(cmd, &eng, &base, &engineName, &speak, &lastText, &narrator, line); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "error: %v\n", err)
		}
	}

	if err := saveHistory(l, histPath); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "warning: saving history: %v\n", err)
	}
	return nil
}

func processLine(cmd *cobra.Command, eng *engine.Engine, base *engine.TranslateInput, engName *string, speak *bool, lastText *string, narrator *string, line string) error {
	if strings.HasPrefix(line, ":") {
		// Known meta-commands take priority over SRC:TL shorthand
		if isKnownMetaCommand(line) {
			return handleMetaCommand(cmd, eng, base, engName, speak, lastText, narrator, line)
		}
		// Try :TL text shorthand (e.g. ":en hello", "ja:de こんにちは")
		if text, tl, ok := parseShellLangSpec(line); ok {
			if text == "" {
				// Just set target, no text to translate (like :target)
				base.Target = tl
				fmt.Fprintf(cmd.OutOrStdout(), "target: %s\n", tl)
				return nil
			}
			base.Target = tl
			*lastText = text
			in := *base
			in.Text = text
			out, err := (*eng).Translate(cmd.Context(), in)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "[%s] ", tl)
			fmt.Fprintln(cmd.OutOrStdout(), out.Text)
			if out.Dictionary != "" {
				fmt.Fprintln(cmd.OutOrStdout(), "--")
				fmt.Fprintln(cmd.OutOrStdout(), out.Dictionary)
			}
			if *speak {
				ttsIn := in
				if n := strings.TrimSpace(*narrator); n != "" {
					ttsIn.Target = n
				}
				u, werr := ttsURLForEngine(*eng, ttsIn, out.Text)
				if werr != nil {
					return fmt.Errorf("tts: %w", werr)
				}
				if werr := playTTS(cmd.Context(), u); werr != nil {
					return fmt.Errorf("play: %w", werr)
				}
			}
			return nil
		}
		return fmt.Errorf("unknown command %q — use :target <code> or :<code> text (type :help)", strings.Fields(line)[0])
	}
	*lastText = line
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
		ttsIn := in
		if n := strings.TrimSpace(*narrator); n != "" {
			ttsIn.Target = n
		}
		u, werr := ttsURLForEngine(*eng, ttsIn, out.Text)
		if werr != nil {
			return fmt.Errorf("tts: %w", werr)
		}
		if werr := playTTS(cmd.Context(), u); werr != nil {
			return fmt.Errorf("play: %w", werr)
		}
	}
	return nil
}

func handleMetaCommand(cmd *cobra.Command, eng *engine.Engine, base *engine.TranslateInput, engName *string, speak *bool, lastText *string, narrator *string, line string) error {
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
  :narrator <code> set TTS voice language (or clear)
  :browser [text]  open translation in web browser
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
		code := strings.TrimSpace(parts[1])
		if !lang.IsKnownLanguage(code) {
			return fmt.Errorf("unknown target language code %q", code)
		}
		base.Target = code
		fmt.Fprintf(cmd.OutOrStdout(), "target: %s\n", base.Target)
	case ":source":
		if len(parts) < 2 {
			return fmt.Errorf("usage: :source <code>")
		}
		code := strings.TrimSpace(parts[1])
		if code != "auto" && !lang.IsKnownLanguage(code) {
			return fmt.Errorf("unknown source language code %q", code)
		}
		base.Source = code
		fmt.Fprintf(cmd.OutOrStdout(), "source: %s\n", base.Source)
	case ":host":
		if len(parts) < 2 {
			return fmt.Errorf("usage: :host <code>")
		}
		code := strings.TrimSpace(parts[1])
		if !lang.IsKnownLanguage(code) {
			return fmt.Errorf("unknown host language code %q", code)
		}
		base.HostLang = code
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
		nar := ""
		if narrator != nil {
			nar = *narrator
		}
		fmt.Fprintf(cmd.OutOrStdout(),
			"engine: %s  source: %s  target: %s  host: %s  brief: %v  dict: %v  speak: %v  narrator: %s  dump: %v  noautocorrect: %v  debug: %v\n",
			*engName, base.Source, base.Target, base.HostLang, base.Brief,
			base.Dictionary, *speak, nar, base.Dump, base.NoAutocorrect, base.Debug)
	case ":browser":
		text := strings.Join(parts[1:], " ")
		if text == "" {
			text = *lastText
		}
		if text == "" {
			return fmt.Errorf("no text to open in browser")
		}
		target := base.Target
		if target == "" {
			target = "en"
		}
		return openBrowser(cmd.Context(), base.Source, target, text)
	case ":narrator":
		if len(parts) < 2 {
			*narrator = ""
			fmt.Fprintln(cmd.OutOrStdout(), "narrator: cleared")
			return nil
		}
		*narrator = strings.TrimSpace(parts[1])
		fmt.Fprintf(cmd.OutOrStdout(), "narrator: %s\n", *narrator)
		return nil
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
		":info", ":help", ":browser", ":narrator",
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
		case ":target", ":source", ":host", ":narrator":
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

// isKnownMetaCommand returns true if the line starts with a known REPL meta-command.
func isKnownMetaCommand(line string) bool {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return false
	}
	cmd := strings.ToLower(parts[0])
	switch cmd {
	case ":help", ":engine", ":target", ":source", ":host",
		":brief", ":nobrief", ":dict", ":nodict",
		":speak", ":nospeak", ":dump", ":nodump",
		":noautocorrect", ":autocorrect", ":debug", ":nodebug", ":info", ":browser", ":narrator":
		return true
	}
	return false
}

// parseShellLangSpec tries to parse a line starting with : as :TL text or SRC:TL text.
// Returns the remaining text, the first target language, and true on success.
func parseShellLangSpec(line string) (text, target string, ok bool) {
	parts := strings.Fields(line)
	if len(parts) < 1 {
		return "", "", false
	}
	src, tgts, ok := parseLangPairToken(parts[0])
	if !ok || len(tgts) == 0 {
		return "", "", false
	}
	target = tgts[0]
	_ = src // source override handled via :source meta-command
	text = strings.Join(parts[1:], " ")
	return text, target, true
}
