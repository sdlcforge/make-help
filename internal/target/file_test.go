package target

import (
	"os"
	"path/filepath"
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
