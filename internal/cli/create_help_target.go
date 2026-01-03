package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sdlcforge/make-help/internal/discovery"
	"github.com/sdlcforge/make-help/internal/model"
	"github.com/sdlcforge/make-help/internal/ordering"
	"github.com/sdlcforge/make-help/internal/parser"
	"github.com/sdlcforge/make-help/internal/summary"
	"github.com/sdlcforge/make-help/internal/target"
)

// filterOutHelpFiles removes help file paths from the makefiles list.
// This ensures MAKE_HELP_MAKEFILES only contains source files, not the generated output.
func filterOutHelpFiles(makefiles []string, helpFiles ...string) []string {
	// Create a set of help file paths to exclude (cleaned/normalized)
	exclude := make(map[string]bool)
	for _, hf := range helpFiles {
		if hf != "" {
			exclude[filepath.Clean(hf)] = true
		}
	}

	// Initialize with empty slice to ensure non-nil return value
	filtered := make([]string, 0, len(makefiles))
	for _, mf := range makefiles {
		if !exclude[filepath.Clean(mf)] {
			filtered = append(filtered, mf)
		}
	}
	return filtered
}

// runCreateHelpTarget generates and writes the help target file.
func runCreateHelpTarget(config *Config) error {
	// Recursion detection: if MAKE_HELP_GENERATING is set, we're being called
	// from within a make process that was spawned by make-help. This indicates
	// infinite recursion (make-help -> make -p -> auto-regen rule -> make-help).
	if os.Getenv("MAKE_HELP_GENERATING") == "1" {
		return fmt.Errorf("recursion detected: make-help was invoked from within a make process spawned by make-help. " +
			"This usually happens when help.mk contains an auto-regeneration rule. " +
			"Regenerate help.mk with the latest make-help to fix this issue")
	}

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
		return fmt.Errorf("makefile validation failed: %w", err)
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
		Dependencies:    targetsResult.Dependencies,
		HasRecipe:       targetsResult.HasRecipe,
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
		config.KeepOrderFiles,
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

	// 8. Determine target file location
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

	// 9.5. Check for existing help.mk file and restore options if no options were provided
	existingFile, err := target.FindExistingHelpFile(makefilePath, config.HelpFileRelPath)
	if err != nil {
		return fmt.Errorf("failed to check for existing help file: %w", err)
	}

	// If we found an existing file and no options were provided, restore options from it
	if existingFile != "" && !HasAnyOptions() {
		cmdLine, err := target.ExtractCommandLineFromHelpFile(existingFile)
		if err != nil {
			if config.Verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to read command line from %s: %v\n", existingFile, err)
			}
		} else if cmdLine != "" && strings.HasPrefix(cmdLine, "make-help") {
			if config.Verbose {
				fmt.Fprintf(os.Stderr, "Restoring options from existing help file: %s\n", existingFile)
				fmt.Fprintf(os.Stderr, "Command line: %s\n", cmdLine)
			}
			// Parse and apply the command line options using Cobra
			if err := ParseCommandLineFromHelpFile(cmdLine, config); err != nil {
				if config.Verbose {
					fmt.Fprintf(os.Stderr, "Warning: failed to parse command line from help file: %v\n", err)
				}
				// Don't fail the whole operation if we can't restore options
			}
			// Note: We don't override config.CommandLine here - we always use
			// the actual invocation command, not what was stored in the file
		}
	}

	if existingFile != "" && existingFile != targetFile {
		if config.Verbose {
			fmt.Fprintf(os.Stderr, "Found existing help file: %s (will create: %s)\n", existingFile, targetFile)
		}
		// Note: We continue anyway - the user may want to move/rename the help file
	}

	// Filter out help files from the makefiles list
	filteredMakefiles := filterOutHelpFiles(makefiles, targetFile, existingFile)

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "Total makefiles discovered: %d, after filtering help files: %d\n", len(makefiles), len(filteredMakefiles))
	}

	// 10. Generate help file content
	// Use the raw command line (always captured from os.Args in PreRunE)
	genConfig := &target.GeneratorConfig{
		UseColor:            config.UseColor,
		Makefiles:           filteredMakefiles,
		HelpModel:           helpModel,
		MakefileDir:         filepath.Dir(makefilePath),
		KeepOrderCategories: config.KeepOrderCategories,
		KeepOrderTargets:    config.KeepOrderTargets,
		CategoryOrder:       config.CategoryOrder,
		DefaultCategory:     config.DefaultCategory,
		HelpCategory:        config.HelpCategory,
		IncludeTargets:      parseIncludeTargets(config.IncludeTargets),
		IncludeAllPhony:     config.IncludeAllPhony,
		CommandLine:         config.CommandLine,
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

		// Determine include directive based on file location
		var includeDirective string
		if filepath.Dir(relPath) == "make" {
			// For files in make/, use pattern include
			suffix := filepath.Ext(targetFile)
			if suffix == "" {
				suffix = ".mk"
			}
			includeDirective = fmt.Sprintf("\n-include make/*%s\n", suffix)
		} else {
			// For files outside make/, use self-referential include
			includeDirective = fmt.Sprintf("\n-include $(dir $(lastword $(MAKEFILE_LIST)))%s\n", relPath)
		}

		fmt.Println()
		fmt.Printf("--- Append to %s ---\n", makefilePath)
		fmt.Print(includeDirective)
		fmt.Println("--- end ---")
	}

	return nil
}
