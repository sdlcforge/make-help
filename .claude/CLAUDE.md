# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test Commands

```bash
# Build the binary
go build ./cmd/make-help

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run a specific package's tests
go test ./internal/parser/...

# Run a specific test by name
go test ./internal/parser/... -run TestScanFile

# Run integration tests only
go test ./test/integration/...

# Run the built binary
./make-help --makefile-path path/to/Makefile
```

## Architecture Overview

`make-help` is a CLI tool that generates formatted help output from specially-formatted Makefile comments. The processing pipeline flows through these stages:

```
CLI Layer → Discovery → Parser → Model Builder → Ordering → Summary → Formatter → Output
```

### Key Packages

- **`internal/cli/`**: Cobra-based CLI using flags (not subcommands). `help.go` contains the main orchestration in `runHelp()`. `root.go` handles flag-based command dispatching.
- **`internal/discovery/`**: Uses `make` commands to discover included files (via `MAKEFILE_LIST`) and targets (via `make -p`). Uses `CommandExecutor` interface for testability.
- **`internal/parser/`**: Stateful scanner that extracts `@file`, `@category`, `@var`, `@alias` directives and associates documentation with targets.
- **`internal/model/`**: `HelpModel` contains categories, targets, aliases, and variables. Builder constructs it from parsed files; validator enforces categorization rules.
- **`internal/ordering/`**: Applies sorting strategies (alphabetical vs discovery order) to categories and targets.
- **`internal/summary/`**: Extracts first sentence from documentation, stripping markdown. Ported from `extract-topic` JS library.
- **`internal/format/`**: Renders help output with optional ANSI colors. Supports both summary and detailed target views.
- **`internal/target/`**: Handles help target generation and removal with atomic file operations.

### CLI Flags

The CLI uses flags instead of subcommands:

- `--create-help-target` - generates help target file with local binary installation
- `--remove-help-target` - removes generated help targets and files
- `--target <name>` - shows detailed help for a single target
- `--include-target` - includes undocumented targets in help output (repeatable, comma-separated)
- `--include-all-phony` - includes all .PHONY targets in help output

### Mode Flags (mutually exclusive)

1. **Default (no special flags)**: Display help output
2. **`--target <name>`**: Display detailed help for single target
3. **`--create-help-target`**: Generate help target file
4. **`--remove-help-target`**: Remove help targets

### Documentation Syntax

The parser recognizes these directives in `## ` comments:
- `@file` - File-level documentation (appears before targets list)
- `@category <name>` - Groups subsequent targets under a category
- `@var <NAME> [description]` - Documents an environment variable
- `@alias <name1>, <name2>` - Alternative target names

### Generated Help File Format

When `--create-help-target` is used, it generates a Makefile include with:

- `GOBIN ?= .bin` - configurable binary directory
- Binary installation target using `go install`
- `.PHONY: help` target for summary view
- `.PHONY: help-<target>` for each documented target (detailed view)

### Test Fixtures

Test Makefiles are in `test/fixtures/makefiles/` and expected outputs in `test/fixtures/expected/`. Integration tests in `test/integration/cli_test.go` build the binary and run it against fixtures.

## Design Reference

`docs/design.md` contains the detailed design specification including data structures, algorithms, and component responsibilities.
