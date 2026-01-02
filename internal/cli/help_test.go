package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunDetailedHelp_DocumentedTarget(t *testing.T) {
	// Use the with_undocumented.mk fixture which has the "build" target with full docs
	fixturePath := filepath.Join("..", "..", "test", "fixtures", "makefiles", "with_undocumented.mk")
	absPath, err := filepath.Abs(fixturePath)
	require.NoError(t, err)

	// Verify fixture exists
	_, err = os.Stat(absPath)
	require.NoError(t, err, "Fixture file should exist")

	config := &Config{
		MakefilePath: absPath,
		Target:       "build",
		UseColor:     false,
		Format:       "text",
	}

	err = runDetailedHelp(config)
	require.NoError(t, err)
	// We can't easily capture stdout in this test, but we can verify it doesn't error
}

func TestRunDetailedHelp_UndocumentedTarget(t *testing.T) {
	// Use the with_undocumented.mk fixture which has an "undocumented" target
	fixturePath := filepath.Join("..", "..", "test", "fixtures", "makefiles", "with_undocumented.mk")
	absPath, err := filepath.Abs(fixturePath)
	require.NoError(t, err)

	config := &Config{
		MakefilePath: absPath,
		Target:       "undocumented",
		UseColor:     false,
		Format:       "text",
	}

	err = runDetailedHelp(config)
	require.NoError(t, err)
	// Should not error even for undocumented targets that exist
}

func TestRunDetailedHelp_NonexistentTarget(t *testing.T) {
	// Use any valid Makefile
	fixturePath := filepath.Join("..", "..", "test", "fixtures", "makefiles", "basic.mk")
	absPath, err := filepath.Abs(fixturePath)
	require.NoError(t, err)

	config := &Config{
		MakefilePath: absPath,
		Target:       "nonexistent_target_xyz",
		UseColor:     false,
	}

	err = runDetailedHelp(config)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "target 'nonexistent_target_xyz' not found")
}

func TestRunDetailedHelp_InvalidMakefile(t *testing.T) {
	config := &Config{
		MakefilePath: "/nonexistent/path/to/Makefile",
		Target:       "build",
		UseColor:     false,
	}

	err := runDetailedHelp(config)
	require.Error(t, err)
	// Should fail to resolve or validate the Makefile path
}

func TestRunDetailedHelp_WithColor(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "test", "fixtures", "makefiles", "with_undocumented.mk")
	absPath, err := filepath.Abs(fixturePath)
	require.NoError(t, err)

	config := &Config{
		MakefilePath: absPath,
		Target:       "build",
		UseColor:     true, // Enable colors
		Format:       "text",
	}

	err = runDetailedHelp(config)
	require.NoError(t, err)
	// Should work with colors enabled
}
