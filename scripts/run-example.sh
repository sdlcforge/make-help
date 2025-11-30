#!/bin/bash
#
# Run an example Makefile's help target from the project root.
#
# Usage: ./scripts/run-example.sh examples/categorized-project [target]
#
# This script sets GOBIN to the project root's .bin directory so the
# make-help binary is installed in a shared location rather than in
# each example's directory.
#
# Arguments:
#   $1 - Path to the example directory (e.g., examples/categorized-project)
#   $2 - Optional: Target to run (default: help)

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <example-dir> [target]"
    echo "Example: $0 examples/categorized-project"
    echo "         $0 examples/full-featured help-build"
    exit 1
fi

EXAMPLE_DIR="$1"
TARGET="${2:-help}"

# Verify example directory exists
if [ ! -d "$EXAMPLE_DIR" ]; then
    echo "Error: Directory '$EXAMPLE_DIR' does not exist"
    exit 1
fi

# Verify Makefile exists
MAKEFILE="$EXAMPLE_DIR/Makefile"
if [ ! -f "$MAKEFILE" ]; then
    echo "Error: Makefile not found at '$MAKEFILE'"
    exit 1
fi

# Set GOBIN to project root's .bin and run make
export GOBIN="$PWD/.bin"
echo "Running: make -f $MAKEFILE $TARGET (GOBIN=$GOBIN)"
make -f "$MAKEFILE" "$TARGET"
