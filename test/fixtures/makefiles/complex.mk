## @file
## Complex Makefile with all features
## This demonstrates the full capabilities of make-help.

## @category Build
## @alias b
## @var CC C compiler to use
## @var CFLAGS Compiler flags
## Build the entire project.
## This compiles all sources and links them.
build:
	@echo building

## @category Build
## @alias c
## Compile source files only
compile:
	@echo compiling

## @category Test
## @alias t
## @var TEST_FILTER Filter for test names
## Run all tests.
## Uses go test under the hood.
test:
	@echo testing

## @category Test
## Run integration tests
integration:
	@echo integration

## @category Deploy
## @var ENV Target environment (dev, staging, prod)
## Deploy to environment
deploy:
	@echo deploying

## @category Utility
## Clean build artifacts
clean:
	@echo cleaning

## @category Utility
## @alias h
## Show this help
help:
	@echo help
