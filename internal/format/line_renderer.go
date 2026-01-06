package format

import "github.com/sdlcforge/make-help/internal/model"

// LineRenderer is an interface for formatters that can render help content
// as individual lines suitable for embedding in generated files (e.g., Makefiles).
// This abstraction allows the generator to work with any formatter that supports
// line-based rendering without depending on concrete formatter implementations.
type LineRenderer interface {
	// RenderHelpLines generates help output lines suitable for embedding.
	// Returns a slice of strings, each representing one line to be output.
	// Each line is properly escaped for the target format context.
	RenderHelpLines(helpModel *model.HelpModel) ([]string, error)

	// RenderDetailedTargetLines renders detailed help for a single target.
	// Returns a slice of strings, each representing one line to be output.
	// Each line is properly escaped for the target format context.
	RenderDetailedTargetLines(target *model.Target) []string
}
