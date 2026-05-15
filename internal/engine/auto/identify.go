package auto

import (
	"context"
	"fmt"

	"github.com/hobbymarks/gtr/internal/engine"
)

// IdentifyLanguage delegates to the same backend [PickBackend] would use for auto→English.
func (e *Engine) IdentifyLanguage(ctx context.Context, text, hostLang string) (string, error) {
	backend := PickBackend("auto", "en")
	f, ok := engine.Lookup(backend)
	if !ok {
		return "", fmt.Errorf("auto: backend %q is not registered", backend)
	}
	eng, err := f()
	if err != nil {
		return "", fmt.Errorf("auto: backend %q: %w", backend, err)
	}
	li, ok := eng.(engine.LanguageIdentifier)
	if !ok {
		return "", fmt.Errorf("auto: backend %q does not support language identification", backend)
	}
	return li.IdentifyLanguage(ctx, text, hostLang)
}

var _ engine.LanguageIdentifier = (*Engine)(nil)
