## !file
## Simple project without categories.
##
## This example shows basic target documentation without using !category.
## All targets appear in a flat list.

## Builds the application binary.
## Creates an optimized build in the ./bin directory.
## !var BUILD_FLAGS Additional flags to pass to the compiler
## !var OUTPUT_DIR Output directory for the binary (default: ./bin)
build:
	@echo "Building application with flags: $(BUILD_FLAGS)"
	@echo "Output directory: $(OUTPUT_DIR)"

## Runs the test suite.
## !var TEST_FLAGS Additional flags for go test
## !var COVERAGE Enable coverage reporting (set to 1)
test:
	@echo "Running tests with flags: $(TEST_FLAGS)"

## Removes build artifacts and temporary files.
clean:
	@echo "Cleaning build artifacts..."

## Installs the application to GOPATH/bin.
install: build
	@echo "Installing application..."

## Runs the application in development mode.
## !alias run, dev
## !var PORT Port to listen on (default: 8080)
serve:
	@echo "Starting server on port $(PORT)..."

.PHONY: build test clean install serve

# Default values
OUTPUT_DIR ?= ./bin
PORT ?= 8080



-include $(dir $(lastword $(MAKEFILE_LIST)))help.mk
