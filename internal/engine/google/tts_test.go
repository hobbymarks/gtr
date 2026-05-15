package google

import (
	"strings"
	"testing"
)

func TestBuildTTSURL_ok(t *testing.T) {
	u, err := BuildTTSURL("hello", "fr")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(u, "translate.googleapis.com/translate_tts") {
		t.Fatalf("unexpected URL: %s", u)
	}
	if !strings.Contains(u, "tl=fr") {
		t.Fatalf("missing target language: %s", u)
	}
	if !strings.Contains(u, "client=gtx") {
		t.Fatalf("missing client: %s", u)
	}
	if !strings.Contains(u, "q=hello") {
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

func TestBuildTTSURL_specialChars(t *testing.T) {
	u, err := BuildTTSURL("Hallo, Welt!", "de")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(u, "tl=de") {
		t.Fatalf("missing target: %s", u)
	}
}
