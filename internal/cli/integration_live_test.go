//go:build integration

package cli

import "testing"

// Live HTTP / end-to-end tests belong here so default `go test ./...` stays off the network.
// Run (when implemented): go test -tags=integration -count=1 ./internal/cli/...
func TestIntegrationLiveHTTPReserved(t *testing.T) {
	t.Skip("no live translator tests yet; add behind -tags=integration when you have a stable harness")
}
