package lint

import (
	"fmt"
	"regexp"
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

		// Skip if target is a generated help target
		if ctx.GeneratedHelpTargets[targetName] {
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
			File:      ctx.MakefilePath,
			Line:      0, // Line number not available from discovery
			Severity:  SeverityWarning,
			CheckName: "undocumented-phony",
			Message:   fmt.Sprintf("undocumented phony target '%s'", targetName),
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
					File:      target.SourceFile,
					Line:      target.LineNumber,
					Severity:  SeverityWarning,
					CheckName: "summary-punctuation",
					Message:   fmt.Sprintf("summary for '%s' does not end with punctuation", target.Name),
					Context:   summary,
				})
			}
		}
	}

	return warnings
}

// CheckOrphanAliases checks for !alias directives that point to non-existent targets.
// An alias is considered orphaned if it refers to a target that doesn't exist
// in the discovered targets (either as a documented target or a phony target).
func CheckOrphanAliases(ctx *CheckContext) []Warning {
	var warnings []Warning

	// Build a set of all known targets (documented + all phony targets)
	allTargets := make(map[string]bool)

	// Add all documented targets
	for targetName := range ctx.DocumentedTargets {
		allTargets[targetName] = true
	}

	// Add all phony targets (discovered by make)
	for targetName := range ctx.PhonyTargets {
		allTargets[targetName] = true
	}

	// Add all targets that have recipes (file targets)
	for targetName := range ctx.HasRecipe {
		allTargets[targetName] = true
	}

	// Check each target's aliases
	for _, category := range ctx.HelpModel.Categories {
		for _, target := range category.Targets {
			for _, alias := range target.Aliases {
				// Check if the alias points to a known target
				if !allTargets[alias] {
					warnings = append(warnings, Warning{
						File:      target.SourceFile,
						Line:      target.LineNumber,
						Severity:  SeverityWarning,
						CheckName: "orphan-alias",
						Message:   fmt.Sprintf("alias '%s' points to non-existent target (referenced by '%s')", alias, target.Name),
						Context:   fmt.Sprintf("!alias %s", alias),
					})
				}
			}
		}
	}

	// Sort warnings by alias name for deterministic output
	sort.Slice(warnings, func(i, j int) bool {
		return warnings[i].Message < warnings[j].Message
	})

	return warnings
}

// CheckLongSummaries checks for target summaries that exceed 80 characters.
// Long summaries make help output harder to read.
func CheckLongSummaries(ctx *CheckContext) []Warning {
	var warnings []Warning
	const maxLength = 80

	for _, category := range ctx.HelpModel.Categories {
		for _, target := range category.Targets {
			summary := strings.TrimSpace(target.Summary)
			if summary == "" {
				continue
			}

			// Check if summary exceeds maximum length
			if len(summary) > maxLength {
				warnings = append(warnings, Warning{
					File:      target.SourceFile,
					Line:      target.LineNumber,
					Severity:  SeverityWarning,
					CheckName: "long-summary",
					Message:   fmt.Sprintf("summary for '%s' is too long (%d characters, max %d)", target.Name, len(summary), maxLength),
					Context:   summary,
				})
			}
		}
	}

	return warnings
}

// CheckEmptyDocumentation checks for empty documentation lines at the beginning or end of documentation.
// Internal blank lines (between paragraphs) are acceptable.
func CheckEmptyDocumentation(ctx *CheckContext) []Warning {
	var warnings []Warning

	for _, category := range ctx.HelpModel.Categories {
		for _, target := range category.Targets {
			docs := target.Documentation
			if len(docs) == 0 {
				continue
			}

			// Check first line for empty/whitespace-only content
			if strings.TrimSpace(docs[0]) == "" {
				warnings = append(warnings, Warning{
					File:      target.SourceFile,
					Line:      target.LineNumber,
					Severity:  SeverityWarning,
					CheckName: "empty-doc",
					Message:   fmt.Sprintf("target '%s' has empty documentation line at the beginning", target.Name),
					Context:   "##",
				})
			}

			// Check last line for empty/whitespace-only content
			if len(docs) > 1 && strings.TrimSpace(docs[len(docs)-1]) == "" {
				warnings = append(warnings, Warning{
					File:      target.SourceFile,
					Line:      target.LineNumber,
					Severity:  SeverityWarning,
					CheckName: "empty-doc",
					Message:   fmt.Sprintf("target '%s' has empty documentation line at the end", target.Name),
					Context:   "##",
				})
			}
		}
	}

	return warnings
}

// CheckMissingVarDescriptions checks for !var directives that lack descriptions.
// Variables should have descriptive documentation to explain their purpose and default values.
func CheckMissingVarDescriptions(ctx *CheckContext) []Warning {
	var warnings []Warning

	for _, category := range ctx.HelpModel.Categories {
		for _, target := range category.Targets {
			for _, variable := range target.Variables {
				// Check if variable has an empty description
				if strings.TrimSpace(variable.Description) == "" {
					warnings = append(warnings, Warning{
						File:      target.SourceFile,
						Line:      target.LineNumber,
						Severity:  SeverityWarning,
						CheckName: "missing-var-desc",
						Message:   fmt.Sprintf("variable '%s' in target '%s' is missing a description", variable.Name, target.Name),
					})
				}
			}
		}
	}

	return warnings
}

// kebabCasePattern matches valid kebab-case names.
// Valid format: lowercase letters and numbers separated by hyphens.
// Examples: build, test, build-all, run-tests, docker-build-image
var kebabCasePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// CheckInconsistentNaming checks that all target names follow kebab-case convention.
// Kebab-case means lowercase letters and numbers separated by hyphens.
func CheckInconsistentNaming(ctx *CheckContext) []Warning {
	var warnings []Warning

	for _, category := range ctx.HelpModel.Categories {
		for _, target := range category.Targets {
			if !kebabCasePattern.MatchString(target.Name) {
				warnings = append(warnings, Warning{
					File:      target.SourceFile,
					Line:      target.LineNumber,
					Severity:  SeverityWarning,
					CheckName: "naming",
					Message:   fmt.Sprintf("target '%s' does not follow kebab-case naming convention", target.Name),
					Context:   target.Name,
				})
			}
		}
	}

	return warnings
}

// CheckCircularAliases detects implicit alias chains that form loops.
// An implicit alias is a phony target with a single phony dependency and no recipe.
// For example: a → b → c → a creates a circular dependency chain.
func CheckCircularAliases(ctx *CheckContext) []Warning {
	var warnings []Warning

	// Build a map of implicit alias relationships
	// A target is an implicit alias if:
	// 1. It's a .PHONY target
	// 2. It has exactly one dependency
	// 3. That dependency is also a .PHONY target
	// 4. The target has no recipe
	implicitAliases := make(map[string]string)

	for targetName, isPhony := range ctx.PhonyTargets {
		if !isPhony {
			continue
		}

		// Check if target has a recipe
		if ctx.HasRecipe[targetName] {
			continue
		}

		// Check if target has exactly one dependency
		deps := ctx.Dependencies[targetName]
		if len(deps) != 1 {
			continue
		}

		// Check if the dependency is also phony
		dependencyName := deps[0]
		if !ctx.PhonyTargets[dependencyName] {
			continue
		}

		// This is an implicit alias
		implicitAliases[targetName] = dependencyName
	}

	// Detect cycles using DFS
	// Track visited nodes and nodes in current path
	visited := make(map[string]bool)
	inPath := make(map[string]bool)
	cycles := make(map[string][]string) // Map cycle start to full cycle path

	var dfs func(node string, path []string) bool
	dfs = func(node string, path []string) bool {
		if inPath[node] {
			// Found a cycle - extract the cycle portion
			cycleStart := -1
			for i, n := range path {
				if n == node {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cycle := append([]string{}, path[cycleStart:]...)
				cycle = append(cycle, node) // Complete the cycle
				// Use the first node in the cycle as the key to avoid duplicates
				cycleKey := cycle[0]
				if _, exists := cycles[cycleKey]; !exists {
					cycles[cycleKey] = cycle
				}
			}
			return true
		}

		if visited[node] {
			return false
		}

		visited[node] = true
		inPath[node] = true
		path = append(path, node)

		// Follow the alias chain
		if next, isAlias := implicitAliases[node]; isAlias {
			dfs(next, path)
		}

		inPath[node] = false
		return false
	}

	// Run DFS from each implicit alias
	for targetName := range implicitAliases {
		if !visited[targetName] {
			dfs(targetName, []string{})
		}
	}

	// Create warnings for each unique cycle
	// Sort cycle keys for deterministic output
	var cycleKeys []string
	for key := range cycles {
		cycleKeys = append(cycleKeys, key)
	}
	sort.Strings(cycleKeys)

	for _, key := range cycleKeys {
		cycle := cycles[key]
		cycleStr := strings.Join(cycle, " → ")
		warnings = append(warnings, Warning{
			File:      ctx.MakefilePath,
			Line:      0, // Line number not available from discovery
			Severity:  SeverityWarning,
			CheckName: "circular-alias",
			Message:   fmt.Sprintf("circular alias chain detected: %s", cycleStr),
		})
	}

	return warnings
}
