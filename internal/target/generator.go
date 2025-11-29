package target

import (
	"strings"
)

// generateHelpTarget creates help target content with flag pass-through.
// The generated target invokes make-help with all relevant configuration flags.
func generateHelpTarget(config *Config) string {
	var buf strings.Builder

	buf.WriteString(".PHONY: help\n")
	buf.WriteString("help:\n")
	buf.WriteString("\t@make-help")

	// Add flags from config
	if config.KeepOrderCategories {
		buf.WriteString(" --keep-order-categories")
	}
	if config.KeepOrderTargets {
		buf.WriteString(" --keep-order-targets")
	}
	if len(config.CategoryOrder) > 0 {
		buf.WriteString(" --category-order ")
		buf.WriteString(strings.Join(config.CategoryOrder, ","))
	}
	if config.DefaultCategory != "" {
		buf.WriteString(" --default-category ")
		buf.WriteString(config.DefaultCategory)
	}

	buf.WriteString("\n")

	return buf.String()
}
