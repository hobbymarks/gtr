package engine

import (
	"sort"
	"sync"
)

var (
	regMu    sync.RWMutex
	registry = map[string]Factory{}
)

// Register adds an engine factory under a canonical name (lowercase).
func Register(name string, f Factory) {
	regMu.Lock()
	defer regMu.Unlock()
	registry[name] = f
}

// Lookup returns the factory for name, if registered.
func Lookup(name string) (Factory, bool) {
	regMu.RLock()
	defer regMu.RUnlock()
	f, ok := registry[name]
	return f, ok
}

// Names returns registered engine names in sorted order.
func Names() []string {
	regMu.RLock()
	defer regMu.RUnlock()
	out := make([]string, 0, len(registry))
	for n := range registry {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}
