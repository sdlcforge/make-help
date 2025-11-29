## @file
## Makefile with documented and undocumented targets

## Build the project.
## This compiles all source files and generates
## the binary in the output directory.
## @alias b, compile
## @var BUILD_FLAGS - Flags passed to go build
## @var OUTPUT_DIR - Directory for build output
build:
	@echo building

# This target has no documentation
undocumented:
	@echo undocumented

## Run all tests
test:
	@echo testing
