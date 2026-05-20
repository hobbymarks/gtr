// Package lang embeds translate-shell language support metadata and provides
// canonical resolution, engine support queries, and language code validation for gtr.
package lang

import (
	_ "embed"
	"encoding/json"
	"regexp"
	"strings"
	"sync"
)

//go:embed data/language_support.json
//go:generate python3 scripts/gen_language_support.py ../translate-shell/include/LanguageData.awk
var embedded []byte

var (
	loadOnce sync.Once
	st       *store
	loadErr  error
)

type store struct {
	Support map[string]struct {
		Google bool `json:"google"`
		Bing   bool `json:"bing"`
	} `json:"support"`
	Aliases map[string]string `json:"aliases"`
}

func data() *store {
	loadOnce.Do(func() {
		var s store
		if err := json.Unmarshal(embedded, &s); err != nil {
			loadErr = err
			return
		}
		st = &s
	})
	return st
}

// Err returns a non-nil error if embedded language data failed to load.
func Err() error {
	_ = data()
	return loadErr
}

// stripRegionTag mirrors translate-shell getCode’s trailing-tag strip:
// ^([[:alpha:]][[:alpha:]][[:alpha:]]?)-(.*)$
var stripRegionTag = regexp.MustCompile(`^([A-Za-z]{2,3})-(.*)$`)

// ResolveCanonical mirrors translate-shell getCode for lookup keys present in
// LanguageData / LocaleAlias (enough for engine routing).
func ResolveCanonical(code string) string {
	s := data()
	if loadErr != nil || s == nil {
		return ""
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return ""
	}
	if code == "auto" {
		return "auto"
	}
	if _, ok := s.Support[code]; ok {
		return code
	}
	if v, ok := s.Aliases[strings.ToLower(code)]; ok {
		return v
	}
	if v, ok := s.Aliases[code]; ok {
		return v
	}
	if m := stripRegionTag.FindStringSubmatch(code); m != nil {
		return m[1]
	}
	return ""
}

// IsGoogleSupported mirrors translate-shell isSupportedByGoogle(getCode(code)).
func IsGoogleSupported(code string) bool {
	s := data()
	if loadErr != nil || s == nil {
		return false
	}
	c := ResolveCanonical(code)
	if c == "" || c == "auto" {
		return false
	}
	ent, ok := s.Support[c]
	return ok && ent.Google
}

// IsBingSupported mirrors translate-shell isSupportedByBing(getCode(code)).
func IsBingSupported(code string) bool {
	s := data()
	if loadErr != nil || s == nil {
		return false
	}
	c := ResolveCanonical(code)
	if c == "" || c == "auto" {
		return false
	}
	ent, ok := s.Support[c]
	return ok && ent.Bing
}

// IsKnownLanguage returns true if code resolves to a known language code
// in the embedded language support data. "auto" always returns true.
func IsKnownLanguage(code string) bool {
	code = strings.TrimSpace(code)
	if code == "" || code == "auto" {
		return true
	}
	return ResolveCanonical(code) != ""
}

// AllCodes returns all known language codes from the embedded data.
func AllCodes() map[string]bool {
	s := data()
	if loadErr != nil || s == nil {
		return nil
	}
	codes := make(map[string]bool, len(s.Support))
	for k := range s.Support {
		codes[k] = true
	}
	return codes
}
