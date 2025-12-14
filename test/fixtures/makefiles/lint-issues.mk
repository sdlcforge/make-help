# Makefile with lint issues

.PHONY: build test setup check documented-no-punct

## !category Build
## Build the project.
build:
	@echo "Building..."

## !category Test
## Run all tests
## This summary has no punctuation at the end
test:
	@echo "Running tests..."

# Undocumented phony target (should trigger warning)
setup:
	@echo "Setting up..."

# Another undocumented phony target
check:
	@echo "Checking..."

## Summary without proper punctuation
documented-no-punct:
	@echo "This target is documented but summary has no punctuation"
