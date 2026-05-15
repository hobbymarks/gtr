package bing

import (
	"github.com/hobbymarks/gtr/internal/engine"
	"github.com/hobbymarks/gtr/internal/httpx"
)

func init() {
	engine.Register("bing", func() (engine.Engine, error) {
		return New(httpx.NewSharedClient()), nil
	}, engine.Capabilities{SupportsTTS: false, SupportsDictionary: true})
}
