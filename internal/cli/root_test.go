package cli

import (
	"bytes"
	"os"
	"path/filepath"
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
