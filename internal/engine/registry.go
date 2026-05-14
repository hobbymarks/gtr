package engine

import (
	"sort"
	"strings"
	"sync"
)

var (
	regMu    sync.RWMutex
	registry = map[string]Factory{}
	capsMap  = map[string]Capabilities{}
)

// Register adds an engine factory and optional capabilities under a canonical name (lowercase).
func Register(name string, f Factory, caps ...Capabilities) {
	regMu.Lock()
	defer regMu.Unlock()
	name = strings.ToLower(strings.TrimSpace(name))
	registry[name] = f
	if len(caps) > 0 {
		capsMap[name] = caps[0]
	} else {
		capsMap[name] = Capabilities{}
	}
}

// CapabilitiesOf returns metadata for a registered engine (zero value if unknown).
func CapabilitiesOf(name string) Capabilities {
	regMu.RLock()
	defer regMu.RUnlock()
	return capsMap[strings.ToLower(strings.TrimSpace(name))]
}

// Lookup returns the factory for name, if registered.
func Lookup(name string) (Factory, bool) {
	regMu.RLock()
	defer regMu.RUnlock()
	f, ok := registry[strings.ToLower(strings.TrimSpace(name))]
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
