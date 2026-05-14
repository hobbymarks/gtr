package yandex

import "testing"

func TestParseDetectedSourceFromLangJSON(t *testing.T) {
	raw := `{"code":200,"lang":"de-en","text":["hello"]}`
	got, err := parseDetectedSourceFromLangJSON([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if got != "de" {
		t.Fatalf("got %q", got)
	}
}
