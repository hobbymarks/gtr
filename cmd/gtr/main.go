package main

import "github.com/ueki/gtr/internal/cli"

// version is overridden via -ldflags '-X main.version=...' at link time.
var version = "dev"

func main() {
	cli.Version = version
	cli.Main()
}
