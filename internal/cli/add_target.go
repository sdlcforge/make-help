package cli

import (
	"fmt"

	"github.com/sdlcforge/make-help/internal/discovery"
	"github.com/sdlcforge/make-help/internal/target"
	"github.com/spf13/cobra"
)

// newAddTargetCmd creates the add-target subcommand.
// This command generates and injects a help target into the Makefile.
func newAddTargetCmd(config *Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-target",
		Short: "Add a help target to the Makefile",
		Long: `Add a help target to the Makefile that invokes make-help.

The target file location is determined by a three-tier strategy:
  1. Use --target-file if specified
  2. Create make/01-help.mk if "include make/*.mk" pattern exists
  3. Append directly to the main Makefile

All configuration flags are passed through to the generated help target.`,
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
				MakefilePath:        config.MakefilePath,
				TargetFile:          config.TargetFile,
				KeepOrderCategories: config.KeepOrderCategories,
				KeepOrderTargets:    config.KeepOrderTargets,
				CategoryOrder:       config.CategoryOrder,
				DefaultCategory:     config.DefaultCategory,
			}

			// Create add service
			executor := discovery.NewDefaultExecutor()
			addService := target.NewAddService(targetConfig, executor, config.Verbose)

			// Add target
			return addService.AddTarget()
		},
	}

	// Add-target specific flag
	cmd.Flags().StringVar(&config.TargetFile,
		"target-file", "", "Explicit path for the help target file")

	// Inherit help generation flags
	cmd.Flags().BoolVar(&config.KeepOrderCategories,
		"keep-order-categories", false, "Preserve category discovery order")
	cmd.Flags().BoolVar(&config.KeepOrderTargets,
		"keep-order-targets", false, "Preserve target discovery order within categories")

	// --keep-order-all is a convenience flag
	var keepOrderAll bool
	cmd.Flags().BoolVar(&keepOrderAll,
		"keep-order-all", false, "Preserve both category and target discovery order")

	cmd.Flags().StringSliceVar(&config.CategoryOrder,
		"category-order", []string{}, "Explicit category order (comma-separated)")
	cmd.Flags().StringVar(&config.DefaultCategory,
		"default-category", "", "Default category for uncategorized targets")

	// Process --keep-order-all flag
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if keepOrderAll {
			config.KeepOrderCategories = true
			config.KeepOrderTargets = true
		}
		return nil
	}

	return cmd
}

// runAddTarget executes the add-target command.
func runAddTarget(config *Config) error {
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
		MakefilePath:        config.MakefilePath,
		TargetFile:          config.TargetFile,
		KeepOrderCategories: config.KeepOrderCategories,
		KeepOrderTargets:    config.KeepOrderTargets,
		CategoryOrder:       config.CategoryOrder,
		DefaultCategory:     config.DefaultCategory,
	}

	// Create add service
	executor := discovery.NewDefaultExecutor()
	addService := target.NewAddService(targetConfig, executor, config.Verbose)

	// Add target
	return addService.AddTarget()
}
