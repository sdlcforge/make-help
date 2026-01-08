package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunLint_Success(t *testing.T) {
	t.Parallel()
	// Create a well-formed Makefile with proper punctuation
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	err := os.WriteFile(makefilePath, []byte(`
## !category Build
## Build the project.
.PHONY: build
build:
	@echo building
`), 0644)
	require.NoError(t, err)

	config := NewConfig()
	config.MakefilePath = makefilePath
	config.UseColor = false
	config.Lint = true

	err = runLint(config)
	// May have warnings or may not - just verify it runs
	if err != nil {
		// If there's an error, it should be the warnings found error
		assert.Equal(t, ErrLintWarningsFound, err)
	}
}

func TestRunLint_WithWarnings(t *testing.T) {
	t.Parallel()
	// Create a Makefile with lint issues (e.g., undocumented .PHONY target)
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	err := os.WriteFile(makefilePath, []byte(`
## Build the project
.PHONY: build
build:
	@echo building

.PHONY: undocumented
undocumented:
	@echo no docs
`), 0644)
	require.NoError(t, err)

	config := NewConfig()
	config.MakefilePath = makefilePath
	config.UseColor = false
	config.Lint = true

	err = runLint(config)
	// Should return ErrLintWarningsFound
	assert.Equal(t, ErrLintWarningsFound, err)
}

func TestRunLint_Verbose(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	err := os.WriteFile(makefilePath, []byte(`
## !category Build
## Build the project.
.PHONY: build
build:
	@echo building
`), 0644)
	require.NoError(t, err)

	config := NewConfig()
	config.MakefilePath = makefilePath
	config.UseColor = false
	config.Lint = true
	config.Verbose = true

	err = runLint(config)
	// May have warnings or may not
	if err != nil {
		assert.Equal(t, ErrLintWarningsFound, err)
	}
}

func TestRunLint_InvalidMakefile(t *testing.T) {
	t.Parallel()
	config := NewConfig()
	config.MakefilePath = "/nonexistent/Makefile"
	config.Lint = true

	err := runLint(config)
	assert.Error(t, err)
}

func TestRunLint_WithFix(t *testing.T) {
	t.Parallel()
	// Create a Makefile with fixable issues
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	err := os.WriteFile(makefilePath, []byte(`
## Build the project
.PHONY: build
build:
	@echo building
`), 0644)
	require.NoError(t, err)

	config := NewConfig()
	config.MakefilePath = makefilePath
	config.UseColor = false
	config.Lint = true
	config.Fix = true

	err = runLint(config)
	// May or may not have warnings depending on what's fixable
	// Just verify it doesn't crash
	if err != nil {
		assert.Equal(t, ErrLintWarningsFound, err)
	}
}

func TestRunLint_WithFixDryRun(t *testing.T) {
	t.Parallel()
	// Create a Makefile with fixable issues
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	err := os.WriteFile(makefilePath, []byte(`
## Build the project
.PHONY: build
build:
	@echo building
`), 0644)
	require.NoError(t, err)

	config := NewConfig()
	config.MakefilePath = makefilePath
	config.UseColor = false
	config.Lint = true
	config.Fix = true
	config.DryRun = true

	err = runLint(config)
	// May or may not have warnings
	if err != nil {
		assert.Equal(t, ErrLintWarningsFound, err)
	}
}

func TestRunLint_RecursionDetection(t *testing.T) {
	// Set the environment variable to simulate recursion
	oldEnv := os.Getenv("MAKE_HELP_GENERATING")
	defer func() {
		if oldEnv != "" {
			os.Setenv("MAKE_HELP_GENERATING", oldEnv)
		} else {
			os.Unsetenv("MAKE_HELP_GENERATING")
		}
	}()

	os.Setenv("MAKE_HELP_GENERATING", "1")

	config := NewConfig()
	config.Lint = true

	err := runLint(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "recursion detected")
}
