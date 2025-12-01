# CLAUDE.md

Quick reference for AI agents working on this codebase. For comprehensive details, see the documentation links below.

## What This Project Does

`make-help` is a Go CLI tool that generates formatted help output from specially-formatted Makefile comments. It processes Makefiles through a pipeline: CLI → Discovery → Parser → Model Builder → Ordering → Summary → Formatter → Output.

## Essential Commands

```bash
# Build and run
go build ./cmd/make-help
./make-help --makefile-path path/to/Makefile

# Testing
go test ./...                          # All tests
go test -cover ./...                   # With coverage
go test ./internal/parser/... -run TestScanFile  # Specific test
go test ./test/integration/...         # Integration tests only

# Development
./make-help --verbose --makefile-path test.mk  # Debug mode
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
| `internal/parser/` | Extract `@file`, `@category`, `@var`, `@alias` | `Scanner.ScanFile()` |
| `internal/model/` | Build & validate help model | `Builder.Build()`, `HelpModel` |
| `internal/ordering/` | Sort categories/targets | `Service.ApplyOrdering()` |
| `internal/summary/` | Extract first sentence (strip markdown) | `Extractor.Extract()` |
| `internal/format/` | Render with optional ANSI colors | `Renderer.Render()` |
| `internal/target/` | Generate/remove help targets | `AddService`, `RemoveService` |

### Important Design Patterns

1. **CLI uses flags, not subcommands**: Mode detection via flag combinations (`--create-help-target`, `--remove-help-target`, `--target <name>`)
2. **Testability via interfaces**: `CommandExecutor` interface for mocking `make` commands
3. **Security-first**: No shell injection; atomic file writes; 30s command timeouts
4. **Stateful parser**: `parser.Scanner` maintains state across lines to associate docs with targets
5. **Immutable model**: `HelpModel` is built once, not mutated

## Documentation Syntax (for parser)

```makefile
## @file                     # File-level docs (appears before targets)
## @category Build           # Groups subsequent targets under "Build"
## @var CC [description]     # Documents environment variable
## @alias b, build-all       # Alternative target names
## Build the project         # Target documentation (first sentence = summary)
build:
	go build ./...
```

## Test Fixtures

- **Input Makefiles**: `test/fixtures/makefiles/*.mk`
- **Expected outputs**: `test/fixtures/expected/*.txt`
- **Integration tests**: `test/integration/cli_test.go` (builds binary, runs against fixtures)

## Common Development Tasks

### Adding a new directive type
1. Update `internal/parser/directive.go` (add constant to DirectiveType)
2. Add parsing logic in `internal/parser/scanner.go` `parseDirective()` method
3. Handle in `internal/model/builder.go` `processFile()` method (around line 197)
4. Update formatter if needed (`internal/format/renderer.go`)
5. Add tests (parser unit test + integration fixture)
6. Update `README.md` and `docs/architecture.md`

### Changing output format
1. Modify templates in `internal/format/renderer.go`
2. Update integration test fixtures in `test/fixtures/expected/`
3. Regenerate example outputs in `examples/*/help.mk`

### Adding a CLI flag
1. Add to `Config` struct in `internal/cli/config.go`
2. Register in `internal/cli/root.go` `NewRootCmd()`
3. Use in appropriate service
4. Add integration test coverage
5. Update `README.md` flags table

## Critical Context for AI Agents

- **All code is in `internal/`**: Not a library, prevents API commitment
- **Only Cobra is an external dependency**: Everything else uses stdlib
- **Single-pass parsing**: Parser reads each file once; no backtracking
- **Security-conscious**: Validate paths, sanitize inputs, atomic writes, timeouts
- **Mixed categorization is an error**: Either all targets categorized or none (use `--default-category` to resolve)
- **Generated help files are self-referential**: Use `$(dir $(lastword $(MAKEFILE_LIST)))` pattern to work from any directory

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
**Color codes appearing in output**: Use --no-color when piping to files
**"make command timed out"**: Check for infinite recursion in Makefile includes
