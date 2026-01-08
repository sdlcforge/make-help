package target

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveService_RemoveTarget_InlineTarget(t *testing.T) {
	t.Parallel()
	// Setup
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	// Create Makefile with inline help target
	makefileContent := `all:
	@echo "Building..."

.PHONY: help
help:
	@make-help --keep-order-all

test:
	@go test ./...
`
	err := os.WriteFile(makefilePath, []byte(makefileContent), 0644)
	require.NoError(t, err)

	// Create config
	config := &Config{
		MakefilePath: makefilePath,
	}

	// Create mock executor
	executor := NewMockExecutor()
	executor.outputs["make -n -f "+makefilePath] = ""

	// Create service
	service := NewRemoveService(config, executor, false)

	// Execute
	err = service.RemoveTarget()
	require.NoError(t, err)

	// Verify help target was removed
	content, err := os.ReadFile(makefilePath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.NotContains(t, contentStr, ".PHONY: help")
	assert.NotContains(t, contentStr, "help:")
	assert.NotContains(t, contentStr, "@make-help")

	// Verify other targets remain
	assert.Contains(t, contentStr, "all:")
	assert.Contains(t, contentStr, "test:")
}

func TestRemoveService_RemoveTarget_IncludeDirective(t *testing.T) {
	t.Parallel()
	// Setup
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	// Create Makefile with include directive
	makefileContent := `include make/01-help.mk

all:
	@echo "Building..."

test:
	@go test ./...
`
	err := os.WriteFile(makefilePath, []byte(makefileContent), 0644)
	require.NoError(t, err)

	// Create make directory and help file
	makeDir := filepath.Join(tmpDir, "make")
	err = os.MkdirAll(makeDir, 0755)
	require.NoError(t, err)

	helpFile := filepath.Join(makeDir, "01-help.mk")
	helpContent := `.PHONY: help
help:
	@make-help
`
	err = os.WriteFile(helpFile, []byte(helpContent), 0644)
	require.NoError(t, err)

	// Create config
	config := &Config{
		MakefilePath: makefilePath,
	}

	// Create mock executor
	executor := NewMockExecutor()
	executor.outputs["make -n -f "+makefilePath] = ""

	// Create service
	service := NewRemoveService(config, executor, false)

	// Execute
	err = service.RemoveTarget()
	require.NoError(t, err)

	// Verify include directive was removed
	content, err := os.ReadFile(makefilePath)
	require.NoError(t, err)
	assert.NotContains(t, string(content), "include make/01-help.mk")

	// Verify help file was deleted
	_, err = os.Stat(helpFile)
	assert.True(t, os.IsNotExist(err))

	// Verify other content remains
	assert.Contains(t, string(content), "all:")
	assert.Contains(t, string(content), "test:")
}

func TestRemoveService_RemoveTarget_BothInlineAndInclude(t *testing.T) {
	t.Parallel()
	// Setup
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	// Create Makefile with both inline and include
	makefileContent := `include custom-help.mk

all:
	@echo "Building..."

.PHONY: help
help:
	@echo "Help target"

test:
	@go test ./...
`
	err := os.WriteFile(makefilePath, []byte(makefileContent), 0644)
	require.NoError(t, err)

	// Create config
	config := &Config{
		MakefilePath: makefilePath,
	}

	// Create mock executor
	executor := NewMockExecutor()
	executor.outputs["make -n -f "+makefilePath] = ""

	// Create service
	service := NewRemoveService(config, executor, false)

	// Execute
	err = service.RemoveTarget()
	require.NoError(t, err)

	// Verify both were removed
	content, err := os.ReadFile(makefilePath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.NotContains(t, contentStr, "include custom-help.mk")
	assert.NotContains(t, contentStr, ".PHONY: help")
	assert.NotContains(t, contentStr, "help:")

	// Verify other targets remain
	assert.Contains(t, contentStr, "all:")
	assert.Contains(t, contentStr, "test:")
}

func TestRemoveService_RemoveTarget_NoHelpTarget(t *testing.T) {
	t.Parallel()
	// Setup
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	// Create Makefile without help target
	makefileContent := `all:
	@echo "Building..."

test:
	@go test ./...
`
	err := os.WriteFile(makefilePath, []byte(makefileContent), 0644)
	require.NoError(t, err)

	// Create config
	config := &Config{
		MakefilePath: makefilePath,
	}

	// Create mock executor
	executor := NewMockExecutor()
	executor.outputs["make -n -f "+makefilePath] = ""

	// Create service
	service := NewRemoveService(config, executor, false)

	// Execute (should not error)
	err = service.RemoveTarget()
	require.NoError(t, err)

	// Verify Makefile unchanged
	content, err := os.ReadFile(makefilePath)
	require.NoError(t, err)
	assert.Equal(t, makefileContent, string(content))
}

func TestRemoveService_RemoveIncludeDirectives(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		input          string
		expectedOutput string
		shouldChange   bool
	}{
		{
			name: "single include directive",
			input: `all:
	@echo test

include make/01-help.mk

test:
	@echo test
`,
			expectedOutput: `all:
	@echo test


test:
	@echo test
`,
			shouldChange: true,
		},
		{
			name: "multiple include directives",
			input: `include make/01-help.mk
include custom-help.mk

all:
	@echo test
`,
			expectedOutput: `
all:
	@echo test
`,
			shouldChange: true,
		},
		{
			name: "self-referential include directive",
			input: `all:
	@echo test

include $(dir $(lastword $(MAKEFILE_LIST)))help.mk

test:
	@echo test
`,
			expectedOutput: `all:
	@echo test


test:
	@echo test
`,
			shouldChange: true,
		},
		{
			name: "self-referential optional include directive",
			input: `all:
	@echo test

-include $(dir $(lastword $(MAKEFILE_LIST)))help.mk

test:
	@echo test
`,
			expectedOutput: `all:
	@echo test


test:
	@echo test
`,
			shouldChange: true,
		},
		{
			name: "no help includes",
			input: `include common.mk

all:
	@echo test
`,
			expectedOutput: `include common.mk

all:
	@echo test
`,
			shouldChange: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			makefilePath := filepath.Join(tmpDir, "Makefile")

			err := os.WriteFile(makefilePath, []byte(tt.input), 0644)
			require.NoError(t, err)

			config := &Config{MakefilePath: makefilePath}
			executor := NewMockExecutor()
			service := NewRemoveService(config, executor, false)

			err = service.removeIncludeDirectives(makefilePath)
			require.NoError(t, err)

			content, err := os.ReadFile(makefilePath)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedOutput, string(content))
		})
	}
}

func TestRemoveService_RemoveInlineHelpTarget(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		input        string
		expected     string
		shouldRemove bool
	}{
		{
			name: "help target with phony",
			input: `.PHONY: help
help:
	@make-help

all:
	@echo test
`,
			expected: `
all:
	@echo test
`,
			shouldRemove: true,
		},
		{
			name: "help target without phony",
			input: `all:
	@echo test

help:
	@echo "Help"
	@echo "More help"

test:
	@echo test
`,
			expected: `all:
	@echo test


test:
	@echo test
`,
			shouldRemove: true,
		},
		{
			name: "no help target",
			input: `all:
	@echo test

test:
	@echo test
`,
			expected: `all:
	@echo test

test:
	@echo test
`,
			shouldRemove: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			makefilePath := filepath.Join(tmpDir, "Makefile")

			err := os.WriteFile(makefilePath, []byte(tt.input), 0644)
			require.NoError(t, err)

			config := &Config{MakefilePath: makefilePath}
			executor := NewMockExecutor()
			service := NewRemoveService(config, executor, false)

			removed, err := service.removeInlineHelpTarget(makefilePath)
			require.NoError(t, err)
			assert.Equal(t, tt.shouldRemove, removed)

			content, err := os.ReadFile(makefilePath)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(content))
		})
	}
}

func TestRemoveService_RemoveHelpTargetFiles(t *testing.T) {
	t.Parallel()
	// Setup with help file
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	makeDir := filepath.Join(tmpDir, "make")
	helpFile := filepath.Join(makeDir, "01-help.mk")

	err := os.MkdirAll(makeDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(helpFile, []byte("help:\n\t@make-help\n"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(makefilePath, []byte("all:\n\t@echo test\n"), 0644)
	require.NoError(t, err)

	config := &Config{MakefilePath: makefilePath}
	executor := NewMockExecutor()
	service := NewRemoveService(config, executor, false)

	// Execute
	removed, err := service.removeHelpTargetFiles(makefilePath)
	require.NoError(t, err)
	assert.True(t, removed)

	// Verify file was deleted
	_, err = os.Stat(helpFile)
	assert.True(t, os.IsNotExist(err))
}

func TestRemoveService_RemoveHelpTargetFiles_NoFile(t *testing.T) {
	t.Parallel()
	// Setup without help file
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	err := os.WriteFile(makefilePath, []byte("all:\n\t@echo test\n"), 0644)
	require.NoError(t, err)

	config := &Config{MakefilePath: makefilePath}
	executor := NewMockExecutor()
	service := NewRemoveService(config, executor, false)

	// Execute (should not error)
	removed, err := service.removeHelpTargetFiles(makefilePath)
	require.NoError(t, err)
	assert.False(t, removed)
}

func TestRemoveService_ValidateMakefile_SyntaxError(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	err := os.WriteFile(makefilePath, []byte("invalid syntax\n"), 0644)
	require.NoError(t, err)

	config := &Config{MakefilePath: makefilePath}

	// Create mock executor that returns error
	executor := NewMockExecutor()
	executor.errors["make -n -f "+makefilePath] = assert.AnError

	service := NewRemoveService(config, executor, false)

	err = service.RemoveTarget()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "makefile validation failed")
}
