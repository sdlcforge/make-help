## !file
## Filtering Demo
##
## This example demonstrates target filtering with --include-target
## and --include-all-phony flags.
##
## Try these commands:
##   make-help --makefile-path examples/filtering-demo/Makefile
##   make-help --makefile-path examples/filtering-demo/Makefile --include-target setup,check
##   make-help --makefile-path examples/filtering-demo/Makefile --include-all-phony

## !category Build
## Compiles the project.
## This is a documented target - always visible.
build:
	@echo "Building..."

## Runs tests.
## Also documented - always visible.
## NOTE: Inherits "Build" category from above
test:
	@echo "Testing..."

## !category Deploy
## Deploys to production.
## Documented and categorized.
## !var DEPLOY_ENV Target environment (staging, production)
deploy:
	@echo "Deploying to $(DEPLOY_ENV)..."

# ============================================================
# UNDOCUMENTED TARGETS BELOW
# These only appear with --include-target or --include-all-phony
# ============================================================

# No ## comment - hidden by default
# Use: --include-target setup
setup:
	@echo "Setting up development environment..."
	@echo "Installing tools..."

# No ## comment - hidden by default
# Use: --include-target check
check:
	@echo "Running pre-commit checks..."

# No ## comment - hidden by default
# Use: --include-target format
format:
	@echo "Formatting code..."

# Internal target - not meant for users
# Still accessible via --include-all-phony
_internal-task:
	@echo "Internal task..."

# Another internal target
_cleanup-temp:
	@echo "Cleaning temp files..."

.PHONY: build test deploy setup check format _internal-task _cleanup-temp

DEPLOY_ENV ?= staging



include $(dir $(lastword $(MAKEFILE_LIST)))help.mk
