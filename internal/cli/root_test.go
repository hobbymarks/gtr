package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hobbymarks/gtr/internal/engine"
	_ "github.com/hobbymarks/gtr/internal/engine/apertium"
	_ "github.com/hobbymarks/gtr/internal/engine/auto"
	_ "github.com/hobbymarks/gtr/internal/engine/bing"
	_ "github.com/hobbymarks/gtr/internal/engine/google"
	_ "github.com/hobbymarks/gtr/internal/engine/spell"
	_ "github.com/hobbymarks/gtr/internal/engine/yandex"
)

func TestVersionFlag_V(t *testing.T) {
	Version = "test-version"
	cmd := newRoot()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"-V"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(buf.String()); got != Version {
		t.Fatalf("printed %q, want %q", got, Version)
	}
}

func TestVersionFlag_long(t *testing.T) {
	Version = "v9"
	cmd := newRoot()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--version"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(buf.String()) != "v9" {
		t.Fatalf("got %q", buf.String())
	}
}

func TestNoArgsShowsHelp(t *testing.T) {
	old := stdinIsTTYFn
	stdinIsTTYFn = func() bool { return true }
	t.Cleanup(func() { stdinIsTTYFn = old })

	cmd := newRoot()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{})
	_ = cmd.Execute()
	if !strings.Contains(out.String(), "gtr") {
		t.Fatalf("expected help output, got %q", out.String())
	}
}

func TestMissingTargetWithText(t *testing.T) {
	cmd := newRoot()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"hello"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "target") {
		t.Fatalf("got %v", err)
	}
}

func TestListEngines(t *testing.T) {
	cmd := newRoot()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--list-engines"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	names := engine.Names()
	if len(names) < 8 {
		t.Fatalf("expected engines registered, got %v", names)
	}
	for _, need := range []string{"ENGINE", "google", "bing", "auto", "yandex", "apertium", "spell", "aspell", "hunspell"} {
		if !strings.Contains(out, need) {
			t.Fatalf("output missing %q: %q", need, out)
		}
	}
}

func TestIdentifyUnsupportedEngine(t *testing.T) {
	cmd := newRoot()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"-e", "apertium", "--identify", "hello"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "identification") {
		t.Fatalf("expected identification error, got %v", err)
	}
}

func TestCannotCombineIdentifyAndDump(t *testing.T) {
	cmd := newRoot()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--identify", "--dump", "-t", "en", "x"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "combine") {
		t.Fatalf("got %v", err)
	}
}

func TestJoinRequiresArgs(t *testing.T) {
	cmd := newRoot()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"-j", "-t", "en"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "-j") {
		t.Fatalf("got %v", err)
	}
}
