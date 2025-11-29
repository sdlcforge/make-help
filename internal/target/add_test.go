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

func TestAddService_AddTarget_AppendToMakefile(t *testing.T) {
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

	// Verify Makefile was updated
	content, err := os.ReadFile(makefilePath)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, ".PHONY: help")
	assert.Contains(t, contentStr, "help:")
	assert.Contains(t, contentStr, "GOBIN ?= .bin")
	assert.Contains(t, contentStr, "MAKE_HELP_BIN := $(GOBIN)/make-help")
	assert.Contains(t, contentStr, "@$(MAKE_HELP_CMD)")
	assert.Contains(t, contentStr, "--keep-order-categories")
	assert.Contains(t, contentStr, "--default-category General")
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

	// Verify 01-help.mk was created
	helpFile := filepath.Join(makeDir, "01-help.mk")
	content, err := os.ReadFile(helpFile)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, ".PHONY: help")
	assert.Contains(t, contentStr, "help:")
	assert.Contains(t, contentStr, "GOBIN ?= .bin")
	assert.Contains(t, contentStr, "@$(MAKE_HELP_CMD)")

	// Verify Makefile was NOT modified (no include directive added)
	makefileContentAfter, err := os.ReadFile(makefilePath)
	require.NoError(t, err)
	assert.Equal(t, makefileContent, string(makefileContentAfter))
}

func TestAddService_AddTarget_ExplicitTargetFile(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	targetFile := filepath.Join(tmpDir, "custom-help.mk")

	// Create simple Makefile
	makefileContent := `all:
	@echo "Building..."
`
	err := os.WriteFile(makefilePath, []byte(makefileContent), 0644)
	require.NoError(t, err)

	// Create config with explicit target file
	config := &Config{
		MakefilePath: makefilePath,
		TargetFile:   targetFile,
	}

	// Create mock executor
	executor := NewMockExecutor()
	executor.outputs["make -n -f "+makefilePath] = ""

	// Create service
	service := NewAddService(config, executor, false)

	// Execute
	err = service.AddTarget()
	require.NoError(t, err)

	// Verify target file was created
	content, err := os.ReadFile(targetFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), ".PHONY: help")

	// Verify include directive was added to Makefile
	makefileContentAfter, err := os.ReadFile(makefilePath)
	require.NoError(t, err)
	assert.Contains(t, string(makefileContentAfter), "include custom-help.mk")
}

func TestAddService_AddTarget_FlagPassThrough(t *testing.T) {
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
	assert.Contains(t, err.Error(), "Makefile validation failed")
}

func TestAddService_VerboseOutput(t *testing.T) {
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	targetFile := filepath.Join(tmpDir, "custom.mk")

	err := os.WriteFile(makefilePath, []byte("all:\n\t@echo test\n"), 0644)
	require.NoError(t, err)

	config := &Config{
		MakefilePath: makefilePath,
		TargetFile:   targetFile,
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
		name          string
		makefileContent string
		targetFile    string
		wantFile      string
		wantInclude   bool
	}{
		{
			name:            "explicit target file",
			makefileContent: "all:\n\t@echo test\n",
			targetFile:      "custom.mk",
			wantFile:        "custom.mk",
			wantInclude:     true,
		},
		{
			name:            "include make/*.mk pattern",
			makefileContent: "include make/*.mk\n\nall:\n\t@echo test\n",
			targetFile:      "",
			wantFile:        "make/01-help.mk",
			wantInclude:     false,
		},
		{
			name:            "-include make/*.mk pattern (optional include)",
			makefileContent: "-include make/*.mk\n\nall:\n\t@echo test\n",
			targetFile:      "",
			wantFile:        "make/01-help.mk",
			wantInclude:     false,
		},
		{
			name:            "no pattern - append to makefile",
			makefileContent: "all:\n\t@echo test\n",
			targetFile:      "",
			wantFile:        "Makefile",
			wantInclude:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			makefilePath := filepath.Join(tmpDir, "Makefile")

			err := os.WriteFile(makefilePath, []byte(tt.makefileContent), 0644)
			require.NoError(t, err)

			config := &Config{
				MakefilePath: makefilePath,
				TargetFile:   tt.targetFile,
			}

			service := &AddService{
				config: config,
			}

			gotFile, gotInclude, err := service.determineTargetFile(makefilePath)
			require.NoError(t, err)

			// Normalize paths for comparison
			if tt.wantFile == "Makefile" {
				assert.Equal(t, makefilePath, gotFile)
			} else if strings.HasPrefix(tt.wantFile, "make/") {
				expected := filepath.Join(tmpDir, tt.wantFile)
				assert.Equal(t, expected, gotFile)
			} else {
				assert.Equal(t, tt.wantFile, gotFile)
			}

			assert.Equal(t, tt.wantInclude, gotInclude)
		})
	}
}
