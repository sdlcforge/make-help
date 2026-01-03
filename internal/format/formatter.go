package format

import (
	"fmt"
	"io"

	"github.com/sdlcforge/make-help/internal/model"
)

// Formatter is the interface that all output format implementations must satisfy.
// Each formatter knows how to render a HelpModel in its specific format.
type Formatter interface {
	// RenderHelp generates the complete help output from a HelpModel.
	// This is the summary view showing all categories and targets.
	RenderHelp(model *model.HelpModel, w io.Writer) error

	// RenderDetailedTarget generates detailed help for a single target.
	// Shows full documentation, variables, aliases, and source location.
	RenderDetailedTarget(target *model.Target, w io.Writer) error

	// RenderBasicTarget generates minimal help for an undocumented target.
	// Shows target name and source location if available.
	RenderBasicTarget(name string, sourceFile string, lineNumber int, w io.Writer) error

	// ContentType returns the MIME type for this format (for HTTP responses, future use).
	ContentType() string

	// DefaultExtension returns the default file extension for this format.
	DefaultExtension() string
}

// FormatterConfig holds configuration options common to all formatters.
type FormatterConfig struct {
	// UseColor enables colored/styled output (where applicable).
	// For terminal formats: ANSI escape codes
	// For HTML: CSS classes
	// For Markdown/Make: no effect (or minimal effect)
	UseColor bool

	// ColorScheme defines colors for different elements (terminal formats only).
	// When UseColor is false, this is nil.
	ColorScheme *ColorScheme
}

// NewFormatter creates a formatter for the specified format type.
// This is the factory function that replaces direct renderer construction.
// Supported format types: "make", "mk", "text", "txt", "html", "markdown", "md"
func NewFormatter(formatType string, config *FormatterConfig) (Formatter, error) {
	switch formatType {
	case "make", "mk":
		return NewMakeFormatter(config), nil
	case "text", "txt":
		return NewTextFormatter(config), nil
	case "html":
		return NewHTMLFormatter(config), nil
	case "markdown", "md":
		return NewMarkdownFormatter(config), nil
	default:
		return nil, fmt.Errorf("unknown format type: %s (supported: make, text, html, markdown)", formatType)
	}
}
