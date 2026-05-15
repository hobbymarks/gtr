// Package httpx provides shared HTTP client, transport, retry, and utility
// functions for all gtr translation engines.
package httpx

import (
	"net/http"
	"time"
)

const defaultTimeout = 30 * time.Second

// NewClient returns an [http.Client] suitable for translator backends.
// It honors HTTP_PROXY / HTTPS_PROXY via the default transport proxy settings,
// applies defaultTimeout, and optionally sets User-Agent from USER_AGENT (same
// env name as translate-shell).
func NewClient() *http.Client {
	t := NewTransport()
	return &http.Client{
		Timeout:   defaultTimeout,
		Transport: t,
	}
}
