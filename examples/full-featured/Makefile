## !file
## Full-Featured Example
##
## This example demonstrates all make-help directives and features:
## - !file for project-level documentation
## - !category for organizing targets (sticky directive)
## - !var for documenting environment variables
## - !alias for target shortcuts
## - Multi-line documentation
##
## NOTE: !category acts as a switch - once set, it applies to all
## subsequent targets until another !category is encountered.
##
## Use `make help` for a summary or `make help-<target>` for details.

## !category Core
## Builds the application with version information embedded.
##
## The build process:
## 1. Generates version info from git
## 2. Compiles with ldflags for version embedding
## 3. Outputs to bin/ directory
##
## !alias b
## !var VERSION Version string (auto-detected from git tag)
## !var COMMIT Git commit hash (auto-detected)
## !var BUILD_DATE Build timestamp (auto-generated)
build:
	@echo "Building version $(VERSION) (commit: $(COMMIT), date: $(BUILD_DATE))"

## Runs the complete test suite with race detection.
##
## Includes unit tests, integration tests, and race condition detection.
## Set INTEGRATION=1 to include integration tests.
##
## NOTE: Inherits "Core" category from above
##
## !alias t
## !var TEST_FLAGS Additional go test flags
## !var INTEGRATION Set to 1 to run integration tests
## !var VERBOSE Set to 1 for verbose test output
test:
	@echo "Running tests with flags: $(TEST_FLAGS)"

## Installs the application to your GOPATH.
## NOTE: Still in "Core" category
## !alias i
install: build
	@echo "Installing application..."

## !category Docker
## Builds the Docker image.
##
## Creates a multi-stage build for minimal image size.
## The image is tagged with both the version and 'latest'.
##
## !var DOCKER_REGISTRY Docker registry URL
## !var DOCKER_IMAGE Image name (default: myapp)
## !var DOCKER_TAG Image tag (default: $(VERSION))
docker-build:
	@echo "Building Docker image: $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)"

## !category Docker
## Pushes the Docker image to the registry.
## !var DOCKER_REGISTRY Registry to push to
docker-push: docker-build
	@echo "Pushing to $(DOCKER_REGISTRY)..."

## !category Docker
## Runs the application in a Docker container.
## !alias docker-run
## !var DOCKER_PORT Host port to bind (default: 8080)
docker-up:
	@echo "Starting container on port $(DOCKER_PORT)..."

## !category Release
## Creates a new release.
##
## This target:
## 1. Runs all tests
## 2. Builds for all platforms
## 3. Creates release archives
## 4. Generates checksums
##
## !var RELEASE_DIR Directory for release artifacts (default: dist)
release:
	@echo "Creating release $(VERSION) in $(RELEASE_DIR)..."

## !category Release
## Publishes release to GitHub.
## !var GITHUB_TOKEN GitHub API token for releases
release-publish: release
	@echo "Publishing release $(VERSION) to GitHub..."

## !category Development
## Starts the development server with hot reload.
##
## Uses air for automatic rebuilds when files change.
## Install air: go install github.com/cosmtrek/air@latest
##
## !alias dev, run, serve
## !var PORT Server port (default: 8080)
## !var LOG_LEVEL Log verbosity (debug, info, warn, error)
dev:
	@echo "Starting development server on port $(PORT) with log level $(LOG_LEVEL)..."

## !category Development
## Opens a REPL for interactive development.
repl:
	@echo "Starting REPL..."

## !category Code Quality
## Runs all linters.
## !var LINT_CONFIG Linter config file (default: .golangci.yml)
lint:
	@echo "Running linters with config: $(LINT_CONFIG)"

## !category Code Quality
## Formats all Go code.
fmt:
	@echo "Formatting code..."

## !category Code Quality
## Checks for security vulnerabilities.
## !alias security
sec:
	@echo "Running security scan..."

## !category Cleanup
## Removes all generated files.
## !alias c
clean:
	@echo "Cleaning build artifacts..."

## !category Cleanup
## Deep clean including caches.
clean-all: clean
	@echo "Deep cleaning including caches..."

.PHONY: build test install docker-build docker-push docker-up release release-publish dev repl lint fmt sec clean clean-all

# Auto-detect version info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Docker settings
DOCKER_REGISTRY ?= ghcr.io/myorg
DOCKER_IMAGE ?= myapp
DOCKER_TAG ?= $(VERSION)
DOCKER_PORT ?= 8080

# Development settings
PORT ?= 8080
LOG_LEVEL ?= info

# Release settings
RELEASE_DIR ?= dist

# Linting
LINT_CONFIG ?= .golangci.yml



-include $(dir $(lastword $(MAKEFILE_LIST)))help.mk

-include make/*.mk
