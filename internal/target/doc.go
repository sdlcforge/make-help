// Package target handles adding and removing help targets from Makefiles.
//
// The add-target command uses a three-tier strategy to determine where
// to place the help target:
//   1. Explicit --target-file path
//   2. make/01-help.mk if include make/*.mk pattern exists
//   3. Append directly to main Makefile
//
// All file modifications use atomic writes (write to temp, then rename)
// to prevent file corruption on process crashes. Makefiles are validated
// with `make -n` before modification to catch syntax errors early.
package target
