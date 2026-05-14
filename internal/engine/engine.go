package engine

import "context"

// Capabilities describes optional engine features (CLI help and future flags).
type Capabilities struct {
	SupportsTTS        bool
	SupportsDictionary bool
}

// TranslateInput is the normalized request passed to engines (expanded in Phase 1).
type TranslateInput struct {
	Text     string
	Source   string
	Target   string
	HostLang string
	Brief    bool
	// NoAutocorrect maps to translate-shell -no-autocorrect (qc vs qca).
	NoAutocorrect bool
	Debug         bool
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
