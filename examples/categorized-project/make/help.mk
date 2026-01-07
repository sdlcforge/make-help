# generated-by: make-help
# command: ./bin/make-help --makefile-path examples/categorized-project/Makefile
# date: 2026-01-07T19:17:46 UTC
# ---
# DO NOT EDIT

MAKE_HELP_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
MAKE_HELP_MAKEFILES := $(MAKE_HELP_DIR)Makefile $(MAKE_HELP_DIR)help.mk

## !category Help
.PHONY: help
## Displays help for available targets.
help:
	@for f in $(MAKE_HELP_MAKEFILES); do \
	  if [ "$$f" -nt "$(MAKE_HELP_DIR)help.mk" ]; then \
	    printf 'Warning: %s is newer than help.mk. Run make update-help to refresh.\n' "$$f"; \
	  fi; \
	done
	@printf '%b\n' "Usage: make [<target>...] [<ENV_VAR>=<value>...]"
	@printf '%b\n' ""
	@printf '%b\n' "Targets:"
	@printf '%b\n' ""
	@printf '%b\n' "Build:"
	@printf '%b\n' "  - build: Builds the production binary with optimizations."
	@printf '%b\n' "    Vars: LDFLAGS Linker flags for build"
	@printf '%b\n' "  - build-all: Cross-compiles for multiple platforms."
	@printf '%b\n' "    Vars: PLATFORMS Target platforms (default: linux/amd64,darwin/amd64)"
	@printf '%b\n' "  - build-debug: Builds with debug symbols and race detector."
	@printf '%b\n' ""
	@printf '%b\n' "Development:"
	@printf '%b\n' "  - generate: Generates code (mocks, protobuf, etc)."
	@printf '%b\n' "  - serve run, dev: Runs the application locally."
	@printf '%b\n' "    Vars: PORT Server port (default: 8080), DEBUG Enable debug mode (default: false)"
	@printf '%b\n' "  - watch: Watches for changes and rebuilds."
	@printf '%b\n' ""
	@printf '%b\n' "Help:"
	@printf '%b\n' "  - help: Displays help for available targets."
	@printf '%b\n' "  - update-help: Regenerates help.mk from source Makefiles."
	@printf '%b\n' ""
	@printf '%b\n' "Maintenance:"
	@printf '%b\n' "  - clean: Removes all build artifacts."
	@printf '%b\n' "  - deps: Updates dependencies."
	@printf '%b\n' "  - lint: Runs linters and formatters."
	@printf '%b\n' "    Vars: LINT_FLAGS Additional golangci-lint flags"
	@printf '%b\n' ""
	@printf '%b\n' "Test:"
	@printf '%b\n' "  - bench: Runs benchmarks."
	@printf '%b\n' "    Vars: BENCH_TIME Duration for each benchmark (default: 3s)"
	@printf '%b\n' "  - test: Runs the full test suite."
	@printf '%b\n' "    Vars: TEST_TIMEOUT Timeout for tests (default: 5m)"
	@printf '%b\n' "  - test-coverage: Runs tests with coverage report."

.PHONY: help-build
help-build:
	@printf '%b\n' "Target: build"
	@printf '%b\n' "Variables:"
	@printf '%b\n' "  - LDFLAGS Linker flags for build"
	@printf '%b\n' ""
	@printf '%b\n' "Builds the production binary with optimizations."
	@printf '%b\n' ""
	@printf '%b\n' "Source: Makefile:13"

.PHONY: help-build-all
help-build-all:
	@printf '%b\n' "Target: build-all"
	@printf '%b\n' "Variables:"
	@printf '%b\n' "  - PLATFORMS Target platforms (default: linux/amd64,darwin/amd64)"
	@printf '%b\n' ""
	@printf '%b\n' "Cross-compiles for multiple platforms."
	@printf '%b\n' "NOTE: Still in \"Build\" category"
	@printf '%b\n' ""
	@printf '%b\n' "Source: Makefile:24"

.PHONY: help-build-debug
help-build-debug:
	@printf '%b\n' "Target: build-debug"
	@printf '%b\n' "Builds with debug symbols and race detector."
	@printf '%b\n' "NOTE: Inherits \"Build\" category from above (no !category needed)"
	@printf '%b\n' ""
	@printf '%b\n' "Source: Makefile:18"

.PHONY: help-generate
help-generate:
	@printf '%b\n' "Target: generate"
	@printf '%b\n' "Generates code (mocks, protobuf, etc)."
	@printf '%b\n' ""
	@printf '%b\n' "Source: Makefile:59"

.PHONY: help-serve
help-serve:
	@printf '%b\n' "Target: serve"
	@printf '%b\n' "Aliases: run, dev"
	@printf '%b\n' "Variables:"
	@printf '%b\n' "  - PORT Server port (default: 8080)"
	@printf '%b\n' "  - DEBUG Enable debug mode (default: false)"
	@printf '%b\n' ""
	@printf '%b\n' "Runs the application locally."
	@printf '%b\n' ""
	@printf '%b\n' "Source: Makefile:49"

.PHONY: help-watch
help-watch:
	@printf '%b\n' "Target: watch"
	@printf '%b\n' "Watches for changes and rebuilds."
	@printf '%b\n' ""
	@printf '%b\n' "Source: Makefile:54"

.PHONY: help-help
help-help:
	@printf '%b\n' "Target: help"
	@printf '%b\n' "Displays help for available targets."
	@printf '%b\n' ""
	@printf '%b\n' "Source: help.mk:10"

.PHONY: help-update-help
help-update-help:
	@printf '%b\n' "Target: update-help"
	@printf '%b\n' "Regenerates help.mk from source Makefiles."
	@printf '%b\n' ""
	@printf '%b\n' "Source: help.mk:156"

.PHONY: help-clean
help-clean:
	@printf '%b\n' "Target: clean"
	@printf '%b\n' "Removes all build artifacts."
	@printf '%b\n' ""
	@printf '%b\n' "Source: Makefile:64"

.PHONY: help-deps
help-deps:
	@printf '%b\n' "Target: deps"
	@printf '%b\n' "Updates dependencies."
	@printf '%b\n' ""
	@printf '%b\n' "Source: Makefile:69"

.PHONY: help-lint
help-lint:
	@printf '%b\n' "Target: lint"
	@printf '%b\n' "Variables:"
	@printf '%b\n' "  - LINT_FLAGS Additional golangci-lint flags"
	@printf '%b\n' ""
	@printf '%b\n' "Runs linters and formatters."
	@printf '%b\n' ""
	@printf '%b\n' "Source: Makefile:75"

.PHONY: help-bench
help-bench:
	@printf '%b\n' "Target: bench"
	@printf '%b\n' "Variables:"
	@printf '%b\n' "  - BENCH_TIME Duration for each benchmark (default: 3s)"
	@printf '%b\n' ""
	@printf '%b\n' "Runs benchmarks."
	@printf '%b\n' ""
	@printf '%b\n' "Source: Makefile:41"

.PHONY: help-test
help-test:
	@printf '%b\n' "Target: test"
	@printf '%b\n' "Variables:"
	@printf '%b\n' "  - TEST_TIMEOUT Timeout for tests (default: 5m)"
	@printf '%b\n' ""
	@printf '%b\n' "Runs the full test suite."
	@printf '%b\n' ""
	@printf '%b\n' "Source: Makefile:30"

.PHONY: help-test-coverage
help-test-coverage:
	@printf '%b\n' "Target: test-coverage"
	@printf '%b\n' "Runs tests with coverage report."
	@printf '%b\n' ""
	@printf '%b\n' "Source: Makefile:35"

# Explicit target to regenerate help.mk
## !category Help
.PHONY: update-help
## Regenerates help.mk from source Makefiles.
update-help:
	@make-help --makefile-path $(MAKE_HELP_DIR)Makefile --no-color || \
	 npx make-help --makefile-path $(MAKE_HELP_DIR)Makefile --no-color || \
	 echo "make-help not found; install with 'go install github.com/sdlcforge/make-help/cmd/make-help@latest' or 'npm install -g make-help'"
