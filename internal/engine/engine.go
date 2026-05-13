package engine

import "context"

// TranslateInput is the normalized request passed to engines (expanded in Phase 1).
type TranslateInput struct {
	Text     string
	Source   string
	Target   string
	HostLang string
	Brief    bool
}

// TranslateOutput is the normalized response from an engine.
type TranslateOutput struct {
	Text string
}

// Engine performs translation for a single backend (google, bing, etc.).
type Engine interface {
	Name() string
	Translate(ctx context.Context, in TranslateInput) (TranslateOutput, error)
}

// Factory constructs an Engine instance, for example after reading config.
type Factory func() (Engine, error)
