package cli

import (
	"errors"
	"fmt"
	"os"

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

	for _, category := range helpModel.Categories {
		for _, target := range category.Targets {
			documentedTargets[target.Name] = true
			for _, alias := range target.Aliases {
				aliases[alias] = true
			}
		}
	}

	checkCtx := &lint.CheckContext{
		HelpModel:         helpModel,
		MakefilePath:      makefilePath,
		PhonyTargets:      targetsResult.IsPhony,
		Dependencies:      targetsResult.Dependencies,
		HasRecipe:         targetsResult.HasRecipe,
		DocumentedTargets: documentedTargets,
		Aliases:           aliases,
	}

	// Step 8: Run all lint checks
	checks := []lint.CheckFunc{
		lint.CheckUndocumentedPhony,
		lint.CheckSummaryPunctuation,
		lint.CheckOrphanAliases,
		lint.CheckLongSummaries,
		lint.CheckEmptyDocumentation,
		lint.CheckMissingVarDescriptions,
		lint.CheckInconsistentNaming,
		lint.CheckCircularAliases,
	}

	result := lint.Lint(checkCtx, checks)

	// Step 9: Output warnings
	if result.HasWarnings {
		for _, warning := range result.Warnings {
			fmt.Println(lint.FormatWarning(warning))
		}

		// Proper pluralization
		count := len(result.Warnings)
		if count == 1 {
			fmt.Println("\nFound 1 warning")
		} else {
			fmt.Printf("\nFound %d warnings\n", count)
		}

		// Return sentinel error to indicate warnings were found (Cobra translates to exit code 1)
		return ErrLintWarningsFound
	}

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "No warnings found\n")
	}

	return nil
}
