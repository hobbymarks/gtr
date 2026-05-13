package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestTextFromArgsOrStdin_argv(t *testing.T) {
	got, err := textFromArgsOrStdin([]string{"a", "b"}, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if got != "a b" {
		t.Fatalf("got %q", got)
	}
}

func TestTextFromArgsOrStdin_stdinPipe(t *testing.T) {
	r := bytes.NewBufferString("  hello \n")
	got, err := textFromArgsOrStdin(nil, r, false)
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Fatalf("got %q", got)
	}
}

func TestTextFromArgsOrStdin_ttyNoArgs(t *testing.T) {
	_, err := textFromArgsOrStdin(nil, bytes.NewReader(nil), true)
	if err == nil || !strings.Contains(err.Error(), "no text") {
		t.Fatalf("got %v", err)
	}
}

func TestTextFromArgsOrStdin_emptyStdin(t *testing.T) {
	_, err := textFromArgsOrStdin(nil, bytes.NewReader([]byte("  \n")), false)
	if err == nil {
		t.Fatal("expected error")
	}
}
