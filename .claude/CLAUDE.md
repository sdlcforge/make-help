# CLAUDE.md

Quick reference for AI agents working on this codebase. For comprehensive details, see the documentation links below.

## What This Project Does

`make-help` is a Go CLI tool that generates static help files for Makefiles from specially-formatted comments. By default, it creates `./make/help.mk` (with automatic directory creation and include directive insertion) with embedded help text. It can also display help dynamically. The tool processes Makefiles through a pipeline: CLI → Discovery → Parser → Model Builder → Ordering → Summary → Formatter → Output.

## Essential Commands

```bash
# Build and run
go build ./cmd/make-help
./bin/make-help                            # Generate ./make/help.mk (default)
./bin/make-help --output -                 # Display help dynamically to stdout
./bin/make-help --output - --target build  # Show detailed target help
./bin/make-help --version                  # Display version information

# Testing
go test ./...                          # All tests
go test -cover ./...                   # With coverage
go test ./internal/parser/... -run TestScanFile  # Specific test
go test -tags=integration ./test/integration/...  # Integration tests only

# Development
./bin/make-help --verbose                  # Debug mode (generates ./make/help.mk)
./bin/make-help --output - --verbose       # Debug dynamic help display
```

## Key Entry Points

- **`internal/cli/root.go`**: CLI flag registration and command routing (uses Cobra)
- **`internal/cli/help.go`**: Main orchestration in `runHelp()` function
- **`cmd/make-help/main.go`**: Binary entry point (thin wrapper)

## Architecture Quick Reference

### Processing Pipeline

```
CLI Layer → Discovery → Parser → Model Builder → Ordering → Summary → Formatter → Output
```

### Package Map (what does what)

| Package | Purpose | Key Type/Function |
|---------|---------|-------------------|
| `internal/cli/` | Cobra CLI, flag validation, routing | `root.go`, `help.go` |
| `internal/discovery/` | Find Makefiles & targets via `make` commands | `Service.DiscoverTargets()` |
| `internal/parser/` | Extract `!file`, `!category`, `!var`, `!alias` | `Scanner.ScanFile()` |
| `internal/model/` | Build & validate help model | `Builder.Build()`, `HelpModel` |
| `internal/ordering/` | Sort categories/targets | `Service.ApplyOrdering()` |
| `internal/summary/` | Extract first sentence (strip markdown) | `Extractor.Extract()` |
| `internal/format/` | Format help output in multiple formats | `Formatter.RenderHelp()`, `NewFormatter()` |
| `internal/target/` | Generate/remove help targets | `AddService`, `RemoveService` |

### Important Design Patterns

1. **Static help generation by default**: Running `make-help` generates `./make/help.mk` with embedded help text (use `--output -` for dynamic display)
2. **Smart file placement**: Defaults to `./make/help.mk` with automatic directory creation, numbered prefix detection, and include directive insertion
3. **CLI uses flags, not subcommands**: Mode detection via flag combinations (`--output -`, `--remove-help`, `--target <name>`)
4. **Testability via interfaces**: `CommandExecutor` interface for mocking `make` commands
5. **Security-first**: No shell injection; atomic file writes; 30s command timeouts
6. **Stateful parser**: `parser.Scanner` maintains state across lines to associate docs with targets
7. **Immutable model**: `HelpModel` is built once, not mutated

## Documentation Syntax (for parser)

```makefile
## !file                     # File-level docs (entry point: before targets; included: in "Included Files:" section)
## !category Build           # Switch: all subsequent targets use "Build" category
## !var CC [description]     # Documents environment variable
## !alias b, build-all       # Alternative target names
## Build the project         # Target documentation (first sentence = summary)
build:
	go build ./...

## Test the project         # Still in "Build" category (inherited)
test:
	go test ./...

## !category _               # Reset to uncategorized (underscore = reset)
## Standalone task           # No category (will error if mixed with categorized)
standalone:
	@echo "Task"
```

**Key !file Behavior:**
- **Entry point Makefile**: `!file` documentation appears at top of help output (full text, not just summary)
- **Included files**: `!file` documentation appears in "Included Files:" section (full text)
- **Multiple directives**: Multiple `!file` directives in same file are concatenated with blank line
- **File ordering**: Files sorted alphabetically by default; use `--keep-order-files` to preserve discovery order

**Key !category Behavior:**
- **Sticky directive**: Once set, applies to all subsequent targets until another `!category` is encountered
- **Category inheritance**: Targets inherit the current category; no need to repeat `!category` for each target
- **Reset syntax**: Use `!category _` to reset to uncategorized state
- **Mixed categorization error**: Can't mix categorized and uncategorized targets (use `--default-category` to resolve)

## Test Fixtures

- **Input Makefiles**: `test/fixtures/makefiles/*.mk`
- **Expected outputs**: `test/fixtures/expected/*.txt`
- **Integration tests**: `test/integration/cli_test.go` (builds binary, runs against fixtures)

## Common Development Tasks

### Adding a new directive type
1. Update `internal/parser/directive.go` (add constant to DirectiveType)
2. Add parsing logic in `internal/parser/scanner.go` `parseDirective()` method
3. Handle in `internal/model/builder.go` `processFile()` method (around line 197)
4. Update formatter if needed (`internal/format/make_formatter.go` or appropriate formatter)
5. Add tests (parser unit test + integration fixture)
6. Update `README.md` and `docs/architecture.md`

### Changing output format
1. Modify templates in the appropriate formatter (e.g., `internal/format/make_formatter.go`, `text_formatter.go`)
2. Update static help generation in `internal/target/generator.go` (for embedded @printf statements)
3. Update integration test fixtures in `test/fixtures/expected/`
4. Regenerate example outputs in `examples/*/help.mk`

### Adding a CLI flag
1. Add to `Config` struct in `internal/cli/config.go`
2. Register in `internal/cli/root.go` `NewRootCmd()`
3. Use in appropriate service
4. Add integration test coverage
5. Update `README.md` flags table and `docs/architecture/components.md`

## Critical Context for AI Agents

- **All code is in `internal/`**: Not a library, prevents API commitment
- **Only Cobra is an external dependency**: Everything else uses stdlib
- **Single-pass parsing**: Parser reads each file once; no backtracking
- **Security-conscious**: Validate paths, sanitize inputs, atomic writes, timeouts
- **Mixed categorization is an error**: Either all targets categorized or none (use `--default-category` to resolve)
- **Generated help files contain static text**: Help text is embedded as @echo statements, not generated dynamically
- **Auto-regeneration**: Generated help files include targets that regenerate when source Makefiles change
- **Fallback chain**: Generated files try `make-help`, then `npx make-help`, then error
- **Default location is `./make/help.mk`**: The `make/` directory is created automatically if needed
- **Numbered prefix support**: If files in `./make/` use numeric prefixes (e.g., `10-foo.mk`), help file uses matching prefix (e.g., `00-help.mk`)
- **Auto-add include directive**: If no `include make/*.mk` pattern exists, one is automatically added to the Makefile
- **Fixed warnings are hidden**: When using `--lint --fix`, only unfixed warnings are displayed in output

## Comprehensive Documentation

| Document | Purpose |
|----------|---------|
| [README.md](../README.md) | End-user documentation |
| [docs/architecture.md](../docs/architecture.md) | Architecture overview with system diagram |
| [docs/architecture/components.md](../docs/architecture/components.md) | Detailed component specs |
| [docs/architecture/data-models.md](../docs/architecture/data-models.md) | Go type definitions |
| [docs/architecture/algorithms.md](../docs/architecture/algorithms.md) | Key algorithms (discovery, summary extraction) |
| [docs/architecture/program-flow.md](../docs/architecture/program-flow.md) | Processing pipelines |
| [docs/architecture/error-handling.md](../docs/architecture/error-handling.md) | Error strategies |
| [docs/architecture/testing-strategy.md](../docs/architecture/testing-strategy.md) | Testing approach |
| [docs/developer-brief.md](../docs/developer-brief.md) | Contributor guide with common tasks |

## Quick Troubleshooting

**Mixed categorization error**: Use `--default-category Misc`
**Tests failing after changes**: Regenerate fixtures by running binary manually and saving output
**Need to debug discovery**: Use `--verbose` flag to see Makefile resolution and target discovery
**Target not appearing in help**: Check if it's .PHONY (use --include-all-phony or --include-target)
**Want dynamic help instead of file generation**: Use `--output -` flag
**Need detailed target help**: Use `--output - --target <name>`
**Generated help not regenerating**: Check that `./make/help.mk` (or custom location) has the auto-regeneration target
**"make command timed out"**: Check for infinite recursion in Makefile includes
**Help file in wrong location**: Default is now `./make/help.mk`; use `--help-file-rel-path` to override
**Want numbered prefix for help file**: Place other numbered files (e.g., `10-foo.mk`) in `./make/` directory first
**Lint shows fixed warnings**: With `--lint --fix`, only unfixed warnings should appear; fixed ones are hidden
**Check version**: Use `--version` flag to display version information

Last reviewed: 2026-01-03T00:00Z
