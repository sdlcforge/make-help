// Package format provides formatters for rendering HelpModel in various output formats.
//
// The package implements a Formatter interface that supports multiple output formats:
//   - Make format: Generates Makefile content with embedded help targets
//   - Text format: Plain text or ANSI-colored terminal output
//   - HTML format: Browser-ready HTML with embedded styles
//   - Markdown format: GitHub-flavored markdown documentation
//   - JSON format: Structured JSON for programmatic consumption
//
// # Architecture
//
// Each format is implemented by a dedicated formatter type:
//   - MakeFormatter: Generates @printf statements for Makefile inclusion
//   - TextFormatter: Produces human-readable text output
//   - HTMLFormatter: Creates styled HTML pages
//   - MarkdownFormatter: Outputs structured markdown documentation
//   - JSONFormatter: Produces structured JSON output
//
// All formatters implement the Formatter interface:
//
//	type Formatter interface {
//	    RenderHelp(model *HelpModel, w io.Writer) error
//	    RenderDetailedTarget(target *Target, w io.Writer) error
//	    RenderBasicTarget(name, sourceFile string, lineNumber int, w io.Writer) error
//	    ContentType() string
//	    DefaultExtension() string
//	}
//
// For formatters that support line-based rendering (e.g., for embedding in generated files),
// there is also a LineRenderer interface:
//
//	type LineRenderer interface {
//	    RenderHelpLines(model *HelpModel) ([]string, error)
//	    RenderDetailedTargetLines(target *Target) []string
//	}
//
// MakeFormatter implements both Formatter and LineRenderer. The LineRenderer interface
// is used by the generator package to embed help text in generated Makefile targets
// without depending on concrete formatter implementations.
//
// # Color Support
//
// Text and Make formatters support ANSI color output, controlled via FormatterConfig.
// Colors can be forced on/off via --color and --no-color flags. The default color scheme:
//   - Bold Cyan for category names
//   - Bold Green for target names
//   - Yellow for aliases
//   - Magenta for variable names
//   - White for documentation text
//
// # Standard Output Format
//
// The basic help output structure (used by Text and Make formatters):
//
//	Usage: make [<target>...] [<ENV_VAR>=<value>...]
//
//	<!file documentation>
//
//	Targets:
//
//	[Category Name:]
//	  - <target>[ <alias1>, ...]: <summary>
//	    [Vars: <VAR1>, <VAR2>...]
//
// # Rich Text Rendering
//
// Different formatters handle rich text (markdown formatting in documentation)
// differently based on their output context:
//
//   - HTMLFormatter: Preserves markdown formatting by converting it to HTML
//     elements (bold becomes <strong>, italic becomes <em>, etc.)
//   - MarkdownFormatter: Preserves original markdown formatting unchanged
//     (e.g., **bold**, *italic*, `code`)
//   - TextFormatter: Strips markdown to plain text for clean terminal output
//   - MakeFormatter: Strips markdown to plain text for Makefile embedding
//   - JSONFormatter: Strips markdown to plain text for programmatic consumers
//
// This design is intentional: terminal and file-based outputs benefit from
// plain text readability, while web and document outputs benefit from rich
// formatting. The richtext package provides both PlainText() and Markdown()
// methods, allowing formatters to choose the appropriate representation for
// their output context.
package format
