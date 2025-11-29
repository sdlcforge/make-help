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

	// Add-target options

	// HelpFilePath specifies explicit path for the generated help target file.
	// If empty, location is determined automatically.
	HelpFilePath string

	// CreateHelpTarget indicates whether to generate help target file.
	CreateHelpTarget bool

	// RemoveHelpTarget indicates whether to remove help target from Makefile.
	RemoveHelpTarget bool

	// Version for go install (e.g., "v1.2.3"), empty = @latest.
	Version string

	// IncludeTargets lists undocumented targets to include in help.
	// Populated from --include-target flag (repeatable, comma-separated).
	IncludeTargets []string

	// IncludeAllPhony includes all .PHONY targets in help output.
	IncludeAllPhony bool

	// Target specifies a target name for detailed help view.
	Target string

	// Derived state (computed at runtime)

	// UseColor is the resolved color setting based on ColorMode and terminal detection.
	UseColor bool
}

// NewConfig creates a new Config with default values.
func NewConfig() *Config {
	return &Config{
		ColorMode:     ColorAuto,
		CategoryOrder: []string{},
	}
}
