package main

import (
	"github.com/ueki/gtr/internal/cli"

	_ "github.com/ueki/gtr/internal/engine/google" // register engines
)

// version is overridden via -ldflags '-X main.version=...' at link time.
var version = "dev"

func main() {
	cli.Version = version
	cli.Main()
}
