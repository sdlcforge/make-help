package format

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/sdlcforge/make-help/internal/model"
)

// Renderer is the interface for generating formatted output.
// It handles the actual rendering of help content in a specific format.
type Renderer interface {
	// RenderHelp generates the complete help output from a HelpModel.
	// This is the summary view showing all categories and targets.
	RenderHelp(model *model.HelpModel, w io.Writer) error

	// RenderDetailedTarget generates detailed help for a single target.
	// Shows full documentation, variables, aliases, and source location.
	RenderDetailedTarget(target *model.Target, w io.Writer) error

	// RenderBasicTarget generates minimal help for an undocumented target.
	// Shows target name and source location if available.
	RenderBasicTarget(name string, sourceFile string, lineNumber int, w io.Writer) error
}

// FormatMetadata provides information about a format's properties.
// This includes content type and file extension for output purposes.
type FormatMetadata interface {
	// ContentType returns the MIME type for this format (for HTTP responses, future use).
	ContentType() string

	// DefaultExtension returns the default file extension for this format.
	DefaultExtension() string
}

// Formatter is the interface that all output format implementations must satisfy.
// Each formatter knows how to render a HelpModel in its specific format.
// This interface combines rendering capabilities with format metadata.
// It is kept for backward compatibility and to provide a complete formatter contract.
type Formatter interface {
	Renderer
	FormatMetadata
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

	// MakefileDir is the directory containing the main Makefile.
	// Used to convert absolute paths to relative paths in Source: lines.
	// If empty, absolute paths are used.
	MakefileDir string
}

// Validate checks that the FormatterConfig is valid.
// Returns an error if the configuration is invalid.
// Note: UseColor=true with ColorScheme=nil is valid - formatters will create
// default color schemes. This method is provided for future validation needs.
func (c *FormatterConfig) Validate() error {
	// Note: We do NOT validate UseColor + nil ColorScheme because all formatters
	// already handle this case gracefully by creating default color schemes.
	// This maintains backward compatibility with existing code.
	return nil
}

// normalizeConfig returns a non-nil config with defaults applied.
// If the provided config is nil, returns a default config with UseColor=false.
// If the provided config is non-nil, returns it unchanged.
func normalizeConfig(config *FormatterConfig) *FormatterConfig {
	if config == nil {
		return &FormatterConfig{UseColor: false}
	}
	return config
}

// makeRelativePath converts an absolute path to a path relative to the Makefile directory.
// If makefileDir is empty or the path cannot be made relative, returns the original path.
func makeRelativePath(absolutePath, makefileDir string) string {
	if makefileDir == "" {
		return absolutePath
	}

	relPath, err := filepath.Rel(makefileDir, absolutePath)
	if err != nil {
		// If we can't make it relative, return the original path
		return absolutePath
	}

	return relPath
}

// NewFormatter creates a formatter for the specified format type.
// This is the factory function that replaces direct renderer construction.
// Supported format types: "make", "mk", "text", "txt", "html", "markdown", "md", "json"
func NewFormatter(formatType string, config *FormatterConfig) (Formatter, error) {
	// Validate config if provided
	if config != nil {
		if err := config.Validate(); err != nil {
			return nil, err
		}
	}

	switch formatType {
	case "make", "mk":
		return NewMakeFormatter(config), nil
	case "text", "txt":
		return NewTextFormatter(config), nil
	case "html":
		return NewHTMLFormatter(config), nil
	case "markdown", "md":
		return NewMarkdownFormatter(config), nil
	case "json":
		return NewJSONFormatter(config), nil
	default:
		return nil, fmt.Errorf("unknown format type: %s (supported: make, text, html, markdown, json)", formatType)
	}
}
