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

func TestCreateHelpTarget_DryRun(t *testing.T) {
	// Create a temp Makefile
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	err := os.WriteFile(makefilePath, []byte(`
## !category Build
## Build the project
build:
	@echo building

## !category Test
## Run tests
test:
	@echo testing
`), 0644)
	require.NoError(t, err)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		"--makefile-path", makefilePath,
		"",
		"--dry-run",
	})

	// Capture stdout (dry-run output goes to stdout via fmt.Print)
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = cmd.Execute()
	require.NoError(t, err)

	_ = w.Close()
	os.Stdout = oldStdout

	var outBuf bytes.Buffer
	_, _ = outBuf.ReadFrom(r)
	output := outBuf.String()

	// Verify dry-run message
	assert.Contains(t, output, "Dry run mode - no files will be modified")

	// Verify it shows what would be created
	assert.Contains(t, output, "Would create:")

	// Verify content preview is shown (new static format)
	assert.Contains(t, output, ".PHONY: help")
	assert.Contains(t, output, "@printf") // New static format uses printf statements
	assert.Contains(t, output, "--- end ---")

	// Verify no files were actually created
	// Check that make/help.mk was NOT created (should create make/help.mk in dry-run)
	helpMkPath := filepath.Join(tmpDir, "make", "help.mk")
	_, err = os.Stat(helpMkPath)
	assert.True(t, os.IsNotExist(err), "make/help.mk should not be created in dry-run mode")

	// Verify Makefile was NOT modified
	content, err := os.ReadFile(makefilePath)
	require.NoError(t, err)
	assert.NotContains(t, string(content), "include", "Makefile should not be modified in dry-run mode")
}

func TestCreateHelpTarget_DryRunWithHelpFileRelPath(t *testing.T) {
	// Create a temp Makefile
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	err := os.WriteFile(makefilePath, []byte(`
## Build the project
build:
	@echo building
`), 0644)
	require.NoError(t, err)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		"--makefile-path", makefilePath,
		"",
		"--dry-run",
		"--help-file-rel-path", "custom-help.mk",
	})

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = cmd.Execute()
	require.NoError(t, err)

	_ = w.Close()
	os.Stdout = oldStdout

	var outBuf bytes.Buffer
	_, _ = outBuf.ReadFrom(r)
	output := outBuf.String()

	// Verify dry-run message
	assert.Contains(t, output, "Dry run mode - no files will be modified")

	// Verify it shows the custom file path
	assert.Contains(t, output, "custom-help.mk")

	// Verify it would append include directive
	assert.Contains(t, output, "Would append to:")
	assert.Contains(t, output, "Append to")
	assert.Contains(t, output, "-include $(dir $(lastword $(MAKEFILE_LIST)))custom-help.mk")

	// Verify no files were actually created
	customHelpPath := filepath.Join(tmpDir, "custom-help.mk")
	_, err = os.Stat(customHelpPath)
	assert.True(t, os.IsNotExist(err), "custom-help.mk should not be created in dry-run mode")

	// Verify Makefile was NOT modified
	content, err := os.ReadFile(makefilePath)
	require.NoError(t, err)
	assert.NotContains(t, string(content), "include", "Makefile should not be modified in dry-run mode")
}

func TestCreateHelpTarget_DryRunWithMakeDirectory(t *testing.T) {
	// Create a temp Makefile with include make/*.mk pattern
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	// Use -include (optional include) to avoid errors during validation
	// since the make/ directory doesn't exist yet
	err := os.WriteFile(makefilePath, []byte(`
-include make/*.mk

## Build the project
build:
	@echo building
`), 0644)
	require.NoError(t, err)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		"--makefile-path", makefilePath,
		"",
		"--dry-run",
	})

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = cmd.Execute()
	require.NoError(t, err)

	_ = w.Close()
	os.Stdout = oldStdout

	var outBuf bytes.Buffer
	_, _ = outBuf.ReadFrom(r)
	output := outBuf.String()

	// Verify dry-run message
	assert.Contains(t, output, "Dry run mode - no files will be modified")

	// Verify it shows make/help.mk would be created (no numbered files, so no prefix)
	assert.Contains(t, output, "make/help.mk")

	// Should NOT show "Would append to" since it uses the include pattern
	assert.NotContains(t, output, "Would append to:")

	// Verify make directory was NOT created
	makeDir := filepath.Join(tmpDir, "make")
	_, err = os.Stat(makeDir)
	assert.True(t, os.IsNotExist(err), "make directory should not be created in dry-run mode")

	// Verify help.mk was NOT created
	helpFile := filepath.Join(makeDir, "help.mk")
	_, err = os.Stat(helpFile)
	assert.True(t, os.IsNotExist(err), "help.mk should not be created in dry-run mode")
}

func TestCreateHelpTarget_DryRunWithOptions(t *testing.T) {
	// Create a temp Makefile
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	err := os.WriteFile(makefilePath, []byte(`
## !category Build
## Build the project
build:
	@echo building

## !category Test
## Run tests
test:
	@echo testing
`), 0644)
	require.NoError(t, err)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		"--makefile-path", makefilePath,
		"--dry-run",
		"--keep-order-categories",
		"--keep-order-targets",
		"--default-category", "General",
	})

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = cmd.Execute()
	require.NoError(t, err)

	_ = w.Close()
	os.Stdout = oldStdout

	var outBuf bytes.Buffer
	_, _ = outBuf.ReadFrom(r)
	output := outBuf.String()

	// Verify dry-run message
	assert.Contains(t, output, "Dry run mode - no files will be modified")

	// Verify options are included in the generated content
	assert.Contains(t, output, "--keep-order-categories")
	assert.Contains(t, output, "--keep-order-targets")
	assert.Contains(t, output, "--default-category General")
}

func TestCreateHelpTarget_ActualCreation(t *testing.T) {
	t.Parallel()
	// Test that without --dry-run, files are actually created
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	err := os.WriteFile(makefilePath, []byte(`
## Build the project
build:
	@echo building
`), 0644)
	require.NoError(t, err)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		"--makefile-path", makefilePath,
		"",
		"--help-file-rel-path", "help.mk",
	})

	err = cmd.Execute()
	require.NoError(t, err)

	// Verify help.mk was created
	helpMkPath := filepath.Join(tmpDir, "help.mk")
	_, err = os.Stat(helpMkPath)
	assert.NoError(t, err, "help.mk should be created")

	// Verify Makefile was modified with include directive
	content, err := os.ReadFile(makefilePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "-include $(dir $(lastword $(MAKEFILE_LIST)))help.mk")
}

func TestPrintDryRunOutput(t *testing.T) {
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	targetFile := filepath.Join(tmpDir, "help.mk")
	content := "test content"

	tests := []struct {
		name         string
		needsInclude bool
		wantAppend   bool
	}{
		{
			name:         "with include directive",
			needsInclude: true,
			wantAppend:   true,
		},
		{
			name:         "without include directive",
			needsInclude: false,
			wantAppend:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := printDryRunOutput(makefilePath, targetFile, tt.needsInclude, content)

			_ = w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			assert.NoError(t, err)
			assert.Contains(t, output, "Dry run mode - no files will be modified")
			assert.Contains(t, output, "Would create:")
			assert.Contains(t, output, targetFile)
			assert.Contains(t, output, "test content")

			if tt.wantAppend {
				assert.Contains(t, output, "Would append to:")
				assert.Contains(t, output, "Append to")
				assert.Contains(t, output, "-include $(dir $(lastword $(MAKEFILE_LIST)))")
			} else {
				assert.NotContains(t, output, "Would append to:")
			}
		})
	}
}

func TestDryRunOutputFormat(t *testing.T) {
	// Test that the dry-run output matches the expected format
	tmpDir := t.TempDir()
	makefilePath := filepath.Join(tmpDir, "Makefile")
	err := os.WriteFile(makefilePath, []byte(`
## Build the project
build:
	@echo building
`), 0644)
	require.NoError(t, err)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		"--makefile-path", makefilePath,
		"",
		"--dry-run",
		"--help-file-rel-path", "make/01-help.mk",
	})

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = cmd.Execute()
	require.NoError(t, err)

	_ = w.Close()
	os.Stdout = oldStdout

	var outBuf bytes.Buffer
	_, _ = outBuf.ReadFrom(r)
	output := outBuf.String()

	// Check the output format matches the spec
	lines := strings.Split(output, "\n")

	// Should contain the header
	assert.Contains(t, output, "Dry run mode - no files will be modified")

	// Should contain file markers
	containsStartMarker := false
	containsEndMarker := false
	for _, line := range lines {
		if strings.HasPrefix(line, "---") && strings.Contains(line, "make/01-help.mk") {
			containsStartMarker = true
		}
		if line == "--- end ---" {
			containsEndMarker = true
		}
	}
	assert.True(t, containsStartMarker, "Should contain start marker for file content")
	assert.True(t, containsEndMarker, "Should contain end marker")
}

func TestFilterOutHelpFiles(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		makefiles []string
		helpFiles []string
		want      []string
	}{
		{
			name: "filter single help file",
			makefiles: []string{
				"/path/to/Makefile",
				"/path/to/make/help.mk",
				"/path/to/make/build.mk",
			},
			helpFiles: []string{"/path/to/make/help.mk"},
			want: []string{
				"/path/to/Makefile",
				"/path/to/make/build.mk",
			},
		},
		{
			name: "filter multiple help files",
			makefiles: []string{
				"/path/to/Makefile",
				"/path/to/make/help.mk",
				"/path/to/make/old-help.mk",
				"/path/to/make/build.mk",
			},
			helpFiles: []string{"/path/to/make/help.mk", "/path/to/make/old-help.mk"},
			want: []string{
				"/path/to/Makefile",
				"/path/to/make/build.mk",
			},
		},
		{
			name: "no help files to filter",
			makefiles: []string{
				"/path/to/Makefile",
				"/path/to/make/build.mk",
			},
			helpFiles: []string{},
			want: []string{
				"/path/to/Makefile",
				"/path/to/make/build.mk",
			},
		},
		{
			name: "empty string in help files",
			makefiles: []string{
				"/path/to/Makefile",
				"/path/to/make/help.mk",
				"/path/to/make/build.mk",
			},
			helpFiles: []string{"", "/path/to/make/help.mk"},
			want: []string{
				"/path/to/Makefile",
				"/path/to/make/build.mk",
			},
		},
		{
			name: "path normalization",
			makefiles: []string{
				"/path/to/Makefile",
				"/path/to/make/help.mk",
				"/path/to/make/build.mk",
			},
			helpFiles: []string{"/path/to/make/../make/help.mk"},
			want: []string{
				"/path/to/Makefile",
				"/path/to/make/build.mk",
			},
		},
		{
			name: "duplicate help files",
			makefiles: []string{
				"/path/to/Makefile",
				"/path/to/make/help.mk",
				"/path/to/make/build.mk",
			},
			helpFiles: []string{"/path/to/make/help.mk", "/path/to/make/help.mk"},
			want: []string{
				"/path/to/Makefile",
				"/path/to/make/build.mk",
			},
		},
		{
			name:      "empty makefiles list",
			makefiles: []string{},
			helpFiles: []string{"/path/to/make/help.mk"},
			want:      []string{},
		},
		{
			name: "help file not in makefiles list",
			makefiles: []string{
				"/path/to/Makefile",
				"/path/to/make/build.mk",
			},
			helpFiles: []string{"/path/to/make/help.mk"},
			want: []string{
				"/path/to/Makefile",
				"/path/to/make/build.mk",
			},
		},
		{
			name: "all makefiles are help files",
			makefiles: []string{
				"/path/to/make/help.mk",
				"/path/to/make/old-help.mk",
			},
			helpFiles: []string{"/path/to/make/help.mk", "/path/to/make/old-help.mk"},
			want:      []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := filterOutHelpFiles(tt.makefiles, tt.helpFiles...)
			assert.Equal(t, tt.want, got)
		})
	}
}
