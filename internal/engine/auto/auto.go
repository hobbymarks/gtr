package auto

import (
	"context"
	"fmt"

	"github.com/ueki/gtr/internal/engine"
	"github.com/ueki/gtr/internal/lang"
)

// PickBackend mirrors translate-shell Auto.awk routing (Google first, then
// Bing, else Google fallback for Google-only vs Bing-only pairs).
func PickBackend(sl, tl string) string {
	if (sl == "auto" || lang.IsGoogleSupported(sl)) && (tl == "auto" || lang.IsGoogleSupported(tl)) {
		return "google"
	}
	if (sl == "auto" || lang.IsBingSupported(sl)) && (tl == "auto" || lang.IsBingSupported(tl)) {
		return "bing"
	}
	return "google"
}

// Engine delegates to google or bing based on language support tables.
type Engine struct{}

func (e *Engine) Name() string { return "auto" }

func (e *Engine) Translate(ctx context.Context, in engine.TranslateInput) (engine.TranslateOutput, error) {
	backend := PickBackend(in.Source, in.Target)
	f, ok := engine.Lookup(backend)
	if !ok {
		return engine.TranslateOutput{}, fmt.Errorf("auto: backend %q is not registered", backend)
	}
	eng, err := f()
	if err != nil {
		return engine.TranslateOutput{}, fmt.Errorf("auto: backend %q: %w", backend, err)
	}
	return eng.Translate(ctx, in)
}

func init() {
	engine.Register("auto", func() (engine.Engine, error) {
		return &Engine{}, nil
	}, engine.Capabilities{SupportsTTS: false, SupportsDictionary: true})
}
