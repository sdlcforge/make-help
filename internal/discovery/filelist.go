package discovery

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// discoverMakefileList discovers all Makefiles using the MAKEFILE_LIST variable.
// It creates a temporary file with the main Makefile content plus a discovery target,
// executes make to get the MAKEFILE_LIST, and returns the list of files.
//
// SECURITY: This function uses temporary physical files instead of bash process
// substitution to prevent command injection vulnerabilities.
func (s *Service) discoverMakefileList(mainPath string) ([]string, error) {
	// Read main Makefile content
	mainContent, err := os.ReadFile(mainPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Makefile: %w", err)
	}

	// Create temporary file in the same directory as the Makefile
	// This ensures relative includes work correctly
	dir := filepath.Dir(mainPath)
	tmpFile, err := os.CreateTemp(dir, ".makefile-discovery-*.mk")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmpFile.Name()

	// Clean up temporary file when done
	defer os.Remove(tmpName)

	// Write main content + discovery target
	discoveryTarget := "\n\n.PHONY: _list_makefiles\n_list_makefiles:\n\t@echo $(MAKEFILE_LIST)\n"

	if _, err := tmpFile.Write(mainContent); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	if _, err := tmpFile.WriteString(discoveryTarget); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("failed to write discovery target: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to close temp file: %w", err)
	}

	// Execute make with timeout to prevent indefinite hangs
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use -s (silent) and --no-print-directory to prevent make from adding
	// extra output like job server info or directory messages that could
	// corrupt the MAKEFILE_LIST output when running from within another make
	stdout, stderr, err := s.executor.ExecuteContext(ctx, "make", "-s", "--no-print-directory", "-f", tmpName, "_list_makefiles")
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("make command timed out after 30s")
		}
		return nil, fmt.Errorf("failed to discover makefiles: %w\nstderr: %s", err, stderr)
	}

	// Parse space-separated file list
	files := strings.Fields(strings.TrimSpace(stdout))

	if len(files) == 0 {
		return nil, fmt.Errorf("no Makefiles found in MAKEFILE_LIST")
	}

	// The first file in MAKEFILE_LIST will be the temp file, replace it with the original
	if len(files) > 0 && filepath.Base(files[0]) == filepath.Base(tmpName) {
		files[0] = mainPath
	}

	// Resolve to absolute paths
	resolved, err := s.resolveAbsolutePaths(files, dir)
	if err != nil {
		return nil, err
	}

	if s.verbose {
		fmt.Printf("Discovered %d Makefile(s):\n", len(resolved))
		for i, f := range resolved {
			fmt.Printf("  %d. %s\n", i+1, f)
		}
	}

	return resolved, nil
}

// resolveAbsolutePaths converts relative paths to absolute paths.
// Paths are resolved relative to the provided base directory.
func (s *Service) resolveAbsolutePaths(files []string, baseDir string) ([]string, error) {
	resolved := make([]string, 0, len(files))

	for _, file := range files {
		var absPath string
		if filepath.IsAbs(file) {
			absPath = file
		} else {
			absPath = filepath.Join(baseDir, file)
		}

		// Clean the path
		absPath = filepath.Clean(absPath)

		// Validate that the file exists
		if _, err := os.Stat(absPath); err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("Makefile not found: %s", absPath)
			}
			return nil, fmt.Errorf("failed to stat Makefile %s: %w", absPath, err)
		}

		resolved = append(resolved, absPath)
	}

	return resolved, nil
}
