package discovery

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// DiscoverTargetsResult contains discovered targets and their metadata.
type DiscoverTargetsResult struct {
	// Targets contains all discovered target names in order.
	Targets []string

	// IsPhony maps target names to their .PHONY status.
	IsPhony map[string]bool

	// Dependencies maps target names to their prerequisite targets.
	Dependencies map[string][]string

	// HasRecipe maps target names to whether they have a recipe (commands).
	HasRecipe map[string]bool
}

// discoverTargets extracts all targets from make -p output.
// It executes make -p -r to get the database output and parses target names.
func (s *Service) discoverTargets(makefilePath string) (*DiscoverTargetsResult, error) {
	// Execute make with timeout to prevent indefinite hangs
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use -s and --no-print-directory to prevent make from adding
	// extra output when running from within another make.
	// Pass MAKE_HELP_GENERATING=1 to prevent auto-regeneration of help.mk
	// which would cause infinite recursion (make-help -> make -> make-help -> ...)
	stdout, stderr, err := s.executor.ExecuteContext(ctx, "make", "-s", "--no-print-directory", "-f", makefilePath, "-p", "-r", "MAKE_HELP_GENERATING=1")
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("make command timed out after 30s")
		}
		// Empty Makefiles cause "No targets" error - this is acceptable
		if strings.Contains(stderr, "No targets") {
			return &DiscoverTargetsResult{
				Targets: []string{},
				IsPhony: map[string]bool{},
			}, nil
		}
		return nil, fmt.Errorf("failed to discover targets: %w\nstderr: %s", err, stderr)
	}

	result := parseTargetsFromDatabase(stdout)

	if s.verbose {
		fmt.Printf("Discovered %d target(s) from make database\n", len(result.Targets))
	}

	return result, nil
}

// parseTargetsFromDatabase extracts target names, .PHONY status, dependencies,
// and recipe presence from make -p output.
// It filters out comments, whitespace-prefixed lines, and built-in targets.
func parseTargetsFromDatabase(output string) *DiscoverTargetsResult {
	var targets []string
	seen := make(map[string]bool)
	isPhony := make(map[string]bool)
	dependencies := make(map[string][]string)
	hasRecipe := make(map[string]bool)

	// Match target definitions: <target>: [deps...] or <target>:: [deps...]
	// Captures: 1=target name, 2=everything after the colon(s)
	targetRegex := regexp.MustCompile(`^([a-zA-Z0-9_/.@%+-][a-zA-Z0-9_/.@%+-]*)\s*::?\s*(.*)$`)

	// Track current target for recipe detection
	var currentTarget string

	lines := strings.Split(output, "\n")
	for i, line := range lines {
		// Parse .PHONY declarations
		if strings.HasPrefix(line, ".PHONY:") {
			// Extract all targets from .PHONY line
			phonyLine := strings.TrimPrefix(line, ".PHONY:")
			phonyTargets := strings.Fields(phonyLine)
			for _, target := range phonyTargets {
				isPhony[target] = true
			}
			continue
		}

		// Check for "recipe to execute" indicator for the current target
		if currentTarget != "" && strings.Contains(line, "recipe to execute") {
			hasRecipe[currentTarget] = true
			continue
		}

		// Skip comments (but we already checked for recipe indicator)
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Skip whitespace-prefixed lines (recipe lines, continuations)
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}

		// Extract target name and dependencies
		if matches := targetRegex.FindStringSubmatch(line); matches != nil {
			targetName := matches[1]
			depsStr := strings.TrimSpace(matches[2])

			// Skip special/built-in targets
			if isSpecialTarget(targetName) {
				currentTarget = ""
				continue
			}

			// Update current target for recipe detection
			currentTarget = targetName

			// Skip if already seen (avoid duplicates)
			if seen[targetName] {
				continue
			}

			targets = append(targets, targetName)
			seen[targetName] = true

			// Parse dependencies (space-separated)
			if depsStr != "" {
				deps := strings.Fields(depsStr)
				// Filter out special targets from dependencies
				var filteredDeps []string
				for _, dep := range deps {
					if !isSpecialTarget(dep) {
						filteredDeps = append(filteredDeps, dep)
					}
				}
				if len(filteredDeps) > 0 {
					dependencies[targetName] = filteredDeps
				}
			}
		} else {
			// Non-target line, but check if we're still in a target block
			// Empty line or other content might indicate end of target block
			// We keep currentTarget set to continue looking for recipe indicator
			_ = i // suppress unused warning
		}
	}

	return &DiscoverTargetsResult{
		Targets:      targets,
		IsPhony:      isPhony,
		Dependencies: dependencies,
		HasRecipe:    hasRecipe,
	}
}

// isSpecialTarget returns true if the target is a special or built-in Make target.
func isSpecialTarget(name string) bool {
	// Skip Make's special targets
	specialTargets := map[string]bool{
		".SUFFIXES":      true,
		".DEFAULT":       true,
		".PRECIOUS":      true,
		".INTERMEDIATE":  true,
		".SECONDARY":     true,
		".SECONDEXPANSION": true,
		".DELETE_ON_ERROR": true,
		".IGNORE":        true,
		".LOW_RESOLUTION_TIME": true,
		".SILENT":        true,
		".EXPORT_ALL_VARIABLES": true,
		".NOTPARALLEL":   true,
		".ONESHELL":      true,
		".POSIX":         true,
		"Makefile":       true,
		"makefile":       true,
	}

	// Check if it's a known special target
	if specialTargets[name] {
		return true
	}

	// Skip pattern rules (contain %)
	if strings.Contains(name, "%") {
		return true
	}

	// Skip variable assignments that look like targets (contain =)
	if strings.Contains(name, "=") {
		return true
	}

	return false
}
