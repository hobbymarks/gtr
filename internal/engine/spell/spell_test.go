package spell

import (
	"strings"
	"testing"
)

func TestFormatSpellReport_misspelling(t *testing.T) {
	ispell := "& wrd 2 6: word, ward, weird\n"
	got, err := formatSpellReport("hello wrd", ispell)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "wrd:") || !strings.Contains(got, "word") {
		t.Fatalf("got %q", got)
	}
}

func TestFormatSpellReport_clean(t *testing.T) {
	got, err := formatSpellReport("hello", "*\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "no spelling") {
		t.Fatalf("got %q", got)
	}
}
