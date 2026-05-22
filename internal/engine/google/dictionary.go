package google

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// FormatDictionaryPayload extracts non-sentence segments from a translate_a/single
// JSON array (indices used for definitions / alternatives in translate-shell) as
// indented JSON for human inspection.
func FormatDictionaryPayload(raw []byte) (string, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return "", fmt.Errorf("google: empty response body")
	}
	var root []any
	if err := json.Unmarshal(raw, &root); err != nil {
		return "", fmt.Errorf("google: invalid JSON: %w", err)
	}
	var b strings.Builder
	for _, idx := range dictIndices {
		if len(root) <= idx || root[idx] == nil {
			continue
		}
		enc, err := json.MarshalIndent(root[idx], "", "  ")
		if err != nil {
			continue
		}
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "[%d]\n%s", idx, enc)
	}
	return strings.TrimSpace(b.String()), nil
}

// dictIndices names the array positions inside translate_a/single JSON that
// contain auxiliary payload segments (definitions, alternatives, synonyms).
var dictIndices = []int{1, 5, 11, 12}
