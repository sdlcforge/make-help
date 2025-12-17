package cli

import (
	"fmt"
	"os"

	"github.com/sdlcforge/make-help/internal/discovery"
	"github.com/sdlcforge/make-help/internal/target"
)

// runRemoveHelpTarget removes help targets from the Makefile.
func runRemoveHelpTarget(config *Config) error {
	// 1. Resolve Makefile path
	makefilePath, err := discovery.ResolveMakefilePath(config.MakefilePath)
	if err != nil {
		return fmt.Errorf("failed to resolve Makefile path: %w", err)
	}

	if err := discovery.ValidateMakefileExists(makefilePath); err != nil {
		return err
	}

	config.MakefilePath = makefilePath

	if config.Verbose {
		fmt.Fprintf(os.Stderr, "Using Makefile: %s\n", makefilePath)
	}

	// 2. Create remove service and execute
	executor := discovery.NewDefaultExecutor()
	removeConfig := &target.Config{
		MakefilePath: makefilePath,
	}
	removeService := target.NewRemoveService(removeConfig, executor, config.Verbose)

	if err := removeService.RemoveTarget(); err != nil {
		return fmt.Errorf("failed to remove help target: %w", err)
	}

	return nil
}
