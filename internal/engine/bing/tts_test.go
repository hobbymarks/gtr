package bing

import (
	"strings"
	"testing"
)

func TestBuildTTSURL_ok(t *testing.T) {
	u, err := BuildTTSURL("hello", "fr")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(u, "www.bing.com/tspeak") {
		t.Fatalf("unexpected URL: %s", u)
	}
	if !strings.Contains(u, "language=fr") {
		t.Fatalf("missing language: %s", u)
	}
	if !strings.Contains(u, "text=hello") {
		t.Fatalf("missing text: %s", u)
	}
}

func TestBuildTTSURL_emptyText(t *testing.T) {
	_, err := BuildTTSURL("", "fr")
	if err == nil {
		t.Fatal("expected error for empty text")
	}
}

func TestBuildTTSURL_emptyTarget(t *testing.T) {
	_, err := BuildTTSURL("hello", "")
	if err == nil {
		t.Fatal("expected error for empty target")
	}
}

func TestBuildTTSURL_truncation(t *testing.T) {
	long := strings.Repeat("x", maxTTSTextLen+100)
	u, err := BuildTTSURL(long, "de")
	if err != nil {
		t.Fatal(err)
	}
	if len(u) < len(long) {
		return
	}
	t.Fatal("expected truncation")
}
