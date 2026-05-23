// Package config provides default engine selection, config file (~/.gtrrc)
// parsing, and environment variable overrides for gtr.
package config

import (
	"os"
	"path/filepath"
	"strings"
)

// DefaultEngine returns the configured default engine name (env/config override, or "auto").
func DefaultEngine() string {
	if v := EnvOverride("GTR_DEFAULT_ENGINE"); v != "" {
		return v
	}
	return "auto"
}

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

// KnownConfigKeys returns the supported config file keys.
func KnownConfigKeys() []string {
	return []string{"GTR_DEFAULT_ENGINE", "GTR_DEFAULT_TARGET", "GTR_TIMEOUT"}
}

// IsKnownConfigKey reports whether key is a supported config key.
func IsKnownConfigKey(key string) bool {
	for _, k := range KnownConfigKeys() {
		if k == key {
			return true
		}
	}
	return false
}

// ConfigFileValueForPath reads a .gtrrc-style file and returns the value for key.
func ConfigFileValueForPath(path, key string) string {
	data, err := os.ReadFile(path)
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

func configFileValue(key string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return ConfigFileValueForPath(filepath.Join(home, ".gtrrc"), key)
}
