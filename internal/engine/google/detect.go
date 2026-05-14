package google

import (
	"encoding/json"
	"fmt"
)

// ParseDetectedSourceLanguage reads the detected source language from a
// translate_a/single JSON payload (root[2] when present, as in translate-shell).
func ParseDetectedSourceLanguage(raw []byte) (string, error) {
	var root []interface{}
	if err := json.Unmarshal(raw, &root); err != nil {
		return "", fmt.Errorf("google: invalid JSON: %w", err)
	}
	if len(root) <= 2 {
		return "", fmt.Errorf("google: no detected language in response")
	}
	switch v := root[2].(type) {
	case string:
		if v == "" {
			return "", fmt.Errorf("google: empty detected language")
		}
		return v, nil
	default:
		return "", fmt.Errorf("google: unexpected detected language type")
	}
}
