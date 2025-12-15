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

// BuildCommandLine builds a command line string from a Config.
// This is used to store the command line in generated help.mk files.
func BuildCommandLine(config *Config) string {
	var parts []string

	parts = append(parts, "make-help")

	// Color flags
	if config.ColorMode == ColorAlways {
		parts = append(parts, "--color")
	} else if config.ColorMode == ColorNever {
		parts = append(parts, "--no-color")
	}

	// Mode flags
	if config.ShowHelp {
		parts = append(parts, "--show-help")
	}
	if config.RemoveHelpTarget {
		parts = append(parts, "--remove-help")
	}
	if config.DryRun {
		parts = append(parts, "--dry-run")
	}
	if config.Lint {
		parts = append(parts, "--lint")
	}
	if config.Fix {
		parts = append(parts, "--fix")
	}

	// Input flags
	if config.MakefilePath != "" {
		parts = append(parts, "--makefile-path", config.MakefilePath)
	}
	if config.HelpFileRelPath != "" {
		parts = append(parts, "--help-file-rel-path", config.HelpFileRelPath)
	}
	if config.Target != "" {
		parts = append(parts, "--target", config.Target)
	}

	// Output/formatting flags
	if config.KeepOrderCategories {
		parts = append(parts, "--keep-order-categories")
	}
	if config.KeepOrderTargets {
		parts = append(parts, "--keep-order-targets")
	}
	if len(config.CategoryOrder) > 0 {
		parts = append(parts, "--category-order", strings.Join(config.CategoryOrder, ","))
	}
	if config.DefaultCategory != "" {
		parts = append(parts, "--default-category", config.DefaultCategory)
	}
	if config.HelpCategory != "" && config.HelpCategory != "Help" {
		parts = append(parts, "--help-category", config.HelpCategory)
	}
	if config.IncludeAllPhony {
		parts = append(parts, "--include-all-phony")
	}
	for _, target := range config.IncludeTargets {
		parts = append(parts, "--include-target", target)
	}

	// Misc flags
	if config.Verbose {
		parts = append(parts, "--verbose")
	}

	return strings.Join(parts, " ")
}
