package yandex

import (
	"github.com/hobbymarks/gtr/internal/engine"
	"github.com/hobbymarks/gtr/internal/httpx"
)

func init() {
	// Dictionary path exists in translate-shell but is disabled (FIXME upstream).
	engine.Register("yandex", func() (engine.Engine, error) {
		return New(httpx.NewSharedClient()), nil
	}, engine.Capabilities{SupportsTTS: false, SupportsDictionary: false})
}
