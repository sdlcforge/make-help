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
	TargetFile          string
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
//  1. Use explicit --target-file if specified
//  2. Create make/01-help.mk if include make/*.mk pattern found
//  3. Append directly to main Makefile
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
	if err := atomicWriteFile(targetFile, []byte(content), 0644); err != nil {
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Run make -n (dry-run) to check syntax without executing recipes
	_, stderr, err := s.executor.ExecuteContext(ctx, "make", "-n", "-f", makefilePath)
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
	// 1. Explicit --target-file
	if s.config.TargetFile != "" {
		return s.config.TargetFile, true, nil
	}

	// 2. Check for include make/*.mk pattern
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		return "", false, fmt.Errorf("failed to read Makefile: %w", err)
	}

	includeRegex := regexp.MustCompile(`(?m)^include\s+make/\*\.mk`)
	if includeRegex.Match(content) {
		// Create make/01-help.mk
		makeDir := filepath.Join(filepath.Dir(makefilePath), "make")
		if err := os.MkdirAll(makeDir, 0755); err != nil {
			return "", false, fmt.Errorf("failed to create make/ directory: %w", err)
		}
		return filepath.Join(makeDir, "01-help.mk"), false, nil
	}

	// 3. Append directly to Makefile
	return makefilePath, false, nil
}

// addIncludeDirective injects an include statement into the Makefile using atomic write.
func (s *AddService) addIncludeDirective(makefilePath, targetFile string) error {
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		return err
	}

	// Make path relative to Makefile
	relPath, err := filepath.Rel(filepath.Dir(makefilePath), targetFile)
	if err != nil {
		return err
	}

	includeDirective := fmt.Sprintf("\ninclude %s\n", relPath)

	// Append to end of file
	newContent := append(content, []byte(includeDirective)...)

	// Use atomic write to prevent corruption
	return atomicWriteFile(makefilePath, newContent, 0644)
}
