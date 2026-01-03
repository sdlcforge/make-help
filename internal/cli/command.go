package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// setupFlags configures flags on a Cobra command and binds them to a Config.
// This allows reusing the same flag setup logic for both CLI parsing and
// command line restoration from help.mk files.
func setupFlags(cmd *cobra.Command, config *Config) {
	// Flags for color mode (mutually exclusive, processed manually)
	var noColor bool
	var forceColor bool

	// --keep-order-all is a convenience flag that sets both order flags
	var keepOrderAll bool

	// Mode flags
	cmd.Flags().BoolVar(&config.RemoveHelpTarget,
		"remove-help", false, "Remove help target from Makefile")
	cmd.Flags().BoolVar(&config.DryRun,
		"dry-run", false, "Show what files would be created/modified without making changes")
	cmd.Flags().BoolVar(&config.Lint,
		"lint", false, "Check documentation quality and report issues")
	cmd.Flags().BoolVar(&config.Fix,
		"fix", false, "Automatically fix auto-fixable lint issues (requires --lint)")
	cmd.Flags().StringVar(&config.Target,
		"target", "", "Show detailed help for a specific target (requires --output -)")

	// Input flags
	cmd.PersistentFlags().StringVar(&config.MakefilePath,
		"makefile-path", "", "Path to Makefile (defaults to ./Makefile)")
	cmd.Flags().StringVar(&config.HelpFileRelPath,
		"help-file-rel-path", "", "Relative path for generated help target file (e.g., help.mk or make/help.mk)")

	// Output/formatting flags
	cmd.Flags().StringVar(&config.Format,
		"format", "make", "Output format (make, text, html, markdown)")
	cmd.Flags().StringVar(&config.Output,
		"output", "", "Output destination (file path or - for stdout). Default depends on format.")
	// Note: Color flags are bound to local variables, not config directly,
	// because they need special processing (mutually exclusive)
	cmd.PersistentFlags().BoolVar(&forceColor,
		"color", false, "Force colored output")
	cmd.PersistentFlags().BoolVar(&noColor,
		"no-color", false, "Disable colored output")
	cmd.Flags().StringSliceVar(&config.IncludeTargets,
		"include-target", []string{}, "Include undocumented target in help (repeatable, comma-separated)")
	cmd.Flags().BoolVar(&config.IncludeAllPhony,
		"include-all-phony", false, "Include all .PHONY targets in help output")
	cmd.Flags().BoolVar(&config.KeepOrderCategories,
		"keep-order-categories", false, "Preserve category discovery order")
	cmd.Flags().BoolVar(&config.KeepOrderTargets,
		"keep-order-targets", false, "Preserve target discovery order within categories")
	cmd.Flags().BoolVar(&config.KeepOrderFiles,
		"keep-order-files", false, "Preserve file discovery order")
	// Note: keep-order-all is bound to local variable for special processing
	cmd.Flags().BoolVar(&keepOrderAll,
		"keep-order-all", false, "Preserve category, target, and file discovery order")
	cmd.Flags().StringSliceVar(&config.CategoryOrder,
		"category-order", []string{}, "Explicit category order (comma-separated)")
	cmd.Flags().StringVar(&config.DefaultCategory,
		"default-category", "", "Default category for uncategorized targets")
	cmd.Flags().StringVar(&config.HelpCategory,
		"help-category", "Help", "Category name for generated help targets (help, update-help)")

	// Misc flags
	cmd.PersistentFlags().BoolVarP(&config.Verbose,
		"verbose", "v", false, "Enable verbose output for debugging")

}

// processFlagsAfterParse processes flags that need special handling after Cobra parsing.
func processFlagsAfterParse(cmd *cobra.Command, config *Config) error {
	// Process color flags
	noColor := cmd.Flags().Lookup("no-color").Changed
	forceColor := cmd.Flags().Lookup("color").Changed

	if noColor && forceColor {
		return fmt.Errorf("cannot use both --color and --no-color flags")
	}

	if forceColor {
		config.ColorMode = ColorAlways
	} else if noColor {
		config.ColorMode = ColorNever
	} else {
		config.ColorMode = ColorAuto
	}

	// Process --keep-order-all flag
	if cmd.Flags().Lookup("keep-order-all").Changed {
		config.KeepOrderCategories = true
		config.KeepOrderTargets = true
		config.KeepOrderFiles = true
	}

	// Normalize IncludeTargets from comma-separated + repeatable flags
	config.IncludeTargets = parseIncludeTargets(config.IncludeTargets)

	return nil
}

// ParseCommandLineFromHelpFile parses a command line string from a help.mk file
// and applies only allowed options to a Config. Only output/formatting, input,
// and verbose flags are allowed. Mode flags would cause an error.
func ParseCommandLineFromHelpFile(cmdLine string, config *Config) error {
	if cmdLine == "" {
		return nil
	}

	// Remove leading "make-help" if present
	cmdLine = strings.TrimSpace(cmdLine)
	if strings.HasPrefix(cmdLine, "make-help") {
		cmdLine = strings.TrimSpace(cmdLine[len("make-help"):])
	}

	// Split into arguments
	args := strings.Fields(cmdLine)
	if len(args) == 0 {
		return nil
	}

	// Create a temporary command to parse flags
	cmd := &cobra.Command{
		Use: "make-help",
		// Disable running the command - we only want flag parsing
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	// Set up flags on the command
	setupFlags(cmd, config)

	// Set the arguments and parse
	cmd.SetArgs(args)

	// Check for disallowed mode flags before parsing
	disallowedFlags := []string{"--remove-help", "--dry-run", "--lint", "--fix", "--target"}
	for _, arg := range args {
		for _, disallowed := range disallowedFlags {
			if arg == disallowed || strings.HasPrefix(arg, disallowed+"=") {
				return fmt.Errorf("mode flag %s is not allowed when restoring options from help.mk file", disallowed)
			}
		}
	}

	// Parse flags (this will populate config via the flag bindings)
	if err := cmd.Execute(); err != nil {
		return fmt.Errorf("failed to parse command line: %w", err)
	}

	// Process flags that need special handling
	if err := processFlagsAfterParse(cmd, config); err != nil {
		return err
	}

	return nil
}
