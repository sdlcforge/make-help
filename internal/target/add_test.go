package target

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockExecutor implements CommandExecutor for testing
type MockExecutor struct {
	outputs map[string]string
	errors  map[string]error
}

func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		outputs: make(map[string]string),
		errors:  make(map[string]error),
	}
}

func (m *MockExecutor) Execute(cmd string, args ...string) (string, string, error) {
	return m.ExecuteContext(context.Background(), cmd, args...)
}

func (m *MockExecutor) ExecuteContext(ctx context.Context, cmd string, args ...string) (string, string, error) {
	key := cmd + " " + strings.Join(args, " ")
	if err, ok := m.errors[key]; ok {
		return "", "error output", err
	}
	if out, ok := m.outputs[key]; ok {
		return out, "", nil
	}
	return "", "", nil
}

func TestAddService_AddTarget_CreateHelpMk(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	// Create simple Makefile
	makefileContent := `all:
	@echo "Building..."

test:
	@go test ./...
`
	err := os.WriteFile(makefilePath, []byte(makefileContent), 0644)
	require.NoError(t, err)

	// Create config
	config := &Config{
		MakefilePath:        makefilePath,
		KeepOrderCategories: true,
		DefaultCategory:     "General",
	}

	// Create mock executor
	executor := NewMockExecutor()
	executor.outputs["make -n -f "+makefilePath] = ""

	// Create service
	service := NewAddService(config, executor, false)

	// Execute
	err = service.AddTarget()
	require.NoError(t, err)

	// Verify make/help.mk was created
	helpMkPath := filepath.Join(tmpDir, "make", "help.mk")
	content, err := os.ReadFile(helpMkPath)
	require.NoError(t, err)

	contentStr := string(content)
	// NOTE: AddService uses the deprecated generateHelpTarget placeholder.
	// The full static generator is used via CLI orchestration (runCreateHelpTarget).
	assert.Contains(t, contentStr, ".PHONY: help")
	assert.Contains(t, contentStr, "help:")
	assert.Contains(t, contentStr, "MAKE_HELP_DIR := $(dir $(lastword $(MAKEFILE_LIST)))")
	assert.Contains(t, contentStr, "This is a placeholder")

	// Verify include directive was added to Makefile (pattern include)
	makefileContentAfter, err := os.ReadFile(makefilePath)
	require.NoError(t, err)
	assert.Contains(t, string(makefileContentAfter), "-include make/*.mk")
}

func TestAddService_AddTarget_CreateMakeDirectory(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	// Create Makefile with include pattern
	makefileContent := `include make/*.mk

all:
	@echo "Building..."
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
	service := NewAddService(config, executor, false)

	// Execute
	err = service.AddTarget()
	require.NoError(t, err)

	// Verify make directory was created
	makeDir := filepath.Join(tmpDir, "make")
	info, err := os.Stat(makeDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Verify help.mk was created (no numbered files exist, so no prefix)
	helpFile := filepath.Join(makeDir, "help.mk")
	content, err := os.ReadFile(helpFile)
	require.NoError(t, err)

	contentStr := string(content)
	// NOTE: AddService uses the deprecated generateHelpTarget placeholder.
	// The full static generator is used via CLI orchestration (runCreateHelpTarget).
	assert.Contains(t, contentStr, ".PHONY: help")
	assert.Contains(t, contentStr, "help:")
	assert.Contains(t, contentStr, "MAKE_HELP_DIR := $(dir $(lastword $(MAKEFILE_LIST)))")
	assert.Contains(t, contentStr, "This is a placeholder")

	// Verify Makefile was NOT modified (no include directive added)
	makefileContentAfter, err := os.ReadFile(makefilePath)
	require.NoError(t, err)
	assert.Equal(t, makefileContent, string(makefileContentAfter))
}

func TestAddService_AddTarget_ExplicitTargetFile(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	targetFileRelPath := "custom-help.mk" // Relative path

	// Create simple Makefile
	makefileContent := `all:
	@echo "Building..."
`
	err := os.WriteFile(makefilePath, []byte(makefileContent), 0644)
	require.NoError(t, err)

	// Create config with explicit relative target file path
	config := &Config{
		MakefilePath:      makefilePath,
		TargetFileRelPath: targetFileRelPath,
	}

	// Create mock executor
	executor := NewMockExecutor()
	executor.outputs["make -n -f "+makefilePath] = ""

	// Create service
	service := NewAddService(config, executor, false)

	// Execute
	err = service.AddTarget()
	require.NoError(t, err)

	// Verify target file was created (absolute path computed from relative)
	absTargetFile := filepath.Join(tmpDir, targetFileRelPath)
	content, err := os.ReadFile(absTargetFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), ".PHONY: help")

	// Verify include directive was added to Makefile with self-referential pattern
	makefileContentAfter, err := os.ReadFile(makefilePath)
	require.NoError(t, err)
	assert.Contains(t, string(makefileContentAfter), "-include $(dir $(lastword $(MAKEFILE_LIST)))custom-help.mk")
}

func TestAddService_AddTarget_FlagPassThrough(t *testing.T) {
	// NOTE: This test is skipped because AddService uses the deprecated generateHelpTarget
	// placeholder. Flag pass-through is tested via CLI integration tests in
	// create_help_target_test.go which exercises the full generator pipeline.
	t.Skip("Flag pass-through tested via CLI integration tests")

	// Setup
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	err := os.WriteFile(makefilePath, []byte("all:\n\t@echo test\n"), 0644)
	require.NoError(t, err)

	tests := []struct {
		name     string
		config   *Config
		expected []string
	}{
		{
			name: "keep order flags",
			config: &Config{
				MakefilePath:        makefilePath,
				KeepOrderCategories: true,
				KeepOrderTargets:    true,
			},
			expected: []string{"--keep-order-categories", "--keep-order-targets"},
		},
		{
			name: "category order",
			config: &Config{
				MakefilePath:  makefilePath,
				CategoryOrder: []string{"Build", "Test", "Deploy"},
			},
			expected: []string{"--category-order Build,Test,Deploy"},
		},
		{
			name: "default category",
			config: &Config{
				MakefilePath:    makefilePath,
				DefaultCategory: "General",
			},
			expected: []string{"--default-category General"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock executor
			executor := NewMockExecutor()
			executor.outputs["make -n -f "+makefilePath] = ""

			// Create service
			service := NewAddService(tt.config, executor, false)

			// Execute
			err := service.AddTarget()
			require.NoError(t, err)

			// Read generated content
			content, err := os.ReadFile(makefilePath)
			require.NoError(t, err)

			// Verify all expected flags are present
			for _, flag := range tt.expected {
				assert.Contains(t, string(content), flag)
			}
		})
	}
}

func TestAddService_ValidateMakefile_SyntaxError(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	// Create Makefile with syntax error
	err := os.WriteFile(makefilePath, []byte("invalid syntax here\n"), 0644)
	require.NoError(t, err)

	// Create config
	config := &Config{
		MakefilePath: makefilePath,
	}

	// Create mock executor that returns error
	executor := NewMockExecutor()
	executor.errors["make -n -f "+makefilePath] = assert.AnError

	// Create service
	service := NewAddService(config, executor, false)

	// Execute
	err = service.AddTarget()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "makefile validation failed")
}

func TestAddService_VerboseOutput(t *testing.T) {
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	targetFileRelPath := "custom.mk" // Relative path

	err := os.WriteFile(makefilePath, []byte("all:\n\t@echo test\n"), 0644)
	require.NoError(t, err)

	config := &Config{
		MakefilePath:      makefilePath,
		TargetFileRelPath: targetFileRelPath,
	}

	executor := NewMockExecutor()
	executor.outputs["make -n -f "+makefilePath] = ""

	// Create service with verbose=true
	service := NewAddService(config, executor, true)

	// Execute (should print verbose output to stdout)
	err = service.AddTarget()
	require.NoError(t, err)
}

func TestAddService_DetermineTargetFileReadError(t *testing.T) {
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	// Don't create the file - reading it should fail
	config := &Config{
		MakefilePath: makefilePath,
	}

	executor := NewMockExecutor()

	service := NewAddService(config, executor, false)

	// Execute should fail when trying to read non-existent Makefile
	err := service.AddTarget()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read Makefile")
}

func TestDetermineTargetFile(t *testing.T) {
	tests := []struct {
		name              string
		makefileContent   string
		targetFileRelPath string
		wantFile          string // relative to tmpDir or "Makefile" for append
		wantInclude       bool
	}{
		{
			name:              "explicit relative target file",
			makefileContent:   "all:\n\t@echo test\n",
			targetFileRelPath: "custom.mk",
			wantFile:          "custom.mk",
			wantInclude:       true,
		},
		{
			name:              "explicit relative target file in subdir",
			makefileContent:   "all:\n\t@echo test\n",
			targetFileRelPath: "make/help.mk",
			wantFile:          "make/help.mk",
			wantInclude:       true,
		},
		{
			name:              "include make/*.mk pattern",
			makefileContent:   "include make/*.mk\n\nall:\n\t@echo test\n",
			targetFileRelPath: "",
			wantFile:          "make/help.mk",
			wantInclude:       false,
		},
		{
			name:              "-include make/*.mk pattern (optional include)",
			makefileContent:   "-include make/*.mk\n\nall:\n\t@echo test\n",
			targetFileRelPath: "",
			wantFile:          "make/help.mk",
			wantInclude:       false,
		},
		{
			name:              "no pattern - create make/help.mk",
			makefileContent:   "all:\n\t@echo test\n",
			targetFileRelPath: "",
			wantFile:          "make/help.mk",
			wantInclude:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			makefilePath := filepath.Join(tmpDir, "Makefile")

			err := os.WriteFile(makefilePath, []byte(tt.makefileContent), 0644)
			require.NoError(t, err)

			config := &Config{
				MakefilePath:      makefilePath,
				TargetFileRelPath: tt.targetFileRelPath,
			}

			service := &AddService{
				config: config,
			}

			gotFile, gotInclude, err := service.determineTargetFile(makefilePath)
			require.NoError(t, err)

			// Normalize paths for comparison - all returned paths are absolute
			expectedFile := filepath.Join(tmpDir, tt.wantFile)
			assert.Equal(t, expectedFile, gotFile)
			assert.Equal(t, tt.wantInclude, gotInclude)
		})
	}
}

func TestDetermineTargetFile_NumberedFiles(t *testing.T) {
	tests := []struct {
		name            string
		makefileContent string
		setupFiles      []string // Files to create in make/ directory
		wantFile        string   // Expected filename (relative to make/)
		wantInclude     bool
	}{
		{
			name:            "numbered file exists - 10-constants.mk",
			makefileContent: "include make/*.mk\n\nall:\n\t@echo test\n",
			setupFiles:      []string{"10-constants.mk"},
			wantFile:        "make/00-help.mk", // 2 digits, so 00- prefix
			wantInclude:     false,
		},
		{
			name:            "numbered file exists - 1-setup.mk",
			makefileContent: "include make/*.mk\n\nall:\n\t@echo test\n",
			setupFiles:      []string{"1-setup.mk"},
			wantFile:        "make/0-help.mk", // 1 digit, so 0- prefix
			wantInclude:     false,
		},
		{
			name:            "numbered file exists - 100-utils.mk",
			makefileContent: "include make/*.mk\n\nall:\n\t@echo test\n",
			setupFiles:      []string{"100-utils.mk"},
			wantFile:        "make/000-help.mk", // 3 digits, so 000- prefix
			wantInclude:     false,
		},
		{
			name:            "multiple numbered files - use max digits",
			makefileContent: "include make/*.mk\n\nall:\n\t@echo test\n",
			setupFiles:      []string{"1-setup.mk", "10-constants.mk", "100-utils.mk"},
			wantFile:        "make/000-help.mk", // Max is 3 digits
			wantInclude:     false,
		},
		{
			name:            "non-numbered files don't affect prefix",
			makefileContent: "include make/*.mk\n\nall:\n\t@echo test\n",
			setupFiles:      []string{"constants.mk", "utils.mk"},
			wantFile:        "make/help.mk", // No numbered files, no prefix
			wantInclude:     false,
		},
		{
			name:            "numbered files with different suffix ignored",
			makefileContent: "include make/*.mk\n\nall:\n\t@echo test\n",
			setupFiles:      []string{"10-constants.txt"}, // Different suffix
			wantFile:        "make/help.mk",               // Doesn't match .mk pattern
			wantInclude:     false,
		},
		{
			name:            "no pattern - numbered files still create prefix",
			makefileContent: "all:\n\t@echo test\n",
			setupFiles:      []string{"10-constants.mk"},
			wantFile:        "make/00-help.mk",
			wantInclude:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			makefilePath := filepath.Join(tmpDir, "Makefile")
			makeDir := filepath.Join(tmpDir, "make")

			// Create Makefile
			err := os.WriteFile(makefilePath, []byte(tt.makefileContent), 0644)
			require.NoError(t, err)

			// Create make directory and setup files
			err = os.MkdirAll(makeDir, 0755)
			require.NoError(t, err)
			for _, filename := range tt.setupFiles {
				filePath := filepath.Join(makeDir, filename)
				err = os.WriteFile(filePath, []byte("# test file\n"), 0644)
				require.NoError(t, err)
			}

			config := &Config{
				MakefilePath:      makefilePath,
				TargetFileRelPath: "",
			}

			service := &AddService{
				config: config,
			}

			gotFile, gotInclude, err := service.determineTargetFile(makefilePath)
			require.NoError(t, err)

			// Normalize paths for comparison - all returned paths are absolute
			expectedFile := filepath.Join(tmpDir, tt.wantFile)
			assert.Equal(t, expectedFile, gotFile, "File path mismatch")
			assert.Equal(t, tt.wantInclude, gotInclude, "Include directive requirement mismatch")
		})
	}
}
