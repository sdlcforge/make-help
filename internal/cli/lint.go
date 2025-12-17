package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sdlcforge/make-help/internal/discovery"
	"github.com/sdlcforge/make-help/internal/lint"
	"github.com/sdlcforge/make-help/internal/model"
	"github.com/sdlcforge/make-help/internal/parser"
	"github.com/sdlcforge/make-help/internal/summary"
)

// ErrLintWarningsFound is a sentinel error returned when lint warnings are found.
// Cobra will translate this into exit code 1.
var ErrLintWarningsFound = errors.New("lint warnings found")

// runLint performs static analysis on Makefiles and their documentation.
// It orchestrates the following steps:
//  1. Discovery - Find Makefile and all included files
//  2. Discovery - Get all targets with .PHONY status
//  3. Parsing - Extract documentation directives
//  4. Building - Construct the help model
//  5. Lint - Run all lint checks
//  6. Output - Display warnings
//
// Exit codes:
//   0 - No warnings
//   1 - Warnings found
//   2 - Error (invalid flags, file not found, etc.)
func runLint(config *Config) error {
	// Check for recursion: prevent make-help from running if we're already in a make-help process
	if os.Getenv("MAKE_HELP_GENERATING") == "1" {
		return fmt.Errorf("recursion detected: make-help was invoked from within a make process spawned by make-help")
	}

	// Step 1: Resolve and validate Makefile path
	makefilePath, err := discovery.ResolveMakefilePath(config.MakefilePath)
	if err != nil {
		return fmt.Errorf("failed to resolve Makefile path: %w", err)
	}

	if err := discovery.ValidateMakefileExists(makefilePath); err != nil {
		return err
	}

	config.MakefilePath = makefilePath

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "Using Makefile: %s\n", makefilePath)
	}

	// Step 2: Discover all Makefiles (main + included)
	discoveryService := discovery.NewService(discovery.NewDefaultExecutor(), config.Verbose)

	makefiles, err := discoveryService.DiscoverMakefiles(makefilePath)
	if err != nil {
		return fmt.Errorf("failed to discover Makefiles: %w", err)
	}

	// Step 3: Parse all Makefiles
	scanner := parser.NewScanner()
	var parsedFiles []*parser.ParsedFile

	for _, mf := range makefiles {
		parsed, err := scanner.ScanFile(mf)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", mf, err)
		}
		parsedFiles = append(parsedFiles, parsed)
	}

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "Parsed %d Makefile(s)\n", len(parsedFiles))
	}

	// Step 4: Discover targets with .PHONY status, dependencies, and recipes
	targetsResult, err := discoveryService.DiscoverTargets(makefilePath)
	if err != nil {
		return fmt.Errorf("failed to discover targets: %w", err)
	}

	// Step 5: Build the help model
	// For lint mode, we don't want to include undocumented targets
	builderConfig := &model.BuilderConfig{
		DefaultCategory: config.DefaultCategory,
		IncludeTargets:  []string{},
		IncludeAllPhony: false,
		PhonyTargets:    targetsResult.IsPhony,
		Dependencies:    targetsResult.Dependencies,
		HasRecipe:       targetsResult.HasRecipe,
	}
	builder := model.NewBuilder(builderConfig)
	helpModel, err := builder.Build(parsedFiles)
	if err != nil {
		return fmt.Errorf("failed to build help model: %w", err)
	}

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "Built help model with %d category/categories\n", len(helpModel.Categories))
	}

	// Step 6: Extract summaries for all targets
	extractor := summary.NewExtractor()
	for i := range helpModel.Categories {
		for j := range helpModel.Categories[i].Targets {
			target := &helpModel.Categories[i].Targets[j]
			target.Summary = extractor.Extract(target.Documentation)
		}
	}

	// Step 7: Build CheckContext
	documentedTargets := make(map[string]bool)
	aliases := make(map[string]bool)
	generatedHelpTargets := make(map[string]bool)
	targetLocations := make(map[string]lint.TargetLocation)

	// Build target locations from parsed files
	for _, pf := range parsedFiles {
		for targetName, lineNum := range pf.TargetMap {
			targetLocations[targetName] = lint.TargetLocation{
				File: pf.Path,
				Line: lineNum,
			}
		}
	}

	// Add the standard generated help targets
	generatedHelpTargets["help"] = true
	generatedHelpTargets["update-help"] = true

	for _, category := range helpModel.Categories {
		for _, target := range category.Targets {
			documentedTargets[target.Name] = true
			// Add help-<target> as a generated target
			generatedHelpTargets["help-"+target.Name] = true
			for _, alias := range target.Aliases {
				aliases[alias] = true
			}
		}
	}

	checkCtx := &lint.CheckContext{
		HelpModel:            helpModel,
		MakefilePath:         makefilePath,
		PhonyTargets:         targetsResult.IsPhony,
		Dependencies:         targetsResult.Dependencies,
		HasRecipe:            targetsResult.HasRecipe,
		DocumentedTargets:    documentedTargets,
		Aliases:              aliases,
		GeneratedHelpTargets: generatedHelpTargets,
		TargetLocations:      targetLocations,
		NotAliasTargets:      builder.NotAliasTargets(),
	}

	// Step 8: Run all lint checks
	checks := lint.AllChecks()
	result := lint.Lint(checkCtx, checks)

	// Step 9: Apply fixes if --fix is set (before displaying warnings)
	var fixResult *lint.FixResult
	fixableCount := 0
	for _, w := range result.Warnings {
		if w.Fixable {
			fixableCount++
		}
	}

	if config.Fix && fixableCount > 0 {
		fixes := lint.CollectFixes(checks, result.Warnings)

		fixer := &lint.Fixer{DryRun: config.DryRun}
		var err error
		fixResult, err = fixer.ApplyFixes(fixes)
		if err != nil {
			return fmt.Errorf("failed to apply fixes: %w", err)
		}
	}

	// Step 10: Determine which warnings to display
	// If fixes were applied (not dry-run), filter out fixed warnings
	warningsToDisplay := result.Warnings
	if fixResult != nil && !config.DryRun && fixResult.TotalFixed > 0 {
		// Filter out fixable warnings that were fixed
		var remaining []lint.Warning
		for _, w := range result.Warnings {
			if !w.Fixable {
				remaining = append(remaining, w)
			}
		}
		warningsToDisplay = remaining
	}

	// Step 11: Output warnings
	if len(warningsToDisplay) > 0 {
		// Get current working directory for relative paths
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "" // Fall back to absolute paths if we can't get cwd
		}

		// Count fixable warnings in displayed set
		displayFixableCount := 0
		for _, w := range warningsToDisplay {
			if w.Fixable {
				displayFixableCount++
			}
		}

		// Group warnings by file
		var currentFile string
		for _, warning := range warningsToDisplay {
			// Convert to relative path if possible
			displayPath := warning.File
			if cwd != "" {
				if rel, err := filepath.Rel(cwd, warning.File); err == nil {
					displayPath = rel
				}
			}

			// Print file header when file changes
			if warning.File != currentFile {
				if currentFile != "" {
					fmt.Println() // Blank line between files
				}
				fmt.Println(displayPath)
				currentFile = warning.File
			}

			// Print warning: "line: message [fixable]"
			fixableTag := ""
			if warning.Fixable {
				fixableTag = " [fixable]"
			}
			if warning.Line > 0 {
				fmt.Printf("  %d: %s%s\n", warning.Line, warning.Message, fixableTag)
			} else {
				fmt.Printf("  %s%s\n", warning.Message, fixableTag)
			}
		}

		// Summary line
		count := len(warningsToDisplay)
		fmt.Println()
		if displayFixableCount > 0 {
			fmt.Printf("Found %d warning(s) (%d fixable)\n", count, displayFixableCount)
		} else if count == 1 {
			fmt.Println("Found 1 warning")
		} else {
			fmt.Printf("Found %d warnings\n", count)
		}
	}

	// Step 12: Report fix results
	if fixResult != nil {
		if len(warningsToDisplay) > 0 {
			fmt.Println()
		}
		if config.DryRun {
			fmt.Printf("Would fix %d issue(s) in %d file(s)\n",
				fixResult.TotalFixed, len(fixResult.FilesModified))
		} else {
			fmt.Printf("Fixed %d issue(s) in %d file(s)\n",
				fixResult.TotalFixed, len(fixResult.FilesModified))
		}
	}

	// Step 13: Determine exit code
	// If there are remaining warnings (unfixed), return error (exit code 1)
	if len(warningsToDisplay) > 0 {
		return ErrLintWarningsFound
	}

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "No warnings found\n")
	}

	return nil
}
