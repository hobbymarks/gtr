package cli

import (
	"fmt"
	"io"
	"strings"
)

const maxStdinBytes = 1 << 20

// textFromArgsOrStdin returns joined argv or stdin body when argv is empty.
// If argv is empty and stdin is a terminal, it returns an error (caller should
// have handled the “bare gtr” help case before requiring -t).
func textFromArgsOrStdin(args []string, stdin io.Reader, stdinIsTTY bool) (string, error) {
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}
	if stdinIsTTY {
		return "", fmt.Errorf("no text to translate: provide arguments or pipe stdin")
	}
	b, err := io.ReadAll(io.LimitReader(stdin, maxStdinBytes+1))
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}
	if len(b) > maxStdinBytes {
		return "", fmt.Errorf("stdin exceeds %d bytes", maxStdinBytes)
	}
	s := strings.TrimSpace(string(b))
	if s == "" {
		return "", fmt.Errorf("empty stdin")
	}
	return s, nil
}
