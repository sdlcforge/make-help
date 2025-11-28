// Package discovery handles finding Makefiles and extracting targets.
//
// It uses Make's MAKEFILE_LIST variable to discover all included files and
// the `make -p` database output to enumerate available targets. All external
// command execution uses context with timeout to prevent indefinite hangs.
//
// # File Discovery
//
// The DiscoverMakefiles function finds all Makefiles by:
//  1. Creating a temporary file with the main Makefile content plus a
//     _list_makefiles target that echoes $(MAKEFILE_LIST)
//  2. Executing make -f <temp> _list_makefiles
//  3. Parsing the space-separated output
//
// Security note: This package uses temporary physical files instead of bash
// process substitution to prevent command injection vulnerabilities.
//
// # Target Discovery
//
// The DiscoverTargets function extracts target names by:
//  1. Running make -p -r to get the make database
//  2. Parsing lines matching the pattern ^<name>:
//  3. Filtering out comments and recipe lines
//
// # Timeouts
//
// All make command executions use a 30-second timeout to prevent
// indefinite hangs on malformed or pathological Makefiles.
package discovery
