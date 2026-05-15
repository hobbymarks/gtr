package httpx

import (
	"bytes"
)

// Truncate returns a string from b, trimmed of whitespace, capped at n bytes
// with an ellipsis appended when truncation occurs. Used for safe error messages
// from HTTP response bodies.
func Truncate(b []byte, n int) string {
	s := string(bytes.TrimSpace(b))
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
