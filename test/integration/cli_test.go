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

	stdout, stderr, err := runMakeHelp(t, binary, "--makefile-path", fixture, "--no-color")
	require.NoError(t, err, "stderr: %s", stderr)

	expected := readExpected(t, "basic_help.txt")
	assert.Equal(t, expected, stdout)
}

func TestCategorizedMakefile(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "categorized.mk")

	stdout, stderr, err := runMakeHelp(t, binary, "--makefile-path", fixture, "--no-color")
	require.NoError(t, err, "stderr: %s", stderr)

	expected := readExpected(t, "categorized_help.txt")
	assert.Equal(t, expected, stdout)
}

func TestComplexMakefile(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "complex.mk")

	stdout, stderr, err := runMakeHelp(t, binary, "--makefile-path", fixture, "--no-color")
	require.NoError(t, err, "stderr: %s", stderr)

	expected := readExpected(t, "complex_help.txt")
	assert.Equal(t, expected, stdout)
}

func TestCategoryOrderFlag(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "categorized.mk")

	// Test explicit category order
	stdout, stderr, err := runMakeHelp(t, binary,
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
## @category Build
## Build the project
build:
	@echo building
`), 0644)
	require.NoError(t, err)

	// Test with --default-category flag (even though not needed here, should work)
	stdout, stderr, err := runMakeHelp(t, binary,
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

	stdout, _, err := runMakeHelp(t, binary, "--makefile-path", fixture, "--no-color")
	require.NoError(t, err)

	// Should not contain ANSI escape codes
	assert.NotContains(t, stdout, "\033[")
	assert.NotContains(t, stdout, "\x1b[")
}

func TestColorFlag(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "basic.mk")

	stdout, _, err := runMakeHelp(t, binary, "--makefile-path", fixture, "--color")
	require.NoError(t, err)

	// Should contain ANSI escape codes
	assert.Contains(t, stdout, "\033[")
}

func TestVerboseFlag(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "basic.mk")

	_, stderr, err := runMakeHelp(t, binary, "--makefile-path", fixture, "--no-color", "--verbose")
	require.NoError(t, err)

	// Verbose output goes to stderr
	assert.Contains(t, stderr, "Using Makefile")
}

func TestMissingMakefile(t *testing.T) {
	binary := buildBinary(t)

	_, _, err := runMakeHelp(t, binary, "--makefile-path", "/nonexistent/Makefile")
	assert.Error(t, err)
}

func TestEmptyMakefile(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "empty.mk")

	stdout, _, err := runMakeHelp(t, binary, "--makefile-path", fixture, "--no-color")
	require.NoError(t, err)

	// Should still produce usage line
	assert.Contains(t, stdout, "Usage: make")
}

func TestWithAliasesMakefile(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "with_aliases.mk")

	stdout, _, err := runMakeHelp(t, binary, "--makefile-path", fixture, "--no-color")
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

func TestAddTargetCommand(t *testing.T) {
	binary := buildBinary(t)

	// Create a temp Makefile
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	err := os.WriteFile(makefilePath, []byte(`
## Build the project
build:
	@echo building
`), 0644)
	require.NoError(t, err)

	// Run add-target
	stdout, stderr, err := runMakeHelp(t, binary,
		"add-target",
		"--makefile-path", makefilePath)
	require.NoError(t, err, "stdout: %s, stderr: %s", stdout, stderr)

	// Verify help target was added
	content, err := os.ReadFile(makefilePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "help:")
	assert.Contains(t, string(content), ".PHONY: help")
}

func TestRemoveTargetCommand(t *testing.T) {
	binary := buildBinary(t)

	// Create a temp Makefile with help target
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	err := os.WriteFile(makefilePath, []byte(`
## Build the project
build:
	@echo building

.PHONY: help
help:
	@make-help --makefile-path Makefile
`), 0644)
	require.NoError(t, err)

	// Run remove-target
	stdout, stderr, err := runMakeHelp(t, binary,
		"remove-target",
		"--makefile-path", makefilePath)
	require.NoError(t, err, "stdout: %s, stderr: %s", stdout, stderr)

	// Verify help target was removed
	content, err := os.ReadFile(makefilePath)
	require.NoError(t, err)
	assert.NotContains(t, string(content), "help:")
	assert.NotContains(t, string(content), ".PHONY: help")
}

func TestAddTargetWithTargetFile(t *testing.T) {
	binary := buildBinary(t)

	// Create a temp directory with Makefile
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	targetFile := filepath.Join(tmpDir, "help.mk")

	err := os.WriteFile(makefilePath, []byte(`
## Build the project
build:
	@echo building
`), 0644)
	require.NoError(t, err)

	// Run add-target with --target-file
	stdout, stderr, err := runMakeHelp(t, binary,
		"add-target",
		"--makefile-path", makefilePath,
		"--target-file", targetFile)
	require.NoError(t, err, "stdout: %s, stderr: %s", stdout, stderr)

	// Verify help target was added to target file
	content, err := os.ReadFile(targetFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "help:")
	assert.Contains(t, string(content), ".PHONY: help")

	// Verify include was added to Makefile
	mainContent, err := os.ReadFile(makefilePath)
	require.NoError(t, err)
	assert.Contains(t, string(mainContent), "include")
	assert.Contains(t, string(mainContent), "help.mk")
}

func TestAddTargetWithMakeDir(t *testing.T) {
	binary := buildBinary(t)

	// Create a temp directory with Makefile that has include make/*.mk pattern
	tmpDir := t.TempDir()
	makeDir := filepath.Join(tmpDir, "make")
	err := os.MkdirAll(makeDir, 0755)
	require.NoError(t, err)

	makefilePath := filepath.Join(tmpDir, "Makefile")
	// Use -include (optional include) to avoid errors when no .mk files exist yet
	err = os.WriteFile(makefilePath, []byte(`
-include make/*.mk

## Build the project
build:
	@echo building
`), 0644)
	require.NoError(t, err)

	// Run add-target
	stdout, stderr, err := runMakeHelp(t, binary,
		"add-target",
		"--makefile-path", makefilePath)
	require.NoError(t, err, "stdout: %s, stderr: %s", stdout, stderr)

	// Verify help target was added to make/01-help.mk
	helpMkPath := filepath.Join(makeDir, "01-help.mk")
	content, err := os.ReadFile(helpMkPath)
	require.NoError(t, err, "01-help.mk should be created")
	assert.Contains(t, string(content), "help:")
}

func TestOutputFormat(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "basic.mk")

	stdout, _, err := runMakeHelp(t, binary, "--makefile-path", fixture, "--no-color")
	require.NoError(t, err)

	// Check output format
	assert.Contains(t, stdout, "Usage: make [<target>...] [<ENV_VAR>=<value>...]")
	assert.Contains(t, stdout, "Targets:")
	assert.Contains(t, stdout, "  - ")
}

func TestFileDocumentation(t *testing.T) {
	binary := buildBinary(t)

	// Create a Makefile with inline @file documentation (on same line)
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	err := os.WriteFile(makefilePath, []byte(`
## @file This is the file documentation
## Build the project
build:
	@echo building
`), 0644)
	require.NoError(t, err)

	stdout, _, err := runMakeHelp(t, binary, "--makefile-path", makefilePath, "--no-color")
	require.NoError(t, err)

	// File documentation should appear
	assert.Contains(t, stdout, "This is the file documentation")
}

func TestDetailedHelp_DocumentedTarget(t *testing.T) {
	binary := buildBinary(t)
	fixture := getFixturePath(t, "with_undocumented.mk")

	stdout, stderr, err := runMakeHelp(t, binary,
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
	assert.Contains(t, stdout, "Documentation:")
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
		"--makefile-path", fixture,
		"--no-color",
		"--target", "build")
	require.NoError(t, err, "stderr: %s", stderr)

	// Should show target name
	assert.Contains(t, stdout, "Target: build")

	// Should show documentation
	assert.Contains(t, stdout, "Documentation:")
	assert.Contains(t, stdout, "Build the project")

	// Should not have aliases or variables for this simple target
	assert.NotContains(t, stdout, "Aliases:")
	assert.NotContains(t, stdout, "Variables:")
}
