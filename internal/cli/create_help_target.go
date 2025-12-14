package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sdlcforge/make-help/internal/discovery"
	"github.com/sdlcforge/make-help/internal/model"
	"github.com/sdlcforge/make-help/internal/ordering"
	"github.com/sdlcforge/make-help/internal/parser"
	"github.com/sdlcforge/make-help/internal/summary"
	"github.com/sdlcforge/make-help/internal/target"
)

// runCreateHelpTarget generates and writes the help target file.
func runCreateHelpTarget(config *Config) error {
	// 1. Resolve Makefile path
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

	// 2. Validate Makefile syntax
	executor := discovery.NewDefaultExecutor()
	if err := target.ValidateMakefile(executor, makefilePath); err != nil {
		return fmt.Errorf("Makefile validation failed: %w", err)
	}

	// 3. Discover files and targets
	discoveryService := discovery.NewService(executor, config.Verbose)

	makefiles, err := discoveryService.DiscoverMakefiles(makefilePath)
	if err != nil {
		return fmt.Errorf("failed to discover Makefile includes: %w", err)
	}

	targetsResult, err := discoveryService.DiscoverTargets(makefilePath)
	if err != nil {
		return fmt.Errorf("failed to discover targets: %w", err)
	}

	// 4. Parse and build model to get documented targets
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

	builderConfig := &model.BuilderConfig{
		DefaultCategory: config.DefaultCategory,
		IncludeTargets:  parseIncludeTargets(config.IncludeTargets),
		IncludeAllPhony: config.IncludeAllPhony,
		PhonyTargets:    targetsResult.IsPhony,
	}
	builder := model.NewBuilder(builderConfig)
	helpModel, err := builder.Build(parsedFiles)
	if err != nil {
		return err
	}

	// 5. Apply ordering rules to the model
	orderingService := ordering.NewService(
		config.KeepOrderCategories,
		config.KeepOrderTargets,
		config.CategoryOrder,
	)
	if err := orderingService.ApplyOrdering(helpModel); err != nil {
		return fmt.Errorf("failed to apply ordering: %w", err)
	}

	// 6. Extract summaries for all targets
	extractor := summary.NewExtractor()
	for i := range helpModel.Categories {
		for j := range helpModel.Categories[i].Targets {
			target := &helpModel.Categories[i].Targets[j]
			target.Summary = extractor.Extract(target.Documentation)
		}
	}

	// 7. Collect documented target names
	var documentedTargets []string
	for _, category := range helpModel.Categories {
		for _, t := range category.Targets {
			documentedTargets = append(documentedTargets, t.Name)
		}
	}

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "Found %d documented target(s)\n", len(documentedTargets))
	}

	// 8. Check for help-<target> conflicts
	existingTargets := make(map[string]bool)
	for _, t := range targetsResult.Targets {
		existingTargets[t] = true
	}

	for _, t := range documentedTargets {
		helpTargetName := "help-" + t
		if existingTargets[helpTargetName] {
			return fmt.Errorf("cannot generate %s target - target already exists in Makefile\nConsider renaming your existing %s target", helpTargetName, helpTargetName)
		}
	}

	// Also check for 'help' conflict
	if existingTargets["help"] {
		return fmt.Errorf("cannot generate help target - target already exists in Makefile")
	}

	// 9. Determine target file location
	var targetFile string
	var needsInclude bool
	if config.DryRun {
		// Use no-dirs version in dry-run mode to avoid creating directories
		targetFile, needsInclude, err = target.DetermineTargetFileNoDirs(makefilePath, config.HelpFileRelPath)
	} else {
		targetFile, needsInclude, err = target.DetermineTargetFile(makefilePath, config.HelpFileRelPath)
	}
	if err != nil {
		return err
	}

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "Target file: %s (needs include: %v)\n", targetFile, needsInclude)
	}

	// 10. Generate help file content
	genConfig := &target.GeneratorConfig{
		UseColor:            config.UseColor,
		Makefiles:           makefiles,
		HelpModel:           helpModel,
		MakefileDir:         filepath.Dir(makefilePath),
		KeepOrderCategories: config.KeepOrderCategories,
		KeepOrderTargets:    config.KeepOrderTargets,
		CategoryOrder:       config.CategoryOrder,
		DefaultCategory:     config.DefaultCategory,
		IncludeTargets:      parseIncludeTargets(config.IncludeTargets),
		IncludeAllPhony:     config.IncludeAllPhony,
	}
	content, err := target.GenerateHelpFile(genConfig)
	if err != nil {
		return fmt.Errorf("failed to generate help file: %w", err)
	}

	// 11. Handle dry-run mode
	if config.DryRun {
		return printDryRunOutput(makefilePath, targetFile, needsInclude, content)
	}

	// 12. Write file atomically
	if err := target.AtomicWriteFile(targetFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write help target file %s: %w", targetFile, err)
	}

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "Created help target file: %s\n", targetFile)
	}

	// 13. Add include directive if needed
	if needsInclude {
		if err := target.AddIncludeDirective(makefilePath, targetFile); err != nil {
			return err
		}
		if config.Verbose {
			fmt.Fprintf(os.Stderr, "Added include directive to: %s\n", makefilePath)
		}
	}

	fmt.Printf("Successfully created help target: %s\n", targetFile)
	return nil
}

// printDryRunOutput displays what would be created/modified in dry-run mode.
func printDryRunOutput(makefilePath, targetFile string, needsInclude bool, content string) error {
	fmt.Println("Dry run mode - no files will be modified")
	fmt.Println()
	fmt.Printf("Would create: %s\n", targetFile)
	if needsInclude {
		fmt.Printf("Would append to: %s\n", makefilePath)
	}
	fmt.Println()
	fmt.Printf("--- %s ---\n", targetFile)
	fmt.Print(content)
	fmt.Println("--- end ---")

	if needsInclude {
		// Compute relative path for include directive
		makefileDir := filepath.Dir(makefilePath)
		relPath, err := filepath.Rel(makefileDir, targetFile)
		if err != nil {
			// Fallback to just the filename if we can't compute relative path
			relPath = filepath.Base(targetFile)
		}

		includeDirective := fmt.Sprintf("\n-include $(dir $(lastword $(MAKEFILE_LIST)))%s\n", relPath)

		fmt.Println()
		fmt.Printf("--- Append to %s ---\n", makefilePath)
		fmt.Print(includeDirective)
		fmt.Println("--- end ---")
	}

	return nil
}
