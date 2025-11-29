package cli

import (
	"fmt"

	"github.com/sdlcforge/make-help/internal/discovery"
	"github.com/sdlcforge/make-help/internal/target"
	"github.com/spf13/cobra"
)

// newRemoveTargetCmd creates the remove-target subcommand.
// This command removes all help target artifacts from the Makefile.
func newRemoveTargetCmd(config *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-target",
		Short: "Remove the help target from the Makefile",
		Long: `Remove all help target artifacts from the Makefile.

This command performs the following cleanup:
  1. Remove include directives for help target files
  2. Remove inline help: target and .PHONY: help
  3. Delete help target files (make/01-help.mk)`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve Makefile path
			makefilePath, err := discovery.ResolveMakefilePath(config.MakefilePath)
			if err != nil {
				return err
			}
			config.MakefilePath = makefilePath

			// Validate Makefile exists
			if err := discovery.ValidateMakefileExists(makefilePath); err != nil {
				return err
			}

			// Convert to target config
			targetConfig := &target.Config{
				MakefilePath: config.MakefilePath,
			}

			// Create remove service
			executor := discovery.NewDefaultExecutor()
			removeService := target.NewRemoveService(targetConfig, executor, config.Verbose)

			// Remove target
			return removeService.RemoveTarget()
		},
	}

	return cmd
}

// runRemoveTarget executes the remove-target command.
func runRemoveTarget(config *Config) error {
	// Resolve Makefile path
	makefilePath, err := discovery.ResolveMakefilePath(config.MakefilePath)
	if err != nil {
		return err
	}
	config.MakefilePath = makefilePath

	// Validate Makefile exists
	if err := discovery.ValidateMakefileExists(makefilePath); err != nil {
		return fmt.Errorf("Makefile not found: %w", err)
	}

	// Convert to target config
	targetConfig := &target.Config{
		MakefilePath: config.MakefilePath,
	}

	// Create remove service
	executor := discovery.NewDefaultExecutor()
	removeService := target.NewRemoveService(targetConfig, executor, config.Verbose)

	// Remove target
	return removeService.RemoveTarget()
}
