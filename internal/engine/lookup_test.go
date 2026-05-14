package engine

import (
	"testing"
)

func TestLookupFuzzy(t *testing.T) {
	regMu.Lock()
	prev := registry
	registry = map[string]Factory{
		"google": func() (Engine, error) { return nil, nil },
		"bing":   func() (Engine, error) { return nil, nil },
		"auto":   func() (Engine, error) { return nil, nil },
	}
	regMu.Unlock()
	t.Cleanup(func() {
		regMu.Lock()
		registry = prev
		regMu.Unlock()
	})

	canon, _, ok := LookupFuzzy("goo")
	if !ok || canon != "google" {
		t.Fatalf("LookupFuzzy goo)=%v,%v", canon, ok)
	}
	canon, _, ok = LookupFuzzy("bi")
	if !ok || canon != "bing" {
		t.Fatalf("LookupFuzzy bi)=%v,%v", canon, ok)
	}
	_, _, ok = LookupFuzzy("zzz")
	if ok {
		t.Fatal("expected miss")
	}
}

func TestLookupFuzzy_ambiguousShortest(t *testing.T) {
	regMu.Lock()
	prev := registry
	registry = map[string]Factory{
		"goog":   func() (Engine, error) { return nil, nil },
		"google": func() (Engine, error) { return nil, nil },
	}
	regMu.Unlock()
	t.Cleanup(func() {
		regMu.Lock()
		registry = prev
		regMu.Unlock()
	})
	canon, _, ok := LookupFuzzy("go")
	if !ok {
		t.Fatal("expected hit")
	}
	if canon != "goog" {
		t.Fatalf("want shortest prefix match, got %q", canon)
	}
}
