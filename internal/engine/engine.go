package engine

import "context"

// MaxReadBody caps HTTP response body reads across all engines to prevent
// unbounded memory consumption from malicious or buggy upstream servers.
const MaxReadBody = 4 << 20

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
	Debug           bool
	// Dump returns the raw HTTP response body as Text without parsing (translate-shell -dump).
	Dump bool
	// Dictionary asks engines that support it to fill TranslateOutput.Dictionary (e.g. Google extras).
	Dictionary bool
}

// TranslateOutput is the normalized response from an engine.
type TranslateOutput struct {
	Text string
	// Dictionary holds auxiliary payload (definitions / alternatives) when requested.
	Dictionary string
	// Phonetic holds romanization / transliteration of the translated text (e.g. pinyin),
	// populated by engines that support it (Google with dt=rm).
	Phonetic string
}

// Engine performs translation for a single backend (google, bing, etc.).
type Engine interface {
	Name() string
	Translate(ctx context.Context, in TranslateInput) (TranslateOutput, error)
}

// Factory constructs an Engine instance, for example after reading config.
type Factory func() (Engine, error)

// LanguageIdentifier is implemented by engines that can detect the language of arbitrary text.
type LanguageIdentifier interface {
	IdentifyLanguage(ctx context.Context, text string, hostLang string) (langCode string, err error)
}
