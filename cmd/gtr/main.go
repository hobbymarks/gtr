package main

import (
	"github.com/ueki/gtr/internal/cli"

	_ "github.com/ueki/gtr/internal/engine/auto"   // register auto
	_ "github.com/ueki/gtr/internal/engine/bing"   // register bing
	_ "github.com/ueki/gtr/internal/engine/google" // register google
)

// version is overridden via -ldflags '-X main.version=...' at link time.
var version = "dev"

func main() {
	cli.Version = version
	cli.Main()
}
