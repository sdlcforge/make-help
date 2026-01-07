# Developer Brief: make-help

A practical guide for contributors and maintainers.

## Table of contents

- [Quick Start for Developers](#quick-start-for-developers)
- [Contributing Guidelines](#contributing-guidelines)
- [Architecture Overview](#architecture-overview)
- [Project Structure](#project-structure)
- [Common Development Tasks](#common-development-tasks)
- [Debugging Tips](#debugging-tips)
- [Design Document Reference](#design-document-reference)

## Quick start for developers

### Prerequisites

- Go 1.21 or later
- GNU Make 4.x installed
- Familiarity with Makefile syntax

### Building and testing

```bash
# Build the binary (outputs to ./bin/make-help)
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
./bin/make-help --makefile-path path/to/Makefile
```

## Contributing guidelines

### Code conventions

Follow these conventions when contributing:

1. **Use `gofmt` for formatting** - All code must be formatted
2. **Write godoc comments** - All exported symbols need documentation
3. **Meaningful variable names** - Prioritize clarity over brevity
4. **Add tests for new functionality** - Aim for >90% coverage on new code
5. **Update documentation** - Keep README.md, architecture.md, and this file in sync

### Pull request process

1. **Fork the repository** and create a feature branch
2. **Write tests first** for new functionality (TDD approach)
3. **Ensure all tests pass** before submitting
4. **Update documentation** if adding features or changing behavior
5. **Submit a pull request** with a clear description of changes

### Testing standards

- **Unit tests:** Cover individual functions and edge cases
- **Integration tests:** Use fixtures in `test/fixtures/`
- **Coverage target:** >90% overall, 95%+ for critical packages (parser, summary)
- **Test naming:** Use descriptive names that explain the scenario

## Architecture overview

### Processing pipeline

```
CLI → Discovery → Parser → Model Builder → Ordering → Summary → Formatter → Output
```

Each stage is independent and testable. See [architecture.md](architecture.md) for detailed architecture.

### Key design principles

1. **Separation of concerns** - Each package has a single, well-defined responsibility
2. **Testability first** - External commands use `CommandExecutor` interface for mocking
3. **Immutability** - Data structures are built once, not mutated
4. **Clear error boundaries** - Custom error types with actionable messages
5. **Security-conscious** - No shell injection, atomic file writes, command timeouts

## Project structure

```
make-help/
├── cmd/make-help/        # CLI entry point (thin wrapper)
├── internal/
│   ├── cli/             # Command-line interface (Cobra-based)
│   ├── discovery/       # Makefile and target discovery
│   ├── parser/          # Documentation parsing (stateful scanner)
│   ├── model/           # Data structures and builder
│   ├── ordering/        # Sorting strategies
│   ├── summary/         # Summary extraction (extract-topic port)
│   ├── format/          # Output rendering with colors
│   ├── target/          # Help file generation/removal
│   └── errors/          # Custom error types
├── examples/            # Working example projects
│   ├── uncategorized-targets/
│   ├── categorized-project/
│   ├── full-featured/
│   └── filtering-demo/
├── scripts/             # Helper scripts
│   └── run-example.sh   # Run examples with shared GOBIN
├── test/
│   ├── fixtures/        # Test Makefiles and expected outputs
│   └── integration/     # End-to-end tests
└── docs/                # Design and developer documentation
```

### Why `internal/`?

All code lives in `internal/` because `make-help` is a CLI tool, not a library. This prevents accidental API commitment and allows freedom to refactor without breaking external dependencies.

### Package responsibilities

| Package | Responsibility | Key Types | External Dependencies |
|---------|---------------|-----------|---------------------|
| `cli` | Argument parsing, flag validation, routing | `Config`, `RootCmd` | `spf13/cobra` |
| `discovery` | Find Makefiles and targets via `make` | `Service`, `DiscoverTargetsResult` | None (uses stdlib `exec`) |
| `parser` | Extract directives from Makefile content | `Scanner`, `Directive` | None |
| `model` | Build help model from directives | `HelpModel`, `Builder` | None |
| `ordering` | Sort categories and targets | `Service` | None |
| `summary` | Extract first sentence from docs | `Extractor` | None |
| `format` | Render help output with colors | `Renderer`, `ColorScheme` | None |
| `target` | Generate/remove help targets with smart location detection | `AddService`, `IncludePattern` | None |
| `lint` | Documentation quality checking and auto-fixing | `Check`, `Fix`, `Fixer` | None |
| `version` | Build-time version information | `Version` variable | None |
| `errors` | Custom error types | `MixedCategorizationError`, etc. | None |

**Key insight:** Only `cli` has external dependencies. All other packages use only stdlib.

### Test organization

```
test/
├── fixtures/
│   ├── makefiles/           # Input test Makefiles
│   │   ├── basic.mk
│   │   ├── categorized.mk
│   │   └── with_includes.mk
│   └── expected/            # Expected outputs
│       ├── basic_help.txt
│       └── categorized_help.txt
└── integration/
    └── cli_test.go          # Fixture-based end-to-end tests
```

**Adding a test:**
1. Create input Makefile in `fixtures/makefiles/`
2. Run `make-help` manually, verify output
3. Save output to `fixtures/expected/`
4. Add test case to `cli_test.go`

## Common development tasks

### Adding a new directive type

1. **Define the directive** in `internal/parser/directive.go`:
   - Add constant to `DirectiveType` enum
   - Update `parseDirective()` to detect it

2. **Handle in model builder** (`internal/model/builder.go`):
   - Add case in `Build()` switch statement
   - Define any new data structures

3. **Update formatter** (`internal/format/renderer.go`):
   - Add rendering logic if it affects output

4. **Add tests**:
   - Parser test in `internal/parser/scanner_test.go`
   - Integration test with fixture

5. **Update documentation**:
   - README.md "Documentation Syntax" section
   - architecture.md data structures section

### Changing output format

1. **Modify templates** in `internal/format/renderer.go`
2. **Update color scheme** in `internal/format/color.go` if needed
3. **Update integration test fixtures** in `test/fixtures/expected/`
4. **Regenerate example outputs** in `examples/*/help.mk`

### Adding a new CLI flag

1. **Add to Config struct** in `internal/cli/config.go`
2. **Register flag** in `internal/cli/root.go` `NewRootCmd()`
3. **Use in appropriate service** (discovery, ordering, etc.)
4. **Add test coverage** in integration tests
5. **Update documentation**:
   - `README.md` flags table and usage examples
   - `docs/architecture/components.md` CLI Parser section
   - `.claude/CLAUDE.md` if it affects common development tasks

### Modifying ordering behavior

1. **Update logic** in `internal/ordering/service.go`
2. **Add test cases** in `internal/ordering/service_test.go`
3. **Update architecture.md** algorithm section if significantly changed

### Adding a new lint check

1. **Define the check** in `internal/lint/checks.go`:
   - Create CheckFunc that scans for issues
   - Optionally create FixFunc that generates fixes
   - Register check in check registry

2. **Add test cases** in `internal/lint/checks_test.go`:
   - Test check detection (should find warnings)
   - Test fix generation (if fixable)
   - Test fix application

3. **Update documentation**:
   - Add to lint check list in README.md
   - Update architecture docs if introducing new concepts

## Debugging tips

### Enable verbose output

```bash
./make-help --verbose --makefile-path test.mk
```

This shows:
- Discovered Makefiles and their order
- Discovered targets from `make -p`
- File operations during file generation or removal

### Common issues

**Issue:** "Makefile not found"
- Check current directory
- Use `--makefile-path` with absolute path
- Verify file permissions

**Issue:** "Mixed categorization" error
- Check if some targets have `!category`, others don't
- Use `--default-category Misc` to resolve
- Or categorize all targets

**Issue:** Tests failing after parser changes
- Regenerate fixtures: run `./bin/make-help` manually and save new expected output
- Update integration test expectations
- Check for whitespace differences in output

**Issue:** Help file created in wrong location
- Check for existing `include make/*.mk` directive in Makefile
- Use `--help-file-rel-path` to specify exact location if needed
- Verify make/ directory is created automatically

**Issue:** Lint fixes not being applied
- Ensure `--fix` is used with `--lint`
- Check that warnings are fixable (some are error-only)
- Use `--fix --dry-run` to preview changes first

### Useful development commands

```bash
# Run specific test
go test ./internal/parser/... -run TestScanFile

# Check test coverage for a package
go test -cover ./internal/parser/...

# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Build and test all examples
./scripts/run-example.sh examples/full-featured

# Lint code (if using golangci-lint)
golangci-lint run

# Run make-help lint on examples
./bin/make-help --lint examples/full-featured/Makefile

# Apply lint fixes
./bin/make-help --lint --fix examples/full-featured/Makefile

# Preview lint fixes without applying
./bin/make-help --lint --fix --dry-run examples/full-featured/Makefile

# Check version
./bin/make-help --version
```

## Design document reference

For detailed information on:
- Component architecture and data flow
- Specific algorithms (file discovery, summary extraction)
- Error handling strategies
- Comprehensive testing approach

See [architecture.md](architecture.md).

Last reviewed: 2026-01-06
