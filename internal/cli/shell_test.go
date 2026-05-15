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
	shellStdinFn = func() io.Reader { return r }
	t.Cleanup(func() { shellStdinFn = orig })
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

	RunShell(cmd, eng, base)

	output := out.String()
	if !strings.Contains(output, "gtr> [fr] hello") {
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

	RunShell(cmd, eng, base)

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

	RunShell(cmd, eng, base)

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

	RunShell(cmd, eng, base)

	if strings.Contains(out.String(), "[it]") {
		t.Fatal("should not have translated anything")
	}
}
