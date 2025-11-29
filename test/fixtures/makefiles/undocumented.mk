## @file
## Test fixture with documented, undocumented, and .PHONY targets

## Documented target.
## This target has documentation and should always appear.
build:
	@echo "build"

# No documentation - should be hidden by default
clean:
	@echo "clean"

## Test target with docs
test:
	@echo "test"

# Also undocumented - should be hidden by default
lint:
	@echo "lint"

# Undocumented file target
output.txt:
	@echo "hello" > output.txt

.PHONY: build test clean lint
