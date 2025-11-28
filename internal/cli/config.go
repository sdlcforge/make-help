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

	// TargetFile specifies explicit path for the generated help target file.
	// If empty, location is determined automatically.
	TargetFile string

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
