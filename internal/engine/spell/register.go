package spell

import "github.com/hobbymarks/gtr/internal/engine"

func init() {
	caps := engine.Capabilities{SupportsTTS: false, SupportsDictionary: false}
	engine.Register("spell", func() (engine.Engine, error) {
		return New(ModeAuto, "spell"), nil
	}, caps)
	engine.Register("aspell", func() (engine.Engine, error) {
		return New(ModeAspell, "aspell"), nil
	}, caps)
	engine.Register("hunspell", func() (engine.Engine, error) {
		return New(ModeHunspell, "hunspell"), nil
	}, caps)
}
