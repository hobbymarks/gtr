package main

import (
	"github.com/ueki/gtr/internal/cli"

	_ "github.com/ueki/gtr/internal/engine/apertium" // register apertium
	_ "github.com/ueki/gtr/internal/engine/auto"     // register auto
	_ "github.com/ueki/gtr/internal/engine/bing"     // register bing
	_ "github.com/ueki/gtr/internal/engine/google"   // register google
	_ "github.com/ueki/gtr/internal/engine/spell"    // register spell / aspell / hunspell
	_ "github.com/ueki/gtr/internal/engine/yandex"   // register yandex
)

// version is overridden via -ldflags '-X main.version=...' at link time.
var version = "dev"

func main() {
	cli.Version = version
	cli.Main()
}
