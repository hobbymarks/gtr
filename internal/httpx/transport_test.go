package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewTransport_USER_AGENT(t *testing.T) {
	t.Setenv("USER_AGENT", "gtr-test/1.0")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != "gtr-test/1.0" {
			t.Errorf("User-Agent = %q, want gtr-test/1.0", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	c := &http.Client{Transport: NewTransport()}
	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}
