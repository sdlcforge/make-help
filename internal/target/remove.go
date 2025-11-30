package target

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sdlcforge/make-help/internal/discovery"
)

// RemoveService handles removing help targets from Makefiles.
type RemoveService struct {
	config   *Config
	executor discovery.CommandExecutor
	verbose  bool
}

// NewRemoveService creates a new RemoveService instance.
func NewRemoveService(config *Config, executor discovery.CommandExecutor, verbose bool) *RemoveService {
	return &RemoveService{
		config:   config,
		executor: executor,
		verbose:  verbose,
	}
}

// RemoveTarget removes help target artifacts from the Makefile.
// It performs the following cleanup steps:
//  1. Remove include directives for help target files
//  2. Remove inline help: target and .PHONY: help
//  3. Delete help target files (make/01-help.mk)
func (s *RemoveService) RemoveTarget() error {
	makefilePath := s.config.MakefilePath

	// Validate Makefile syntax before modifying
	if err := s.validateMakefile(makefilePath); err != nil {
		return fmt.Errorf("Makefile validation failed: %w", err)
	}

	removed := false

	// Remove include directives
	if err := s.removeIncludeDirectives(makefilePath); err != nil {
		return err
	}

	// Remove inline help target
	inlineRemoved, err := s.removeInlineHelpTarget(makefilePath)
	if err != nil {
		return err
	}
	if inlineRemoved {
		removed = true
	}

	// Remove help target files
	filesRemoved, err := s.removeHelpTargetFiles(makefilePath)
	if err != nil {
		return err
	}
	if filesRemoved {
		removed = true
	}

	if removed {
		fmt.Printf("Successfully removed help target from: %s\n", makefilePath)
	} else {
		fmt.Printf("No help target found in: %s\n", makefilePath)
	}

	return nil
}

// validateMakefile runs `make -n` to check for syntax errors.
func (s *RemoveService) validateMakefile(makefilePath string) error {
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

// removeIncludeDirectives removes include lines for help targets using atomic write.
// Matches both simple includes (include help.mk) and self-referential includes
// (include $(dir $(lastword $(MAKEFILE_LIST)))help.mk).
func (s *RemoveService) removeIncludeDirectives(makefilePath string) error {
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	filtered := []string{}

	// Match both patterns (with optional - prefix for silent include):
	// - include help.mk / -include help.mk
	// - include $(dir $(lastword $(MAKEFILE_LIST)))help.mk / -include ...
	includeRegex := regexp.MustCompile(`^-?include\s+(\$\(dir \$\(lastword \$\(MAKEFILE_LIST\)\)\))?.*help.*\.mk`)
	removed := false

	for _, line := range lines {
		if !includeRegex.MatchString(line) {
			filtered = append(filtered, line)
		} else {
			removed = true
			if s.verbose {
				fmt.Printf("Removed include directive: %s\n", line)
			}
		}
	}

	if !removed {
		return nil // No changes needed
	}

	newContent := strings.Join(filtered, "\n")
	return AtomicWriteFile(makefilePath, []byte(newContent), 0644)
}

// removeInlineHelpTarget removes help target from Makefile using atomic write.
// Returns true if a help target was found and removed.
func (s *RemoveService) removeInlineHelpTarget(makefilePath string) (bool, error) {
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		return false, err
	}

	lines := strings.Split(string(content), "\n")
	filtered := []string{}

	inHelpTarget := false
	removed := false

	for _, line := range lines {
		// Detect start of help target
		if strings.HasPrefix(line, "help:") || strings.HasPrefix(line, ".PHONY: help") {
			inHelpTarget = true
			removed = true
			if s.verbose {
				fmt.Printf("Removing help target starting at: %s\n", line)
			}
			continue
		}

		// Detect end of help target (next target or non-recipe line)
		if inHelpTarget {
			if strings.HasPrefix(line, "\t") || strings.HasPrefix(line, "  ") {
				continue // Skip recipe lines
			}
			inHelpTarget = false
		}

		filtered = append(filtered, line)
	}

	if !removed {
		return false, nil // No changes needed
	}

	newContent := strings.Join(filtered, "\n")
	return true, AtomicWriteFile(makefilePath, []byte(newContent), 0644)
}

// removeHelpTargetFiles deletes help target files.
// Returns true if any files were removed.
func (s *RemoveService) removeHelpTargetFiles(makefilePath string) (bool, error) {
	makeDir := filepath.Join(filepath.Dir(makefilePath), "make")
	helpFile := filepath.Join(makeDir, "01-help.mk")

	removed := false

	if _, err := os.Stat(helpFile); err == nil {
		if err := os.Remove(helpFile); err != nil {
			return false, fmt.Errorf("failed to remove %s: %w", helpFile, err)
		}
		if s.verbose {
			fmt.Printf("Removed help target file: %s\n", helpFile)
		}
		removed = true
	}

	return removed, nil
}
