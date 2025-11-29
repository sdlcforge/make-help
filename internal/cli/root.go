package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command for make-help.
// The default action is to run the help command.
func NewRootCmd() *cobra.Command {
	config := NewConfig()

	// Flags for color mode (mutually exclusive, processed manually)
	var noColor bool
	var forceColor bool

	rootCmd := &cobra.Command{
		Use:   "make-help",
		Short: "Dynamic help generation for Makefiles",
		Long: `make-help scans Makefile documentation and generates formatted help output.

It supports special directives:
  @file         - File-level documentation
  @category     - Group targets into categories
  @var          - Document environment variables
  @alias        - Define target aliases

Documentation lines start with ## and are associated with the next target.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Process color flags
			if err := processColorFlags(&config.ColorMode, noColor, forceColor); err != nil {
				return err
			}

			// Resolve color mode
			config.UseColor = ResolveColorMode(config)

			// Run the help command
			return runHelp(config)
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&config.MakefilePath,
		"makefile-path", "", "Path to Makefile (defaults to ./Makefile)")
	rootCmd.PersistentFlags().BoolVar(&noColor,
		"no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().BoolVar(&forceColor,
		"color", false, "Force colored output")
	rootCmd.PersistentFlags().BoolVarP(&config.Verbose,
		"verbose", "v", false, "Enable verbose output for debugging")

	// Help generation flags (only on root command, not subcommands)
	rootCmd.Flags().BoolVar(&config.KeepOrderCategories,
		"keep-order-categories", false, "Preserve category discovery order")
	rootCmd.Flags().BoolVar(&config.KeepOrderTargets,
		"keep-order-targets", false, "Preserve target discovery order within categories")

	// --keep-order-all is a convenience flag that sets both
	var keepOrderAll bool
	rootCmd.Flags().BoolVar(&keepOrderAll,
		"keep-order-all", false, "Preserve both category and target discovery order")

	rootCmd.Flags().StringSliceVar(&config.CategoryOrder,
		"category-order", []string{}, "Explicit category order (comma-separated)")
	rootCmd.Flags().StringVar(&config.DefaultCategory,
		"default-category", "", "Default category for uncategorized targets")

	// Process --keep-order-all flag
	rootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if keepOrderAll {
			config.KeepOrderCategories = true
			config.KeepOrderTargets = true
		}
		return nil
	}

	// Add subcommands
	rootCmd.AddCommand(newAddTargetCmd(config))
	rootCmd.AddCommand(newRemoveTargetCmd(config))

	return rootCmd
}

// processColorFlags validates and processes color mode flags.
func processColorFlags(mode *ColorMode, noColor, forceColor bool) error {
	if noColor && forceColor {
		return fmt.Errorf("cannot use both --color and --no-color flags")
	}

	if forceColor {
		*mode = ColorAlways
	} else if noColor {
		*mode = ColorNever
	} else {
		*mode = ColorAuto
	}

	return nil
}

// parseCategoryOrder parses a comma-separated list of categories.
// This handles both comma-separated values from the flag and shell-provided lists.
func parseCategoryOrder(input []string) []string {
	var result []string
	for _, item := range input {
		// Split on commas and trim whitespace
		parts := strings.Split(item, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
	}
	return result
}
