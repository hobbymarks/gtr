package cli

import (
	"testing"
)

func TestParseLangPairToken(t *testing.T) {
	cases := []struct {
		in       string
		wantSrc  string
		wantTgts []string
		wantOK   bool
	}{
		{":en", "auto", []string{"en"}, true},
		{"ja:en", "ja", []string{"en"}, true},
		{"auto:en+de", "auto", []string{"en", "de"}, true},
		{"hello", "", nil, false},
		{"http://x", "", nil, false},
		{"a:b:c", "", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			src, tgts, ok := parseLangPairToken(tc.in)
			if ok != tc.wantOK {
				t.Fatalf("ok=%v want %v (src=%q tgts=%v)", ok, tc.wantOK, src, tgts)
			}
			if !tc.wantOK {
				return
			}
			if src != tc.wantSrc {
				t.Fatalf("src got %q want %q", src, tc.wantSrc)
			}
			if len(tgts) != len(tc.wantTgts) {
				t.Fatalf("targets %v want %v", tgts, tc.wantTgts)
			}
			for i := range tc.wantTgts {
				if tgts[i] != tc.wantTgts[i] {
					t.Fatalf("targets[%d] got %q want %q", i, tgts[i], tc.wantTgts[i])
				}
			}
		})
	}
}

func TestStripLeadingLangSpec(t *testing.T) {
	args := []string{":en", "hi"}
	rest, src, tgts, stripped := stripLeadingLangSpec(args, false, false)
	if !stripped || src != "auto" || len(tgts) != 1 || tgts[0] != "en" {
		t.Fatalf("strip got rest=%v src=%q tgts=%v stripped=%v", rest, src, tgts, stripped)
	}
	if len(rest) != 1 || rest[0] != "hi" {
		t.Fatalf("rest %v", rest)
	}
	rest2, _, _, stripped2 := stripLeadingLangSpec(args, true, false)
	if stripped2 || len(rest2) != 2 {
		t.Fatalf("should not strip when source flag changed")
	}
}
