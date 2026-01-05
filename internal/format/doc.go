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
package format
