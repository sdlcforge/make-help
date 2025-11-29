## @file
## A basic make setup. Any generated artifacts will be deleted by default on failure.

MAKE_HELP_BIN:=make-help
SRC_FILES:=$(shell find cmd internal -name "*.go")

## Builds the make-help binary.
build: $(MAKE_HELP_BIN)
.PHONY: build

make-help:go.mod go.sum $(SRC_FILES)
	go build -o make-help cmd/make-help/main.go

## Deletes all built artifacts.
clean:
	rm -f $(MAKE_HELP_BIN)
.PHONY: clean

.DELETE_ON_ERROR:

SHELL:=bash

default: all

all: build
.PHONY: all