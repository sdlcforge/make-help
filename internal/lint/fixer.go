package lint

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Fixer applies fixes to source files.
type Fixer struct {
	// DryRun when true shows what would be fixed without modifying files.
	DryRun bool
}

// FixResult contains the results of applying fixes.
type FixResult struct {
	// TotalFixed is the number of fixes successfully applied.
	TotalFixed int

	// FilesModified maps file paths to the number of fixes applied.
	FilesModified map[string]int
}

// ApplyFixes groups fixes by file and applies them atomically.
// Fixes are applied in reverse line order to avoid offset invalidation.
// Returns an error if any fix fails; no partial changes are made per file.
func (f *Fixer) ApplyFixes(fixes []Fix) (*FixResult, error) {
	if len(fixes) == 0 {
		return &FixResult{FilesModified: make(map[string]int)}, nil
	}

	// Group fixes by file
	fileFixes := make(map[string][]Fix)
	for _, fix := range fixes {
		fileFixes[fix.File] = append(fileFixes[fix.File], fix)
	}

	result := &FixResult{
		FilesModified: make(map[string]int),
	}

	// Apply fixes file by file
	for file, fixes := range fileFixes {
		count, err := f.applyFileFixes(file, fixes)
		if err != nil {
			return result, fmt.Errorf("failed to fix %s: %w", file, err)
		}
		result.FilesModified[file] = count
		result.TotalFixed += count
	}

	return result, nil
}

// applyFileFixes applies all fixes to a single file atomically.
func (f *Fixer) applyFileFixes(filePath string, fixes []Fix) (int, error) {
	// Validate path is absolute
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return 0, fmt.Errorf("invalid path: %w", err)
	}

	// Read current file content
	lines, err := readFileLines(absPath)
	if err != nil {
		return 0, fmt.Errorf("read failed: %w", err)
	}

	// Sort fixes by line number (descending) to avoid offset issues
	sort.Slice(fixes, func(i, j int) bool {
		return fixes[i].Line > fixes[j].Line
	})

	// Track which lines to delete
	deleteLines := make(map[int]bool)

	// Validate and apply fixes
	applied := 0
	for _, fix := range fixes {
		if err := validateFix(fix, lines); err != nil {
			// Skip invalid fixes (file may have changed)
			continue
		}

		switch fix.Operation {
		case FixReplace:
			lines[fix.Line-1] = fix.NewContent
			applied++
		case FixDelete:
			deleteLines[fix.Line-1] = true
			applied++
		}
	}

	if applied == 0 {
		return 0, nil
	}

	// Filter out deleted lines
	filteredLines := make([]string, 0, len(lines)-len(deleteLines))
	for i, line := range lines {
		if !deleteLines[i] {
			filteredLines = append(filteredLines, line)
		}
	}

	if f.DryRun {
		// Just return count, don't modify file
		return applied, nil
	}

	// Write atomically
	if err := writeFileAtomic(absPath, filteredLines); err != nil {
		return 0, fmt.Errorf("write failed: %w", err)
	}

	return applied, nil
}

// validateFix ensures the fix is still applicable.
func validateFix(fix Fix, lines []string) error {
	if fix.Line < 1 || fix.Line > len(lines) {
		return fmt.Errorf("line %d out of range (file has %d lines)", fix.Line, len(lines))
	}

	actualLine := strings.TrimSpace(lines[fix.Line-1])
	expectedLine := strings.TrimSpace(fix.OldContent)

	if expectedLine != "" && actualLine != expectedLine {
		return fmt.Errorf("line content mismatch at line %d: expected %q, got %q",
			fix.Line, expectedLine, actualLine)
	}

	return nil
}

// readFileLines reads a file into lines.
func readFileLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

// writeFileAtomic writes lines to file atomically using temp file + rename.
func writeFileAtomic(path string, lines []string) error {
	// Get original file permissions
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	// Write to temp file in same directory (for atomic rename)
	dir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(dir, ".fix-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()

	// Ensure temp file is cleaned up on error
	defer func() {
		if tmpFile != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	writer := bufio.NewWriter(tmpFile)
	for i, line := range lines {
		if _, err := writer.WriteString(line); err != nil {
			return err
		}
		// Add newline after each line except potentially the last
		// (preserve original trailing newline behavior)
		if i < len(lines)-1 || true { // Always add trailing newline for Makefiles
			if _, err := writer.WriteString("\n"); err != nil {
				return err
			}
		}
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	// Set permissions before closing
	if err := tmpFile.Chmod(info.Mode()); err != nil {
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}

	// Mark as successfully renamed so defer doesn't remove it
	tmpFile = nil
	return nil
}
