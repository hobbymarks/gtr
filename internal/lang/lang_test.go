package lang

import "testing"

func TestResolveCanonical(t *testing.T) {
	if Err() != nil {
		t.Fatal(Err())
	}
	cases := []struct{ in, want string }{
		{"en", "en"},
		{"zh", "zh-CN"},
		{"eng", "en"},
		{"auto", "auto"},
		{"yue", "yue"},
	}
	for _, tc := range cases {
		if got := ResolveCanonical(tc.in); got != tc.want {
			t.Errorf("ResolveCanonical(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}

func TestSupportFlags(t *testing.T) {
	if !IsGoogleSupported("en") || !IsBingSupported("en") {
		t.Fatal("en should be google+bing")
	}
	if IsGoogleSupported("yue") || !IsBingSupported("yue") {
		t.Fatal("yue is bing-only in reference data")
	}
	if IsGoogleSupported("ba") || !IsBingSupported("ba") {
		t.Fatal("ba is bing+yandex only")
	}
	if !IsGoogleSupported("ay") || IsBingSupported("ay") {
		t.Fatal("ay is google-only")
	}
	if IsGoogleSupported("cv") || IsBingSupported("cv") {
		t.Fatal("cv is yandex-only in reference data")
	}
}
