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

	// --keep-order-all is a convenience flag that sets both order flags
	var keepOrderAll bool

	rootCmd := &cobra.Command{
		Use:   "make-help",
		Short: "Dynamic help generation for Makefiles",
		Long: `make-help generates formatted help output from Makefile documentation.

Default behavior displays help. Use flags for other operations:
  --target <name>       Show detailed help for a target
  --create-help-target  Generate help target file
  --remove-help-target  Remove help targets

Documentation directives (in ## comments):
  @file         File-level documentation
  @category     Group targets into categories
  @var          Document environment variables
  @alias        Define target aliases`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Process color flags
			if err := processColorFlags(&config.ColorMode, noColor, forceColor); err != nil {
				return err
			}

			// Process --keep-order-all flag
			if keepOrderAll {
				config.KeepOrderCategories = true
				config.KeepOrderTargets = true
			}

			// Mutual exclusivity validation
			if config.CreateHelpTarget && config.RemoveHelpTarget {
				return fmt.Errorf("cannot use both --create-help-target and --remove-help-target")
			}

			// --remove-help-target only allows --verbose and --makefile-path
			if config.RemoveHelpTarget {
				// Check that no other flags are set
				if config.Target != "" {
					return fmt.Errorf("--remove-help-target only accepts --verbose and --makefile-path flags")
				}
				if len(config.IncludeTargets) > 0 {
					return fmt.Errorf("--remove-help-target only accepts --verbose and --makefile-path flags")
				}
				if config.IncludeAllPhony {
					return fmt.Errorf("--remove-help-target only accepts --verbose and --makefile-path flags")
				}
				if config.CreateHelpTarget {
					return fmt.Errorf("--remove-help-target only accepts --verbose and --makefile-path flags")
				}
				if config.Version != "" {
					return fmt.Errorf("--remove-help-target only accepts --verbose and --makefile-path flags")
				}
				if config.HelpFilePath != "" {
					return fmt.Errorf("--remove-help-target only accepts --verbose and --makefile-path flags")
				}
				if config.KeepOrderCategories {
					return fmt.Errorf("--remove-help-target only accepts --verbose and --makefile-path flags")
				}
				if config.KeepOrderTargets {
					return fmt.Errorf("--remove-help-target only accepts --verbose and --makefile-path flags")
				}
				if len(config.CategoryOrder) > 0 {
					return fmt.Errorf("--remove-help-target only accepts --verbose and --makefile-path flags")
				}
				if config.DefaultCategory != "" {
					return fmt.Errorf("--remove-help-target only accepts --verbose and --makefile-path flags")
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Normalize IncludeTargets from comma-separated + repeatable flags
			config.IncludeTargets = parseIncludeTargets(config.IncludeTargets)

			// Resolve color mode
			config.UseColor = ResolveColorMode(config)

			// Dispatch to appropriate handler
			if config.RemoveHelpTarget {
				return runRemoveHelpTarget(config)
			} else if config.CreateHelpTarget {
				return runCreateHelpTarget(config)
			} else if config.Target != "" {
				return runDetailedHelp(config)
			} else {
				// Default behavior: run help command
				return runHelp(config)
			}
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
	rootCmd.Flags().BoolVar(&keepOrderAll,
		"keep-order-all", false, "Preserve both category and target discovery order")

	rootCmd.Flags().StringSliceVar(&config.CategoryOrder,
		"category-order", []string{}, "Explicit category order (comma-separated)")
	rootCmd.Flags().StringVar(&config.DefaultCategory,
		"default-category", "", "Default category for uncategorized targets")

	// New flags for target creation/removal and filtering
	rootCmd.Flags().BoolVar(&config.CreateHelpTarget,
		"create-help-target", false, "Generate help target file with local binary installation")
	rootCmd.Flags().BoolVar(&config.RemoveHelpTarget,
		"remove-help-target", false, "Remove help target from Makefile")
	rootCmd.Flags().StringVar(&config.Version,
		"version", "", "Version to pin in generated go install (e.g., v1.2.3)")
	rootCmd.Flags().StringSliceVar(&config.IncludeTargets,
		"include-target", []string{}, "Include undocumented target in help (repeatable, comma-separated)")
	rootCmd.Flags().BoolVar(&config.IncludeAllPhony,
		"include-all-phony", false, "Include all .PHONY targets in help output")
	rootCmd.Flags().StringVar(&config.Target,
		"target", "", "Show detailed help for a specific target")
	rootCmd.Flags().StringVar(&config.HelpFilePath,
		"help-file-path", "", "Explicit path for generated help target file")

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
