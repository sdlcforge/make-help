package lint

// Check represents a lint check with optional auto-fix capability.
type Check struct {
	// Name is a unique identifier for the check (e.g., "summary-punctuation").
	Name string

	// CheckFunc performs the check and returns any warnings found.
	CheckFunc CheckFunc

	// FixFunc generates a fix for a warning. May be nil if the check is not auto-fixable.
	// Returns nil if the specific warning instance cannot be fixed.
	FixFunc FixFunc
}

// FixFunc generates a fix for a warning.
// Returns nil if the warning cannot be auto-fixed.
type FixFunc func(w Warning) *Fix

// Fix represents a single file modification to fix a lint warning.
type Fix struct {
	// File is the absolute path to the file to modify.
	File string

	// Line is the 1-indexed line number to modify.
	Line int

	// Operation specifies the type of modification.
	Operation FixOperation

	// OldContent is the expected current content of the line (for validation).
	// Used to detect if the file has changed since the check was run.
	OldContent string

	// NewContent is the replacement content (for FixReplace operation).
	NewContent string
}

// FixOperation specifies the type of file modification.
type FixOperation int

const (
	// FixReplace replaces the entire line with new content.
	FixReplace FixOperation = iota

	// FixDelete removes the line entirely.
	FixDelete
)
