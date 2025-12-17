// Package cli provides the command-line interface for make-help using Cobra.
//
// This package handles argument parsing, flag validation, terminal detection,
// and delegates to the appropriate service packages for actual functionality.
// It is the only package that interacts with os.Args and stdout/stderr.
//
// # Commands
//
// The CLI provides three commands:
//   - make-help (default): Generate help output from Makefile documentation
//   - make-help add-target: Add a help target to the Makefile
//   - make-help remove-target: Remove help target artifacts
//
// # Color Detection
//
// Color output is automatically enabled when stdout is a terminal.
// This can be overridden with --color (force on) or --no-color (force off).
// When output is piped, colors are disabled by default.
//
// # Configuration
//
// The Config struct holds all CLI configuration and is passed to
// service packages. It includes both user-provided flags and derived
// state computed at runtime (e.g., UseColor).
package cli
