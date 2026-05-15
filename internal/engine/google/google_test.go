package google

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseTranslateSingleResponse_golden(t *testing.T) {
	dir := "testdata"
	ents, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range ents {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		name := e.Name()
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			var doc struct {
				Response json.RawMessage `json:"response"`
				Want     string          `json:"want"`
			}
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatal(err)
			}
			got, _, err := ParseTranslateSingleResponse(doc.Response)
			if err != nil {
				t.Fatal(err)
			}
			if got != doc.Want {
				t.Fatalf("got %q want %q", got, doc.Want)
			}
		})
	}
}

func TestParseDetectedSourceLanguage(t *testing.T) {
	raw := `[[],[], "de", [], "hello"]`
	got, err := ParseDetectedSourceLanguage([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if got != "de" {
		t.Fatalf("got %q", got)
	}
}

func TestParseTranslateSingleResponse_errors(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{"empty", ""},
		{"not_array", `{"a":1}`},
		{"empty_root_array", `[]`},
		{"no_strings", `[[[]]]`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := ParseTranslateSingleResponse([]byte(tc.raw))
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestBuildSingleRequestURL(t *testing.T) {
	got, err := buildSingleRequestURL("hello world", "auto", "fr", "en", false)
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if u.Path != "/translate_a/single" {
		t.Fatalf("path %q", u.Path)
	}
	q := u.Query()
	if q.Get("client") != "gtx" || q.Get("sl") != "auto" || q.Get("tl") != "fr" || q.Get("hl") != "en" {
		t.Fatalf("query mismatch: %v", q)
	}
	if q.Get("q") != "hello world" {
		t.Fatalf("q=%q", q.Get("q"))
	}
	dts := q["dt"]
	if !containsAll(dts, []string{"t", "qca"}) {
		t.Fatalf("dt values: %v", dts)
	}

	got2, err := buildSingleRequestURL("x", "en", "de", "en", true)
	if err != nil {
		t.Fatal(err)
	}
	u2, _ := url.Parse(got2)
	dts2 := u2.Query()["dt"]
	if !containsAll(dts2, []string{"t", "qc"}) || containsString(dts2, "qca") {
		t.Fatalf("expected qc not qca in dt: %v", dts2)
	}
}

func containsAll(hay, need []string) bool {
	for _, n := range need {
		found := false
		for _, h := range hay {
			if h == n {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func containsString(hay []string, s string) bool {
	for _, h := range hay {
		if h == s {
			return true
		}
	}
	return false
}
