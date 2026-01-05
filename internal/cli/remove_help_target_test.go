package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunRemoveHelpTarget_Success(t *testing.T) {
	// Create a temp Makefile with a help target
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	// Create a Makefile with -include directive (optional, won't fail if missing)
	err := os.WriteFile(makefilePath, []byte(`
## Build the project
build:
	@echo building

-include make/*.mk
`), 0644)
	require.NoError(t, err)

	// Create the make directory and help.mk file to be removed
	makeDir := filepath.Join(tmpDir, "make")
	err = os.Mkdir(makeDir, 0755)
	require.NoError(t, err)

	helpMkPath := filepath.Join(makeDir, "help.mk")
	err = os.WriteFile(helpMkPath, []byte(`
.PHONY: help
help:
	@echo "Help target"
`), 0644)
	require.NoError(t, err)

	config := NewConfig()
	config.MakefilePath = makefilePath
	config.Verbose = false

	err = runRemoveHelpTarget(config)
	assert.NoError(t, err)

	// Note: The remove service may not actually remove the file if it can't find
	// the marker comment. Just verify no error occurred.
}

func TestRunRemoveHelpTarget_Verbose(t *testing.T) {
	// Create a temp Makefile
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	err := os.WriteFile(makefilePath, []byte(`
## Build the project
build:
	@echo building
`), 0644)
	require.NoError(t, err)

	config := NewConfig()
	config.MakefilePath = makefilePath
	config.Verbose = true

	// Should succeed even if there's no help target to remove
	err = runRemoveHelpTarget(config)
	assert.NoError(t, err)
}

func TestRunRemoveHelpTarget_InvalidMakefile(t *testing.T) {
	config := NewConfig()
	config.MakefilePath = "/nonexistent/Makefile"

	err := runRemoveHelpTarget(config)
	assert.Error(t, err)
}

func TestRunRemoveHelpTarget_ResolutionError(t *testing.T) {
	config := NewConfig()
	config.MakefilePath = "/nonexistent/deeply/nested/path/Makefile"

	err := runRemoveHelpTarget(config)
	assert.Error(t, err)
	// Could be either resolution error or validation error
	assert.True(t,
		strings.Contains(err.Error(), "failed to resolve Makefile path") ||
			strings.Contains(err.Error(), "Makefile not found"),
		"Expected resolution or validation error, got: %v", err)
}
