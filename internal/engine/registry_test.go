package engine

import "testing"

func TestNamesSorted(t *testing.T) {
	prev := registry
	t.Cleanup(func() { registry = prev })
	registry = map[string]Factory{}
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
