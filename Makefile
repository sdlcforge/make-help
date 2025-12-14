## !file
## A basic make setup. Any generated artifacts will be deleted by default on failure.

MAKE_HELP_BIN:=make-help
SRC_FILES:=$(shell find cmd internal -name "*.go")

## !category Build
##
## Builds the make-help binary.
build: $(MAKE_HELP_BIN)
.PHONY: build

make-help:go.mod go.sum $(SRC_FILES)
	go build -o make-help cmd/make-help/main.go

## Deletes all built artifacts.
clean:
	rm -f $(MAKE_HELP_BIN)
.PHONY: clean

## !category Test
##
## Run unit tests (excludes integration tests)
test.unit:
	go test ./...
.PHONY: test.unit

## !alias t
##
## Run unit tests
test: test.unit
.PHONY: test

t: test
.PHONY: t

## Run integration tests only
test.integration:
	go test -tags=integration ./test/integration/...
.PHONY: test.integration

## Run all tests (unit + integration)
test.all: test.unit test.integration
.PHONY: test.all

## !category Quality
##
## Run golangci-lint
lint:
	golangci-lint run
.PHONY: lint

## Run golangci-lint with auto-fix
lint-fix:
	golangci-lint run --fix
.PHONY: lint-fix

## Run all quality checks (tests + lint)
qa: test.all lint
.PHONY: qa

.DELETE_ON_ERROR:

SHELL:=bash

default: all

all: build
.PHONY: all