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
		// Look up location from parsed files if available
		file := ctx.MakefilePath
		line := 0
		if loc, ok := ctx.TargetLocations[targetName]; ok {
			file = loc.File
			line = loc.Line
		}

		warnings = append(warnings, Warning{
			File:      file,
			Line:      line,
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
			// Skip if no documentation
			if len(target.Documentation) == 0 {
				continue
			}

			// The summary is the first documentation line
			firstDocLine := strings.TrimSpace(target.Documentation[0])
			if firstDocLine == "" {
				continue
			}

			// Check if first doc line ends with proper punctuation
			lastChar := firstDocLine[len(firstDocLine)-1]
			if lastChar != '.' && lastChar != '!' && lastChar != '?' {
				// Calculate the line number of the first doc line
				// Docs are directly above the target, so first doc is at:
				// target.LineNumber - len(Documentation)
				docLineNumber := target.LineNumber - len(target.Documentation)

				warnings = append(warnings, Warning{
					File:      target.SourceFile,
					Line:      docLineNumber,
					Severity:  SeverityWarning,
					CheckName: "summary-punctuation",
					Message:   fmt.Sprintf("summary for '%s' does not end with punctuation", target.Name),
					Context:   "## " + firstDocLine, // Store full line for fix validation
				})
			}
		}
	}

	return warnings
}

// fixSummaryPunctuation generates a fix for a summary-punctuation warning.
// It appends a period to the end of the first documentation line.
func fixSummaryPunctuation(w Warning) *Fix {
	if w.Context == "" {
		return nil // Can't fix without context
	}

	return &Fix{
		File:       w.File,
		Line:       w.Line,
		Operation:  FixReplace,
		OldContent: w.Context,
		NewContent: w.Context + ".",
	}
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

			// Calculate the first doc line number
			// Docs are directly above the target, so first doc is at:
			// target.LineNumber - len(Documentation)
			firstDocLine := target.LineNumber - len(docs)

			// Check first line for empty/whitespace-only content
			if strings.TrimSpace(docs[0]) == "" {
				warnings = append(warnings, Warning{
					File:      target.SourceFile,
					Line:      firstDocLine,
					Severity:  SeverityWarning,
					CheckName: "empty-doc",
					Message:   fmt.Sprintf("target '%s' has empty documentation line at the beginning", target.Name),
					Context:   "##", // Empty doc line content for fix validation
				})
			}

			// Check last line for empty/whitespace-only content
			if len(docs) > 1 && strings.TrimSpace(docs[len(docs)-1]) == "" {
				// Last doc line is at: target.LineNumber - 1
				lastDocLine := target.LineNumber - 1
				warnings = append(warnings, Warning{
					File:      target.SourceFile,
					Line:      lastDocLine,
					Severity:  SeverityWarning,
					CheckName: "empty-doc",
					Message:   fmt.Sprintf("target '%s' has empty documentation line at the end", target.Name),
					Context:   "##", // Empty doc line content for fix validation
				})
			}
		}
	}

	return warnings
}

// fixEmptyDocumentation generates a fix for an empty-doc warning.
// It deletes the empty documentation line.
func fixEmptyDocumentation(w Warning) *Fix {
	return &Fix{
		File:       w.File,
		Line:       w.Line,
		Operation:  FixDelete,
		OldContent: "##", // Validate the line is an empty comment
	}
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

// CheckCircularDependencies detects circular dependency chains in targets.
// Uses the dependency graph from `make -p` to detect cycles.
// For example: a → b → c → a creates a circular dependency chain.
func CheckCircularDependencies(ctx *CheckContext) []Warning {
	var warnings []Warning

	// Detect cycles using DFS on the actual dependency graph
	// Track visited nodes and nodes in current path
	visited := make(map[string]bool)
	inPath := make(map[string]bool)
	cycles := make(map[string][]string) // Map cycle start to full cycle path

	var dfs func(node string, path []string)
	dfs = func(node string, path []string) {
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
				// Use the lexicographically smallest node as key to avoid duplicate reports
				minNode := cycle[0]
				for _, n := range cycle {
					if n < minNode {
						minNode = n
					}
				}
				if _, exists := cycles[minNode]; !exists {
					cycles[minNode] = cycle
				}
			}
			return
		}

		if visited[node] {
			return
		}

		visited[node] = true
		inPath[node] = true
		path = append(path, node)

		// Follow all dependencies
		for _, dep := range ctx.Dependencies[node] {
			dfs(dep, path)
		}

		inPath[node] = false
	}

	// Run DFS from each target with dependencies
	for targetName := range ctx.Dependencies {
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
			CheckName: "circular-dependency",
			Message:   fmt.Sprintf("circular dependency chain detected: %s", cycleStr),
		})
	}

	return warnings
}

// CheckRedundantDirectives detects redundant or ineffective !notalias and !alias directives.
// A !notalias is redundant when the target wouldn't be an implicit alias anyway:
// - Target has documentation (documented targets are never implicit aliases)
// - Target has a recipe (targets with recipes are never implicit aliases)
// - Target is not .PHONY (non-phony targets are never implicit aliases)
// - Target has multiple dependencies (only single-dep targets can be aliases)
func CheckRedundantDirectives(ctx *CheckContext) []Warning {
	var warnings []Warning

	// Check each target marked with !notalias
	for targetName := range ctx.NotAliasTargets {
		loc := ctx.TargetLocations[targetName]

		// Check if target has documentation
		if ctx.DocumentedTargets[targetName] {
			warnings = append(warnings, Warning{
				File:      loc.File,
				Line:      loc.Line,
				Severity:  SeverityWarning,
				CheckName: "redundant-notalias",
				Message:   fmt.Sprintf("!notalias on '%s' is redundant: documented targets are never implicit aliases", targetName),
				Fixable:   true,
			})
			continue
		}

		// Check if target has a recipe
		if ctx.HasRecipe[targetName] {
			warnings = append(warnings, Warning{
				File:      loc.File,
				Line:      loc.Line,
				Severity:  SeverityWarning,
				CheckName: "redundant-notalias",
				Message:   fmt.Sprintf("!notalias on '%s' is redundant: targets with recipes are never implicit aliases", targetName),
				Fixable:   true,
			})
			continue
		}

		// Check if target is not .PHONY
		if !ctx.PhonyTargets[targetName] {
			warnings = append(warnings, Warning{
				File:      loc.File,
				Line:      loc.Line,
				Severity:  SeverityWarning,
				CheckName: "redundant-notalias",
				Message:   fmt.Sprintf("!notalias on '%s' is redundant: non-phony targets are never implicit aliases", targetName),
				Fixable:   true,
			})
			continue
		}

		// Check if target has multiple dependencies
		deps := ctx.Dependencies[targetName]
		if len(deps) != 1 {
			warnings = append(warnings, Warning{
				File:      loc.File,
				Line:      loc.Line,
				Severity:  SeverityWarning,
				CheckName: "redundant-notalias",
				Message:   fmt.Sprintf("!notalias on '%s' is redundant: only targets with exactly one dependency can be implicit aliases", targetName),
				Fixable:   true,
			})
			continue
		}

		// Check if the single dependency is not .PHONY
		depName := deps[0]
		if !ctx.PhonyTargets[depName] {
			warnings = append(warnings, Warning{
				File:      loc.File,
				Line:      loc.Line,
				Severity:  SeverityWarning,
				CheckName: "redundant-notalias",
				Message:   fmt.Sprintf("!notalias on '%s' is redundant: its dependency '%s' is not phony, so it can't be an implicit alias", targetName, depName),
				Fixable:   true,
			})
		}
	}

	// Check for self-referencing aliases in explicit aliases
	for _, cat := range ctx.HelpModel.Categories {
		for _, target := range cat.Targets {
			for _, alias := range target.Aliases {
				if alias == target.Name {
					loc := ctx.TargetLocations[target.Name]
					warnings = append(warnings, Warning{
						File:      loc.File,
						Line:      loc.Line,
						Severity:  SeverityWarning,
						CheckName: "redundant-alias",
						Message:   fmt.Sprintf("target '%s' has itself as an alias", target.Name),
						Fixable:   true,
					})
				}
			}
		}
	}

	return warnings
}

// AllChecks returns all available lint checks.
func AllChecks() []Check {
	return []Check{
		{Name: "undocumented-phony", CheckFunc: CheckUndocumentedPhony, FixFunc: nil},
		{Name: "summary-punctuation", CheckFunc: CheckSummaryPunctuation, FixFunc: fixSummaryPunctuation},
		{Name: "orphan-alias", CheckFunc: CheckOrphanAliases, FixFunc: nil},
		{Name: "long-summary", CheckFunc: CheckLongSummaries, FixFunc: nil},
		{Name: "empty-doc", CheckFunc: CheckEmptyDocumentation, FixFunc: fixEmptyDocumentation},
		{Name: "missing-var-desc", CheckFunc: CheckMissingVarDescriptions, FixFunc: nil},
		{Name: "naming", CheckFunc: CheckInconsistentNaming, FixFunc: nil},
		{Name: "circular-dependency", CheckFunc: CheckCircularDependencies, FixFunc: nil},
		{Name: "redundant-notalias", CheckFunc: CheckRedundantDirectives, FixFunc: nil},
	}
}
