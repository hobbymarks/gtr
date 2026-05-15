package google

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestIdentifyLanguage_ok(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "sl=auto") {
			t.Error("missing sl=auto")
		}
		w.Write([]byte(`[[], [], "fr", [], "hello"]`))
	}))
	defer svr.Close()

	eng := &Engine{HTTP: &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			r.URL.Scheme = "http"
			r.URL.Host = svr.Listener.Addr().String()
			return http.DefaultTransport.RoundTrip(r)
		}),
	}}

	lang, err := eng.IdentifyLanguage(context.Background(), "bonjour", "en")
	if err != nil {
		t.Fatal(err)
	}
	if lang != "fr" {
		t.Fatalf("lang=%q want fr", lang)
	}
}

func TestIdentifyLanguage_emptyText(t *testing.T) {
	eng := &Engine{HTTP: &http.Client{}}
	_, err := eng.IdentifyLanguage(context.Background(), "", "en")
	if err == nil {
		t.Fatal("expected error for empty text")
	}
}

func TestIdentifyLanguage_defaultHostLang(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("hl") != "en" {
			t.Errorf("hl=%q want en", q.Get("hl"))
		}
		w.Write([]byte(`[[], [], "de", [], "hallo"]`))
	}))
	defer svr.Close()

	eng := &Engine{HTTP: &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			r.URL.Scheme = "http"
			r.URL.Host = svr.Listener.Addr().String()
			return http.DefaultTransport.RoundTrip(r)
		}),
	}}

	lang, err := eng.IdentifyLanguage(context.Background(), "hallo", "")
	if err != nil {
		t.Fatal(err)
	}
	if lang != "de" {
		t.Fatalf("lang=%q want de", lang)
	}
}

func TestIdentifyLanguage_httpError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	eng := &Engine{HTTP: &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			r.URL.Scheme = "http"
			r.URL.Host = svr.Listener.Addr().String()
			return http.DefaultTransport.RoundTrip(r)
		}),
	}}

	_, err := eng.IdentifyLanguage(context.Background(), "text", "en")
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
	if !strings.Contains(err.Error(), "HTTP 500") {
		t.Fatalf("error should mention HTTP 500: %v", err)
	}
}

func TestBuildSingleRequestURL_identify(t *testing.T) {
	u, err := buildSingleRequestURL("test", "auto", "en", "en", false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(u, "sl=auto") {
		t.Fatalf("missing sl=auto: %s", u)
	}
	if !strings.Contains(u, "tl=en") {
		t.Fatalf("missing tl=en: %s", u)
	}
	if !strings.Contains(u, "q=test") {
		t.Fatalf("missing q=test: %s", u)
	}
}
