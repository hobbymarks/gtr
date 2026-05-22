package cli

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/hobbymarks/gtr/internal/engine"
	"github.com/spf13/cobra"
)

type mockEngine struct{}

func (m *mockEngine) Name() string { return "mock" }
func (m *mockEngine) Translate(ctx context.Context, in engine.TranslateInput) (engine.TranslateOutput, error) {
	return engine.TranslateOutput{Text: "[" + in.Target + "] " + in.Text}, nil
}

func setShellStdin(t *testing.T, r io.Reader) {
	orig := shellStdinFn
	origTTY := stdinIsTTYFn
	shellStdinFn = func() io.Reader { return r }
	stdinIsTTYFn = func() bool { return false }
	t.Cleanup(func() {
		shellStdinFn = orig
		stdinIsTTYFn = origTTY
	})
}

func TestRunShell_basic(t *testing.T) {
	setShellStdin(t, strings.NewReader("hello\nexit\n"))

	eng := &mockEngine{}
	base := engine.TranslateInput{
		Source:   "auto",
		Target:   "fr",
		HostLang: "en",
	}
	cmd := &cobra.Command{}
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	RunShell(cmd, eng, base, "mock")

	output := out.String()
	if !strings.Contains(output, "[fr] hello") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestRunShell_quit(t *testing.T) {
	setShellStdin(t, strings.NewReader("quit\n"))

	eng := &mockEngine{}
	base := engine.TranslateInput{
		Source:   "auto",
		Target:   "de",
		HostLang: "en",
	}
	cmd := &cobra.Command{}
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	RunShell(cmd, eng, base, "mock")

	if strings.Contains(out.String(), "[de]") {
		t.Fatal("should not have translated anything")
	}
}

func TestRunShell_emptyLines(t *testing.T) {
	setShellStdin(t, strings.NewReader("\n\nhola\nexit\n"))

	eng := &mockEngine{}
	base := engine.TranslateInput{
		Source:   "auto",
		Target:   "es",
		HostLang: "en",
	}
	cmd := &cobra.Command{}
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	RunShell(cmd, eng, base, "mock")

	if !strings.Contains(out.String(), "[es] hola") {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestRunShell_caseInsensitiveQuit(t *testing.T) {
	setShellStdin(t, strings.NewReader("QUIT\n"))

	eng := &mockEngine{}
	base := engine.TranslateInput{
		Source:   "auto",
		Target:   "it",
		HostLang: "en",
	}
	cmd := &cobra.Command{}
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	RunShell(cmd, eng, base, "mock")

	if strings.Contains(out.String(), "[it]") {
		t.Fatal("should not have translated anything")
	}
}

func TestRunShell_metaCommands(t *testing.T) {
	setShellStdin(t, strings.NewReader(":target de\nhallo\n:info\nexit\n"))

	eng := &mockEngine{}
	base := engine.TranslateInput{
		Source:   "auto",
		Target:   "fr",
		HostLang: "en",
	}
	cmd := &cobra.Command{}
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	RunShell(cmd, eng, base, "mock")

	output := out.String()
	if !strings.Contains(output, "target: de") {
		t.Fatalf("missing target change: %q", output)
	}
	if !strings.Contains(output, "[de] hallo") {
		t.Fatalf("missing translation with new target: %q", output)
	}
	if !strings.Contains(output, "target: de") {
		t.Fatalf("missing info output: %q", output)
	}
}

func TestRunShell_metaToggle(t *testing.T) {
	setShellStdin(t, strings.NewReader(":brief\n:info\nexit\n"))

	eng := &mockEngine{}
	base := engine.TranslateInput{
		Source:   "auto",
		Target:   "es",
		HostLang: "en",
	}
	cmd := &cobra.Command{}
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	RunShell(cmd, eng, base, "mock")

	output := out.String()
	if !strings.Contains(output, "brief: on") {
		t.Fatalf("missing brief enable: %q", output)
	}
	if !strings.Contains(output, "brief: true") {
		t.Fatalf("missing brief in info: %q", output)
	}
}

func TestShellComplete_metaCommands(t *testing.T) {
	completions := shellComplete(":eng", "mock")
	found := false
	for _, c := range completions {
		if strings.HasPrefix(c, ":engine ") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected :engine in completions, got %v", completions)
	}
}

func TestShellPrompt(t *testing.T) {
	p := shellPrompt("google", "fr")
	if p != "[google:fr]> " {
		t.Fatalf("unexpected prompt: %q", p)
	}
	p = shellPrompt("auto", "")
	if p != "[auto]> " {
		t.Fatalf("unexpected empty-target prompt: %q", p)
	}
}

func TestProcessLine_langShorthand(t *testing.T) {
	mock := &mockEngine{}
	var eng engine.Engine = mock
	base := engine.TranslateInput{
		Source:   "auto",
		Target:   "",
		HostLang: "en",
	}
	var engName string = "mock"
	var speak bool
	var lastText string
	var narrator string
	cmd := &cobra.Command{}
	var out strings.Builder
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := processLine(cmd, &eng, &base, &engName, &speak, &lastText, &narrator, ":en")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "target: en") {
		t.Fatalf("expected target set, got %q", out.String())
	}
	if base.Target != "en" {
		t.Fatalf("expected target=en, got %q", base.Target)
	}

	// :de hello should translate
	out.Reset()
	err = processLine(cmd, &eng, &base, &engName, &speak, &lastText, &narrator, ":de hallo")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "[de] [de] hallo") {
		t.Fatalf("expected translation, got %q", out.String())
	}
}
