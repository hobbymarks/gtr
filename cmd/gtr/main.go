// Package main is the entry point for the gtr multi-engine translation CLI.
package main

import (
	"os"
	"runtime/debug"
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
	if version != "dev" {
		if commit != "" {
			if version == commit || strings.HasPrefix(version, commit+"-") {
				return version
			}
			return version + "-" + commit
		}
		return version
	}
	if commit != "" {
		return commit
	}
	if v := strings.TrimSpace(os.Getenv("GTR_VERSION")); v != "" {
		return v
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return "dev"
}

func main() {
	cli.Version = versionString()
	cli.Main()
}
