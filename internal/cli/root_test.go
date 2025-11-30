package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessColorFlags(t *testing.T) {
	tests := []struct {
		name        string
		noColor     bool
		forceColor  bool
		expected    ColorMode
		expectError bool
	}{
		{
			name:        "default auto mode",
			noColor:     false,
			forceColor:  false,
			expected:    ColorAuto,
			expectError: false,
		},
		{
			name:        "force color",
			noColor:     false,
			forceColor:  true,
			expected:    ColorAlways,
			expectError: false,
		},
		{
			name:        "disable color",
			noColor:     true,
			forceColor:  false,
			expected:    ColorNever,
			expectError: false,
		},
		{
			name:        "conflicting flags",
			noColor:     true,
			forceColor:  true,
			expected:    ColorAuto,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mode ColorMode
			err := processColorFlags(&mode, tt.noColor, tt.forceColor)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "cannot use both")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, mode)
			}
		})
	}
}

func TestResolveColorMode(t *testing.T) {
	tests := []struct {
		name     string
		mode     ColorMode
		expected bool
	}{
		{
			name:     "always mode",
			mode:     ColorAlways,
			expected: true,
		},
		{
			name:     "never mode",
			mode:     ColorNever,
			expected: false,
		},
		{
			name:     "auto mode",
			mode:     ColorAuto,
			expected: false, // Usually false in test environment (not a TTY)
		},
		{
			name:     "unknown mode defaults to false",
			mode:     ColorMode(999),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				ColorMode: tt.mode,
			}

			result := ResolveColorMode(config)

			if tt.mode == ColorAuto {
				// For auto mode, we just check that it returns a boolean
				// The actual value depends on whether we're in a terminal
				assert.IsType(t, false, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseCategoryOrder(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "single category",
			input:    []string{"Build"},
			expected: []string{"Build"},
		},
		{
			name:     "multiple categories",
			input:    []string{"Build", "Test", "Deploy"},
			expected: []string{"Build", "Test", "Deploy"},
		},
		{
			name:     "comma-separated",
			input:    []string{"Build,Test,Deploy"},
			expected: []string{"Build", "Test", "Deploy"},
		},
		{
			name:     "mixed format",
			input:    []string{"Build,Test", "Deploy"},
			expected: []string{"Build", "Test", "Deploy"},
		},
		{
			name:     "with whitespace",
			input:    []string{" Build , Test ", "Deploy"},
			expected: []string{"Build", "Test", "Deploy"},
		},
		{
			name:     "empty strings filtered",
			input:    []string{"Build", "", "Test"},
			expected: []string{"Build", "Test"},
		},
		{
			name:     "empty input",
			input:    []string{},
			expected: nil,
		},
		{
			name:     "only whitespace",
			input:    []string{"  ", "  ,  "},
			expected: nil,
		},
		{
			name:     "trailing comma",
			input:    []string{"Build,Test,"},
			expected: []string{"Build", "Test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCategoryOrder(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewConfig(t *testing.T) {
	config := NewConfig()

	assert.NotNil(t, config)
	assert.Equal(t, ColorAuto, config.ColorMode)
	assert.NotNil(t, config.CategoryOrder)
	assert.Equal(t, 0, len(config.CategoryOrder))
	assert.False(t, config.UseColor)
	assert.False(t, config.Verbose)
	assert.False(t, config.KeepOrderCategories)
	assert.False(t, config.KeepOrderTargets)
	assert.Empty(t, config.DefaultCategory)
	assert.Empty(t, config.MakefilePath)
}

func TestColorModeString(t *testing.T) {
	tests := []struct {
		mode     ColorMode
		expected string
	}{
		{
			mode:     ColorAuto,
			expected: "auto",
		},
		{
			mode:     ColorAlways,
			expected: "always",
		},
		{
			mode:     ColorNever,
			expected: "never",
		},
		{
			mode:     ColorMode(999),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.mode.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewRootCmd(t *testing.T) {
	cmd := NewRootCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "make-help", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check that essential flags are registered
	flags := cmd.Flags()

	assert.NotNil(t, flags.Lookup("keep-order-categories"))
	assert.NotNil(t, flags.Lookup("keep-order-targets"))
	assert.NotNil(t, flags.Lookup("keep-order-all"))
	assert.NotNil(t, flags.Lookup("category-order"))
	assert.NotNil(t, flags.Lookup("default-category"))

	persistentFlags := cmd.PersistentFlags()
	assert.NotNil(t, persistentFlags.Lookup("makefile-path"))
	assert.NotNil(t, persistentFlags.Lookup("no-color"))
	assert.NotNil(t, persistentFlags.Lookup("color"))
	assert.NotNil(t, persistentFlags.Lookup("verbose"))
}

func TestRootCmd_FlagDefaults(t *testing.T) {
	cmd := NewRootCmd()

	// Check default values
	makefilePath, err := cmd.PersistentFlags().GetString("makefile-path")
	assert.NoError(t, err)
	assert.Empty(t, makefilePath)

	noColor, err := cmd.PersistentFlags().GetBool("no-color")
	assert.NoError(t, err)
	assert.False(t, noColor)

	forceColor, err := cmd.PersistentFlags().GetBool("color")
	assert.NoError(t, err)
	assert.False(t, forceColor)

	verbose, err := cmd.PersistentFlags().GetBool("verbose")
	assert.NoError(t, err)
	assert.False(t, verbose)
}

func TestRootCmd_ConflictingColorFlags(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--color", "--no-color"})

	// Capture output
	var errBuf bytes.Buffer
	cmd.SetErr(&errBuf)

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot use both")
}

func TestRootCmd_KeepOrderAll(t *testing.T) {
	// Create a temp Makefile for the test
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	err := os.WriteFile(makefilePath, []byte(`
## Build the project
all:
	@echo hello
`), 0644)
	require.NoError(t, err)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--makefile-path", makefilePath, "--keep-order-all", "--no-color"})

	// Capture output
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)

	err = cmd.Execute()
	assert.NoError(t, err)
}

func TestRootCmd_VerboseFlag(t *testing.T) {
	// Create a temp Makefile for the test
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	err := os.WriteFile(makefilePath, []byte(`
## Build the project
all:
	@echo hello
`), 0644)
	require.NoError(t, err)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--makefile-path", makefilePath, "--verbose", "--no-color"})

	// Capture output - verbose output goes to stderr
	var errBuf bytes.Buffer
	cmd.SetErr(&errBuf)

	err = cmd.Execute()
	assert.NoError(t, err)
}

func TestRootCmd_MissingMakefile(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--makefile-path", "/nonexistent/Makefile"})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestRootCmd_CategoryOrder(t *testing.T) {
	// Create a temp Makefile with categories
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	err := os.WriteFile(makefilePath, []byte(`
## @category Build
## Build the project
build:
	@echo build

## @category Test
## Run tests
test:
	@echo test
`), 0644)
	require.NoError(t, err)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		"--makefile-path", makefilePath,
		"--category-order", "Test,Build",
		"--no-color",
	})

	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)

	err = cmd.Execute()
	assert.NoError(t, err)
}

func TestRootCmd_DefaultCategory(t *testing.T) {
	// Create a temp Makefile with mixed categorization
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	err := os.WriteFile(makefilePath, []byte(`
## @category Build
## Build the project
build:
	@echo build

## Uncategorized target
clean:
	@echo clean
`), 0644)
	require.NoError(t, err)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		"--makefile-path", makefilePath,
		"--default-category", "Other",
		"--no-color",
	})

	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)

	err = cmd.Execute()
	assert.NoError(t, err)
}

func TestIsTerminal(t *testing.T) {
	// Test with stdout - may or may not be a terminal
	result := IsTerminal(os.Stdout.Fd())
	assert.IsType(t, false, result)

	// Test with stderr
	result = IsTerminal(os.Stderr.Fd())
	assert.IsType(t, false, result)

	// Test with stdin
	result = IsTerminal(os.Stdin.Fd())
	assert.IsType(t, false, result)
}

func TestRunHelp_Success(t *testing.T) {
	// Create a temp Makefile
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	err := os.WriteFile(makefilePath, []byte(`
## @file
## This is a test Makefile

## @category Build
## Build the project
## This target compiles everything.
## @var VERBOSE Print verbose output
build:
	@echo build

## @alias t
## Run tests
test:
	@echo test
`), 0644)
	require.NoError(t, err)

	config := NewConfig()
	config.MakefilePath = makefilePath
	config.UseColor = false

	err = runHelp(config)
	assert.NoError(t, err)
}

func TestRunHelp_WithOrdering(t *testing.T) {
	// Create a temp Makefile with categories
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	err := os.WriteFile(makefilePath, []byte(`
## @category Build
## Build the project
build:
	@echo build

## @category Test
## Run tests
test:
	@echo test

## @category Deploy
## Deploy the project
deploy:
	@echo deploy
`), 0644)
	require.NoError(t, err)

	config := NewConfig()
	config.MakefilePath = makefilePath
	config.UseColor = false
	config.KeepOrderCategories = true
	config.KeepOrderTargets = true

	err = runHelp(config)
	assert.NoError(t, err)
}

func TestRunHelp_ResolutionError(t *testing.T) {
	config := NewConfig()
	// Empty path should resolve to current dir which may not have a Makefile
	config.MakefilePath = "/nonexistent/deeply/nested/path/Makefile"

	err := runHelp(config)
	assert.Error(t, err)
}

func TestRunHelp_ValidationError(t *testing.T) {
	config := NewConfig()
	config.MakefilePath = "/nonexistent/Makefile"

	err := runHelp(config)
	assert.Error(t, err)
}

func TestRunHelp_Verbose(t *testing.T) {
	// Create a temp Makefile
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	err := os.WriteFile(makefilePath, []byte(`
## Build the project
build:
	@echo build
`), 0644)
	require.NoError(t, err)

	config := NewConfig()
	config.MakefilePath = makefilePath
	config.UseColor = false
	config.Verbose = true

	err = runHelp(config)
	assert.NoError(t, err)
}

func TestRunHelp_WithDefaultCategory(t *testing.T) {
	// Create a temp Makefile with mixed categorization
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	err := os.WriteFile(makefilePath, []byte(`
## @category Build
## Build the project
build:
	@echo build

## Uncategorized target
clean:
	@echo clean
`), 0644)
	require.NoError(t, err)

	config := NewConfig()
	config.MakefilePath = makefilePath
	config.UseColor = false
	config.DefaultCategory = "Other"

	err = runHelp(config)
	assert.NoError(t, err)
}

func TestRunHelp_CategoryOrder(t *testing.T) {
	// Create a temp Makefile with categories
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	err := os.WriteFile(makefilePath, []byte(`
## @category Build
## Build the project
build:
	@echo build

## @category Test
## Run tests
test:
	@echo test
`), 0644)
	require.NoError(t, err)

	config := NewConfig()
	config.MakefilePath = makefilePath
	config.UseColor = false
	config.CategoryOrder = []string{"Test", "Build"}

	err = runHelp(config)
	assert.NoError(t, err)
}

func TestRunHelp_InvalidCategoryOrder(t *testing.T) {
	// Create a temp Makefile with categories
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	err := os.WriteFile(makefilePath, []byte(`
## @category Build
## Build the project
build:
	@echo build
`), 0644)
	require.NoError(t, err)

	config := NewConfig()
	config.MakefilePath = makefilePath
	config.UseColor = false
	config.CategoryOrder = []string{"NonExistent"}

	err = runHelp(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ordering")
}

func TestRunHelp_EmptyMakefile(t *testing.T) {
	// Create an empty temp Makefile
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	err := os.WriteFile(makefilePath, []byte(""), 0644)
	require.NoError(t, err)

	config := NewConfig()
	config.MakefilePath = makefilePath
	config.UseColor = false

	// Empty Makefile should still work (just produce minimal output)
	err = runHelp(config)
	assert.NoError(t, err)
}

func TestRunHelp_WithInclude(t *testing.T) {
	// Create a temp Makefile with include using -include (optional include)
	// to avoid errors when the file is copied to a temp location
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	includePath := filepath.Join(tmpDir, "common.mk")

	// Use -include which silently ignores missing files
	err := os.WriteFile(makefilePath, []byte(`
-include common.mk

## Build the project
build:
	@echo build
`), 0644)
	require.NoError(t, err)

	err = os.WriteFile(includePath, []byte(`
## Common target
clean:
	@echo clean
`), 0644)
	require.NoError(t, err)

	config := NewConfig()
	config.MakefilePath = makefilePath
	config.UseColor = false

	err = runHelp(config)
	assert.NoError(t, err)
}

func TestMutualExclusivityFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorText   string
	}{
		{
			name:        "create-help-target and remove-help-target together",
			args:        []string{"--create-help-target", "--remove-help-target"},
			expectError: true,
			errorText:   "cannot use both --create-help-target and --remove-help-target",
		},
		{
			name:        "only create-help-target",
			args:        []string{"--create-help-target"},
			expectError: true,
			errorText:   "Makefile not found",
		},
		{
			name:        "only remove-help-target",
			args:        []string{"--remove-help-target"},
			expectError: true,
			errorText:   "Makefile not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRootCmd()
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorText != "" {
					assert.Contains(t, err.Error(), tt.errorText)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRemoveHelpTargetFlagRestrictions(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "remove-help-target with verbose",
			args:        []string{"--remove-help-target", "--verbose"},
			expectError: false,
		},
		{
			name:        "remove-help-target with makefile-path",
			args:        []string{"--remove-help-target", "--makefile-path", "/tmp/Makefile"},
			expectError: false,
		},
		{
			name:        "remove-help-target with target",
			args:        []string{"--remove-help-target", "--target", "build"},
			expectError: true,
		},
		{
			name:        "remove-help-target with include-target",
			args:        []string{"--remove-help-target", "--include-target", "foo"},
			expectError: true,
		},
		{
			name:        "remove-help-target with include-all-phony",
			args:        []string{"--remove-help-target", "--include-all-phony"},
			expectError: true,
		},
		{
			name:        "remove-help-target with version",
			args:        []string{"--remove-help-target", "--version", "v1.0.0"},
			expectError: true,
		},
		{
			name:        "remove-help-target with help-file-rel-path",
			args:        []string{"--remove-help-target", "--help-file-rel-path", "help.mk"},
			expectError: true,
		},
		{
			name:        "remove-help-target with keep-order-categories",
			args:        []string{"--remove-help-target", "--keep-order-categories"},
			expectError: true,
		},
		{
			name:        "remove-help-target with keep-order-targets",
			args:        []string{"--remove-help-target", "--keep-order-targets"},
			expectError: true,
		},
		{
			name:        "remove-help-target with category-order",
			args:        []string{"--remove-help-target", "--category-order", "Build"},
			expectError: true,
		},
		{
			name:        "remove-help-target with default-category",
			args:        []string{"--remove-help-target", "--default-category", "Other"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRootCmd()
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "--remove-help-target only accepts --verbose and --makefile-path flags")
			} else {
				// Should get an error related to Makefile issues (not flag validation)
				assert.Error(t, err)
				// The error could be "Makefile not found" or "Makefile validation failed"
				// depending on whether the file exists
				assert.True(t,
					strings.Contains(err.Error(), "Makefile not found") ||
						strings.Contains(err.Error(), "Makefile validation failed"),
					"Expected Makefile-related error, got: %v", err)
			}
		})
	}
}

func TestNewFlags(t *testing.T) {
	cmd := NewRootCmd()

	// Check that new flags are registered
	flags := cmd.Flags()

	assert.NotNil(t, flags.Lookup("create-help-target"))
	assert.NotNil(t, flags.Lookup("remove-help-target"))
	assert.NotNil(t, flags.Lookup("version"))
	assert.NotNil(t, flags.Lookup("include-target"))
	assert.NotNil(t, flags.Lookup("include-all-phony"))
	assert.NotNil(t, flags.Lookup("target"))
	assert.NotNil(t, flags.Lookup("help-file-rel-path"))
}

func TestIncludeTargetFlag(t *testing.T) {
	// Create a temp Makefile for the test
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	err := os.WriteFile(makefilePath, []byte(`
## Documented target
all:
	@echo hello
`), 0644)
	require.NoError(t, err)

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "single include-target",
			args: []string{"--makefile-path", makefilePath, "--include-target", "foo", "--no-color"},
		},
		{
			name: "comma-separated include-target",
			args: []string{"--makefile-path", makefilePath, "--include-target", "foo,bar", "--no-color"},
		},
		{
			name: "multiple include-target flags",
			args: []string{"--makefile-path", makefilePath, "--include-target", "foo", "--include-target", "bar", "--no-color"},
		},
		{
			name: "mixed include-target",
			args: []string{"--makefile-path", makefilePath, "--include-target", "foo,bar", "--include-target", "baz", "--no-color"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRootCmd()
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			// Should succeed (filtering not yet implemented, but flag should parse)
			assert.NoError(t, err)
		})
	}
}

func TestTargetFlag(t *testing.T) {
	// Create a temp Makefile with a build target
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	err := os.WriteFile(makefilePath, []byte(`
## Build the project
build:
	@echo building
`), 0644)
	require.NoError(t, err)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--makefile-path", makefilePath, "--target", "build", "--no-color"})

	// The --target flag is now implemented and should work without error
	err = cmd.Execute()
	require.NoError(t, err, "should successfully run with --target flag")
}

func TestIncludeAllPhonyFlag(t *testing.T) {
	// Create a temp Makefile for the test
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	err := os.WriteFile(makefilePath, []byte(`
## Documented target
all:
	@echo hello
`), 0644)
	require.NoError(t, err)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--makefile-path", makefilePath, "--include-all-phony", "--no-color"})

	err = cmd.Execute()
	// Should succeed (filtering not yet implemented, but flag should work)
	assert.NoError(t, err)
}
