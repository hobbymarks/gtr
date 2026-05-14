package engine

import (
	"sort"
	"strings"
)

// LookupFuzzy resolves an engine name case-insensitively, then by shortest
// registered prefix (translate-shell-style fuzzy match).
func LookupFuzzy(name string) (canon string, factory Factory, ok bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return "", nil, false
	}
	regMu.RLock()
	defer regMu.RUnlock()
	if f, ok := registry[name]; ok {
		return name, f, true
	}
	var keys []string
	for k := range registry {
		if strings.HasPrefix(k, name) {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return "", nil, false
	}
	sort.Slice(keys, func(i, j int) bool {
		if len(keys[i]) != len(keys[j]) {
			return len(keys[i]) < len(keys[j])
		}
		return keys[i] < keys[j]
	})
	k := keys[0]
	return k, registry[k], true
}
