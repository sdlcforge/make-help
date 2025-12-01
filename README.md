# make-help

Dynamic help generation for Makefiles with rich documentation support.

`make-help` is a command-line tool that generates beautiful, organized help output from specially-formatted comments in your Makefiles. It supports categories, aliases, environment variables, target filtering, and can generate project-local help targets.

## Features

- **Automatic Target Discovery**: Scans your Makefiles (including included files) to find documented targets
- **Rich Documentation Syntax**: Support for file-level docs, categories, aliases, and environment variables
- **Target Filtering**: Control which targets appear in help output
- **Detailed Target Help**: Show full documentation for individual targets
- **Generated Help Targets**: Create `make help` and `make help-<target>` targets with local binary installation
- **Flexible Ordering**: Alphabetical or discovery order for both categories and targets
- **Colored Output**: Terminal-aware colored output with override flags
- **Summary Extraction**: Automatically extracts first sentence from multi-line documentation
- **Include Support**: Discovers and processes Makefiles via `include` directives

## Installation

### Project-Local (Recommended)

Run once to set up help targets in your project:

```bash
make-help --create-help-target
```

This automatically creates a help file with:
- Automatic binary installation via `go install`
- `make help` - displays help summary
- `make help-<target>` - detailed help for each documented target

The file location is auto-detected:
- If your Makefile has `include make/*.mk`, creates `make/01-help.mk`
- Otherwise, creates `help.mk` in the same directory as the Makefile

### Global

```bash
go install github.com/sdlcforge/make-help/cmd/make-help@latest
```

## Quick Start

### 1. Document Your Makefile

Add documentation comments using the `##` prefix:

```makefile
## @file
## My Project Makefile
## This file contains build and deployment targets.

## @category Build
## Build the application
build:
	go build -o myapp ./cmd/myapp

## @category Test
## Run all tests
test:
	go test ./...

## @category Deploy
## @var ENV Target environment (dev, staging, prod)
## Deploy the application
deploy:
	./scripts/deploy.sh $(ENV)
```

### 2. Generate Help

```bash
make-help
```

Output:

```
Usage: make [<target>...] [<ENV_VAR>=<value>...]

Targets:

Build:
  - build: Build the application

Deploy:
  - deploy: Deploy the application
    Vars: ENV Target environment (dev, staging, prod)

Test:
  - test: Run all tests
```

### 3. Add Help Targets to Makefile

```bash
make-help --create-help-target
```

Now you can run:
- `make help` - displays help summary
- `make help-build` - detailed documentation for the build target

## Usage

### Display Help (default)

```bash
make-help                              # Show help for ./Makefile
make-help --makefile-path path/to/Makefile
```

### Detailed Target Help

```bash
make-help --target build               # Full docs for 'build' target
```

### Target Filtering

By default, only documented targets appear in help output.

```bash
make-help --include-target clean       # Include undocumented 'clean'
make-help --include-target foo,bar     # Include multiple (comma-separated)
make-help --include-all-phony          # Include all .PHONY targets
```

### Generate Help Targets

```bash
# Auto-detects location: creates help.mk or make/01-help.mk
make-help --create-help-target

# Pin specific version for go install
make-help --create-help-target --version v1.2.3

# Specify custom relative path for the help file
make-help --create-help-target --help-file-rel-path custom/help.mk

# Use default category for uncategorized targets
make-help --create-help-target --default-category Misc
```

### Remove Help Targets

```bash
make-help --remove-help-target         # Remove generated help files
```

## Flags Reference

| Flag | Description |
|------|-------------|
| `--makefile-path` | Path to Makefile (default: ./Makefile) |
| `--target` | Show detailed help for specific target |
| `--include-target` | Include undocumented target (repeatable) |
| `--include-all-phony` | Include all .PHONY targets |
| `--create-help-target` | Generate help target file |
| `--remove-help-target` | Remove help target from Makefile |
| `--help-file-rel-path` | Override relative path for generated help file (e.g., `help.mk` or `make/help.mk`). Must be a relative path. |
| `--version` | Pin version in generated go install |
| `--default-category` | Default category for uncategorized targets |
| `--keep-order-categories` | Preserve category discovery order |
| `--keep-order-targets` | Preserve target discovery order |
| `--category-order` | Explicit category order (comma-separated) |
| `--no-color` | Disable colored output |
| `--color` | Force colored output |
| `--verbose` | Enable verbose output |

## Documentation Syntax

### File-Level Documentation

Use `@file` to add file-level documentation that appears before the targets list:

```makefile
## @file
## Project Build System
## This Makefile handles building, testing, and deploying the application.
```

### Target Documentation

Document targets with `##` comments immediately before the target:

```makefile
## Build the entire project.
## This compiles all sources and runs code generation.
build:
	@echo "Building..."
```

The first sentence becomes the summary in the help output.

### Categories

Group related targets using `@category`:

```makefile
## @category Build
## Build the project
build:
	@echo "Building..."

## @category Build
## Compile source files
compile:
	@echo "Compiling..."

## @category Test
## Run unit tests
test:
	@echo "Testing..."
```

**Note**: If you use categories, all documented targets must be categorized. Use `--default-category` to assign uncategorized targets to a default category.

### Aliases

Provide alternative names for targets using `@alias`:

```makefile
## @alias b, build-all
## Build the project
build:
	@echo "Building..."
```

Now users can run `make b` or `make build-all` instead of `make build`.

### Environment Variables

Document required environment variables using `@var`:

```makefile
## @var DATABASE_URL Database connection string
## @var LOG_LEVEL Logging verbosity (debug, info, warn, error)
## Start the application server
server:
	./bin/server
```

Variables appear in the help output under the target:

```
  - server: Start the application server
    Vars: DATABASE_URL Database connection string, LOG_LEVEL Logging verbosity (debug, info, warn, error)
```

### Complete Example

```makefile
## @file
## MyApp Build System
## Targets for building, testing, and deploying MyApp.

## @category Build
## @alias b
## @var CC C compiler to use
## @var CFLAGS Compiler flags
## Build the entire project.
## This compiles all sources and links the final binary.
build:
	$(CC) $(CFLAGS) -o myapp main.c

## @category Test
## @alias t
## @var TEST_FILTER Filter for test names
## Run all tests.
## Uses the native test framework.
test:
	./scripts/test.sh $(TEST_FILTER)

## @category Deploy
## @var ENV Target environment (dev, staging, prod)
## Deploy to specified environment
deploy:
	./scripts/deploy.sh $(ENV)
```

## Examples

### Basic Usage

```bash
# Generate help for current directory's Makefile
make-help

# Generate help for specific Makefile
make-help --makefile-path path/to/Makefile

# Disable colored output
make-help --no-color

# Enable verbose debugging
make-help --verbose
```

### Target Filtering

```bash
# Include specific undocumented targets
make-help --include-target clean,install

# Include all .PHONY targets
make-help --include-all-phony

# Show detailed help for a target
make-help --target build
```

### Ordering Examples

```bash
# Preserve discovery order for categories and targets
make-help --keep-order-categories --keep-order-targets

# Only preserve category order
make-help --keep-order-categories

# Explicit category order (Build, Test, then others alphabetically)
make-help --category-order Build,Test

# Handle mixed categorization by assigning to default
make-help --default-category Uncategorized
```

### Generate and Remove Help Targets

```bash
# Generate help targets (auto-detects best location)
make-help --create-help-target

# Pin specific version
make-help --create-help-target --version v1.2.3

# Specify explicit relative path
make-help --create-help-target --help-file-rel-path custom-help.mk

# Remove help targets and all artifacts
make-help --remove-help-target
```

## Example Projects

The `examples/` directory contains complete working examples demonstrating different features.

### Running Examples

Each example includes a generated `help.mk` file, so you can run them via Make:

```bash
make -f examples/categorized-project/Makefile help
make -f examples/full-featured/Makefile help-build
```

Or use the helper script which sets up a shared binary location:

```bash
./scripts/run-example.sh examples/categorized-project
./scripts/run-example.sh examples/full-featured help-build
```

The helper script installs the make-help binary to the project root's `.bin/` directory, avoiding multiple installations across examples.

### uncategorized-targets

Simple flat target list without categories. Demonstrates basic documentation with `@var` and `@alias` directives.

```bash
make-help --makefile-path examples/uncategorized-targets/Makefile
```

### categorized-project

Uses `@category` to organize targets into logical groups (Build, Test, Development, Maintenance).

```bash
make-help --makefile-path examples/categorized-project/Makefile
```

### full-featured

Comprehensive example using all directives:
- `@file` for project-level documentation
- `@category` for target organization
- `@var` for environment variable documentation
- `@alias` for target shortcuts
- Multi-line documentation with paragraph breaks

```bash
# Show help summary
make-help --makefile-path examples/full-featured/Makefile

# Show detailed help for a target
make-help --makefile-path examples/full-featured/Makefile --target build
```

### filtering-demo

Demonstrates target filtering with `--include-target` and `--include-all-phony` flags. Contains both documented and undocumented targets.

```bash
# Default: only documented targets
make-help --makefile-path examples/filtering-demo/Makefile

# Include specific undocumented targets
make-help --makefile-path examples/filtering-demo/Makefile --include-target setup,check --default-category Misc

# Include all .PHONY targets
make-help --makefile-path examples/filtering-demo/Makefile --include-all-phony --default-category Misc
```

## Integration with Make

After running `make-help --create-help-target`, you can invoke help directly from Make:

```bash
# Show help summary
make help

# Show detailed help for specific target
make help-build
make help-test

# Continue using other targets
make build
make test
```

The generated help targets automatically install the `make-help` binary locally to `.bin/` and use it to generate help output.

### How Generated Help Files Work

The generated `help.mk` file uses a self-referential pattern to locate files correctly regardless of where Make is invoked from:

```makefile
MAKE_HELP_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
GOBIN ?= $(MAKE_HELP_DIR).bin
MAKE_HELP_BIN := $(GOBIN)/make-help
MAKE_HELP_OPTS := --makefile-path $(MAKE_HELP_DIR)Makefile
```

This pattern ensures:
- `MAKE_HELP_DIR` always points to the directory containing the help.mk file
- The binary is installed relative to that directory (in `.bin/`)
- The help command always references the correct Makefile

This allows the help.mk file to work correctly whether:
- It's in the project root (`help.mk`)
- It's in a subdirectory (`make/01-help.mk`)
- Make is invoked with `-f path/to/Makefile` from any directory

You can override `GOBIN` to install the binary elsewhere:

```bash
GOBIN=/usr/local/bin make help
```

## Output Format

The help output follows this structure:

```
Usage: make [<target>...] [<ENV_VAR>=<value>...]

[File-level documentation if present]

Targets:

[Category Name:]
  - <target> [<alias1>, <alias2>...]: <summary>
    [Vars: <VAR1> <description1>, <VAR2> <description2>...]
```

### Color Scheme

When colors are enabled:

- **Category Names**: Bold Cyan
- **Target Names**: Bold Green
- **Aliases**: Yellow
- **Variable Names**: Magenta
- **Documentation**: White

## Advanced Topics

### Working with Included Files

`make-help` automatically discovers all Makefiles via `include` directives using Make's `MAKEFILE_LIST` variable. Documentation from all files is aggregated in discovery order.

```makefile
# Main Makefile
include make/*.mk

## @category Core
## Build everything
all: build test deploy
	@echo "Done!"
```

```makefile
# make/build.mk
## @category Build
## Build the project
build:
	go build ./...
```

Both files are processed and targets are grouped by category.

### Split Categories

You can define the same category in multiple files. Targets are merged together:

```makefile
# Makefile
## @category Build
## Build the application
build:
	go build ./...
```

```makefile
# make/build.mk
## @category Build
## Build documentation
docs:
	go doc ./...
```

The `Build` category will contain both `build` and `docs` targets.

### Summary Extraction Algorithm

`make-help` uses a sophisticated algorithm (ported from the extract-topic JavaScript library) to extract the first sentence from multi-line documentation:

- Strips markdown formatting (`**bold**`, `*italic*`, `` `code` ``, `[links](urls)`)
- Strips HTML tags
- Handles edge cases like ellipsis (`...`) and IP addresses (`127.0.0.1`)
- Extracts the first sentence ending in `.`, `!`, or `?`

Example:

```makefile
## **Build** the entire project.
## This compiles all sources, runs code generation,
## and links the final binary. See docs/build.md for details.
build:
	@echo "Building..."
```

Summary: "Build the entire project."

## Troubleshooting

### Mixed Categorization Error

**Error**: `mixed categorization: found both categorized and uncategorized targets`

**Solution**: Either categorize all targets or use `--default-category`:

```bash
make-help --default-category Miscellaneous
```

### Unknown Category Error

**Error**: `unknown category "Foo" in --category-order`

**Solution**: Check the available categories in your Makefile. The error message lists all available categories.

### Makefile Not Found

**Error**: `Makefile not found: ./Makefile`

**Solution**: Specify the path explicitly:

```bash
make-help --makefile-path path/to/your/Makefile
```

### Make Command Timeout

If your Makefile is very complex or has expensive operations during parsing, you may encounter timeouts. The default timeout is 30 seconds.

**Solution**: Simplify variable expansions or move expensive operations out of variable assignments.

## Documentation

- **[Developer Brief](docs/developer-brief.md)** - Contributing guide, development setup, and common tasks
- **[Architecture Document](docs/architecture.md)** - Comprehensive architecture and implementation details

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE.txt](LICENSE.txt) file for details.

## Credits

The summary extraction algorithm is a Go port of the [extract-topic](https://www.npmjs.com/package/extract-topic) JavaScript library.

## Support

- **Issues**: Report bugs and request features via GitHub Issues
- **Discussions**: Ask questions in GitHub Discussions
