// Package config provides default engine selection, config file (~/.gtrrc)
// parsing, and environment variable overrides for gtr.
package config

import (
	"os"
	"path/filepath"
	"strings"
)

// DefaultEngine is used when -e / --engine is omitted (translate-shell-style auto-router).
const DefaultEngine = "auto"

// EnvOverride returns the value for key from environment variables or a
// config file (~/.gtrrc). Environment variables take precedence.
// Supported keys: GTR_DEFAULT_ENGINE, GTR_DEFAULT_TARGET, GTR_TIMEOUT.
func EnvOverride(key string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return configFileValue(key)
}

// DefaultTarget returns the configured default target language, if any.
func DefaultTarget() string {
	return EnvOverride("GTR_DEFAULT_TARGET")
}

func configFileValue(key string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(home, ".gtrrc"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		if k == key {
			return v
		}
	}
	return ""
}
