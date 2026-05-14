package yandex

import (
	"testing"
)

func TestParseTranslateJSON_ok(t *testing.T) {
	raw := `{"code":200,"lang":"en-ru","text":["Привет"]}`
	got, err := parseTranslateJSON([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if got != "Привет" {
		t.Fatalf("got %q", got)
	}
}

func TestParseTranslateJSON_textString(t *testing.T) {
	raw := `{"code":200,"text":"Hi"}`
	got, err := parseTranslateJSON([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if got != "Hi" {
		t.Fatalf("got %q", got)
	}
}

func TestParseTranslateJSON_errorCode(t *testing.T) {
	raw := `{"code":401,"message":"nope"}`
	_, err := parseTranslateJSON([]byte(raw))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStripDigraph(t *testing.T) {
	if stripDigraph("zh-CN") != "zh" {
		t.Fatal()
	}
	if stripDigraph("auto") != "auto" {
		t.Fatal()
	}
}
