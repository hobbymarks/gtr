package cli

import (
	"bytes"
	"strings"
	"testing"
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
	cmd := newRoot()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{})
	_ = cmd.Execute()
	if !strings.Contains(out.String(), "gtr") {
		t.Fatalf("expected help output, got %q", out.String())
	}
}
