package yandex

import (
	"strings"
	"testing"
)

func TestNewUCID_format(t *testing.T) {
	id := newUCID()
	if len(id) != 32 {
		t.Fatalf("expected 32 hex chars, got %d: %q", len(id), id)
	}
	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Fatalf("invalid hex char %q in %q", c, id)
		}
	}
}

func TestNewUCID_unique(t *testing.T) {
	const n = 100
	seen := make(map[string]bool, n)
	for i := 0; i < n; i++ {
		id := newUCID()
		if seen[id] {
			t.Fatalf("duplicate ucid: %q", id)
		}
		seen[id] = true
	}
}

func TestNewUCID_noDashes(t *testing.T) {
	id := newUCID()
	if strings.Contains(id, "-") {
		t.Fatalf("ucid contains dashes: %q", id)
	}
}
