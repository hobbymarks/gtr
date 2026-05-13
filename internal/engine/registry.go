package engine

import "sort"

var registry = map[string]Factory{}

// Register adds an engine factory under a canonical name (lowercase).
func Register(name string, f Factory) {
	registry[name] = f
}

// Lookup returns the factory for name, if registered.
func Lookup(name string) (Factory, bool) {
	f, ok := registry[name]
	return f, ok
}

// Names returns registered engine names in sorted order.
func Names() []string {
	out := make([]string, 0, len(registry))
	for n := range registry {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}
