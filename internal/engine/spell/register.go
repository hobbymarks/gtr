package spell

import "github.com/ueki/gtr/internal/engine"

func init() {
	caps := engine.Capabilities{SupportsTTS: false, SupportsDictionary: false}
	engine.Register("spell", func() (engine.Engine, error) {
		return New(ModeAuto, "spell")
	}, caps)
	engine.Register("aspell", func() (engine.Engine, error) {
		return New(ModeAspell, "aspell")
	}, caps)
	engine.Register("hunspell", func() (engine.Engine, error) {
		return New(ModeHunspell, "hunspell")
	}, caps)
}
