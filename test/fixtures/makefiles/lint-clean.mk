# Makefile with clean lint (no warnings)

.PHONY: build test clean

## !category Build
## Build the project.
build:
	@echo "Building..."

## !category Test
## Run all tests.
test:
	@echo "Running tests..."

## !category Cleanup
## Clean build artifacts.
clean:
	@echo "Cleaning..."

## !alias b
# This is an implicit alias (no recipe, single phony dep)
.PHONY: b
b: build
