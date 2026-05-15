package google

import (
	"github.com/hobbymarks/gtr/internal/engine"
	"github.com/hobbymarks/gtr/internal/httpx"
)

func init() {
	engine.Register("google", func() (engine.Engine, error) {
		return New(httpx.NewClient()), nil
	}, engine.Capabilities{SupportsTTS: true, SupportsDictionary: true})
}
