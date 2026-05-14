package google

import (
	"strings"
	"testing"
)

func TestFormatDictionaryPayload_indices(t *testing.T) {
	// Minimal root: [0] sentences, [1] extra payload, [5] alternatives slot.
	raw := `[[["Hi","Hola"]],"classes","en",null,null,[["alt1","alt2"]]]`
	got, err := FormatDictionaryPayload([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if got == "" {
		t.Fatal("expected non-empty dictionary payload")
	}
	for _, p := range []string{"[1]", "[5]", "classes", "alt1"} {
		if !strings.Contains(got, p) {
			t.Fatalf("missing %q in %q", p, got)
		}
	}
}
