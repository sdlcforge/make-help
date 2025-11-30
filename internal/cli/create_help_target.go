package cli

import (
	"fmt"
	"os"

	"github.com/sdlcforge/make-help/internal/discovery"
	"github.com/sdlcforge/make-help/internal/model"
	"github.com/sdlcforge/make-help/internal/parser"
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

	// 5. Collect documented target names
	var documentedTargets []string
	for _, category := range helpModel.Categories {
		for _, t := range category.Targets {
			documentedTargets = append(documentedTargets, t.Name)
		}
	}

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "Found %d documented target(s)\n", len(documentedTargets))
	}

	// 6. Check for help-<target> conflicts
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

	// 7. Determine target file location
	targetFile, needsInclude, err := target.DetermineTargetFile(makefilePath, config.HelpFilePath)
	if err != nil {
		return err
	}

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "Target file: %s (needs include: %v)\n", targetFile, needsInclude)
	}

	// 8. Generate help file content
	genConfig := &target.GeneratorConfig{
		KeepOrderCategories: config.KeepOrderCategories,
		KeepOrderTargets:    config.KeepOrderTargets,
		CategoryOrder:       config.CategoryOrder,
		DefaultCategory:     config.DefaultCategory,
		IncludeTargets:      parseIncludeTargets(config.IncludeTargets),
		IncludeAllPhony:     config.IncludeAllPhony,
		Version:             config.Version,
	}
	content := target.GenerateHelpFile(genConfig, documentedTargets)

	// 9. Write file atomically
	if err := target.AtomicWriteFile(targetFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write help target file %s: %w", targetFile, err)
	}

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "Created help target file: %s\n", targetFile)
	}

	// 10. Add include directive if needed
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
