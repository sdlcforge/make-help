package cli

import (
	"fmt"
	"strings"

	"github.com/sdlcforge/make-help/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	modeGroupLabel   = "Mode"
	inputGroupLabel  = "Input"
	outputGroupLabel = "Output/formatting"
	miscGroupLabel   = "Misc"
)

func init() {
	// Register custom template function for flag grouping
	cobra.AddTemplateFunc("flagGroups", flagGroupsFunc)
}

// NewRootCmd creates the root command for make-help.
// The default action is to run the help command.
func NewRootCmd() *cobra.Command {
	config := NewConfig()

	rootCmd := &cobra.Command{
		Use:     "make-help",
		Short:   "Dynamic help generation for Makefiles",
		Version: version.Version,
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
			// Process flags that need special handling
			if err := processFlagsAfterParse(cmd, config); err != nil {
				return err
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

			// --lint validations
			if config.Lint {
				if config.ShowHelp {
					return fmt.Errorf("--lint cannot be used with --show-help")
				}
				if config.RemoveHelpTarget {
					return fmt.Errorf("--lint cannot be used with --remove-help")
				}
				if config.DryRun && !config.Fix {
					return fmt.Errorf("--dry-run with --lint requires --fix")
				}
			}

			// --fix requires --lint
			if config.Fix && !config.Lint {
				return fmt.Errorf("--fix requires --lint")
			}

			// Validate --help-file-rel-path is a relative path (no leading /)
			if config.HelpFileRelPath != "" && strings.HasPrefix(config.HelpFileRelPath, "/") {
				return fmt.Errorf("--help-file-rel-path must be a relative path (no leading '/')")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve color mode
			config.UseColor = ResolveColorMode(config)

			// Dispatch to appropriate handler
			if config.Lint {
				return runLint(config)
			} else if config.ShowHelp {
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

	// Set up flags using shared function
	setupFlags(rootCmd, config)

	// Annotate flags with their groups for custom help display
	annotateFlag(rootCmd, "show-help", modeGroupLabel)
	annotateFlag(rootCmd, "remove-help", modeGroupLabel)
	annotateFlag(rootCmd, "dry-run", modeGroupLabel)
	annotateFlag(rootCmd, "lint", modeGroupLabel)
	annotateFlag(rootCmd, "target", modeGroupLabel)

	annotateFlag(rootCmd, "makefile-path", inputGroupLabel)
	annotateFlag(rootCmd, "help-file-rel-path", inputGroupLabel)

	annotateFlag(rootCmd, "color", outputGroupLabel)
	annotateFlag(rootCmd, "no-color", outputGroupLabel)
	annotateFlag(rootCmd, "include-target", outputGroupLabel)
	annotateFlag(rootCmd, "include-all-phony", outputGroupLabel)
	annotateFlag(rootCmd, "keep-order-categories", outputGroupLabel)
	annotateFlag(rootCmd, "keep-order-targets", outputGroupLabel)
	annotateFlag(rootCmd, "keep-order-all", outputGroupLabel)
	annotateFlag(rootCmd, "category-order", outputGroupLabel)
	annotateFlag(rootCmd, "default-category", outputGroupLabel)
	annotateFlag(rootCmd, "help-category", outputGroupLabel)

	annotateFlag(rootCmd, "verbose", miscGroupLabel)

	// Set custom usage template
	rootCmd.SetUsageTemplate(usageTemplate)

	return rootCmd
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
		{config.Lint, "--lint"},
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

// annotateFlag adds a group annotation to a flag for custom help grouping.
func annotateFlag(cmd *cobra.Command, flagName, group string) {
	// Try local flags first
	flag := cmd.Flags().Lookup(flagName)
	// If not found, try persistent flags
	if flag == nil {
		flag = cmd.PersistentFlags().Lookup(flagName)
	}

	if flag != nil {
		if flag.Annotations == nil {
			flag.Annotations = make(map[string][]string)
		}
		flag.Annotations["group"] = []string{group}
	}
}

// usageTemplate is a custom template that groups flags by their annotations.
const usageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{flagGroups .}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

// flagGroupsFunc generates grouped flag output for the custom usage template.
func flagGroupsFunc(cmd *cobra.Command) string {
	// Define the order of groups
	groupOrder := []string{modeGroupLabel, inputGroupLabel, outputGroupLabel, miscGroupLabel}

	// Collect flags by group
	flagsByGroup := make(map[string][]string)
	seenFlags := make(map[string]bool)

	// Process both local and persistent flags
	processFlags := func(flags *pflag.FlagSet) {
		flags.VisitAll(func(flag *pflag.Flag) {
			if flag.Hidden {
				return
			}

			// Skip if we've already processed this flag
			if seenFlags[flag.Name] {
				return
			}
			seenFlags[flag.Name] = true

			group := miscGroupLabel // default group
			if flag.Annotations != nil {
				if groups, ok := flag.Annotations["group"]; ok && len(groups) > 0 {
					group = groups[0]
				}
			}

			// Format the flag usage
			usage := formatFlagUsage(flag)
			flagsByGroup[group] = append(flagsByGroup[group], usage)
		})
	}

	// Process local flags first, then persistent flags
	processFlags(cmd.Flags())
	processFlags(cmd.PersistentFlags())

	// Build output string with groups in order
	var sb strings.Builder
	for _, group := range groupOrder {
		flags, ok := flagsByGroup[group]
		if !ok || len(flags) == 0 {
			continue
		}

		sb.WriteString(group)
		sb.WriteString(":\n")
		for _, flagUsage := range flags {
			sb.WriteString(flagUsage)
		}
		sb.WriteString("\n")
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// formatFlagUsage formats a single flag for display in the help output.
func formatFlagUsage(flag *pflag.Flag) string {
	var sb strings.Builder

	// Start with shorthand if it exists
	if flag.Shorthand != "" && flag.ShorthandDeprecated == "" {
		sb.WriteString("  -")
		sb.WriteString(flag.Shorthand)
		sb.WriteString(", ")
	} else {
		sb.WriteString("      ")
	}

	// Add the long flag name
	sb.WriteString("--")
	sb.WriteString(flag.Name)

	// Add the type/value info if not a boolean
	if flag.Value.Type() != "bool" {
		sb.WriteString(" ")
		// Normalize type names for better readability
		typeName := flag.Value.Type()
		switch typeName {
		case "stringSlice":
			typeName = "strings"
		case "intSlice":
			typeName = "ints"
		}
		sb.WriteString(typeName)
	}

	// Pad to align descriptions (using 36 as a reasonable width)
	currentLen := sb.Len()
	paddingNeeded := 36 - currentLen
	if paddingNeeded > 0 {
		sb.WriteString(strings.Repeat(" ", paddingNeeded))
	} else {
		sb.WriteString("   ")
	}

	// Add the usage description
	sb.WriteString(flag.Usage)

	// Add default value if meaningful (not empty, false, [], etc.)
	if shouldShowDefault(flag) {
		sb.WriteString(fmt.Sprintf(" (default %s)", flag.DefValue))
	}

	sb.WriteString("\n")

	return sb.String()
}

// shouldShowDefault determines if a flag's default value should be displayed.
func shouldShowDefault(flag *pflag.Flag) bool {
	if flag.DefValue == "" {
		return false
	}
	if flag.Value.Type() == "bool" && flag.DefValue == "false" {
		return false
	}
	// Don't show empty slice/array defaults
	if flag.DefValue == "[]" {
		return false
	}
	return true
}
