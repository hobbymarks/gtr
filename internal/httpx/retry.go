package httpx

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

var (
	sharedTransportOnce sync.Once
	sharedTransport     http.RoundTripper
	// SharedClientTimeout configures the timeout for NewSharedClient().
	// Set before the first call to NewSharedClient() (default 30s).
	SharedClientTimeout = defaultTimeout
)

func sharedRoundTripper() http.RoundTripper {
	sharedTransportOnce.Do(func() {
		sharedTransport = NewTransport()
	})
	return sharedTransport
}

// NewSharedClient returns an http.Client backed by a single shared transport
// (connection pool) for use across all engines. Uses SharedClientTimeout.
func NewSharedClient() *http.Client {
	return &http.Client{
		Timeout:   SharedClientTimeout,
		Transport: &retryTransport{base: sharedRoundTripper()},
	}
}

const (
	maxRetries    = 3
	backoffBase   = 500 * time.Millisecond
	retryableMin  = 429
	retryableMax  = 599
)

type retryTransport struct {
	base http.RoundTripper
}

func (rt *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := backoffBase * time.Duration(1<<(attempt-1))
			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case <-time.After(delay):
			}
		}
		resp, err := rt.base.RoundTrip(req)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode >= retryableMin && resp.StatusCode <= retryableMax {
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			continue
		}
		return resp, nil
	}
	return nil, fmt.Errorf("retry exhausted after %d attempts: %w", maxRetries+1, lastErr)
}
