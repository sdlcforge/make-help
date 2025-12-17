package discovery

import (
	"fmt"
)

// Service provides Makefile and target discovery functionality.
// It uses the CommandExecutor interface for testability.
type Service struct {
	executor CommandExecutor
	verbose  bool
}

// NewService creates a new discovery Service with the given executor and verbose flag.
func NewService(executor CommandExecutor, verbose bool) *Service {
	return &Service{
		executor: executor,
		verbose:  verbose,
	}
}

// DiscoverMakefiles discovers all Makefiles using the MAKEFILE_LIST variable.
// It returns an ordered list of absolute paths to all Makefiles (main and included).
//
// The function creates a temporary file with a special target to extract MAKEFILE_LIST,
// executes make, and parses the output. This approach is secure and avoids shell injection.
func (s *Service) DiscoverMakefiles(mainPath string) ([]string, error) {
	if s.verbose {
		fmt.Printf("Discovering Makefiles starting from: %s\n", mainPath)
	}

	return s.discoverMakefileList(mainPath)
}

// DiscoverTargets discovers all targets in the given Makefile using make -p.
// It returns target names and their .PHONY status extracted from the make database output.
//
// The function filters out special targets, pattern rules, and built-in targets,
// returning only user-defined targets.
func (s *Service) DiscoverTargets(makefilePath string) (*DiscoverTargetsResult, error) {
	if s.verbose {
		fmt.Printf("Discovering targets from: %s\n", makefilePath)
	}

	return s.discoverTargets(makefilePath)
}
