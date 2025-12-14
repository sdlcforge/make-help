package lint

import (
	"fmt"
	"sort"
	"strings"
)

// CheckUndocumentedPhony checks for .PHONY targets that lack documentation.
// It reports warnings for phony targets that are neither documented nor aliases.
func CheckUndocumentedPhony(ctx *CheckContext) []Warning {
	var warnings []Warning

	// Collect undocumented phony target names first
	var undocumentedTargets []string

	for targetName, isPhony := range ctx.PhonyTargets {
		if !isPhony {
			continue
		}

		// Skip if target is documented
		if ctx.DocumentedTargets[targetName] {
			continue
		}

		// Skip if target is an alias (explicit or implicit)
		if ctx.Aliases[targetName] {
			continue
		}

		// This is an undocumented phony target
		undocumentedTargets = append(undocumentedTargets, targetName)
	}

	// Sort target names for deterministic output
	sort.Strings(undocumentedTargets)

	// Create warnings in sorted order
	for _, targetName := range undocumentedTargets {
		warnings = append(warnings, Warning{
			File:     ctx.MakefilePath,
			Line:     0, // Line number not available from discovery
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("undocumented phony target '%s'", targetName),
		})
	}

	return warnings
}

// CheckSummaryPunctuation checks that target summaries end with proper punctuation.
// Valid punctuation: '.', '!', '?'
func CheckSummaryPunctuation(ctx *CheckContext) []Warning {
	var warnings []Warning

	for _, category := range ctx.HelpModel.Categories {
		for _, target := range category.Targets {
			// Extract summary (first sentence)
			summary := strings.TrimSpace(target.Summary)
			if summary == "" {
				continue
			}

			// Check if summary ends with proper punctuation
			lastChar := summary[len(summary)-1]
			if lastChar != '.' && lastChar != '!' && lastChar != '?' {
				warnings = append(warnings, Warning{
					File:     target.SourceFile,
					Line:     target.LineNumber,
					Severity: SeverityWarning,
					Message:  fmt.Sprintf("summary for '%s' does not end with punctuation", target.Name),
					Context:  summary,
				})
			}
		}
	}

	return warnings
}
