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

	// Step 4: Build the help model
	builder := model.NewBuilder(config.DefaultCategory)
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
			target.Summary = extractor.Extract(target.Documentation)
		}
	}

	// Step 7: Render the output
	renderer := format.NewRenderer(config.UseColor)
	output, err := renderer.Render(helpModel)
	if err != nil {
		return fmt.Errorf("failed to render help: %w", err)
	}

	// Step 8: Write to stdout
	fmt.Print(output)

	return nil
}
