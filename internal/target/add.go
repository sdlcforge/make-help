package target

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/sdlcforge/make-help/internal/discovery"
)

// Config holds configuration for target manipulation operations.
type Config struct {
	MakefilePath        string
	TargetFileRelPath   string // Relative path for help target file (e.g., "help.mk" or "make/help.mk")
	KeepOrderCategories bool
	KeepOrderTargets    bool
	CategoryOrder       []string
	DefaultCategory     string
}

// AddService handles adding help targets to Makefiles.
type AddService struct {
	config   *Config
	executor discovery.CommandExecutor
	verbose  bool
}

// NewAddService creates a new AddService instance.
func NewAddService(config *Config, executor discovery.CommandExecutor, verbose bool) *AddService {
	return &AddService{
		config:   config,
		executor: executor,
		verbose:  verbose,
	}
}

// AddTarget generates and injects a help target into the Makefile.
// It follows a three-tier strategy for target file placement:
//  1. Use explicit --help-file-rel-path if specified (needs include directive)
//  2. Create make/01-help.mk if include make/*.mk pattern found (no include needed)
//  3. Otherwise create help.mk in same directory as Makefile (needs include directive)
func (s *AddService) AddTarget() error {
	makefilePath := s.config.MakefilePath

	// Validate Makefile syntax before modifying
	if err := s.validateMakefile(makefilePath); err != nil {
		return fmt.Errorf("Makefile validation failed: %w", err)
	}

	// Determine target file location
	targetFile, needsInclude, err := s.determineTargetFile(makefilePath)
	if err != nil {
		return err
	}

	// Generate help target content
	content := generateHelpTarget(s.config)

	// Write target file using atomic write
	if err := AtomicWriteFile(targetFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write target file %s: %w", targetFile, err)
	}

	if s.verbose {
		fmt.Printf("Created help target file: %s\n", targetFile)
	}

	// Add include directive if needed
	if needsInclude {
		if err := s.addIncludeDirective(makefilePath, targetFile); err != nil {
			return err
		}
		if s.verbose {
			fmt.Printf("Added include directive to: %s\n", makefilePath)
		}
	}

	fmt.Printf("Successfully added help target to: %s\n", targetFile)
	return nil
}

// validateMakefile runs `make -n` to check for syntax errors.
func (s *AddService) validateMakefile(makefilePath string) error {
	return ValidateMakefile(s.executor, makefilePath)
}

// ValidateMakefile runs `make -n` to check for syntax errors.
func ValidateMakefile(executor discovery.CommandExecutor, makefilePath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Run make -n (dry-run) to check syntax without executing recipes
	_, stderr, err := executor.ExecuteContext(ctx, "make", "-n", "-f", makefilePath)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("validation timed out")
		}
		return fmt.Errorf("syntax error in Makefile:\n%s", stderr)
	}
	return nil
}

// determineTargetFile decides where to create the help target.
// Returns: (targetFile path, needsInclude directive, error)
func (s *AddService) determineTargetFile(makefilePath string) (string, bool, error) {
	return DetermineTargetFile(makefilePath, s.config.TargetFileRelPath)
}

// DetermineTargetFile decides where to create the help target.
// explicitRelPath must be a relative path (validated by CLI).
// Returns: (targetFile absolute path, needsInclude directive, error)
func DetermineTargetFile(makefilePath, explicitRelPath string) (string, bool, error) {
	return determineTargetFileImpl(makefilePath, explicitRelPath, true)
}

// DetermineTargetFileNoDirs decides where to create the help target without creating directories.
// Used for dry-run mode. Same as DetermineTargetFile but doesn't create directories.
// Returns: (targetFile absolute path, needsInclude directive, error)
func DetermineTargetFileNoDirs(makefilePath, explicitRelPath string) (string, bool, error) {
	return determineTargetFileImpl(makefilePath, explicitRelPath, false)
}

// determineTargetFileImpl is the shared implementation.
// If createDirs is true, creates parent directories as needed.
func determineTargetFileImpl(makefilePath, explicitRelPath string, createDirs bool) (string, bool, error) {
	makefileDir := filepath.Dir(makefilePath)

	// 1. Explicit --help-file-rel-path (always relative)
	if explicitRelPath != "" {
		// Compute absolute path for file writing
		absPath := filepath.Join(makefileDir, explicitRelPath)
		if createDirs {
			// Create parent directory if needed
			parentDir := filepath.Dir(absPath)
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				return "", false, fmt.Errorf("failed to create directory %s: %w", parentDir, err)
			}
		}
		return absPath, true, nil
	}

	// 2. Check for include make/*.mk pattern
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		return "", false, fmt.Errorf("failed to read Makefile: %w", err)
	}

	includeRegex := regexp.MustCompile(`(?m)^-?include\s+make/\*\.mk`)
	if includeRegex.Match(content) {
		// Create make/01-help.mk
		makeDir := filepath.Join(makefileDir, "make")
		if createDirs {
			if err := os.MkdirAll(makeDir, 0755); err != nil {
				return "", false, fmt.Errorf("failed to create make/ directory: %w", err)
			}
		}
		return filepath.Join(makeDir, "01-help.mk"), false, nil
	}

	// 3. Create help.mk in same directory as Makefile
	return filepath.Join(makefileDir, "help.mk"), true, nil
}

// addIncludeDirective injects an include statement into the Makefile using atomic write.
func (s *AddService) addIncludeDirective(makefilePath, targetFile string) error {
	return AddIncludeDirective(makefilePath, targetFile)
}

// AddIncludeDirective injects an include statement into the Makefile using atomic write.
// Uses the self-referential pattern $(dir $(lastword $(MAKEFILE_LIST))) to ensure
// the include works regardless of the working directory when make is invoked.
// targetFile should be an absolute path; this function computes the relative path
// from the Makefile directory.
func AddIncludeDirective(makefilePath, targetFile string) error {
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		return err
	}

	// Compute relative path from Makefile directory to target file
	makefileDir := filepath.Dir(makefilePath)
	relPath, err := filepath.Rel(makefileDir, targetFile)
	if err != nil {
		// Fallback to just the filename if we can't compute relative path
		relPath = filepath.Base(targetFile)
	}

	// Use self-referential include pattern that works from any directory
	includeDirective := fmt.Sprintf("\ninclude $(dir $(lastword $(MAKEFILE_LIST)))%s\n", relPath)

	// Append to end of file
	newContent := append(content, []byte(includeDirective)...)

	// Use atomic write to prevent corruption
	return AtomicWriteFile(makefilePath, newContent, 0644)
}
