package cli

// ColorMode represents the color output mode for the CLI.
type ColorMode int

const (
	// ColorAuto enables color output when connected to a terminal.
	ColorAuto ColorMode = iota

	// ColorAlways forces color output regardless of terminal detection.
	ColorAlways

	// ColorNever disables color output.
	ColorNever
)

// String returns the string representation of ColorMode.
func (c ColorMode) String() string {
	switch c {
	case ColorAuto:
		return "auto"
	case ColorAlways:
		return "always"
	case ColorNever:
		return "never"
	default:
		return "unknown"
	}
}

// Config holds all CLI configuration options.
type Config struct {
	// Global options

	// MakefilePath is the path to the main Makefile (resolved to absolute path).
	// If empty, defaults to "Makefile" in the current working directory.
	MakefilePath string

	// ColorMode determines when to use colored output.
	ColorMode ColorMode

	// Verbose enables verbose output for debugging file discovery and parsing.
	Verbose bool

	// Help generation options

	// KeepOrderCategories preserves category discovery order instead of alphabetical.
	KeepOrderCategories bool

	// KeepOrderTargets preserves target discovery order within categories.
	KeepOrderTargets bool

	// CategoryOrder specifies explicit category ordering.
	// Categories not in this list are appended alphabetically.
	CategoryOrder []string

	// DefaultCategory is the category name for uncategorized targets.
	// Required when mixing categorized and uncategorized targets.
	DefaultCategory string

	// HelpCategory is the category name for generated help targets (help, update-help).
	// Defaults to "Help" if not specified.
	HelpCategory string

	// Add-target options

	// HelpFileRelPath specifies a relative path for the generated help target file.
	// Must be a relative path (no leading '/'). If empty, location is determined automatically.
	HelpFileRelPath string

	// ShowHelp displays help dynamically instead of generating a help file.
	ShowHelp bool

	// RemoveHelpTarget indicates whether to remove help target from Makefile.
	RemoveHelpTarget bool

	// IncludeTargets lists undocumented targets to include in help.
	// Populated from --include-target flag (repeatable, comma-separated).
	IncludeTargets []string

	// IncludeAllPhony includes all .PHONY targets in help output.
	IncludeAllPhony bool

	// Target specifies a target name for detailed help view.
	Target string

	// DryRun shows what would be created/modified without actually making changes.
	// Valid with CreateHelpTarget or --lint --fix.
	DryRun bool

	// Lint enables lint mode to check documentation quality.
	Lint bool

	// Fix automatically fixes auto-fixable lint issues.
	// Only valid with --lint.
	Fix bool

	// Derived state (computed at runtime)

	// UseColor is the resolved color setting based on ColorMode and terminal detection.
	UseColor bool

	// CommandLine stores the raw command line to be recorded in generated help files.
	// Captured from os.Args in PreRunE.
	CommandLine string
}

// NewConfig creates a new Config with default values.
func NewConfig() *Config {
	return &Config{
		ColorMode:     ColorAuto,
		CategoryOrder: []string{},
		HelpCategory:  "Help",
	}
}
