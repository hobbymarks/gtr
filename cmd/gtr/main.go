// Package main is the entry point for the gtr multi-engine translation CLI.
package main

import (
	"os"
	"strings"

	"github.com/hobbymarks/gtr/internal/cli"

	_ "github.com/hobbymarks/gtr/internal/engine/apertium" // register apertium
	_ "github.com/hobbymarks/gtr/internal/engine/auto"     // register auto
	_ "github.com/hobbymarks/gtr/internal/engine/bing"     // register bing
	_ "github.com/hobbymarks/gtr/internal/engine/google"   // register google
	_ "github.com/hobbymarks/gtr/internal/engine/spell"    // register spell / aspell / hunspell
	_ "github.com/hobbymarks/gtr/internal/engine/yandex"   // register yandex
)

var (
	// version is overridden via -ldflags '-X main.version=...' at link time.
	version = "dev"
	// commit is overridden via -ldflags '-X main.commit=...' at link time.
	commit = ""
)

func versionString() string {
	if commit != "" {
		if version == commit || strings.HasPrefix(version, commit+"-") {
			return version
		}
		return version + "-" + commit
	}
	if v := strings.TrimSpace(os.Getenv("GTR_VERSION")); v != "" {
		return v
	}
	return version
}

func main() {
	cli.Version = versionString()
	cli.Main()
}
