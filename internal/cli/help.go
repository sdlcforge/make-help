package cli

import (
	"fmt"
	"os"

	"github.com/sdlcforge/make-help/internal/discovery"
	"github.com/sdlcforge/make-help/internal/format"
	"github.com/sdlcforge/make-help/internal/model"
	"github.com/sdlcforge/make-help/internal/ordering"
	"github.com/sdlcforge/make-help/internal/parser"
	"github.com/sdlcforge/make-help/internal/summary"
)

// runHelp orchestrates the help generation process.
// It performs the following steps:
//  1. Discovery - Find Makefile and all included files
//  2. Parsing - Extract documentation directives
//  3. Building - Construct the help model
//  4. Ordering - Apply sorting rules
//  5. Summary - Extract topic sentences
//  6. Formatting - Render the output
//  7. Output - Write to stdout
func runHelp(config *Config) error {
	// Recursion detection: if MAKE_HELP_GENERATING is set, we're being called
	// from within a make process that was spawned by make-help. This indicates
	// infinite recursion (make-help -> make -p -> auto-regen rule -> make-help).
	if os.Getenv("MAKE_HELP_GENERATING") == "1" {
		return fmt.Errorf("recursion detected: make-help was invoked from within a make process spawned by make-help. " +
			"This usually happens when help.mk contains an auto-regeneration rule. " +
			"Regenerate help.mk with the latest make-help to fix this issue")
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

	// Step 3.5: Discover targets with .PHONY status
	targetsResult, err := discoveryService.DiscoverTargets(makefilePath)
	if err != nil {
		return fmt.Errorf("failed to discover targets: %w", err)
	}

	// Step 4: Build the help model with filtering
	includeTargets := parseIncludeTargets(config.IncludeTargets)
	builderConfig := &model.BuilderConfig{
		DefaultCategory: config.DefaultCategory,
		IncludeTargets:  includeTargets,
		IncludeAllPhony: config.IncludeAllPhony,
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

	// Step 5: Apply ordering rules
	orderingService := ordering.NewService(
		config.KeepOrderCategories,
		config.KeepOrderTargets,
		config.KeepOrderFiles,
		config.CategoryOrder,
	)
	if err := orderingService.ApplyOrdering(helpModel); err != nil {
		return fmt.Errorf("failed to apply ordering: %w", err)
	}

	// Step 6: Extract summaries for all targets
	extractor := summary.NewExtractor()
	for i := range helpModel.Categories {
		for j := range helpModel.Categories[i].Targets {
			target := &helpModel.Categories[i].Targets[j]
			summaryText := extractor.ExtractPlainText(target.Documentation)
			if summaryText != "" {
				target.Summary = []string{summaryText}
			} else {
				target.Summary = []string{}
			}
		}
	}

	// Step 7: Create formatter and render the output
	formatterConfig := &format.FormatterConfig{
		UseColor: config.UseColor,
	}
	formatter, err := format.NewFormatter(config.Format, formatterConfig)
	if err != nil {
		return fmt.Errorf("failed to create formatter: %w", err)
	}

	// Step 8: Write to stdout
	if err := formatter.RenderHelp(helpModel, os.Stdout); err != nil {
		return fmt.Errorf("failed to render help: %w", err)
	}

	return nil
}

// runDetailedHelp displays detailed information for a single target.
// Shows full documentation, all variables with descriptions, aliases, and source location.
// If the target doesn't exist, returns an error.
// If the target exists but has no documentation, shows basic info.
func runDetailedHelp(config *Config) error {
	// Step 1: Resolve Makefile path
	makefilePath, err := discovery.ResolveMakefilePath(config.MakefilePath)
	if err != nil {
		return fmt.Errorf("failed to resolve Makefile path: %w", err)
	}

	if err := discovery.ValidateMakefileExists(makefilePath); err != nil {
		return err
	}

	config.MakefilePath = makefilePath

	// Step 2: Discover all targets to verify the requested target exists
	discoveryService := discovery.NewService(discovery.NewDefaultExecutor(), config.Verbose)
	targetsResult, err := discoveryService.DiscoverTargets(makefilePath)
	if err != nil {
		return fmt.Errorf("failed to discover targets: %w", err)
	}

	// Step 3: Check if target exists
	targetExists := false
	for _, t := range targetsResult.Targets {
		if t == config.Target {
			targetExists = true
			break
		}
	}
	if !targetExists {
		return fmt.Errorf("target '%s' not found", config.Target)
	}

	// Step 4: Discover and parse all Makefiles to get documentation
	makefiles, err := discoveryService.DiscoverMakefiles(makefilePath)
	if err != nil {
		return fmt.Errorf("failed to discover Makefiles: %w", err)
	}

	scanner := parser.NewScanner()
	var parsedFiles []*parser.ParsedFile

	for _, mf := range makefiles {
		parsed, err := scanner.ScanFile(mf)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", mf, err)
		}
		parsedFiles = append(parsedFiles, parsed)
	}

	// Step 5: Build the help model to get documentation
	// For detailed help, we want to include the specific target even if undocumented
	includeTargets := parseIncludeTargets(config.IncludeTargets)
	includeTargets = append(includeTargets, config.Target) // Always include the requested target
	builderConfig := &model.BuilderConfig{
		DefaultCategory: config.DefaultCategory,
		IncludeTargets:  includeTargets,
		PhonyTargets:    targetsResult.IsPhony,
		Dependencies:    targetsResult.Dependencies,
		HasRecipe:       targetsResult.HasRecipe,
	}
	builder := model.NewBuilder(builderConfig)
	helpModel, err := builder.Build(parsedFiles)
	if err != nil {
		return fmt.Errorf("failed to build help model: %w", err)
	}

	// Step 6: Find the target in the model
	var foundTarget *model.Target
	for i := range helpModel.Categories {
		for j := range helpModel.Categories[i].Targets {
			if helpModel.Categories[i].Targets[j].Name == config.Target {
				foundTarget = &helpModel.Categories[i].Targets[j]
				break
			}
		}
		if foundTarget != nil {
			break
		}
	}

	// Step 7: Create formatter and render the output
	formatterConfig := &format.FormatterConfig{
		UseColor: config.UseColor,
	}
	formatter, err := format.NewFormatter(config.Format, formatterConfig)
	if err != nil {
		return fmt.Errorf("failed to create formatter: %w", err)
	}

	if foundTarget != nil && len(foundTarget.Documentation) > 0 {
		// Target has documentation - use detailed renderer
		if err := formatter.RenderDetailedTarget(foundTarget, os.Stdout); err != nil {
			return fmt.Errorf("failed to render detailed target: %w", err)
		}
	} else {
		// Target exists but has no documentation - show basic info
		// If we found it in the model, use its source info; otherwise leave empty
		sourceFile := ""
		lineNumber := 0
		if foundTarget != nil {
			sourceFile = foundTarget.SourceFile
			lineNumber = foundTarget.LineNumber
		}
		if err := formatter.RenderBasicTarget(config.Target, sourceFile, lineNumber, os.Stdout); err != nil {
			return fmt.Errorf("failed to render basic target: %w", err)
		}
	}

	return nil
}
