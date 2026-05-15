package bing

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetup_regexScraping(t *testing.T) {
	html := `<script>var params_AbusePreventionHelper = [12345,"abc123def","https://www.bing.com"];</script>
<div>IG:"abcdef12"</div>
<input data-iid="translator.5010" />`
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Set-Cookie", "SRCHHPGUSR=xyz; path=/")
		w.Write([]byte(html))
	}))
	defer svr.Close()

	eng := &Engine{HTTP: svr.Client()}
	cookie, ig, iid, token, key, err := eng.setup(context.Background(), svr.URL)
	if err != nil {
		t.Fatal(err)
	}
	if ig != "abcdef12" {
		t.Fatalf("ig=%q want abcdef12", ig)
	}
	if iid != "translator.5010" {
		t.Fatalf("iid=%q want translator.5010", iid)
	}
	if key != "12345" {
		t.Fatalf("key=%q want 12345", key)
	}
	if token != "abc123def" {
		t.Fatalf("token=%q want abc123def", token)
	}
	if !strings.Contains(cookie, "SRCHHPGUSR=xyz") {
		t.Fatalf("missing cookie: %q", cookie)
	}
}

func TestSetup_httpError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer svr.Close()

	eng := &Engine{HTTP: svr.Client()}
	_, _, _, _, _, err := eng.setup(context.Background(), svr.URL)
	if err == nil {
		t.Fatal("expected error for 503")
	}
	if !strings.Contains(err.Error(), "HTTP 503") {
		t.Fatalf("error should mention HTTP 503: %v", err)
	}
}

func TestSetup_missingIG(t *testing.T) {
	html := `<div>no IG here</div>
<script>var params_AbusePreventionHelper = [1,"t",""];</script>`
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
	defer svr.Close()

	eng := &Engine{HTTP: svr.Client()}
	_, _, _, _, _, err := eng.setup(context.Background(), svr.URL)
	if err == nil {
		t.Fatal("expected error for missing IG")
	}
}

func TestSetup_missingToken(t *testing.T) {
	html := `<div>IG:"abc"</div><input data-iid="xyz" />`
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
	defer svr.Close()

	eng := &Engine{HTTP: svr.Client()}
	_, _, _, _, _, err := eng.setup(context.Background(), svr.URL)
	if err == nil {
		t.Fatal("expected error for missing token")
	}
}

func TestSetup_golden_fixtures(t *testing.T) {
	dir := "testdata"
	ents, err := os.ReadDir(dir)
	if err != nil {
		t.Skipf("no testdata dir: %v", err)
	}
	for _, e := range ents {
		if e.IsDir() || !strings.HasSuffix(e.Name(), "_setup.html") {
			continue
		}
		name := e.Name()
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			parts := strings.SplitN(string(raw), "\n---\n", 2)
			html := parts[0]
			var want struct {
				IG    string `json:"ig"`
				IID   string `json:"iid"`
				Key   string `json:"key"`
				Token string `json:"token"`
			}
			if len(parts) > 1 {
				if err := json.Unmarshal([]byte(parts[1]), &want); err != nil {
					t.Fatal(err)
				}
			}

			svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(html))
			}))
			defer svr.Close()

			eng := &Engine{HTTP: svr.Client()}
			_, ig, iid, _, token, err := eng.setup(context.Background(), svr.URL)
			if want.IG != "" && ig != want.IG {
				t.Fatalf("ig=%q want %q", ig, want.IG)
			}
			if want.IID != "" && iid != want.IID {
				t.Fatalf("iid=%q want %q", iid, want.IID)
			}
			if want.Token != "" && token != want.Token {
				t.Fatalf("token=%q want %q", token, want.Token)
			}
			if err != nil && want.IG != "" {
				t.Fatal(err)
			}
		})
	}
}
