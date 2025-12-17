package cli

import (
	"os"
	"strings"
)

// HasAnyOptions checks if any command-line options were provided.
// It checks the os.Args for flags (excluding the program name).
func HasAnyOptions() bool {
	if len(os.Args) <= 1 {
		return false
	}

	// Check if any argument looks like a flag
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-") {
			return true
		}
	}

	return false
}
