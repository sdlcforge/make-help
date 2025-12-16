# generated-by: make-help
# command: ./bin/make-help
# date: 2025-12-16T18:31:10 UTC
# ---
# DO NOT EDIT

MAKE_HELP_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
MAKE_HELP_MAKEFILES := $(MAKE_HELP_DIR)Makefile

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
	@printf '%b\n' "  - all: Builds the make-help binary and generated diagram SVG files."
	@printf '%b\n' "  - build: Builds the make-help binary."
	@printf '%b\n' "  - clean: Deletes all built artifacts."
	@printf '%b\n' "  - clean.all: Deletes all built artifacts and generated diagram SVG files."
	@printf '%b\n' "  - clean.diagrams: Remove generated diagram SVG files."
	@printf '%b\n' "  - diagrams: Generate SVG diagrams from Mermaid files."
	@printf '%b\n' ""
	@printf '%b\n' "Quality:"
	@printf '%b\n' "  - lint: Run golangci-lint."
	@printf '%b\n' "  - lint-fix: Run golangci-lint with auto-fix."
	@printf '%b\n' "  - qa: Run all quality checks (test.all + lint)."
	@printf '%b\n' ""
	@printf '%b\n' "Test:"
	@printf '%b\n' "  - test t, t: Run unit tests."
	@printf '%b\n' "  - test.all: Run all tests (unit + integration)."
	@printf '%b\n' "  - test.integration: Run integration tests."
	@printf '%b\n' "  - test.unit: Run unit tests."

.PHONY: help-all
help-all:
	@printf '%b\n' "Target: all"
	@printf '%b\n' "Builds the make-help binary and generated diagram SVG files."
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:91"

.PHONY: help-build
help-build:
	@printf '%b\n' "Target: build"
	@printf '%b\n' "Builds the make-help binary."
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:66"

.PHONY: help-clean
help-clean:
	@printf '%b\n' "Target: clean"
	@printf '%b\n' "Deletes all built artifacts."
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:70"

.PHONY: help-clean.all
help-clean.all:
	@printf '%b\n' "Target: clean.all"
	@printf '%b\n' "Deletes all built artifacts and generated diagram SVG files."
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:80"

.PHONY: help-clean.diagrams
help-clean.diagrams:
	@printf '%b\n' "Target: clean.diagrams"
	@printf '%b\n' "Remove generated diagram SVG files."
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:75"

.PHONY: help-diagrams
help-diagrams:
	@printf '%b\n' "Target: diagrams"
	@printf '%b\n' "Generate SVG diagrams from Mermaid files."
	@printf '%b\n' "Requires mermaid-cli: npm install -g @mermaid-js/mermaid-cli"
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:59"

.PHONY: help-lint
help-lint:
	@printf '%b\n' "Target: lint"
	@printf '%b\n' "Run golangci-lint."
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:38"

.PHONY: help-lint-fix
help-lint-fix:
	@printf '%b\n' "Target: lint-fix"
	@printf '%b\n' "Run golangci-lint with auto-fix."
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:43"

.PHONY: help-qa
help-qa:
	@printf '%b\n' "Target: qa"
	@printf '%b\n' "Run all quality checks (test.all + lint)."
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:48"

.PHONY: help-test
help-test:
	@printf '%b\n' "Target: test"
	@printf '%b\n' "Aliases: t, t"
	@printf '%b\n' "Run unit tests."
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:21"

.PHONY: help-test.all
help-test.all:
	@printf '%b\n' "Target: test.all"
	@printf '%b\n' "Run all tests (unit + integration)."
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:33"

.PHONY: help-test.integration
help-test.integration:
	@printf '%b\n' "Target: test.integration"
	@printf '%b\n' "Run integration tests. Use 'test.all' to run all tests."
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:28"

.PHONY: help-test.unit
help-test.unit:
	@printf '%b\n' "Target: test.unit"
	@printf '%b\n' "Run unit tests. Use 'test.all' to run all tests."
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:15"

# Explicit target to regenerate help.mk
## !category Help
.PHONY: update-help
## Regenerates help.mk from source Makefiles.
update-help:
	@make-help --makefile-path $(MAKE_HELP_DIR)Makefile --no-color || \
	 npx make-help --makefile-path $(MAKE_HELP_DIR)Makefile --no-color || \
	 echo "make-help not found; install with 'go install github.com/sdlcforge/make-help/cmd/make-help@latest' or 'npm install -g make-help'"
