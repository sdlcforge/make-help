package target

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAtomicWriteFile(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		perm    os.FileMode
		wantErr bool
	}{
		{
			name:    "write simple file",
			data:    []byte("Hello, World!\n"),
			perm:    0644,
			wantErr: false,
		},
		{
			name:    "write empty file",
			data:    []byte(""),
			perm:    0644,
			wantErr: false,
		},
		{
			name:    "write multiline content",
			data:    []byte("Line 1\nLine 2\nLine 3\n"),
			perm:    0600,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir := t.TempDir()
			targetFile := filepath.Join(tmpDir, "test.txt")

			// Write file atomically
			err := AtomicWriteFile(targetFile, tt.data, tt.perm)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify file exists
			info, err := os.Stat(targetFile)
			require.NoError(t, err)

			// Verify permissions
			assert.Equal(t, tt.perm, info.Mode().Perm())

			// Verify content
			content, err := os.ReadFile(targetFile)
			require.NoError(t, err)
			assert.Equal(t, tt.data, content)
		})
	}
}

func TestAtomicWriteFileOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "test.txt")

	// Write initial content
	err := AtomicWriteFile(targetFile, []byte("Initial content\n"), 0644)
	require.NoError(t, err)

	// Overwrite with new content
	err = AtomicWriteFile(targetFile, []byte("New content\n"), 0644)
	require.NoError(t, err)

	// Verify new content
	content, err := os.ReadFile(targetFile)
	require.NoError(t, err)
	assert.Equal(t, []byte("New content\n"), content)
}

func TestAtomicWriteFileInvalidDirectory(t *testing.T) {
	// Attempt to write to non-existent directory
	err := AtomicWriteFile("/nonexistent/directory/test.txt", []byte("test"), 0644)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create temp file")
}

func TestAtomicWriteFileNoTempFilesLeft(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "test.txt")

	// Write file
	err := AtomicWriteFile(targetFile, []byte("test content\n"), 0644)
	require.NoError(t, err)

	// List directory contents
	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)

	// Should only have the target file, no temp files
	assert.Len(t, entries, 1)
	assert.Equal(t, "test.txt", entries[0].Name())
}

func TestAtomicWriteFileErrorCases(t *testing.T) {
	t.Run("invalid permissions on write", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetFile := filepath.Join(tmpDir, "test.txt")

		// First write a file successfully
		err := AtomicWriteFile(targetFile, []byte("initial"), 0644)
		require.NoError(t, err)

		// Make directory read-only to cause permission error on temp file creation
		err = os.Chmod(tmpDir, 0444)
		require.NoError(t, err)
		defer func() { _ = os.Chmod(tmpDir, 0755) }() // Restore for cleanup

		// Attempt to write should fail
		err = AtomicWriteFile(targetFile, []byte("new content"), 0644)
		assert.Error(t, err)
	})

	t.Run("file in subdirectory", func(t *testing.T) {
		tmpDir := t.TempDir()
		subDir := filepath.Join(tmpDir, "subdir")
		err := os.MkdirAll(subDir, 0755)
		require.NoError(t, err)

		targetFile := filepath.Join(subDir, "test.txt")

		// Write to subdirectory should work
		err = AtomicWriteFile(targetFile, []byte("content"), 0600)
		require.NoError(t, err)

		// Verify file exists and has correct permissions
		info, err := os.Stat(targetFile)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	})
}

// TestAtomicWriteFileConcurrentSameFile verifies that multiple goroutines
// writing to the same file concurrently don't cause data corruption.
func TestAtomicWriteFileConcurrentSameFile(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "concurrent.txt")

	const (
		numGoroutines = 10
		iterations    = 100
	)

	// Run multiple iterations to increase chance of catching race conditions
	for iter := 0; iter < iterations; iter++ {
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// Each goroutine writes different content
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				content := []byte(fmt.Sprintf("Content from goroutine %d\n", id))
				err := AtomicWriteFile(targetFile, content, 0644)
				// Errors are acceptable under high concurrency, but should be rare
				if err != nil {
					t.Logf("Write error from goroutine %d (iter %d): %v", id, iter, err)
				}
			}(i)
		}

		wg.Wait()

		// Verify the file exists and contains valid content
		content, err := os.ReadFile(targetFile)
		require.NoError(t, err, "iteration %d: file should exist after concurrent writes", iter)

		// Content should match one of the expected values (no corruption)
		validContent := false
		for i := 0; i < numGoroutines; i++ {
			expected := []byte(fmt.Sprintf("Content from goroutine %d\n", i))
			if string(content) == string(expected) {
				validContent = true
				break
			}
		}
		assert.True(t, validContent, "iteration %d: file content corrupted: %q", iter, content)

		// Verify no temp files left behind
		entries, err := os.ReadDir(tmpDir)
		require.NoError(t, err)
		assert.Len(t, entries, 1, "iteration %d: should only have target file, no temp files", iter)
	}
}

// TestAtomicWriteFileConcurrentDifferentFiles verifies that multiple goroutines
// writing to different files concurrently work correctly.
func TestAtomicWriteFileConcurrentDifferentFiles(t *testing.T) {
	tmpDir := t.TempDir()

	const (
		numGoroutines = 20
		iterations    = 50
	)

	for iter := 0; iter < iterations; iter++ {
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		// Each goroutine writes to its own file
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				targetFile := filepath.Join(tmpDir, fmt.Sprintf("file%d-%d.txt", id, iter))
				content := []byte(fmt.Sprintf("Content for file %d iteration %d\n", id, iter))
				err := AtomicWriteFile(targetFile, content, 0644)
				require.NoError(t, err, "goroutine %d (iter %d): write should succeed", id, iter)
			}(i)
		}

		wg.Wait()

		// Verify all files were created with correct content
		for i := 0; i < numGoroutines; i++ {
			targetFile := filepath.Join(tmpDir, fmt.Sprintf("file%d-%d.txt", i, iter))
			content, err := os.ReadFile(targetFile)
			require.NoError(t, err, "file %d (iter %d) should exist", i, iter)

			expected := []byte(fmt.Sprintf("Content for file %d iteration %d\n", i, iter))
			assert.Equal(t, expected, content, "file %d (iter %d) has wrong content", i, iter)
		}

		// Verify no temp files left behind
		// We should have exactly numGoroutines files (one per goroutine) for this iteration
		entries, err := os.ReadDir(tmpDir)
		require.NoError(t, err)

		// Count non-temp files
		nonTempCount := 0
		for _, entry := range entries {
			if !entry.IsDir() && entry.Name()[0] != '.' {
				nonTempCount++
			}
		}

		// We accumulate files across iterations, so expected count is numGoroutines * (iter + 1)
		expectedCount := numGoroutines * (iter + 1)
		assert.Equal(t, expectedCount, nonTempCount, "iteration %d: unexpected file count", iter)
	}
}

// TestAtomicWriteFileConcurrentNoTempFilesLeftBehind specifically tests
// that no temp files remain after concurrent operations complete.
func TestAtomicWriteFileConcurrentNoTempFilesLeftBehind(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "test.txt")

	const (
		numGoroutines = 50
		iterations    = 20
	)

	for iter := 0; iter < iterations; iter++ {
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				content := []byte(fmt.Sprintf("Data from goroutine %d\n", id))
				_ = AtomicWriteFile(targetFile, content, 0644)
			}(i)
		}

		wg.Wait()

		// List all files in directory
		entries, err := os.ReadDir(tmpDir)
		require.NoError(t, err)

		// Check for temp files (start with .tmp-)
		tempFiles := []string{}
		for _, entry := range entries {
			name := entry.Name()
			if len(name) >= 5 && name[:5] == ".tmp-" {
				tempFiles = append(tempFiles, name)
			}
		}

		assert.Empty(t, tempFiles, "iteration %d: found temp files: %v", iter, tempFiles)

		// Should only have the target file
		regularFiles := []string{}
		for _, entry := range entries {
			if !entry.IsDir() && entry.Name()[0] != '.' {
				regularFiles = append(regularFiles, entry.Name())
			}
		}
		assert.Equal(t, []string{"test.txt"}, regularFiles, "iteration %d: unexpected files in directory", iter)
	}
}

// TestAtomicWriteFileConcurrentMixedOperations tests concurrent writes
// to the same file with different content sizes to stress-test the function.
func TestAtomicWriteFileConcurrentMixedOperations(t *testing.T) {
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "mixed.txt")

	const (
		numGoroutines = 15
		iterations    = 100
	)

	for iter := 0; iter < iterations; iter++ {
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()

				// Vary content size to test different scenarios
				var content []byte
				switch id % 3 {
				case 0:
					// Small content
					content = []byte(fmt.Sprintf("Small %d\n", id))
				case 1:
					// Medium content
					content = []byte(fmt.Sprintf("Medium content from goroutine %d with more text\n", id))
				case 2:
					// Large content
					largeText := make([]byte, 1024)
					for j := range largeText {
						largeText[j] = byte('A' + (id % 26))
					}
					content = largeText
				}

				_ = AtomicWriteFile(targetFile, content, 0644)
			}(i)
		}

		wg.Wait()

		// Verify file exists and is readable (not corrupted)
		content, err := os.ReadFile(targetFile)
		require.NoError(t, err, "iteration %d: file should exist", iter)
		assert.NotEmpty(t, content, "iteration %d: file should not be empty", iter)

		// Verify no temp files
		entries, err := os.ReadDir(tmpDir)
		require.NoError(t, err)
		assert.Len(t, entries, 1, "iteration %d: should only have target file", iter)
	}
}
