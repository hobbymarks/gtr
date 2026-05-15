package apertium

import (
	"github.com/hobbymarks/gtr/internal/engine"
	"github.com/hobbymarks/gtr/internal/httpx"
)

func init() {
	engine.Register("apertium", func() (engine.Engine, error) {
		return New(httpx.NewClient()), nil
	}, engine.Capabilities{SupportsTTS: false, SupportsDictionary: false})
}
