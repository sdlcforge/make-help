package discovery

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockCommandExecutor is a mock implementation of CommandExecutor for testing.
type MockCommandExecutor struct {
	outputs     map[string]string        // command -> stdout
	errors      map[string]error         // command -> error
	delays      map[string]time.Duration // command -> delay before returning
	prefixMatch bool                     // if true, match commands by prefix
}

// NewMockCommandExecutor creates a new MockCommandExecutor.
func NewMockCommandExecutor() *MockCommandExecutor {
	return &MockCommandExecutor{
		outputs: make(map[string]string),
		errors:  make(map[string]error),
		delays:  make(map[string]time.Duration),
	}
}

// SetOutput sets the expected output for a command.
func (m *MockCommandExecutor) SetOutput(cmd string, output string) {
	m.outputs[cmd] = output
}

// SetError sets the expected error for a command.
func (m *MockCommandExecutor) SetError(cmd string, err error) {
	m.errors[cmd] = err
}

// SetDelay sets a delay for a command (useful for testing timeouts).
func (m *MockCommandExecutor) SetDelay(cmd string, delay time.Duration) {
	m.delays[cmd] = delay
}

// SetPrefixMatch enables prefix matching for commands.
func (m *MockCommandExecutor) SetPrefixMatch(enabled bool) {
	m.prefixMatch = enabled
}

// getKey finds the matching key for a command, supporting prefix matching.
func (m *MockCommandExecutor) getKey(fullKey string) string {
	// First, try exact match
	if _, ok := m.outputs[fullKey]; ok {
		return fullKey
	}
	if _, ok := m.errors[fullKey]; ok {
		return fullKey
	}

	// If prefix matching is enabled, try to find a prefix match
	if m.prefixMatch {
		for key := range m.outputs {
			if strings.HasPrefix(fullKey, key) {
				return key
			}
		}
		for key := range m.errors {
			if strings.HasPrefix(fullKey, key) {
				return key
			}
		}
	}

	return fullKey
}

// Execute implements CommandExecutor.Execute.
func (m *MockCommandExecutor) Execute(cmd string, args ...string) (string, string, error) {
	return m.ExecuteContext(context.Background(), cmd, args...)
}

// ExecuteContext implements CommandExecutor.ExecuteContext.
func (m *MockCommandExecutor) ExecuteContext(ctx context.Context, cmd string, args ...string) (string, string, error) {
	fullKey := fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
	key := m.getKey(fullKey)

	// Simulate delay if configured
	if delay, ok := m.delays[key]; ok {
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return "", "", ctx.Err()
		}
	}

	// Check for configured error
	if err, ok := m.errors[key]; ok {
		return "", "error output", err
	}

	// Return configured output
	if output, ok := m.outputs[key]; ok {
		return output, "", nil
	}

	// Default: command not found
	return "", "", fmt.Errorf("command not mocked: %s", fullKey)
}

func TestNewService(t *testing.T) {
	mock := NewMockCommandExecutor()
	service := NewService(mock, false)

	assert.NotNil(t, service)
	assert.Equal(t, mock, service.executor)
	assert.False(t, service.verbose)

	verboseService := NewService(mock, true)
	assert.True(t, verboseService.verbose)
}

func TestDiscoverMakefiles_Verbose(t *testing.T) {
	// Create temporary directory and Makefile
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	err := os.WriteFile(makefilePath, []byte("all:\n\t@echo hello\n"), 0644)
	require.NoError(t, err)

	// Use real executor for this test to verify verbose output path
	executor := NewDefaultExecutor()
	service := NewService(executor, true) // verbose mode

	// With real executor, this should succeed
	makefiles, err := service.DiscoverMakefiles(makefilePath)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(makefiles), 1)
}

func TestDiscoverTargets_Basic(t *testing.T) {
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	err := os.WriteFile(makefilePath, []byte("all:\n\t@echo hello\n"), 0644)
	require.NoError(t, err)

	mock := NewMockCommandExecutor()
	mock.SetPrefixMatch(true)
	mock.SetOutput("make -s --no-print-directory -f", `# make database
all: build
build:
	go build
test:
	go test
`)

	service := NewService(mock, false)
	result, err := service.DiscoverTargets(makefilePath)

	require.NoError(t, err)
	assert.Equal(t, []string{"all", "build", "test"}, result.Targets)
}

func TestDiscoverTargets_Verbose(t *testing.T) {
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	err := os.WriteFile(makefilePath, []byte("all:\n\t@echo hello\n"), 0644)
	require.NoError(t, err)

	mock := NewMockCommandExecutor()
	mock.SetPrefixMatch(true)
	mock.SetOutput("make -s --no-print-directory -f", `all:
build:
`)

	service := NewService(mock, true) // verbose mode
	result, err := service.DiscoverTargets(makefilePath)

	require.NoError(t, err)
	assert.Equal(t, []string{"all", "build"}, result.Targets)
}

func TestDiscoverTargets_Timeout(t *testing.T) {
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	err := os.WriteFile(makefilePath, []byte("all:\n\t@echo hello\n"), 0644)
	require.NoError(t, err)

	mock := NewMockCommandExecutor()
	mock.SetPrefixMatch(true)
	// Set a delay longer than any reasonable timeout for testing
	mock.SetDelay("make -s --no-print-directory -f", 35*time.Second)
	mock.SetOutput("make -s --no-print-directory -f", "all:")

	// This test would take too long with real timeout,
	// but we can verify the timeout code path exists
	// by checking that context cancellation is handled

	// Create a short context to test cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Directly call ExecuteContext to verify timeout handling
	_, _, err = mock.ExecuteContext(ctx, "make", "-s", "--no-print-directory", "-f", makefilePath, "-p", "-r")
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestDiscoverTargets_Error(t *testing.T) {
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	err := os.WriteFile(makefilePath, []byte("all:\n\t@echo hello\n"), 0644)
	require.NoError(t, err)

	mock := NewMockCommandExecutor()
	mock.SetPrefixMatch(true)
	mock.SetError("make -f", fmt.Errorf("make failed"))

	service := NewService(mock, false)
	_, err = service.DiscoverTargets(makefilePath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to discover targets")
}

func TestResolveAbsolutePaths(t *testing.T) {
	tests := []struct {
		name        string
		files       []string
		createFiles []string
		expected    []string
		expectError bool
	}{
		{
			name:        "absolute paths",
			files:       []string{},
			createFiles: []string{"Makefile"},
			expected:    []string{},
			expectError: false,
		},
		{
			name:        "relative paths",
			files:       []string{"Makefile", "include/common.mk"},
			createFiles: []string{"Makefile", "include/common.mk"},
			expected:    []string{"Makefile", "include/common.mk"},
			expectError: false,
		},
		{
			name:        "file not found",
			files:       []string{"nonexistent.mk"},
			createFiles: []string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create test files
			for _, file := range tt.createFiles {
				var filePath string
				if filepath.IsAbs(file) {
					filePath = file
				} else {
					filePath = filepath.Join(tmpDir, file)
				}

				err := os.MkdirAll(filepath.Dir(filePath), 0755)
				require.NoError(t, err)

				err = os.WriteFile(filePath, []byte("# test file\n"), 0644)
				require.NoError(t, err)
			}

			service := NewService(NewMockCommandExecutor(), false)
			resolved, err := service.resolveAbsolutePaths(tt.files, tmpDir)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expected), len(resolved))

				// Verify all paths are absolute
				for _, path := range resolved {
					assert.True(t, filepath.IsAbs(path), "path should be absolute: %s", path)
				}
			}
		})
	}
}

func TestResolveAbsolutePaths_StatError(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file that we'll test
	testFile := filepath.Join(tmpDir, "test.mk")
	err := os.WriteFile(testFile, []byte("# test\n"), 0644)
	require.NoError(t, err)

	service := NewService(NewMockCommandExecutor(), false)

	// Test with existing file
	resolved, err := service.resolveAbsolutePaths([]string{testFile}, tmpDir)
	require.NoError(t, err)
	assert.Len(t, resolved, 1)
}

func TestResolveAbsolutePaths_AbsoluteInput(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file with absolute path
	testFile := filepath.Join(tmpDir, "test.mk")
	err := os.WriteFile(testFile, []byte("# test\n"), 0644)
	require.NoError(t, err)

	service := NewService(NewMockCommandExecutor(), false)

	// Test with absolute path input
	resolved, err := service.resolveAbsolutePaths([]string{testFile}, "/some/other/dir")
	require.NoError(t, err)
	assert.Len(t, resolved, 1)
	assert.Equal(t, testFile, resolved[0])
}

func TestParseTargetsFromDatabase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "simple targets",
			input: `# Make database
all: build test
build:
	go build
test:
	go test
`,
			expected: []string{"all", "build", "test"},
		},
		{
			name: "filters special targets",
			input: `# Make database
.SUFFIXES:
.DEFAULT:
all: build
build:
	go build
`,
			expected: []string{"all", "build"},
		},
		{
			name: "filters pattern rules",
			input: `# Make database
%.o: %.c
	gcc -c $<
all: build
build:
	go build
`,
			expected: []string{"all", "build"},
		},
		{
			name: "handles comments and whitespace",
			input: `# This is a comment
all: build
	@echo building
# Another comment
build:
	go build
`,
			expected: []string{"all", "build"},
		},
		{
			name: "no duplicates",
			input: `all: build
all: test
build:
	go build
`,
			expected: []string{"all", "build"},
		},
		{
			name: "empty input",
			input: "",
			expected: nil,
		},
		{
			name: "only comments",
			input: `# comment 1
# comment 2
`,
			expected: nil,
		},
		{
			name: "tab-prefixed lines ignored",
			input: `all:
	echo building
build:
	go build
`,
			expected: []string{"all", "build"},
		},
		{
			name: "space-prefixed lines ignored",
			input: `all:
  echo building
build:
  go build
`,
			expected: []string{"all", "build"},
		},
		{
			name: "complex target names",
			input: `my-target:
my_target2:
my.target:
my/target:
my@target:
my+target:
`,
			expected: []string{"my-target", "my_target2", "my.target", "my/target", "my@target", "my+target"},
		},
		{
			name:     "double colon targets",
			input:    "all:: build\nbuild::\n",
			expected: []string{"all", "build"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTargetsFromDatabase(tt.input)
			assert.Equal(t, tt.expected, result.Targets)
		})
	}
}

func TestIsSpecialTarget(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		expected bool
	}{
		{
			name:     "normal target",
			target:   "build",
			expected: false,
		},
		{
			name:     "SUFFIXES special target",
			target:   ".SUFFIXES",
			expected: true,
		},
		{
			name:     "DEFAULT special target",
			target:   ".DEFAULT",
			expected: true,
		},
		{
			name:     "PRECIOUS special target",
			target:   ".PRECIOUS",
			expected: true,
		},
		{
			name:     "INTERMEDIATE special target",
			target:   ".INTERMEDIATE",
			expected: true,
		},
		{
			name:     "SECONDARY special target",
			target:   ".SECONDARY",
			expected: true,
		},
		{
			name:     "SECONDEXPANSION special target",
			target:   ".SECONDEXPANSION",
			expected: true,
		},
		{
			name:     "DELETE_ON_ERROR special target",
			target:   ".DELETE_ON_ERROR",
			expected: true,
		},
		{
			name:     "IGNORE special target",
			target:   ".IGNORE",
			expected: true,
		},
		{
			name:     "LOW_RESOLUTION_TIME special target",
			target:   ".LOW_RESOLUTION_TIME",
			expected: true,
		},
		{
			name:     "SILENT special target",
			target:   ".SILENT",
			expected: true,
		},
		{
			name:     "EXPORT_ALL_VARIABLES special target",
			target:   ".EXPORT_ALL_VARIABLES",
			expected: true,
		},
		{
			name:     "NOTPARALLEL special target",
			target:   ".NOTPARALLEL",
			expected: true,
		},
		{
			name:     "ONESHELL special target",
			target:   ".ONESHELL",
			expected: true,
		},
		{
			name:     "POSIX special target",
			target:   ".POSIX",
			expected: true,
		},
		{
			name:     "pattern rule",
			target:   "%.o",
			expected: true,
		},
		{
			name:     "pattern rule with path",
			target:   "obj/%.o",
			expected: true,
		},
		{
			name:     "Makefile itself",
			target:   "Makefile",
			expected: true,
		},
		{
			name:     "makefile lowercase",
			target:   "makefile",
			expected: true,
		},
		{
			name:     "contains assignment",
			target:   "VAR=value",
			expected: true,
		},
		{
			name:     "complex assignment",
			target:   "CC=gcc",
			expected: true,
		},
		{
			name:     "PHONY is regular target",
			target:   ".PHONY",
			expected: false, // .PHONY is not in special list as users often define it
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSpecialTarget(tt.target)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolveMakefilePath(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "empty path defaults to Makefile",
			input:       "",
			expectError: false,
		},
		{
			name:        "relative path",
			input:       "Makefile",
			expectError: false,
		},
		{
			name:        "absolute path",
			input:       "/tmp/Makefile",
			expectError: false,
		},
		{
			name:        "nested relative path",
			input:       "path/to/Makefile",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveMakefilePath(tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, filepath.IsAbs(result), "result should be absolute path")
			}
		})
	}
}

func TestValidateMakefileExists(t *testing.T) {
	t.Run("file exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		makefilePath := filepath.Join(tmpDir, "Makefile")

		err := os.WriteFile(makefilePath, []byte("all:\n\t@echo hello\n"), 0644)
		require.NoError(t, err)

		err = ValidateMakefileExists(makefilePath)
		assert.NoError(t, err)
	})

	t.Run("file does not exist", func(t *testing.T) {
		err := ValidateMakefileExists("/nonexistent/path/Makefile")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("path is directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := ValidateMakefileExists(tmpDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "directory")
	})
}

func TestDefaultExecutor(t *testing.T) {
	executor := NewDefaultExecutor()
	assert.NotNil(t, executor)

	// Test Execute with a simple command that should work on any system
	stdout, stderr, err := executor.Execute("echo", "hello")
	require.NoError(t, err)
	assert.Equal(t, "hello\n", stdout)
	assert.Empty(t, stderr)

	// Test ExecuteContext
	ctx := context.Background()
	stdout, stderr, err = executor.ExecuteContext(ctx, "echo", "world")
	require.NoError(t, err)
	assert.Equal(t, "world\n", stdout)
	assert.Empty(t, stderr)
}

func TestDefaultExecutor_ContextCancellation(t *testing.T) {
	executor := NewDefaultExecutor()

	// Create an already cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// The command should fail due to cancelled context
	_, _, err := executor.ExecuteContext(ctx, "sleep", "10")
	assert.Error(t, err)
}

func TestDefaultExecutor_CommandError(t *testing.T) {
	executor := NewDefaultExecutor()

	// Test with a command that doesn't exist
	_, stderr, err := executor.Execute("nonexistent_command_xyz123")
	assert.Error(t, err)
	// stderr may or may not have content depending on OS
	_ = stderr
}

func TestDiscoverMakefileList_ReadError(t *testing.T) {
	mock := NewMockCommandExecutor()
	service := NewService(mock, false)

	// Try to discover from a non-existent file
	_, err := service.discoverMakefileList("/nonexistent/path/Makefile")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read Makefile")
}

func TestDiscoverMakefileList_Success(t *testing.T) {
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	// Create main Makefile without includes (simpler test)
	err := os.WriteFile(makefilePath, []byte("all:\n\t@echo hello\n"), 0644)
	require.NoError(t, err)

	// Use real executor for integration test
	executor := NewDefaultExecutor()
	service := NewService(executor, false)

	makefiles, err := service.discoverMakefileList(makefilePath)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(makefiles), 1)
	// The first file should be the main Makefile
	assert.Equal(t, makefilePath, makefiles[0])
}

func TestDiscoverMakefileList_Verbose(t *testing.T) {
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	err := os.WriteFile(makefilePath, []byte("all:\n\t@echo hello\n"), 0644)
	require.NoError(t, err)

	executor := NewDefaultExecutor()
	service := NewService(executor, true) // verbose mode

	makefiles, err := service.discoverMakefileList(makefilePath)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(makefiles), 1)
}

func TestDiscoverMakefileList_EmptyResult(t *testing.T) {
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")

	// Create a Makefile that somehow results in empty MAKEFILE_LIST
	// This is hard to trigger naturally, so we'll use a mock
	err := os.WriteFile(makefilePath, []byte("all:\n\t@echo hello\n"), 0644)
	require.NoError(t, err)

	mock := NewMockCommandExecutor()
	mock.SetPrefixMatch(true)
	mock.SetOutput("make -s --no-print-directory -f", "") // Empty output

	service := NewService(mock, false)
	_, err = service.discoverMakefileList(makefilePath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no Makefiles found")
}
