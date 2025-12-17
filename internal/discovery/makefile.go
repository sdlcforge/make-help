package discovery

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sdlcforge/make-help/internal/errors"
)

// ResolveMakefilePath resolves a Makefile path to an absolute path.
// If the path is empty, it defaults to "Makefile" in the current working directory.
// If the path is relative, it is resolved relative to the current working directory.
func ResolveMakefilePath(path string) (string, error) {
	// Default to Makefile in current directory
	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
		path = filepath.Join(cwd, "Makefile")
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	return absPath, nil
}

// ValidateMakefileExists checks if a Makefile exists at the given path.
// Returns MakefileNotFoundError if the file does not exist.
func ValidateMakefileExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.NewMakefileNotFoundError(path)
		}
		return fmt.Errorf("failed to stat Makefile: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", path)
	}

	return nil
}
