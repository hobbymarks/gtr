package httpx

import (
	"net/http"
	"os"
	"strings"
)

// NewTransport returns a cloned [http.DefaultTransport] with the same proxy
// behavior (HTTP_PROXY, HTTPS_PROXY, NO_PROXY). When USER_AGENT is set, a
// wrapping round tripper sets that header on each request unless the request
// already specifies User-Agent.
func NewTransport() http.RoundTripper {
	base := http.DefaultTransport.(*http.Transport).Clone()
	ua := strings.TrimSpace(os.Getenv("USER_AGENT"))
	if ua == "" {
		return base
	}
	return &userAgentTransport{base: base, ua: ua}
}

type userAgentTransport struct {
	base http.RoundTripper
	ua   string
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r := req.Clone(req.Context())
	if r.Header.Get("User-Agent") == "" {
		r.Header.Set("User-Agent", t.ua)
	}
	return t.base.RoundTrip(r)
}
