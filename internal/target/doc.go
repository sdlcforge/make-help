// Package target handles adding and removing help targets from Makefiles.
//
// # Add-Target Strategy
//
// The add-target command uses a three-tier strategy to determine where
// to place the help target:
//  1. Explicit --target-file path: Create file at specified path and add
//     include directive to main Makefile
//  2. make/01-help.mk if include make/*.mk pattern exists: Create the file
//     in the existing make/ directory structure
//  3. Append directly to main Makefile: Add help target at the end of file
//
// # File Safety
//
// All file modifications use atomic writes (write to temp, then rename)
// to prevent file corruption on process crashes. The sequence is:
//  1. Create temporary file in same directory
//  2. Write content to temp file
//  3. Sync to disk
//  4. Atomic rename to target path
//
// # Validation
//
// Makefiles are validated with `make -n` before modification to catch
// syntax errors early. This prevents corrupting a working Makefile due
// to invalid help target content.
//
// # Generated Content
//
// The generated help target includes all relevant flags from the config,
// ensuring that running `make help` produces the same output as the
// configuration used when add-target was run.
package target
