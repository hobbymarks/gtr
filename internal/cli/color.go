package cli

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"

	envNoColor = "NO_COLOR"
)

// UseColor reports whether color output should be used (TTY + !--no-color + !NO_COLOR).
var UseColor = true

func initColorOut(noColorFlag bool) {
	if noColorFlag || os.Getenv(envNoColor) != "" || !term.IsTerminal(int(os.Stdout.Fd())) {
		UseColor = false
	}
}

// Bold wraps s in bold ANSI if color is enabled.
func Bold(s string) string {
	if !UseColor {
		return s
	}
	return colorBold + s + colorReset
}

// Green wraps s in green ANSI if color is enabled.
func Green(s string) string {
	if !UseColor {
		return s
	}
	return colorGreen + s + colorReset
}

// Yellow wraps s in yellow ANSI if color is enabled.
func Yellow(s string) string {
	if !UseColor {
		return s
	}
	return colorYellow + s + colorReset
}

// Cyan wraps s in cyan ANSI if color is enabled.
func Cyan(s string) string {
	if !UseColor {
		return s
	}
	return colorCyan + s + colorReset
}

// ColorHeading formats a heading line.
func ColorHeading(format string, a ...interface{}) string {
	return Bold(fmt.Sprintf(format, a...))
}
