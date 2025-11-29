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

- **`internal/cli/`**: Cobra-based CLI, delegates to services. `help.go` contains the main orchestration in `runHelp()`.
- **`internal/discovery/`**: Uses `make` commands to discover included files (via `MAKEFILE_LIST`) and targets (via `make -p`). Uses `CommandExecutor` interface for testability.
- **`internal/parser/`**: Stateful scanner that extracts `@file`, `@category`, `@var`, `@alias` directives and associates documentation with targets.
- **`internal/model/`**: `HelpModel` contains categories, targets, aliases, and variables. Builder constructs it from parsed files; validator enforces categorization rules.
- **`internal/ordering/`**: Applies sorting strategies (alphabetical vs discovery order) to categories and targets.
- **`internal/summary/`**: Extracts first sentence from documentation, stripping markdown. Ported from `extract-topic` JS library.
- **`internal/format/`**: Renders help output with optional ANSI colors.
- **`internal/target/`**: Handles `add-target` and `remove-target` commands with atomic file operations.

### Documentation Syntax

The parser recognizes these directives in `## ` comments:
- `@file` - File-level documentation (appears before targets list)
- `@category <name>` - Groups subsequent targets under a category
- `@var <NAME> [description]` - Documents an environment variable
- `@alias <name1>, <name2>` - Alternative target names

### Test Fixtures

Test Makefiles are in `test/fixtures/makefiles/` and expected outputs in `test/fixtures/expected/`. Integration tests in `test/integration/cli_test.go` build the binary and run it against fixtures.

## Design Reference

`docs/design.md` contains the detailed design specification including data structures, algorithms, and component responsibilities.
