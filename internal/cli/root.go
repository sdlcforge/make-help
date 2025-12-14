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

Default behavior generates help file. Use flags for other operations:
  --show-help           Display help dynamically
  --target <name>       Show detailed help for a target (requires --show-help)
  --remove-help         Remove help targets

Documentation directives (in ## comments):
  !file         File-level documentation
  !category     Group targets into categories
  !var          Document environment variables
  !alias        Define target aliases`,
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

			// --remove-help only allows --verbose and --makefile-path (check this first)
			if config.RemoveHelpTarget {
				if err := validateRemoveHelpFlags(config); err != nil {
					return err
				}
			}

			// --target only valid with --show-help
			if config.Target != "" && !config.ShowHelp {
				return fmt.Errorf("--target can only be used with --show-help")
			}

			// --dry-run cannot be used with --show-help
			if config.DryRun && config.ShowHelp {
				return fmt.Errorf("--dry-run cannot be used with --show-help")
			}

			// Validate --help-file-rel-path is a relative path (no leading /)
			if config.HelpFileRelPath != "" && strings.HasPrefix(config.HelpFileRelPath, "/") {
				return fmt.Errorf("--help-file-rel-path must be a relative path (no leading '/')")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Normalize IncludeTargets from comma-separated + repeatable flags
			config.IncludeTargets = parseIncludeTargets(config.IncludeTargets)

			// Resolve color mode
			config.UseColor = ResolveColorMode(config)

			// Dispatch to appropriate handler
			if config.ShowHelp {
				if config.Target != "" {
					return runDetailedHelp(config)
				}
				return runHelp(config)
			} else if config.RemoveHelpTarget {
				return runRemoveHelpTarget(config)
			} else {
				// Default behavior: generate help file
				return runCreateHelpTarget(config)
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
	rootCmd.Flags().BoolVar(&config.ShowHelp,
		"show-help", false, "Display help dynamically instead of generating a help file")
	rootCmd.Flags().BoolVar(&config.RemoveHelpTarget,
		"remove-help", false, "Remove help target from Makefile")
	rootCmd.Flags().StringSliceVar(&config.IncludeTargets,
		"include-target", []string{}, "Include undocumented target in help (repeatable, comma-separated)")
	rootCmd.Flags().BoolVar(&config.IncludeAllPhony,
		"include-all-phony", false, "Include all .PHONY targets in help output")
	rootCmd.Flags().StringVar(&config.Target,
		"target", "", "Show detailed help for a specific target (requires --show-help)")
	rootCmd.Flags().StringVar(&config.HelpFileRelPath,
		"help-file-rel-path", "", "Relative path for generated help target file (e.g., help.mk or make/help.mk)")
	rootCmd.Flags().BoolVar(&config.DryRun,
		"dry-run", false, "Show what files would be created/modified without making changes")

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

// validateRemoveHelpFlags checks for incompatible flags with --remove-help.
// It uses a table-driven approach to provide specific error messages for each incompatible flag.
func validateRemoveHelpFlags(config *Config) error {
	// Table of incompatible flags: condition check, flag name
	incompatibleFlags := []struct {
		isSet    bool
		flagName string
	}{
		{config.Target != "", "--target"},
		{len(config.IncludeTargets) > 0, "--include-target"},
		{config.IncludeAllPhony, "--include-all-phony"},
		{config.ShowHelp, "--show-help"},
		{config.DryRun, "--dry-run"},
		{config.HelpFileRelPath != "", "--help-file-rel-path"},
		{config.KeepOrderCategories, "--keep-order-categories"},
		{config.KeepOrderTargets, "--keep-order-targets"},
		{len(config.CategoryOrder) > 0, "--category-order"},
		{config.DefaultCategory != "", "--default-category"},
	}

	for _, flag := range incompatibleFlags {
		if flag.isSet {
			return fmt.Errorf("--remove-help cannot be used with %s", flag.flagName)
		}
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
