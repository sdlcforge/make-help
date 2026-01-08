package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Tests for fix functions

func TestFixSummaryPunctuation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		warning  Warning
		wantFix  *Fix
		wantNil  bool
	}{
		{
			name: "adds period to summary",
			warning: Warning{
				File:      "/path/to/Makefile",
				Line:      10,
				Context:   "## Build the project",
				CheckName: "summary-punctuation",
			},
			wantFix: &Fix{
				File:       "/path/to/Makefile",
				Line:       10,
				Operation:  FixReplace,
				OldContent: "## Build the project",
				NewContent: "## Build the project.",
			},
		},
		{
			name: "returns nil when context is empty",
			warning: Warning{
				File:      "/path/to/Makefile",
				Line:      10,
				Context:   "",
				CheckName: "summary-punctuation",
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fix := fixSummaryPunctuation(tt.warning)
			if tt.wantNil {
				if fix != nil {
					t.Errorf("expected nil fix, got %+v", fix)
				}
				return
			}
			if fix == nil {
				t.Fatal("expected non-nil fix")
			}
			if fix.File != tt.wantFix.File {
				t.Errorf("File = %q, want %q", fix.File, tt.wantFix.File)
			}
			if fix.Line != tt.wantFix.Line {
				t.Errorf("Line = %d, want %d", fix.Line, tt.wantFix.Line)
			}
			if fix.Operation != tt.wantFix.Operation {
				t.Errorf("Operation = %v, want %v", fix.Operation, tt.wantFix.Operation)
			}
			if fix.OldContent != tt.wantFix.OldContent {
				t.Errorf("OldContent = %q, want %q", fix.OldContent, tt.wantFix.OldContent)
			}
			if fix.NewContent != tt.wantFix.NewContent {
				t.Errorf("NewContent = %q, want %q", fix.NewContent, tt.wantFix.NewContent)
			}
		})
	}
}

func TestFixEmptyDocumentation(t *testing.T) {
	t.Parallel()
	warning := Warning{
		File:      "/path/to/Makefile",
		Line:      5,
		Context:   "##",
		CheckName: "empty-doc",
	}

	fix := fixEmptyDocumentation(warning)
	if fix == nil {
		t.Fatal("expected non-nil fix")
	}
	if fix.File != warning.File {
		t.Errorf("File = %q, want %q", fix.File, warning.File)
	}
	if fix.Line != warning.Line {
		t.Errorf("Line = %d, want %d", fix.Line, warning.Line)
	}
	if fix.Operation != FixDelete {
		t.Errorf("Operation = %v, want FixDelete", fix.Operation)
	}
	if fix.OldContent != "##" {
		t.Errorf("OldContent = %q, want %q", fix.OldContent, "##")
	}
}

// Tests for CollectFixes

func TestCollectFixes_Empty(t *testing.T) {
	t.Parallel()
	checks := AllChecks()
	fixes := CollectFixes(checks, []Warning{})
	if len(fixes) != 0 {
		t.Errorf("expected 0 fixes, got %d", len(fixes))
	}
}

func TestCollectFixes_OnlyFixableWarnings(t *testing.T) {
	t.Parallel()
	checks := AllChecks()
	warnings := []Warning{
		{
			CheckName: "summary-punctuation",
			File:      "Makefile",
			Line:      10,
			Context:   "## Build",
			Fixable:   true,
		},
		{
			CheckName: "naming",
			File:      "Makefile",
			Line:      15,
			Fixable:   false, // Not fixable
		},
		{
			CheckName: "empty-doc",
			File:      "Makefile",
			Line:      20,
			Context:   "##",
			Fixable:   true,
		},
	}

	fixes := CollectFixes(checks, warnings)
	if len(fixes) != 2 {
		t.Fatalf("expected 2 fixes, got %d", len(fixes))
	}

	// Verify first fix is summary-punctuation
	if fixes[0].Operation != FixReplace {
		t.Errorf("fix[0].Operation = %v, want FixReplace", fixes[0].Operation)
	}
	if fixes[0].NewContent != "## Build." {
		t.Errorf("fix[0].NewContent = %q, want %q", fixes[0].NewContent, "## Build.")
	}

	// Verify second fix is empty-doc
	if fixes[1].Operation != FixDelete {
		t.Errorf("fix[1].Operation = %v, want FixDelete", fixes[1].Operation)
	}
}

func TestCollectFixes_UnknownCheckName(t *testing.T) {
	t.Parallel()
	checks := AllChecks()
	warnings := []Warning{
		{
			CheckName: "unknown-check",
			File:      "Makefile",
			Line:      10,
			Fixable:   true,
		},
	}

	fixes := CollectFixes(checks, warnings)
	if len(fixes) != 0 {
		t.Errorf("expected 0 fixes for unknown check, got %d", len(fixes))
	}
}

// Tests for Fixer

func TestFixer_ApplyFixes_Empty(t *testing.T) {
	t.Parallel()
	fixer := &Fixer{}
	result, err := fixer.ApplyFixes([]Fix{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalFixed != 0 {
		t.Errorf("TotalFixed = %d, want 0", result.TotalFixed)
	}
	if len(result.FilesModified) != 0 {
		t.Errorf("FilesModified = %d, want 0", len(result.FilesModified))
	}
}

func TestFixer_ApplyFixes_Replace(t *testing.T) {
	t.Parallel()
	// Create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "Makefile")
	content := `## Build the project
build:
	@echo "building"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	fixes := []Fix{
		{
			File:       tmpFile,
			Line:       1,
			Operation:  FixReplace,
			OldContent: "## Build the project",
			NewContent: "## Build the project.",
		},
	}

	fixer := &Fixer{}
	result, err := fixer.ApplyFixes(fixes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalFixed != 1 {
		t.Errorf("TotalFixed = %d, want 1", result.TotalFixed)
	}
	if result.FilesModified[tmpFile] != 1 {
		t.Errorf("FilesModified[%s] = %d, want 1", tmpFile, result.FilesModified[tmpFile])
	}

	// Verify file content
	got, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	want := `## Build the project.
build:
	@echo "building"
`
	if string(got) != want {
		t.Errorf("file content:\ngot:\n%s\nwant:\n%s", string(got), want)
	}
}

func TestFixer_ApplyFixes_Delete(t *testing.T) {
	t.Parallel()
	// Create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "Makefile")
	content := `##
## Build the project.
build:
	@echo "building"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	fixes := []Fix{
		{
			File:       tmpFile,
			Line:       1,
			Operation:  FixDelete,
			OldContent: "##",
		},
	}

	fixer := &Fixer{}
	result, err := fixer.ApplyFixes(fixes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalFixed != 1 {
		t.Errorf("TotalFixed = %d, want 1", result.TotalFixed)
	}

	// Verify file content
	got, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	want := `## Build the project.
build:
	@echo "building"
`
	if string(got) != want {
		t.Errorf("file content:\ngot:\n%s\nwant:\n%s", string(got), want)
	}
}

func TestFixer_ApplyFixes_MultipleFixesSameFile(t *testing.T) {
	t.Parallel()
	// Create temp file with multiple issues
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "Makefile")
	content := `##
## Build the project
build:
	@echo "building"

##
## Run tests
test:
	@echo "testing"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	fixes := []Fix{
		{
			File:       tmpFile,
			Line:       1,
			Operation:  FixDelete,
			OldContent: "##",
		},
		{
			File:       tmpFile,
			Line:       2,
			Operation:  FixReplace,
			OldContent: "## Build the project",
			NewContent: "## Build the project.",
		},
		{
			File:       tmpFile,
			Line:       6,
			Operation:  FixDelete,
			OldContent: "##",
		},
		{
			File:       tmpFile,
			Line:       7,
			Operation:  FixReplace,
			OldContent: "## Run tests",
			NewContent: "## Run tests.",
		},
	}

	fixer := &Fixer{}
	result, err := fixer.ApplyFixes(fixes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalFixed != 4 {
		t.Errorf("TotalFixed = %d, want 4", result.TotalFixed)
	}

	// Verify file content
	got, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	want := `## Build the project.
build:
	@echo "building"

## Run tests.
test:
	@echo "testing"
`
	if string(got) != want {
		t.Errorf("file content:\ngot:\n%s\nwant:\n%s", string(got), want)
	}
}

func TestFixer_ApplyFixes_DryRun(t *testing.T) {
	t.Parallel()
	// Create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "Makefile")
	originalContent := `## Build the project
build:
	@echo "building"
`
	if err := os.WriteFile(tmpFile, []byte(originalContent), 0644); err != nil {
		t.Fatal(err)
	}

	fixes := []Fix{
		{
			File:       tmpFile,
			Line:       1,
			Operation:  FixReplace,
			OldContent: "## Build the project",
			NewContent: "## Build the project.",
		},
	}

	fixer := &Fixer{DryRun: true}
	result, err := fixer.ApplyFixes(fixes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalFixed != 1 {
		t.Errorf("TotalFixed = %d, want 1", result.TotalFixed)
	}

	// Verify file was NOT modified
	got, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != originalContent {
		t.Errorf("dry-run should not modify file:\ngot:\n%s\nwant:\n%s", string(got), originalContent)
	}
}

func TestFixer_ApplyFixes_ContentMismatch(t *testing.T) {
	t.Parallel()
	// Create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "Makefile")
	content := `## Build project (changed)
build:
	@echo "building"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Fix expects different content than what's in the file
	fixes := []Fix{
		{
			File:       tmpFile,
			Line:       1,
			Operation:  FixReplace,
			OldContent: "## Build the project",
			NewContent: "## Build the project.",
		},
	}

	fixer := &Fixer{}
	result, err := fixer.ApplyFixes(fixes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Fix should be skipped due to content mismatch
	if result.TotalFixed != 0 {
		t.Errorf("TotalFixed = %d, want 0 (fix should be skipped)", result.TotalFixed)
	}
}

func TestFixer_ApplyFixes_LineOutOfRange(t *testing.T) {
	t.Parallel()
	// Create temp file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "Makefile")
	content := `## Build
build:
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	fixes := []Fix{
		{
			File:       tmpFile,
			Line:       100, // Line doesn't exist
			Operation:  FixReplace,
			OldContent: "something",
			NewContent: "something else",
		},
	}

	fixer := &Fixer{}
	result, err := fixer.ApplyFixes(fixes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Fix should be skipped due to invalid line
	if result.TotalFixed != 0 {
		t.Errorf("TotalFixed = %d, want 0 (fix should be skipped)", result.TotalFixed)
	}
}

func TestFixer_ApplyFixes_PreservesPermissions(t *testing.T) {
	t.Parallel()
	// Create temp file with specific permissions
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "Makefile")
	content := `## Build the project
build:
	@echo "building"
`
	if err := os.WriteFile(tmpFile, []byte(content), 0755); err != nil {
		t.Fatal(err)
	}

	fixes := []Fix{
		{
			File:       tmpFile,
			Line:       1,
			Operation:  FixReplace,
			OldContent: "## Build the project",
			NewContent: "## Build the project.",
		},
	}

	fixer := &Fixer{}
	_, err := fixer.ApplyFixes(fixes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify permissions are preserved
	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("file permissions = %o, want %o", info.Mode().Perm(), 0755)
	}
}

func TestFixer_ApplyFixes_MultipleFiles(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create two files
	file1 := filepath.Join(tmpDir, "Makefile")
	content1 := `## Build
build:
`
	if err := os.WriteFile(file1, []byte(content1), 0644); err != nil {
		t.Fatal(err)
	}

	file2 := filepath.Join(tmpDir, "other.mk")
	content2 := `## Test
test:
`
	if err := os.WriteFile(file2, []byte(content2), 0644); err != nil {
		t.Fatal(err)
	}

	fixes := []Fix{
		{
			File:       file1,
			Line:       1,
			Operation:  FixReplace,
			OldContent: "## Build",
			NewContent: "## Build.",
		},
		{
			File:       file2,
			Line:       1,
			Operation:  FixReplace,
			OldContent: "## Test",
			NewContent: "## Test.",
		},
	}

	fixer := &Fixer{}
	result, err := fixer.ApplyFixes(fixes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalFixed != 2 {
		t.Errorf("TotalFixed = %d, want 2", result.TotalFixed)
	}
	if len(result.FilesModified) != 2 {
		t.Errorf("FilesModified count = %d, want 2", len(result.FilesModified))
	}

	// Verify both files were modified
	got1, _ := os.ReadFile(file1)
	if !strings.Contains(string(got1), "## Build.") {
		t.Errorf("file1 not modified correctly")
	}
	got2, _ := os.ReadFile(file2)
	if !strings.Contains(string(got2), "## Test.") {
		t.Errorf("file2 not modified correctly")
	}
}

// Tests for validateFix helper

func TestValidateFix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		fix     Fix
		lines   []string
		wantErr bool
	}{
		{
			name: "valid fix with matching content",
			fix: Fix{
				Line:       1,
				OldContent: "## Build",
			},
			lines:   []string{"## Build", "build:"},
			wantErr: false,
		},
		{
			name: "content mismatch",
			fix: Fix{
				Line:       1,
				OldContent: "## Build",
			},
			lines:   []string{"## Test", "build:"},
			wantErr: true,
		},
		{
			name: "line number out of range (too high)",
			fix: Fix{
				Line:       100,
				OldContent: "anything",
			},
			lines:   []string{"line1", "line2"},
			wantErr: true,
		},
		{
			name: "line number out of range (zero)",
			fix: Fix{
				Line:       0,
				OldContent: "anything",
			},
			lines:   []string{"line1"},
			wantErr: true,
		},
		{
			name: "empty old content allows any match",
			fix: Fix{
				Line:       1,
				OldContent: "",
			},
			lines:   []string{"anything here"},
			wantErr: false,
		},
		{
			name: "whitespace handling",
			fix: Fix{
				Line:       1,
				OldContent: "  ## Build  ",
			},
			lines:   []string{"## Build"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateFix(tt.fix, tt.lines)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
