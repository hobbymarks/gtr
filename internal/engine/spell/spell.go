package spell

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ueki/gtr/internal/engine"
)

// Mode selects which external checker binary to use.
type Mode int

const (
	// ModeAuto prefers aspell, then hunspell.
	ModeAuto Mode = iota
	ModeAspell
	ModeHunspell
)

// Engine runs an ispell-compatible spell checker subprocess (translate-shell parity).
type Engine struct {
	mode   Mode
	binary string
	name   string // CLI engine name (spell, aspell, hunspell)
}

// New returns a spell engine or an error if no suitable binary exists.
func New(mode Mode, engineName string) (*Engine, error) {
	var (
		bin string
		err error
		eff = mode
	)
	switch mode {
	case ModeAspell:
		bin, err = lookTool("aspell", []string{"--version"})
	case ModeHunspell:
		bin, err = lookTool("hunspell", []string{"--version"})
	default:
		bin, err = lookTool("aspell", []string{"--version"})
		if err != nil {
			bin, err = lookTool("hunspell", []string{"--version"})
			eff = ModeHunspell
		} else {
			eff = ModeAspell
		}
	}
	if err != nil {
		return nil, err
	}
	return &Engine{mode: eff, binary: bin, name: engineName}, nil
}

func (e *Engine) Name() string {
	return e.name
}

func (e *Engine) Translate(ctx context.Context, in engine.TranslateInput) (engine.TranslateOutput, error) {
	if in.Dump {
		return engine.TranslateOutput{}, fmt.Errorf("%s: --dump is not supported for spell engines", e.Name())
	}
	if in.Dictionary {
		return engine.TranslateOutput{}, fmt.Errorf("%s: dictionary mode (-d) is not applicable", e.Name())
	}
	lang := strings.TrimSpace(in.Source)
	if lang == "" || lang == "auto" {
		lang = "en"
	}
	args := e.buildArgs(lang)
	cmd := exec.CommandContext(ctx, e.binary, args...)
	cmd.Stdin = strings.NewReader(in.Text)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return engine.TranslateOutput{}, fmt.Errorf("%s: %s", e.Name(), msg)
	}
	report, err := formatSpellReport(in.Text, stdout.String())
	if err != nil {
		return engine.TranslateOutput{}, err
	}
	if in.Brief {
		report = strings.TrimSpace(report)
	}
	return engine.TranslateOutput{Text: report}, nil
}

func (e *Engine) buildArgs(lang string) []string {
	if filepath.Base(e.binary) == "hunspell" {
		return []string{"-a", "-d", hunspellDictName(lang)}
	}
	return []string{"-a", "-l", lang, "--encoding=utf-8"}
}

func hunspellDictName(lang string) string {
	if m := hunspellDictOverrides[strings.ToLower(lang)]; m != "" {
		return m
	}
	return lang
}

var hunspellDictOverrides = map[string]string{
	"en": "en_US",
}

var reMisspell = regexp.MustCompile(`^& ([^ ]+) [0-9]+ [0-9]+: (.+)$`)

func formatSpellReport(original, ispellOut string) (string, error) {
	var issues []string
	for _, line := range strings.Split(ispellOut, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "@") {
			continue
		}
		if m := reMisspell.FindStringSubmatch(line); len(m) == 3 {
			word := m[1]
			rawSug := m[2]
			parts := strings.Split(rawSug, ", ")
			var sug []string
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					sug = append(sug, p)
				}
			}
			issues = append(issues, fmt.Sprintf("%s: %s", word, strings.Join(sug, ", ")))
		}
	}
	if len(issues) == 0 {
		return original + "\n(no spelling issues reported)", nil
	}
	var b strings.Builder
	b.WriteString(original)
	b.WriteString("\n\nSpelling:\n")
	for _, s := range issues {
		b.WriteString("- ")
		b.WriteString(s)
		b.WriteString("\n")
	}
	return b.String(), nil
}

func lookTool(name string, versionArg []string) (string, error) {
	p, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("%s not found in PATH", name)
	}
	cmd := exec.Command(p, versionArg...)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s failed self-check: %w", name, err)
	}
	return p, nil
}

var _ engine.Engine = (*Engine)(nil)
