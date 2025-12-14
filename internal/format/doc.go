// Package format renders HelpModel as formatted text output.
//
// It supports colorized output using ANSI escape codes, with automatic
// detection of terminal capabilities. Colors can be forced on or off
// via --color and --no-color flags respectively.
//
// # Output Format
//
// The standard help output format is:
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
// # Color Scheme
//
// The default color scheme uses:
//   - Bold Cyan for category names
//   - Bold Green for target names
//   - Yellow for aliases
//   - Magenta for variable names
//   - White for documentation text
//
// # Rendering
//
// The Renderer type uses a template-like approach with string builders
// for efficient concatenation. All color codes are injected conditionally
// based on the UseColor configuration.
package format
