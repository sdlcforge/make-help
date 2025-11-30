# Makefile with existing help-build target to test conflict detection

## Build target.
build:
	@echo "build"

## Existing help-build target.
help-build:
	@echo "This conflicts with generated help-build target"
