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
		return fmt.Errorf("makefile validation failed: %w", err)
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

// IncludePattern holds information about a detected include directive pattern.
type IncludePattern struct {
	// Suffix is the file extension (e.g., ".mk" or "")
	Suffix string
	// FullPattern is the complete include pattern (e.g., "make/*.mk")
	FullPattern string
	// PatternPrefix is the prefix part before the wildcard (e.g., "make/" or "./make/")
	PatternPrefix string
}

// determineTargetFileImpl is the shared implementation.
// If createDirs is true, creates parent directories as needed.
//
// Strategy:
//  1. If explicit --help-file-rel-path is provided, use that (needs include directive)
//  2. Default to make/help.mk (or make/NN-help.mk if numbered files exist)
//  3. Scan Makefile for existing include patterns to determine suffix
//  4. If no include pattern exists, one will be added
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

	// 2. Read Makefile to check for include patterns
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		return "", false, fmt.Errorf("failed to read Makefile: %w", err)
	}

	// 3. Find include pattern for make/* files
	pattern := findMakeIncludePattern(content)

	// 4. Determine the suffix to use for our file
	suffix := ".mk" // default
	if pattern != nil {
		suffix = pattern.Suffix
	}

	// 5. Create make/ directory if needed
	makeDir := filepath.Join(makefileDir, "make")
	if createDirs {
		if err := os.MkdirAll(makeDir, 0755); err != nil {
			return "", false, fmt.Errorf("failed to create make/ directory: %w", err)
		}
	}

	// 6. Check for numbered files in make/ directory using the same pattern matching
	prefix := determineNumberPrefix(makeDir, suffix, pattern)

	// 7. Construct filename
	filename := prefix + "help" + suffix
	targetPath := filepath.Join(makeDir, filename)

	// Need include directive if no existing pattern was found
	needsInclude := pattern == nil

	return targetPath, needsInclude, nil
}

// findMakeIncludePattern scans Makefile content for include directives matching make/*
// Returns nil if no matching pattern found.
func findMakeIncludePattern(content []byte) *IncludePattern {
	// Match patterns like:
	// - include make/*
	// - include make/*.mk
	// - -include make/*
	// - include ./make/*.mk
	// - -include $(dir ...)make/*.mk (less common but possible)
	// Capture groups:
	//   1: Optional $(...) prefix
	//   2: Optional ./ prefix
	//   3: File extension suffix (e.g., .mk)
	includeRegex := regexp.MustCompile(`(?m)^-?include\s+(?:\$\([^)]+\))?(\./)?make/\*(\.[a-zA-Z0-9]+)?(?:\s|$)`)

	matches := includeRegex.FindSubmatch(content)
	if matches == nil {
		return nil
	}

	suffix := ""
	if len(matches) > 2 && len(matches[2]) > 0 {
		suffix = string(matches[2])
	}

	patternPrefix := "make/"
	if len(matches) > 1 && len(matches[1]) > 0 {
		// If ./ prefix was found, use it
		patternPrefix = "./make/"
	}

	return &IncludePattern{
		Suffix:        suffix,
		FullPattern:   string(matches[0]),
		PatternPrefix: patternPrefix,
	}
}

// determineNumberPrefix checks if files in the make directory use numeric prefixes.
// If numbered files exist (e.g., "10-foo.mk"), returns a prefix with matching digit count
// using zeros (e.g., "00-"). Otherwise returns empty string.
// Uses the same pattern matching logic as the include directive - only matches files
// that would be included by the pattern (based on suffix).
func determineNumberPrefix(makeDir, suffix string, pattern *IncludePattern) string {
	// Try to read directory; if it doesn't exist, no prefix needed
	entries, err := os.ReadDir(makeDir)
	if err != nil {
		return ""
	}

	// Pattern to match files starting with digits followed by dash
	// e.g., "01-foo.mk", "10-constants.mk", "100-utils.mk"
	// Only match files that would be included by the pattern (same suffix)
	numberedFileRegex := regexp.MustCompile(`^(\d+)-.*` + regexp.QuoteMeta(suffix) + `$`)

	maxDigits := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// Only check files that match the include pattern suffix
		matches := numberedFileRegex.FindStringSubmatch(entry.Name())
		if matches != nil {
			digitCount := len(matches[1])
			if digitCount > maxDigits {
				maxDigits = digitCount
			}
		}
	}

	if maxDigits == 0 {
		return ""
	}

	// Generate prefix with zeros matching the digit count
	// e.g., if maxDigits is 2, return "00-"
	zeros := ""
	for i := 0; i < maxDigits; i++ {
		zeros += "0"
	}
	return zeros + "-"
}

// FindExistingHelpFile checks for existing help.mk files that were generated by make-help.
// It checks:
//  1. The explicit target file path (if provided)
//  2. Files matching make/(?:0+-)?help.mk pattern in the make directory
//
// Returns the path to an existing file if found, or empty string if none found.
func FindExistingHelpFile(makefilePath, explicitRelPath string) (string, error) {
	makefileDir := filepath.Dir(makefilePath)

	// Check explicit path first if provided
	if explicitRelPath != "" {
		absPath := filepath.Join(makefileDir, explicitRelPath)
		if isGeneratedByMakeHelp(absPath) {
			return absPath, nil
		}
	}

	// Check make/ directory for numbered help files
	makeDir := filepath.Join(makefileDir, "make")
	entries, err := os.ReadDir(makeDir)
	if err != nil {
		// Directory doesn't exist or can't be read - no existing file
		return "", nil
	}

	// Pattern to match: make/help.mk, make/0-help.mk, make/00-help.mk, etc.
	helpFileRegex := regexp.MustCompile(`^(0+-)?help\.mk$`)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if helpFileRegex.MatchString(entry.Name()) {
			helpPath := filepath.Join(makeDir, entry.Name())
			if isGeneratedByMakeHelp(helpPath) {
				return helpPath, nil
			}
		}
	}

	return "", nil
}

// isGeneratedByMakeHelp checks if a file exists and starts with the make-help generation marker.
func isGeneratedByMakeHelp(filePath string) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	// Check if file starts with "# generated-by: make-help" or "# generated by: make-help"
	// (case-insensitive, allow variations)
	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 {
		return false
	}
	firstLine := strings.TrimSpace(lines[0])
	markerRegex := regexp.MustCompile(`(?i)^#\s*generated[- ]by:?\s*make-help`)
	return markerRegex.MatchString(firstLine)
}

// ExtractCommandLineFromHelpFile reads the command line from a help.mk file.
// It looks for a line starting with "# command:" and returns the rest of that line.
func ExtractCommandLineFromHelpFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# command:") {
			// Extract everything after "# command:"
			cmdLine := strings.TrimSpace(strings.TrimPrefix(trimmed, "# command:"))
			return cmdLine, nil
		}
	}

	return "", nil // No command line found
}

// addIncludeDirective injects an include statement into the Makefile using atomic write.
func (s *AddService) addIncludeDirective(makefilePath, targetFile string) error {
	return AddIncludeDirective(makefilePath, targetFile)
}

// AddIncludeDirective injects an include statement into the Makefile using atomic write.
// When targetFile is in the make/ directory and no existing include pattern is found,
// adds a pattern include (-include make/*.mk). Otherwise, uses the self-referential pattern
// $(dir $(lastword $(MAKEFILE_LIST))) to ensure the include works regardless of the working
// directory when make is invoked.
// targetFile should be an absolute path; this function computes the relative path
// from the Makefile directory.
// If an include directive for this file already exists (either include or -include),
// no changes are made.
func AddIncludeDirective(makefilePath, targetFile string) error {
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		return err
	}

	makefileDir := filepath.Dir(makefilePath)

	// Compute relative path from Makefile directory to target file
	relPath, err := filepath.Rel(makefileDir, targetFile)
	if err != nil {
		// Fallback to just the filename if we can't compute relative path
		relPath = filepath.Base(targetFile)
	}

	// Check if target file is in make/ directory
	isInMakeDir := strings.HasPrefix(relPath, "make"+string(filepath.Separator)) || relPath == "make"

	if isInMakeDir {
		// Target is in make/ directory - check for existing pattern
		pattern := findMakeIncludePattern(content)
		if pattern != nil {
			// Pattern already exists, no need to add include directive
			return nil
		}

		// Check if an include directive already exists for make/*.mk pattern
		// Match patterns like:
		// - include make/*.mk / -include make/*.mk
		// - include ./make/*.mk / -include ./make/*.mk
		patternIncludeRegex := regexp.MustCompile(`(?m)^-?include\s+(?:\./)?make/\*\.mk\s*$`)
		if patternIncludeRegex.Match(content) {
			// Pattern include already exists, nothing to do
			return nil
		}

		// No pattern found, add -include make/*.mk
		includeDirective := "\n-include make/*.mk\n"
		newContent := append(content, []byte(includeDirective)...)
		return AtomicWriteFile(makefilePath, newContent, 0644)
	}

	// Target is not in make/ directory - add specific file include
	// Check if an include directive already exists for this file (include or -include)
	// Match patterns like:
	// - include help.mk / -include help.mk
	// - include $(dir $(lastword $(MAKEFILE_LIST)))help.mk / -include ...
	escapedRelPath := regexp.QuoteMeta(relPath)
	includePattern := fmt.Sprintf(`(?m)^-?include\s+(\$\(dir \$\(lastword \$\(MAKEFILE_LIST\)\)\))?%s\s*$`, escapedRelPath)
	existingIncludeRegex := regexp.MustCompile(includePattern)
	if existingIncludeRegex.Match(content) {
		// Include directive already exists, nothing to do
		return nil
	}

	// Use self-referential include pattern that works from any directory
	// Using -include (optional include) allows users to delete help.mk and regenerate via make
	includeDirective := fmt.Sprintf("\n-include $(dir $(lastword $(MAKEFILE_LIST)))%s\n", relPath)

	// Append to end of file
	newContent := append(content, []byte(includeDirective)...)

	// Use atomic write to prevent corruption
	return AtomicWriteFile(makefilePath, newContent, 0644)
}
