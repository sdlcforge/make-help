package cli

import (
	"os"

	"golang.org/x/term"
)

// IsTerminal returns true if the given file descriptor refers to a terminal.
func IsTerminal(fd uintptr) bool {
	return term.IsTerminal(int(fd))
}

// ResolveColorMode determines whether to use colored output based on the config.
// It respects the ColorMode setting and checks if stdout is a terminal.
func ResolveColorMode(config *Config) bool {
	switch config.ColorMode {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	case ColorAuto:
		// Auto-detect based on whether stdout is a terminal
		return IsTerminal(os.Stdout.Fd())
	default:
		return false
	}
}
