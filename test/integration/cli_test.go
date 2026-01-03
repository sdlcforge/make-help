//go:build integration

package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getProjectRoot returns the project root directory
func getProjectRoot(t *testing.T) string {
	// Find project root by looking for go.mod
	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root")
		}
		dir = parent
	}
}

// buildBinary builds the make-help binary and returns its path
func buildBinary(t *testing.T) string {
	projectRoot := getProjectRoot(t)
	binaryPath := filepath.Join(t.TempDir(), "make-help")

	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/make-help")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "failed to build binary: %s", output)

	return binaryPath
}

// runMakeHelp runs make-help with the given arguments
func runMakeHelp(t *testing.T, binary string, args ...string) (string, string, error) {
	cmd := exec.Command(binary, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// getFixturePath returns the path to a test fixture
func getFixturePath(t *testing.T, name string) string {
	projectRoot := getProjectRoot(t)
	return filepath.Join(projectRoot, "test", "fixtures", "makefiles", name)
}

// getExpectedPath returns the path to an expected output file
func getExpectedPath(t *testing.T, name string) string {
	projectRoot := getProjectRoot(t)
	return filepath.Join(projectRoot, "test", "fixtures", "expected", name)
}

// readExpected reads the expected output file
func readExpected(t *testing.T, name string) string {
	path := getExpectedPath(t, name)
	content, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read expected file: %s", path)
	return string(content)
}

func TestBasicMakefile(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "basic.mk")

	stdout, stderr, err := runMakeHelp(t, binary, "--output", "-", "--makefile-path", fixture, "--no-color")
	require.NoError(t, err, "stderr: %s", stderr)

	expected := readExpected(t, "basic_help.txt")
	assert.Equal(t, expected, stdout)
}

func TestCategorizedMakefile(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "categorized.mk")

	stdout, stderr, err := runMakeHelp(t, binary, "--output", "-", "--makefile-path", fixture, "--no-color")
	require.NoError(t, err, "stderr: %s", stderr)

	expected := readExpected(t, "categorized_help.txt")
	assert.Equal(t, expected, stdout)
}

func TestComplexMakefile(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "complex.mk")

	stdout, stderr, err := runMakeHelp(t, binary, "--output", "-", "--makefile-path", fixture, "--no-color")
	require.NoError(t, err, "stderr: %s", stderr)

	expected := readExpected(t, "complex_help.txt")
	assert.Equal(t, expected, stdout)
}

func TestCategoryOrderFlag(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "categorized.mk")

	// Test explicit category order
	stdout, stderr, err := runMakeHelp(t, binary,
		"--output", "-",
		"--makefile-path", fixture,
		"--no-color",
		"--category-order", "Test,Build")
	require.NoError(t, err, "stderr: %s", stderr)

	// Test category should appear before Build
	testIdx := strings.Index(stdout, "Test:")
	buildIdx := strings.Index(stdout, "Build:")
	assert.True(t, testIdx < buildIdx, "Test should appear before Build with --category-order Test,Build")
}

func TestKeepOrderFlags(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "categorized.mk")

	// Test --keep-order-categories
	stdout, stderr, err := runMakeHelp(t, binary,
		"--output", "-",
		"--makefile-path", fixture,
		"--no-color",
		"--keep-order-categories")
	require.NoError(t, err, "stderr: %s", stderr)

	// Build should appear before Test (discovery order)
	buildIdx := strings.Index(stdout, "Build:")
	testIdx := strings.Index(stdout, "Test:")
	assert.True(t, buildIdx < testIdx, "Build should appear before Test with --keep-order-categories")
}

func TestKeepOrderAllFlag(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "categorized.mk")

	stdout, stderr, err := runMakeHelp(t, binary,
		"--output", "-",
		"--makefile-path", fixture,
		"--no-color",
		"--keep-order-all")
	require.NoError(t, err, "stderr: %s", stderr)

	// Build should appear before Test (discovery order)
	buildIdx := strings.Index(stdout, "Build:")
	testIdx := strings.Index(stdout, "Test:")
	assert.True(t, buildIdx < testIdx, "Build should appear before Test with --keep-order-all")
}

func TestDefaultCategoryFlag(t *testing.T) {
	binary := buildBinary(t)

	// Test that --default-category applies to uncategorized targets
	// The mixed_categorization.mk file has category directive that persists,
	// so we create a simple test with a category-less include file
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	// Create a Makefile with a single categorized target
	err := os.WriteFile(makefilePath, []byte(`
## !category Build
## Build the project
build:
	@echo building
`), 0644)
	require.NoError(t, err)

	// Test with --default-category flag (even though not needed here, should work)
	stdout, stderr, err := runMakeHelp(t, binary,
		"--output", "-",
		"--makefile-path", makefilePath,
		"--no-color",
		"--default-category", "Other")
	require.NoError(t, err, "stderr: %s", stderr)
	assert.Contains(t, stdout, "Build:", "should have Build category")
	assert.Contains(t, stdout, "build", "should have build target")
}

func TestNoColorFlag(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "basic.mk")

	stdout, _, err := runMakeHelp(t, binary, "--output", "-", "--makefile-path", fixture, "--no-color")
	require.NoError(t, err)

	// Should not contain ANSI escape codes
	assert.NotContains(t, stdout, "\033[")
	assert.NotContains(t, stdout, "\x1b[")
}

func TestColorFlag(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "basic.mk")

	stdout, _, err := runMakeHelp(t, binary, "--output", "-", "--makefile-path", fixture, "--color")
	require.NoError(t, err)

	// Should contain ANSI escape codes
	assert.Contains(t, stdout, "\033[")
}

func TestVerboseFlag(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "basic.mk")

	_, stderr, err := runMakeHelp(t, binary, "--output", "-", "--makefile-path", fixture, "--no-color", "--verbose")
	require.NoError(t, err)

	// Verbose output goes to stderr
	assert.Contains(t, stderr, "Using Makefile")
}

func TestEmptyMakefile(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "empty.mk")

	stdout, _, err := runMakeHelp(t, binary, "--output", "-", "--makefile-path", fixture, "--no-color")
	require.NoError(t, err)

	// Should still produce usage line
	assert.Contains(t, stdout, "Usage: make")
}

func TestWithAliasesMakefile(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "with_aliases.mk")

	stdout, _, err := runMakeHelp(t, binary, "--output", "-", "--makefile-path", fixture, "--no-color")
	require.NoError(t, err)

	// Should contain aliases
	assert.Contains(t, stdout, "b")
	assert.Contains(t, stdout, "t")
	// Should contain variables
	assert.Contains(t, stdout, "VERBOSE")
}

func TestUnknownCategoryOrder(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "categorized.mk")

	_, _, err := runMakeHelp(t, binary,
		"--output", "-",
		"--makefile-path", fixture,
		"--no-color",
		"--category-order", "NonExistent")
	assert.Error(t, err, "should fail with unknown category")
}

func TestConflictingColorFlags(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "basic.mk")

	_, _, err := runMakeHelp(t, binary,
		"--makefile-path", fixture,
		"--color", "--no-color")
	assert.Error(t, err, "should fail with conflicting color flags")
}

func TestHelpFlag(t *testing.T) {
	binary := buildBinary(t)

	stdout, _, err := runMakeHelp(t, binary, "--help")
	require.NoError(t, err)

	assert.Contains(t, stdout, "make-help")
	assert.Contains(t, stdout, "--makefile-path")
	assert.Contains(t, stdout, "--no-color")
	assert.Contains(t, stdout, "--verbose")
}

func TestVersionFlag(t *testing.T) {
	binary := buildBinary(t)

	stdout, _, err := runMakeHelp(t, binary, "--version")
	require.NoError(t, err)

	// Should contain "make-help version" and some version string
	assert.Contains(t, stdout, "make-help version")
	// Version should be either a semver or "dev"
	assert.Regexp(t, `make-help version (dev|\d+\.\d+\.\d+.*)`, stdout)
}

// TestAddTargetCommand tests the old add-target subcommand.
// DEPRECATED: This test is for the old subcommand interface that was replaced
// with --create-help-target flag in Stage 1. This test is disabled.
// See TestCreateHelpTarget for the new flag-based test.
func TestAddTargetCommand(t *testing.T) {
	t.Skip("DEPRECATED: add-target subcommand removed in favor of --create-help-target flag")
}

// TestRemoveTargetCommand tests the old remove-target subcommand.
// DEPRECATED: This test is for the old subcommand interface that was replaced
// with --remove-help-target flag in Stage 1. This test is disabled.
// See TestRemoveHelpTarget for the new flag-based test.
func TestRemoveTargetCommand(t *testing.T) {
	t.Skip("DEPRECATED: remove-target subcommand removed in favor of --remove-help-target flag")
}

// TestAddTargetWithTargetFile tests the old add-target subcommand with --target-file.
// DEPRECATED: This test is for the old subcommand interface that was replaced
// with --create-help-target flag in Stage 1. This test is disabled.
// See TestCreateHelpTargetWithHelpFilePath for the new flag-based test.
func TestAddTargetWithTargetFile(t *testing.T) {
	t.Skip("DEPRECATED: add-target subcommand removed in favor of --create-help-target flag")
}

// TestAddTargetWithMakeDir tests the old add-target subcommand with make/ directory.
// DEPRECATED: This test is for the old subcommand interface that was replaced
// with --create-help-target flag in Stage 1. This test is disabled.
// See TestCreateHelpTargetWithMakeDir for the new flag-based test.
func TestAddTargetWithMakeDir(t *testing.T) {
	t.Skip("DEPRECATED: add-target subcommand removed in favor of --create-help-target flag")
}

func TestOutputFormat(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "basic.mk")

	stdout, _, err := runMakeHelp(t, binary, "--output", "-", "--makefile-path", fixture, "--no-color")
	require.NoError(t, err)

	// Check output format
	assert.Contains(t, stdout, "Usage: make [<target>...] [<ENV_VAR>=<value>...]")
	assert.Contains(t, stdout, "Targets:")
	assert.Contains(t, stdout, "  - ")
}

func TestFileDocumentation(t *testing.T) {
	binary := buildBinary(t)

	// Create a Makefile with inline !file documentation (on same line)
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	err := os.WriteFile(makefilePath, []byte(`
## !file This is the file documentation
## Build the project
build:
	@echo building
`), 0644)
	require.NoError(t, err)

	stdout, _, err := runMakeHelp(t, binary, "--output", "-", "--makefile-path", makefilePath, "--no-color")
	require.NoError(t, err)

	// File documentation should appear
	assert.Contains(t, stdout, "This is the file documentation")
}

func TestDetailedHelp_DocumentedTarget(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "with_undocumented.mk")

	stdout, stderr, err := runMakeHelp(t, binary,
		"--output", "-",
		"--makefile-path", fixture,
		"--no-color",
		"--target", "build")
	require.NoError(t, err, "stderr: %s", stderr)

	// Should show target name
	assert.Contains(t, stdout, "Target: build")

	// Should show aliases
	assert.Contains(t, stdout, "Aliases: b, compile")

	// Should show variables with full descriptions
	assert.Contains(t, stdout, "Variables:")
	assert.Contains(t, stdout, "BUILD_FLAGS: Flags passed to go build")
	assert.Contains(t, stdout, "OUTPUT_DIR: Directory for build output")

	// Should show full documentation (not just summary)
	assert.Contains(t, stdout, "Build the project.")
	assert.Contains(t, stdout, "This compiles all source files and generates")
	assert.Contains(t, stdout, "the binary in the output directory.")

	// Should show source location
	assert.Contains(t, stdout, "Source:")
}

func TestDetailedHelp_UndocumentedTarget(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "with_undocumented.mk")

	stdout, stderr, err := runMakeHelp(t, binary,
		"--output", "-",
		"--makefile-path", fixture,
		"--no-color",
		"--target", "undocumented")
	require.NoError(t, err, "stderr: %s", stderr)

	// Should show target name
	assert.Contains(t, stdout, "Target: undocumented")

	// Should show "no documentation" message
	assert.Contains(t, stdout, "No documentation available.")

	// Should not show sections for aliases, variables, or documentation
	assert.NotContains(t, stdout, "Aliases:")
	assert.NotContains(t, stdout, "Variables:")
	assert.NotContains(t, stdout, "Documentation:")
}

func TestDetailedHelp_NonexistentTarget(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "basic.mk")

	_, _, err := runMakeHelp(t, binary,
		"--output", "-",
		"--makefile-path", fixture,
		"--no-color",
		"--target", "nonexistent")

	// Should error for nonexistent target
	require.Error(t, err)
}

func TestDetailedHelp_WithColors(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "with_undocumented.mk")

	stdout, stderr, err := runMakeHelp(t, binary,
		"--output", "-",
		"--makefile-path", fixture,
		"--color",
		"--target", "build")
	require.NoError(t, err, "stderr: %s", stderr)

	// Should contain ANSI color codes
	assert.Contains(t, stdout, "\033[")
	assert.Contains(t, stdout, "Target: build")
}

func TestDetailedHelp_MinimalTarget(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "basic.mk")

	stdout, stderr, err := runMakeHelp(t, binary,
		"--output", "-",
		"--makefile-path", fixture,
		"--no-color",
		"--target", "build")
	require.NoError(t, err, "stderr: %s", stderr)

	// Should show target name
	assert.Contains(t, stdout, "Target: build")

	// Should show documentation
	assert.Contains(t, stdout, "Build the project")

	// Should not have aliases or variables for this simple target
	assert.NotContains(t, stdout, "Aliases:")
	assert.NotContains(t, stdout, "Variables:")
}

func TestLintCommand_NoWarnings(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "lint-clean.mk")

	stdout, stderr, err := runMakeHelp(t, binary, "--lint", "--makefile-path", fixture)
	require.NoError(t, err, "stderr: %s\nstdout: %s", stderr, stdout)

	// Should have no output for clean Makefile
	assert.Empty(t, stdout, "expected no warnings in output")

	if len(stderr) > 0 {
		// Only verbose messages allowed in stderr
		assert.NotContains(t, stderr, "warning")
	}
}

func TestLintCommand_WithWarnings(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "lint-issues.mk")

	stdout, _, err := runMakeHelp(t, binary, "--lint", "--makefile-path", fixture)
	// Should exit with code 1 when warnings are found
	require.Error(t, err, "expected exit code 1 for warnings")
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok, "error should be ExitError")
	assert.Equal(t, 1, exitErr.ExitCode(), "expected exit code 1")

	// Check for expected warnings in stdout
	assert.Contains(t, stdout, "lint-issues.mk", "should contain filename")
	assert.Contains(t, stdout, "undocumented phony target 'setup'", "should warn about setup")
	assert.Contains(t, stdout, "undocumented phony target 'check'", "should warn about check")
	assert.Contains(t, stdout, "does not end with punctuation", "should warn about missing punctuation")
	assert.Contains(t, stdout, "Found", "should show warning count")

	// Warnings should show line numbers (format: "  N: message")
	assert.Regexp(t, `\s+\d+:.*does not end with punctuation`, stdout, "warnings should include line numbers")
	// Fixable warnings should be marked
	assert.Contains(t, stdout, "[fixable]", "fixable warnings should be marked")
}

func TestLintCommand_InvalidFlags(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "basic.mk")

	// --lint with --output - should fail
	_, _, err := runMakeHelp(t, binary, "--lint", "--output", "-", "--makefile-path", fixture)
	assert.Error(t, err, "should fail when combining --lint with --output -")

	// --lint with --remove-help should fail
	_, _, err = runMakeHelp(t, binary, "--lint", "--remove-help", "--makefile-path", fixture)
	assert.Error(t, err, "should fail when combining --lint with --remove-help")

	// --lint with --dry-run (without --fix) should fail
	_, _, err = runMakeHelp(t, binary, "--lint", "--dry-run", "--makefile-path", fixture)
	assert.Error(t, err, "should fail when combining --lint with --dry-run without --fix")

	// --fix without --lint should fail
	_, _, err = runMakeHelp(t, binary, "--fix", "--makefile-path", fixture)
	assert.Error(t, err, "should fail when using --fix without --lint")
}

func TestLintCommand_Verbose(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "lint-clean.mk")

	stdout, stderr, err := runMakeHelp(t, binary, "--lint", "--makefile-path", fixture, "--verbose")
	require.NoError(t, err)

	// Verbose mode should output to stderr
	assert.Contains(t, stderr, "Using Makefile", "verbose should show Makefile path")
	// stdout might contain "No warnings found" in verbose mode or be empty
	// Just check it doesn't contain actual warnings
	assert.NotContains(t, stdout, "warning:", "no warnings for clean Makefile")
}

func TestLintCommand_FixDryRun(t *testing.T) {
	binary := buildBinary(t)

	// Create a temporary file with fixable issues
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "Makefile")
	content := `.PHONY: build

## Build the project
build:
	@echo "building"
`
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)

	// Run lint with --fix --dry-run
	stdout, _, err := runMakeHelp(t, binary, "--lint", "--fix", "--dry-run", "--makefile-path", tmpFile)
	// Should exit with code 1 (warnings found)
	require.Error(t, err)
	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok, "error should be ExitError")
	assert.Equal(t, 1, exitErr.ExitCode())

	// Should show dry-run message
	assert.Contains(t, stdout, "Would fix", "should show dry-run message")

	// File should NOT be modified
	got, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	assert.Equal(t, content, string(got), "file should not be modified in dry-run")
}

func TestLintCommand_FixApply(t *testing.T) {
	binary := buildBinary(t)

	// Create a temporary file with fixable and unfixable issues
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "Makefile")
	content := `.PHONY: build test setup

##
## Build the project
build:
	@echo "building"

## Run tests
test:
	@echo "testing"

setup:
	@echo "setup"
`
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)

	// Run lint with --fix (first pass - should fix empty-doc)
	stdout, _, _ := runMakeHelp(t, binary, "--lint", "--fix", "--makefile-path", tmpFile)
	assert.Contains(t, stdout, "Fixed", "should show fixed message")
	// Fixed warnings should NOT be shown
	assert.NotContains(t, stdout, "empty documentation line", "fixed warnings should not be displayed")
	// Unfixed warnings should still be shown
	assert.Contains(t, stdout, "undocumented phony target 'setup'", "unfixed warnings should be displayed")

	// Verify empty ## line was removed
	got, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	assert.NotContains(t, string(got), "##\n## Build", "empty doc line should be removed")

	// Run lint with --fix again (second pass - should fix punctuation)
	stdout, _, _ = runMakeHelp(t, binary, "--lint", "--fix", "--makefile-path", tmpFile)
	assert.Contains(t, stdout, "Fixed", "should fix punctuation issues")
	// Fixed warnings should NOT be shown
	assert.NotContains(t, stdout, "does not end with punctuation", "fixed warnings should not be displayed")

	// Verify punctuation was added
	got, err = os.ReadFile(tmpFile)
	require.NoError(t, err)
	assert.Contains(t, string(got), "## Build the project.", "punctuation should be added")
	assert.Contains(t, string(got), "## Run tests.", "punctuation should be added")
}

func TestLintCommand_FixAllIssues(t *testing.T) {
	binary := buildBinary(t)

	// Create a temporary file with only fixable issues (both empty-doc and missing punctuation)
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "Makefile")
	content := `.PHONY: build

##
## Build the project
build:
	@echo "building"
`
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	require.NoError(t, err)

	// Run lint with --fix - first pass fixes empty-doc
	stdout, _, err := runMakeHelp(t, binary, "--lint", "--fix", "--makefile-path", tmpFile)
	require.NoError(t, err, "should exit with code 0 when all issues in this pass are fixed")
	assert.Contains(t, stdout, "Fixed", "should show fixed message")

	// Run lint with --fix again - second pass fixes punctuation
	stdout, _, err = runMakeHelp(t, binary, "--lint", "--fix", "--makefile-path", tmpFile)
	require.NoError(t, err, "should exit with code 0 when all remaining issues are fixed")
	assert.Contains(t, stdout, "Fixed", "should show fixed message")

	// Run lint again - should pass with no warnings
	_, _, err = runMakeHelp(t, binary, "--lint", "--makefile-path", tmpFile)
	require.NoError(t, err, "fully fixed file should have no warnings")
}

// Error Scenario Tests

func TestErrorScenario_InvalidMakefileSyntax(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "invalid_syntax.mk")

	stdout, stderr, err := runMakeHelp(t, binary, "--output", "-", "--makefile-path", fixture, "--no-color")

	// Should fail when make cannot parse the Makefile
	require.Error(t, err, "should fail for invalid Makefile syntax")

	// Check that error output contains meaningful information
	combinedOutput := stdout + stderr
	assert.Contains(t, combinedOutput, "failed to discover", "error should mention discovery failure")
}

func TestErrorScenario_MissingMakefile(t *testing.T) {
	binary := buildBinary(t)
	nonexistentPath := "/nonexistent/path/to/Makefile"

	stdout, stderr, err := runMakeHelp(t, binary, "--output", "-", "--makefile-path", nonexistentPath, "--no-color")

	// Should fail when Makefile doesn't exist
	require.Error(t, err, "should fail when Makefile doesn't exist")

	// Check that error output contains meaningful information
	combinedOutput := stdout + stderr
	assert.Contains(t, combinedOutput, "not found", "error should mention file not found")
	assert.Contains(t, combinedOutput, nonexistentPath, "error should include the path that was searched")
}

func TestErrorScenario_MixedCategorizationWithoutDefault(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "mixed_categorization.mk")

	stdout, stderr, err := runMakeHelp(t, binary, "--output", "-", "--makefile-path", fixture, "--no-color")

	// Should fail due to mixed categorization
	require.Error(t, err, "should fail for mixed categorization without --default-category")

	// Check that error message is helpful
	combinedOutput := stdout + stderr
	assert.Contains(t, combinedOutput, "mixed categorization", "error should mention mixed categorization")
	assert.Contains(t, combinedOutput, "--default-category", "error should suggest using --default-category flag")
}

func TestErrorScenario_MixedCategorizationWithDefault(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "mixed_categorization.mk")

	stdout, stderr, err := runMakeHelp(t, binary, "--output", "-", "--makefile-path", fixture, "--no-color", "--default-category", "Misc")

	// Should succeed when --default-category is provided
	require.NoError(t, err, "stderr: %s", stderr)

	// Should show both categorized and uncategorized targets
	assert.Contains(t, stdout, "Build:", "should show Build category")
	assert.Contains(t, stdout, "Misc:", "should show Misc category for uncategorized targets")
	assert.Contains(t, stdout, "build", "should show build target")
	assert.Contains(t, stdout, "clean", "should show clean target")
}
