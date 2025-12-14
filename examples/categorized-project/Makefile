## @file
## Categorized Project Example
##
## This example demonstrates using @category to organize targets
## into logical groups. Categories appear as section headers.
##
## NOTE: @category is a "sticky" directive - once set, it applies
## to all subsequent targets until another @category is encountered.

## @category Build
## Builds the production binary with optimizations.
## @var LDFLAGS Linker flags for build
build:
	@echo "Building for production with LDFLAGS: $(LDFLAGS)"

## Builds with debug symbols and race detector.
## NOTE: Inherits "Build" category from above (no @category needed)
build-debug:
	@echo "Building debug version with race detector..."

## Cross-compiles for multiple platforms.
## @var PLATFORMS Target platforms (default: linux/amd64,darwin/amd64)
## NOTE: Still in "Build" category
build-all:
	@echo "Building for platforms: $(PLATFORMS)"

## @category Test
## Runs the full test suite.
## @var TEST_TIMEOUT Timeout for tests (default: 5m)
test:
	@echo "Running tests with timeout: $(TEST_TIMEOUT)"

## @category Test
## Runs tests with coverage report.
test-coverage:
	@echo "Running tests with coverage..."

## @category Test
## Runs benchmarks.
## @var BENCH_TIME Duration for each benchmark (default: 3s)
bench:
	@echo "Running benchmarks for $(BENCH_TIME)..."

## @category Development
## Runs the application locally.
## @alias run, dev
## @var PORT Server port (default: 8080)
## @var DEBUG Enable debug mode (default: false)
serve:
	@echo "Starting development server on port $(PORT)..."

## @category Development
## Watches for changes and rebuilds.
watch:
	@echo "Watching for changes..."

## @category Development
## Generates code (mocks, protobuf, etc).
generate:
	@echo "Generating code..."

## @category Maintenance
## Removes all build artifacts.
clean:
	@echo "Cleaning build artifacts..."

## @category Maintenance
## Updates dependencies.
deps:
	@echo "Updating dependencies..."

## @category Maintenance
## Runs linters and formatters.
## @var LINT_FLAGS Additional golangci-lint flags
lint:
	@echo "Linting with flags: $(LINT_FLAGS)"

.PHONY: build build-debug build-all test test-coverage bench serve watch generate clean deps lint

# Defaults
PLATFORMS ?= linux/amd64,darwin/amd64,windows/amd64
TEST_TIMEOUT ?= 5m
BENCH_TIME ?= 3s
PORT ?= 8080
DEBUG ?= false


include $(dir $(lastword $(MAKEFILE_LIST)))help.mk
