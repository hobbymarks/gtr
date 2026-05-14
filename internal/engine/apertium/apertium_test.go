package apertium

import "testing"

func TestParseTranslateBody_ok(t *testing.T) {
	raw := `{"responseData":{"translatedText":"Hola"}}`
	got, err := parseTranslateBody([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if got != "Hola" {
		t.Fatalf("got %q", got)
	}
}

func TestParseTranslateBody_exception(t *testing.T) {
	raw := `{"exception":"bad pair"}`
	_, err := parseTranslateBody([]byte(raw))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseTranslateBody_emptyData(t *testing.T) {
	raw := `{"responseData":null}`
	_, err := parseTranslateBody([]byte(raw))
	if err == nil {
		t.Fatal("expected error")
	}
}
