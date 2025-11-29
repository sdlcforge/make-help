# make-help

Dynamic help generation for Makefiles with rich documentation support.

`make-help` is a command-line tool that generates beautiful, organized help output from specially-formatted comments in your Makefiles. It supports categories, aliases, environment variables, and more.

## Features

- **Automatic Target Discovery**: Scans your Makefiles (including included files) to find documented targets
- **Rich Documentation Syntax**: Support for file-level docs, categories, aliases, and environment variables
- **Flexible Ordering**: Alphabetical or discovery order for both categories and targets
- **Colored Output**: Terminal-aware colored output with override flags
- **Summary Extraction**: Automatically extracts first sentence from multi-line documentation
- **Include Support**: Discovers and processes Makefiles via `include` directives

## Installation

### From Source

```bash
go install github.com/sdlcforge/make-help/cmd/make-help@latest
```

### Build Locally

```bash
git clone https://github.com/sdlcforge/make-help
cd make-help
go build ./cmd/make-help
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

### 3. Add Help Target to Makefile

```bash
make-help add-target
```

Now you can run `make help` to display help directly from Make!

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

## CLI Reference

### Commands

#### `make-help` (default)

Generate help output from Makefile documentation.

```bash
make-help [flags]
```

#### `make-help add-target`

Add a `help` target to your Makefile that invokes `make-help`.

```bash
make-help add-target [flags]
```

#### `make-help remove-target`

Remove the help target from your Makefile.

```bash
make-help remove-target [flags]
```

### Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--makefile-path <path>` | Path to Makefile | `./Makefile` |
| `--color` | Force colored output | Auto-detect terminal |
| `--no-color` | Disable colored output | Auto-detect terminal |
| `--verbose` | Enable verbose output for debugging | `false` |

### Help Generation Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--keep-order-categories` | Preserve category discovery order | `false` (alphabetical) |
| `--keep-order-targets` | Preserve target discovery order | `false` (alphabetical) |
| `--keep-order-all` | Preserve discovery order for both | `false` |
| `--category-order <list>` | Explicit category order (comma-separated) | None |
| `--default-category <name>` | Default category for uncategorized targets | None (error on mixed) |

### Add-Target Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--target-file <path>` | Explicit path for help target file | Auto-detect |

The `add-target` command uses a three-tier strategy to determine where to place the help target:

1. **Explicit path**: If `--target-file` is specified, creates the file there and adds an include directive
2. **Auto-detect include pattern**: If Makefile contains `include make/*.mk`, creates `make/01-help.mk`
3. **Append to Makefile**: Otherwise, appends the help target directly to the main Makefile

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

### Ordering Examples

```bash
# Preserve discovery order for categories and targets
make-help --keep-order-all

# Only preserve category order
make-help --keep-order-categories

# Explicit category order (Build, Test, then others alphabetically)
make-help --category-order Build,Test

# Handle mixed categorization by assigning to default
make-help --default-category Uncategorized
```

### Add/Remove Help Target

```bash
# Add help target (auto-detects best location)
make-help add-target

# Add help target to specific file
make-help add-target --target-file custom-help.mk

# Add help target with ordering preferences
make-help add-target --keep-order-all

# Remove help target and all artifacts
make-help remove-target
```

## Integration with Make

After running `make-help add-target`, you can invoke help directly from Make:

```bash
# Show help
make help

# Continue using other targets
make build
make test
```

The generated help target passes through all your configuration flags, so the help output will match your preferences.

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

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Fork the repository** and create a feature branch
2. **Write tests** for new functionality (aim for >90% coverage)
3. **Follow Go conventions**:
   - Use `gofmt` for formatting
   - Write godoc comments for exported symbols
   - Use meaningful variable names
4. **Update documentation** if adding features
5. **Submit a pull request** with a clear description

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests
go test ./test/integration/...
```

### Project Structure

```
make-help/
├── cmd/make-help/        # CLI entry point
├── internal/
│   ├── cli/             # Command-line interface
│   ├── discovery/       # Makefile and target discovery
│   ├── parser/          # Documentation parsing
│   ├── model/           # Data structures
│   ├── ordering/        # Sorting strategies
│   ├── summary/         # Summary extraction
│   ├── format/          # Output rendering
│   ├── target/          # Add/remove target logic
│   └── errors/          # Custom error types
├── test/
│   ├── fixtures/        # Test Makefiles and expected outputs
│   └── integration/     # End-to-end tests
└── docs/                # Design documentation
```

## License

[Add your license here]

## Credits

The summary extraction algorithm is a Go port of the [extract-topic](https://www.npmjs.com/package/extract-topic) JavaScript library.

## Support

- **Issues**: Report bugs and request features via GitHub Issues
- **Discussions**: Ask questions in GitHub Discussions
- **Documentation**: See `docs/design.md` for detailed design documentation
