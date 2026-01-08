# make-help

[![coverage: 91.6%](./docs/assets/coverage-badge.svg)](./actions/workflows/) [![license: Apache 2.0](./docs/assets/license-badge.svg)](./blob/main/LICENSE.txt)

Static help generation for Makefiles with rich documentation support and linting.

## Quick start

```makefile
## !file
## My Project Makefile

## !category Build
## Build the application
build:
	go build -o myapp ./cmd/myapp

## !category Test
## !var TEST_TYPES May be 'unit' (default) or 'integeration'
## Run all test types.
test:
	./scripts/run-tests.sh $(TEST_TYPES)
```

```bash
make-help && make help
```
Outputs:
```
Usage: make [<target>...] [<ENV_VAR>=<value>...]

My Project Makefile

Targets:

Build:
  - build: Build the application

Test:
  - test: Run all tests
    Vars: TEST_TYPES May be 'unit' (default) or 'integeration'
```

## Why make-help?

Makefiles are powerful but lack a built-in help system. As projects grow, developers accumulate dozens of targets with no easy way to discover or document them.

`make-help` solves this by extracting documentation from specially-formatted comments and generating static help files displayed with `make help`. The generated help files contain embedded help text and automatically regenerate when source Makefiles change. It supports categories, aliases, environment variables, and target filtering.

## Features

- **Static Help Generation**: Creates help files with embedded help text (no runtime dependencies)
- **Auto-Regeneration**: Generated files automatically regenerate when source Makefiles change
- **Automatic Target Discovery**: Scans your Makefiles (including included files) to find documented targets
- **Rich Documentation Syntax**: Support for file-level docs, categories, aliases, and environment variables
- **Target Filtering**: Control which targets appear in help output
- **Detailed Target Help**: Show full documentation for individual targets via `make help-<target>`
- **Flexible Ordering**: Alphabetical or discovery order for both categories and targets
- **Colored Output**: Terminal-aware colored output with override flags
- **Summary Extraction**: Automatically extracts first sentence from multi-line documentation

## Installation

### Homebrew (macOS/Linux)

```bash
brew install sdlcforge/tap/make-help
```

### Go

```bash
go install github.com/sdlcforge/make-help/cmd/make-help@latest
```

### npm/bun

```bash
# global installation
npm install -g @sdlcforge/make-help
bun install -g @sdlcforge/make-help
# package installation
npm install --save-dev @sdlcforge/make-help
bun install --save-dev --trust true @sdlcforge/make-help
```

## Usage

### Generate static help file (default)

```bash
make-help                              # Generate ./make/help.mk for ./Makefile
make-help --makefile-path path/to/Makefile
make-help --help-file-rel-path custom/path.mk  # Override default location
```

### Lint Makefile and help documentation

```bash
make-help --lint        # find potential red flags
make-help --lint --fix  # fix what can be automatically fixed and report the rest
```

### Display help dynamically

To see help output without generating a file:

```bash
make-help --output -                   # Show help for ./Makefile
make-help --output - --makefile-path path/to/Makefile
```

To get detailed help for a particular target:

```bash
make-help --output - --target build    # Full docs for 'build' target
```

### Target filtering

By default, only documented targets appear in help output.

```bash
make-help --include-target clean       # Include undocumented 'clean'
make-help --include-target foo,bar     # Include multiple (comma-separated)
make-help --include-all-phony          # Include all .PHONY targets
```

### Remove help files

```bash
make-help --remove-help                # Remove generated help files and include
```

## CLI reference

**Mode:**
- `--dry-run` - Preview changes without making them
- `--fix` - Auto-fix lint issues (requires `--lint`)
- `--lint` - Check documentation quality and report issues
- `--remove-help` - Remove generated help files
- `--target <name>` - Show detailed help for specific target (requires `--output -`)

**Input:**
- `--help-file-rel-path <path>` - Override the relative path stored in the generated help file for auto-regeneration (derived from `--output` by default)
- `--makefile-path <path>` - Path to Makefile (default: `./Makefile` in current directory)

**Output/formatting:**
- `--category-order <list>` - Explicit category order (comma-separated)
- `--color` / `--no-color` - Force or disable colored output
- `--default-category <name>` - Default category for uncategorized targets
- `--format <type>` - Output format: make, text, html, markdown (default: make)
- `--help-category <name>` - Category for generated help targets (default: `Help`)
- `--include-all-phony` - Include all .PHONY targets
- `--include-target <list>` - Include undocumented targets (comma-separated, repeatable)
- `--keep-order-all` - Preserve category, target, and file order
- `--keep-order-categories` - Preserve category discovery order
- `--keep-order-files` - Preserve file discovery order (default: alphabetical)
- `--keep-order-targets` - Preserve target discovery order
- `--output <path>` - Output destination (file path or `-` for stdout; default: `./make/help.mk` for make format)

**Misc:**
- `--help` - Displays `make-help` help
- `--verbose` - Enable verbose output
- `--version` - Display version information

## Documentation syntax

- Any line beginning with `##` is considered part of the documentation.
- Lines beginning with a `#` are treated as internal documentation and will be ignored.
- Documentation directives include:
  - `!file` to identify file level documentation.
  - `!category` to specify the category for the following targets within the source file.
  - `!alias` explicitly names another target as an alias for the target being documented. Aliases can usually be inferred and the use of this directive may not be necessary.
  - `!notalias` marks a phony `X: Y` construct as a non-alias.
  - `!var` documents environment variables affecting the target behavior.

### File-level documentation

Use `!file` to add documentation about the Makefile itself:

```makefile
## !file
## Project Build System
## This Makefile handles building, testing, and deploying the application.
```

> **Note: Entry point vs included files**
>
> The `!file` directive behaves differently depending on where it appears:
> - **Entry point Makefile**: Documentation appears at the top of help output, immediately after "Usage:"
> - **Included files**: Documentation appears in the "Included Files:" section with the file path

**Additional behaviors:**
- **Multiple `!file` directives**: Multiple directives in the same file are concatenated with a blank line between them.
- **File ordering**: Included files are sorted alphabetically by default. Use `--keep-order-files` to preserve discovery order.
- **Full text**: All file-level documentation is included, not just a summary.

### Target documentation

Document targets with `##` comments immediately before the target:

```makefile
## Build the entire project.
## This compiles all sources and runs code generation.
build:
	@echo "Building..."
```

The first sentence becomes the summary in the help output.

### Categories

Group related targets using `!category`. The `!category` directive applies to all subsequent targets until changed:

```makefile
## !category Build
## Build the project
build:                    # category: Build
	@echo "Building..."

## Compile source files (still in Build category)
compile:                  # category: Build (inherited)
	@echo "Compiling..."

## !category Test
## Run unit tests
test:                     # category: Test
	@echo "Testing..."

## Run integration tests (still in Test category)
integration:              # category: Test (inherited)
	@echo "Running integration tests..."
```

**Key Behaviors:**
- **Sticky directive**: Once set, `!category` applies to all subsequent targets within that make file until another `!category` is encountered
- **Reset to uncategorized**: Use `!category _` to reset the category to uncategorized (nil)
- **Categories are merged**: If you switch back and forth to the same category in a single or use the same category in mulitple files, all targets in that category will be grouped together.
- **Mixed categorization**: If you use categories, all documented targets must be categorized. Use `--default-category` to assign uncategorized targets to a default category

### Aliases

An 'alias' is just an alternate name for a target. There are two ways to create an alias.

**Implicit aliases**: A single phony target with a single phony target dependency is recognized as an alias; e.g., `test: test.unit` recognized `test` as an alias for `test.unit`. This can be suppressed by placing the `## !notalias` directive before the target. E.g.:

```makefile
## !notalias
all: build
```

You can also explicitly name one or more aliases with the `!alias` directive:

```makefile
## !alias b, build-all
## Build the project
build:
	@echo "Building..."
```

### Environment variables

Document behavior affecting environment variables using `!var`. This is similar to how parameters would be documented for function documentation.

```makefile
## !var DATABASE_URL Database connection string
## !var LOG_LEVEL Logging verbosity (debug, info, warn, error)
## Start the application server
server:
	./bin/server
```

Variables appear in the help output under the target:

```
  - server: Start the application server
    Vars: DATABASE_URL Database connection string, LOG_LEVEL Logging verbosity (debug, info, warn, error)
```

## Examples

The `examples/` directory contains complete working examples demonstrating different features. Each example includes a
generated `help.mk` file, so you can run them via Make:

```bash
make -f examples/categorized-project/Makefile help
make -f examples/full-featured/Makefile help-build
```

## Output format

The help output follows this structure:

```
Usage: make [<target>...] [<ENV_VAR>=<value>...]

[Entry point Makefile's !file documentation - full text, if present]

[Included files:
  path/to/file.mk
    Full documentation from the file's !file directive
    Can be multiple lines

  path/to/another.mk
    Documentation here]

Targets:

[Category name:]
  - <target> [<alias1>, <alias2>...]: <summary>
    [Vars: <VAR1> <description1>, <VAR2> <description2>...]
```

### Color scheme

When colors are enabled:

- **Category Names**: Bold Cyan
- **Target Names**: Bold Green
- **Aliases**: Yellow
- **Variable Names**: Magenta
- **Documentation**: White

## Advanced topics

### Working with included files

`make-help` automatically discovers all Makefiles via `include` directives using Make's `MAKEFILE_LIST` variable. Documentation from all files is aggregated in discovery order.

```makefile
# Main Makefile
include make/*.mk

## !category Core
## Build everything
all: build test deploy
	@echo "Done!"
```

```makefile
# make/build.mk
## !category Build
## Build the project
build:
	go build ./...
```

Both files are processed and targets are grouped by category.

## Uninstalling

### Removing generated help files

To remove the generated help files and artifacts created by `make-help`:

```bash
make-help --remove-help
```

**What gets removed**:
- The generated help file (e.g., `./make/help.mk` or `./make/00-help.mk`)
- The include directive that was automatically added to your Makefile (e.g., `-include make/*.mk`)

**What does NOT get removed**:
- Any targets or content you wrote yourself
- The `bin/` directory (where the binary is built locally during development)
- Your documentation comments (`##` comments in your Makefiles)

### Removing the binary

**If installed via Homebrew**:
```bash
brew uninstall make-help
```

**If installed via Go**:
```bash
rm $(go env GOPATH)/bin/make-help
```

**If installed via npm or bun**:
```bash
# global installation
npm uninstall -g @sdlcforge/make-help
bun uninstall -g @sdlcforge/make-help
# package installation
npm uninstall @sdlcforge/make-help
bun uninstall @sdlcforge/make-help
```

**If built locally during development**:
```bash
make clean    # Removes ./bin/make-help
```

## Troubleshooting

### Mixed categorization error

**Error**: `mixed categorization: found both categorized and uncategorized targets`

**Solution**: Either categorize all targets or use `--default-category`:

```bash
make-help --default-category Miscellaneous
```

### Unknown category error

**Error**: `unknown category "Foo" in --category-order`

**Solution**: Check the available categories in your Makefile. The error message lists all available categories.

### Makefile not found

**Error**: `Makefile not found: ./Makefile`

**Solution**: Specify the path explicitly:

```bash
make-help --makefile-path path/to/your/Makefile
```

### Make command timeout

If your Makefile is very complex or has expensive operations during parsing, you may encounter timeouts. The default timeout is 30 seconds.

**Solution**: Simplify variable expansions or move expensive operations out of variable assignments.

## Developer documentation

- **[Developer Brief](docs/developer-brief.md)** - Contributing guide, development setup, and common tasks
- **[Architecture Document](docs/architecture.md)** - Comprehensive architecture and implementation details

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE.txt](LICENSE.txt) file for details.

## Credits

The summary extraction algorithm is a Go port of the [extract-topic](https://www.npmjs.com/package/extract-topic) JavaScript library.

## Support

- **Issues**: Report bugs and request features via GitHub Issues
- **Discussions**: Ask questions in GitHub Discussions

Last reviewed: 2026-01-07
