package google

import (
	"github.com/ueki/gtr/internal/engine"
	"github.com/ueki/gtr/internal/httpx"
)

func init() {
	engine.Register("google", func() (engine.Engine, error) {
		return New(httpx.NewClient()), nil
	})
}
