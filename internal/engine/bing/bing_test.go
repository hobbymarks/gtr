package bing

import "testing"

func TestParseTranslateResponse_ok(t *testing.T) {
	raw := `[{"translations":[{"text":"Hola"}]}]`
	got, err := parseTranslateResponse([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if got != "Hola" {
		t.Fatalf("got %q", got)
	}
}

func TestParseBingResponse_detected(t *testing.T) {
	raw := `[{"detectedLanguage":{"language":"de","score":1},"translations":[{"text":"hello"}]}]`
	text, detected, err := parseBingResponse([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if text != "hello" || detected != "de" {
		t.Fatalf("text=%q detected=%q", text, detected)
	}
}

func TestParseTranslateResponse_status400(t *testing.T) {
	raw := `[{"statusCode":400}]`
	_, err := parseTranslateResponse([]byte(raw))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPatchFromToLang(t *testing.T) {
	sl, tl := "auto", "pt-PT"
	patchLangCodes(&sl, &tl)
	if sl != "auto-detect" || tl != "pt-pt" {
		t.Fatalf("sl=%q tl=%q", sl, tl)
	}
}
