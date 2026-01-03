# generated-by: make-help
# command: ./bin/make-help
# date: 2026-01-03T20:58:45 UTC
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
	    printf '\033[0;33mWarning: %s is newer than help.mk. Run make update-help to refresh.\033[0m\n' "$$f"; \
	  fi; \
	done
	@printf '%b\n' "Usage: make [<target>...] [<ENV_VAR>=<value>...]"
	@printf '%b\n' ""
	@printf '%b\n' "Targets:"
	@printf '%b\n' ""
	@printf '%b\n' "\033[1;36mBuild:\033[0m"
	@printf '%b\n' "  - \033[1;32mall\033[0m: \033[0;37mBuilds the make-help binary and generated diagram SVG files.\033[0m"
	@printf '%b\n' "  - \033[1;32mbuild\033[0m: \033[0;37mBuilds the make-help binary.\033[0m"
	@printf '%b\n' "  - \033[1;32mclean\033[0m: \033[0;37mDeletes all built artifacts.\033[0m"
	@printf '%b\n' "  - \033[1;32mclean.all\033[0m: \033[0;37mDeletes all built artifacts and generated diagram SVG files.\033[0m"
	@printf '%b\n' "  - \033[1;32mclean.diagrams\033[0m: \033[0;37mRemove generated diagram SVG files.\033[0m"
	@printf '%b\n' "  - \033[1;32mdiagrams\033[0m: \033[0;37mGenerate SVG diagrams from Mermaid files.\033[0m"
	@printf '%b\n' ""
	@printf '%b\n' "\033[1;36mHelp:\033[0m"
	@printf '%b\n' "  - \033[1;32mhelp\033[0m: \033[0;37mDisplays help for available targets.\033[0m"
	@printf '%b\n' "  - \033[1;32mupdate-help\033[0m: \033[0;37mRegenerates help.mk from source Makefiles.\033[0m"
	@printf '%b\n' ""
	@printf '%b\n' "\033[1;36mQuality:\033[0m"
	@printf '%b\n' "  - \033[1;32mlint\033[0m: \033[0;37mRun golangci-lint.\033[0m"
	@printf '%b\n' "  - \033[1;32mlint-fix\033[0m: \033[0;37mRun golangci-lint with auto-fix.\033[0m"
	@printf '%b\n' "  - \033[1;32mqa\033[0m: \033[0;37mRun all quality checks (test.all + lint).\033[0m"
	@printf '%b\n' ""
	@printf '%b\n' "\033[1;36mTest:\033[0m"
	@printf '%b\n' "  - \033[1;32mtest\033[0m \033[0;33mt, t\033[0m: \033[0;37mRun unit tests.\033[0m"
	@printf '%b\n' "  - \033[1;32mtest.all\033[0m: \033[0;37mRun all tests (unit + integration).\033[0m"
	@printf '%b\n' "  - \033[1;32mtest.integration\033[0m: \033[0;37mRun integration tests.\033[0m"
	@printf '%b\n' "  - \033[1;32mtest.unit\033[0m: \033[0;37mRun unit tests.\033[0m"

.PHONY: help-all
help-all:
	@printf '%b\n' "\033[1;32mTarget: all\033[0m"
	@printf '%b\n' "\033[0;37mBuilds the make-help binary and generated diagram SVG files.\033[0m"
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:91"

.PHONY: help-build
help-build:
	@printf '%b\n' "\033[1;32mTarget: build\033[0m"
	@printf '%b\n' "\033[0;37mBuilds the make-help binary.\033[0m"
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:66"

.PHONY: help-clean
help-clean:
	@printf '%b\n' "\033[1;32mTarget: clean\033[0m"
	@printf '%b\n' "\033[0;37mDeletes all built artifacts.\033[0m"
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:70"

.PHONY: help-clean.all
help-clean.all:
	@printf '%b\n' "\033[1;32mTarget: clean.all\033[0m"
	@printf '%b\n' "\033[0;37mDeletes all built artifacts and generated diagram SVG files.\033[0m"
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:80"

.PHONY: help-clean.diagrams
help-clean.diagrams:
	@printf '%b\n' "\033[1;32mTarget: clean.diagrams\033[0m"
	@printf '%b\n' "\033[0;37mRemove generated diagram SVG files.\033[0m"
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:75"

.PHONY: help-diagrams
help-diagrams:
	@printf '%b\n' "\033[1;32mTarget: diagrams\033[0m"
	@printf '%b\n' "\033[0;37mGenerate SVG diagrams from Mermaid files.\033[0m"
	@printf '%b\n' "\033[0;37mRequires mermaid-cli: npm install -g @mermaid-js/mermaid-cli\033[0m"
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:59"

.PHONY: help-help
help-help:
	@printf '%b\n' "\033[1;32mTarget: help\033[0m"
	@printf '%b\n' "\033[0;37mDisplays help for available targets.\033[0m"
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/make/00-help.mk:13"

.PHONY: help-update-help
help-update-help:
	@printf '%b\n' "\033[1;32mTarget: update-help\033[0m"
	@printf '%b\n' "\033[0;37mRegenerates help.mk from source Makefiles.\033[0m"
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/make/00-help.mk:157"

.PHONY: help-lint
help-lint:
	@printf '%b\n' "\033[1;32mTarget: lint\033[0m"
	@printf '%b\n' "\033[0;37mRun golangci-lint.\033[0m"
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:38"

.PHONY: help-lint-fix
help-lint-fix:
	@printf '%b\n' "\033[1;32mTarget: lint-fix\033[0m"
	@printf '%b\n' "\033[0;37mRun golangci-lint with auto-fix.\033[0m"
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:43"

.PHONY: help-qa
help-qa:
	@printf '%b\n' "\033[1;32mTarget: qa\033[0m"
	@printf '%b\n' "\033[0;37mRun all quality checks (test.all + lint).\033[0m"
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:48"

.PHONY: help-test
help-test:
	@printf '%b\n' "\033[1;32mTarget: test\033[0m"
	@printf '%b\n' "\033[0;33mAliases: t, t\033[0m"
	@printf '%b\n' "\033[0;37mRun unit tests.\033[0m"
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:21"

.PHONY: help-test.all
help-test.all:
	@printf '%b\n' "\033[1;32mTarget: test.all\033[0m"
	@printf '%b\n' "\033[0;37mRun all tests (unit + integration).\033[0m"
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:33"

.PHONY: help-test.integration
help-test.integration:
	@printf '%b\n' "\033[1;32mTarget: test.integration\033[0m"
	@printf '%b\n' "\033[0;37mRun integration tests. Use 'test.all' to run all tests.\033[0m"
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:28"

.PHONY: help-test.unit
help-test.unit:
	@printf '%b\n' "\033[1;32mTarget: test.unit\033[0m"
	@printf '%b\n' "\033[0;37mRun unit tests. Use 'test.all' to run all tests.\033[0m"
	@printf '%b\n' ""
	@printf '%b\n' "Source: /Users/zane/playground/sdlcforge/make-help/Makefile:15"

# Explicit target to regenerate help.mk
## !category Help
.PHONY: update-help
## Regenerates help.mk from source Makefiles.
update-help:
	@make-help --makefile-path $(MAKE_HELP_DIR)Makefile || \
	 npx make-help --makefile-path $(MAKE_HELP_DIR)Makefile || \
	 echo "make-help not found; install with 'go install github.com/sdlcforge/make-help/cmd/make-help@latest' or 'npm install -g make-help'"
