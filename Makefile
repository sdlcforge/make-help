## !file
## A basic make setup. Any generated artifacts will be deleted by default on failure.

MAKE_HELP_BIN:=bin/make-help
SRC_FILES:=$(shell find cmd internal -name "*.go" -not -name "*_test.go")
VERSION:=$(shell node -p "require('./package.json').version")
LDFLAGS:=-ldflags "-X github.com/sdlcforge/make-help/internal/version.Version=$(VERSION)"

$(MAKE_HELP_BIN): go.mod go.sum $(SRC_FILES) package.json
	@mkdir -p $(dir $@)
	go build $(LDFLAGS) -o $@ cmd/make-help/main.go

## !category Test
## Run unit tests. Use 'test.all' to run all tests.
test.unit:
	go test ./... -race -cover
.PHONY: test.unit

## !alias t
## Run unit tests.
test: test.unit
.PHONY: test

t: test
.PHONY: t

## Run integration tests. Use 'test.all' to run all tests.
test.integration:
	go test -tags=integration ./test/integration/...
.PHONY: test.integration

## Run all tests (unit + integration).
test.all: test.unit test.integration
.PHONY: test.all

## !category Quality
## Run golangci-lint.
lint:
	golangci-lint run
	go vet ./...
	@which -s staticcheck && { echo "Running staticcheck..."; staticcheck ./...; } || { echo "staticcheck not found"; true; }
.PHONY: lint

## Run golangci-lint with auto-fix.
lint-fix:
	golangci-lint run --fix
.PHONY: lint-fix

## Run all quality checks (test.all + lint).
qa: test.all lint
.PHONY: qa

## !category Documentation
DIAGRAM_DIR:=docs/architecture/diagrams
MMD_FILES:=$(wildcard $(DIAGRAM_DIR)/*.mmd)
SVG_FILES:=$(patsubst $(DIAGRAM_DIR)/%.mmd,$(DIAGRAM_DIR)/%.svg,$(MMD_FILES))

## !category Build
## Generate SVG diagrams from Mermaid files.
## Requires mermaid-cli: npm install -g @mermaid-js/mermaid-cli
diagrams: $(SVG_FILES)
.PHONY: diagrams

$(SVG_FILES): $(DIAGRAM_DIR)/%.svg: $(DIAGRAM_DIR)/%.mmd
	npx mmdc -i $< -o $@

## Builds the make-help binary.
build: $(MAKE_HELP_BIN)
.PHONY: build

## Deletes all built artifacts.
clean:
	rm -f $(MAKE_HELP_BIN)
.PHONY: clean

## Remove generated diagram SVG files.
clean.diagrams:
	rm -f $(SVG_FILES)
.PHONY: clean.diagrams

## Deletes all built artifacts and generated diagram SVG files.
clean.all: clean clean.diagrams
.PHONY: clean.all

.DELETE_ON_ERROR:

SHELL:=bash

.DEFAULT_GOAL:=all

## !category Build
## Builds the make-help binary and generated diagram SVG files.
all: build diagrams
.PHONY: all

## Prepares the package for publishing (generates coverage badge, placeholder binary).
prepack:
	./scripts/prepack.sh
.PHONY: prepack

-include make/*.mk
