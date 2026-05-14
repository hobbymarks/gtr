package yandex

import (
	"github.com/ueki/gtr/internal/engine"
	"github.com/ueki/gtr/internal/httpx"
)

func init() {
	// Dictionary path exists in translate-shell but is disabled (FIXME upstream).
	engine.Register("yandex", func() (engine.Engine, error) {
		return New(httpx.NewClient()), nil
	}, engine.Capabilities{SupportsTTS: true, SupportsDictionary: false})
}
