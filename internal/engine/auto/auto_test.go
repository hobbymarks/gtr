package auto

import "testing"

func TestPickBackend(t *testing.T) {
	cases := []struct {
		sl, tl, want string
	}{
		{"en", "fr", "google"},
		{"auto", "fr", "google"},
		{"yue", "en", "bing"},
		{"ba", "en", "bing"},
		{"ay", "en", "google"},
		{"cv", "en", "google"},
		{"auto", "yue", "bing"},
	}
	for _, tc := range cases {
		if got := PickBackend(tc.sl, tc.tl); got != tc.want {
			t.Errorf("PickBackend(%q,%q)=%q want %q", tc.sl, tc.tl, got, tc.want)
		}
	}
}
