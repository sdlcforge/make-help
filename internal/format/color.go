package format

// ANSI color codes
const (
	reset         = "\033[0m"
	boldCyan      = "\033[1;36m"
	boldGreen     = "\033[1;32m"
	yellow        = "\033[0;33m"
	magenta       = "\033[0;35m"
	white         = "\033[0;37m"
)

// ColorScheme defines ANSI color codes for different help output elements.
// When colors are disabled, all fields are empty strings.
type ColorScheme struct {
	// CategoryName colors category headers
	CategoryName string

	// TargetName colors target names
	TargetName string

	// Alias colors target aliases
	Alias string

	// Variable colors environment variable names
	Variable string

	// Documentation colors documentation text
	Documentation string

	// Reset resets color to default
	Reset string
}

// NewColorScheme creates a ColorScheme with colors enabled or disabled.
// When useColor is true, ANSI color codes are assigned to each field.
// When useColor is false, all fields are empty strings (no colors).
func NewColorScheme(useColor bool) *ColorScheme {
	if !useColor {
		return &ColorScheme{}
	}

	return &ColorScheme{
		CategoryName:  boldCyan,
		TargetName:    boldGreen,
		Alias:         yellow,
		Variable:      magenta,
		Documentation: white,
		Reset:         reset,
	}
}
