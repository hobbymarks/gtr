package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
)

const maxStdinBytes = 1 << 20

// stripFileURLPrefix turns file:///path and file://path into a filesystem path.
func stripFileURLPrefix(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "file://") {
		s = strings.TrimPrefix(s, "file://")
		if strings.HasPrefix(s, "//") {
			// file://hostname/path → drop leading "//hostname"; best-effort for localhost.
			if idx := strings.Index(s[2:], "/"); idx >= 0 {
				s = s[2+idx:]
			}
		}
	}
	return s
}

// readTextFile reads a UTF-8 text file with the same size cap as stdin.
func readTextFile(path string) (string, error) {
	path = stripFileURLPrefix(path)
	if path == "" {
		return "", fmt.Errorf("empty input path")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read input file: %w", err)
	}
	if len(b) > maxStdinBytes {
		return "", fmt.Errorf("input file exceeds %d bytes", maxStdinBytes)
	}
	s := strings.TrimSpace(string(b))
	if s == "" {
		return "", fmt.Errorf("empty input file")
	}
	return s, nil
}

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
