package engine

import "testing"

func TestNamesSorted(t *testing.T) {
	regMu.Lock()
	prev := registry
	prevCaps := capsMap
	registry = map[string]Factory{}
	capsMap = map[string]Capabilities{}
	regMu.Unlock()
	t.Cleanup(func() {
		regMu.Lock()
		registry = prev
		capsMap = prevCaps
		regMu.Unlock()
	})

	Register("bing", func() (Engine, error) { return nil, nil })
	Register("google", func() (Engine, error) { return nil, nil })
	got := Names()
	want := []string{"bing", "google"}
	if len(got) != len(want) {
		t.Fatalf("Names() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Names() = %v, want %v", got, want)
		}
	}
}
