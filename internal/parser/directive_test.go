package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsDocumentationLine(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{
			name:     "documentation line",
			line:     "## This is documentation",
			expected: true,
		},
		{
			name:     "documentation line with directive",
			line:     "## !category Build",
			expected: true,
		},
		{
			name:     "documentation line with only prefix",
			line:     "## ",
			expected: true,
		},
		{
			name:     "single hash",
			line:     "# Not documentation",
			expected: false,
		},
		{
			name:     "triple hash",
			line:     "### Also not documentation",
			expected: false,
		},
		{
			name:     "no hash",
			line:     "Regular line",
			expected: false,
		},
		{
			name:     "hash hash without space",
			line:     "##No space",
			expected: false,
		},
		{
			name:     "empty line",
			line:     "",
			expected: false,
		},
		{
			name:     "whitespace only",
			line:     "   ",
			expected: false,
		},
		{
			name:     "target line",
			line:     "build:",
			expected: false,
		},
		{
			name:     "recipe line",
			line:     "\tgo build",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := IsDocumentationLine(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsTargetLine(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{
			name:     "simple target",
			line:     "build:",
			expected: true,
		},
		{
			name:     "target with prerequisites",
			line:     "build: deps lint",
			expected: true,
		},
		{
			name:     "grouped target operator",
			line:     "build&:",
			expected: true,
		},
		{
			name:     "multiple targets",
			line:     "all build compile:",
			expected: true,
		},
		{
			name:     "variable target",
			line:     "$(VAR):",
			expected: true,
		},
		{
			name:     "target with recipe on same line",
			line:     "build: ; go build",
			expected: true,
		},
		{
			name:     "recipe line with tab",
			line:     "\tgo build",
			expected: false,
		},
		{
			name:     "recipe line with spaces",
			line:     "    go build",
			expected: false,
		},
		{
			name:     "variable assignment with colon",
			line:     "PATH := /usr/bin:/usr/local/bin",
			expected: true, // Contains colon, will be caught by ExtractTargetName
		},
		{
			name:     "comment line",
			line:     "# This is a comment",
			expected: false,
		},
		{
			name:     "documentation line",
			line:     "## Build the project",
			expected: false,
		},
		{
			name:     "empty line",
			line:     "",
			expected: false,
		},
		{
			name:     "no colon",
			line:     "just text",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := IsTargetLine(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractTargetName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "simple target",
			line:     "build:",
			expected: "build",
		},
		{
			name:     "target with whitespace",
			line:     "  build:  ",
			expected: "",
		},
		{
			name:     "target with prerequisites",
			line:     "build: deps lint",
			expected: "build",
		},
		{
			name:     "grouped target operator",
			line:     "build&:",
			expected: "build",
		},
		{
			name:     "grouped target operator with space",
			line:     "build &:",
			expected: "build",
		},
		{
			name:     "multiple targets",
			line:     "all build compile:",
			expected: "all",
		},
		{
			name:     "multiple targets with tabs",
			line:     "all\tbuild\tcompile:",
			expected: "all",
		},
		{
			name:     "variable target",
			line:     "$(VAR):",
			expected: "$(VAR)",
		},
		{
			name:     "complex variable target",
			line:     "$(shell uname):",
			expected: "$(shell",
		},
		{
			name:     "target with recipe on same line",
			line:     "build: ; go build",
			expected: "build",
		},
		{
			name:     "target with whitespace before colon",
			line:     "build :",
			expected: "build",
		},
		{
			name:     "recipe line with tab",
			line:     "\tgo build",
			expected: "",
		},
		{
			name:     "recipe line with spaces",
			line:     "    go build",
			expected: "",
		},
		{
			name:     "no colon",
			line:     "just text",
			expected: "",
		},
		{
			name:     "empty line",
			line:     "",
			expected: "",
		},
		{
			name:     "colon at start",
			line:     ": invalid",
			expected: "",
		},
		{
			name:     "PHONY target",
			line:     ".PHONY: build test",
			expected: ".PHONY",
		},
		{
			name:     "special variable assignment",
			line:     ".DEFAULT_GOAL := build",
			expected: "", // This is a variable assignment, not a target
		},
		{
			name:     "pattern rule",
			line:     "%.o: %.c",
			expected: "%.o",
		},
		{
			name:     "static pattern rule",
			line:     "objects = foo.o bar.o",
			expected: "",
		},
		{
			name:     "variable assignment :=",
			line:     "VAR := value",
			expected: "",
		},
		{
			name:     "variable assignment ?=",
			line:     "VAR ?= value",
			expected: "",
		},
		{
			name:     "variable assignment +=",
			line:     "VAR += value",
			expected: "",
		},
		{
			name:     "variable assignment !=",
			line:     "VAR != value",
			expected: "",
		},
		{
			name:     "recursive variable with ::=",
			line:     "VAR ::= value",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ExtractTargetName(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractTargetName_EdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "unicode target name",
			line:     "日本語:",
			expected: "日本語",
		},
		{
			name:     "target with special characters",
			line:     "build-all:",
			expected: "build-all",
		},
		{
			name:     "target with underscore",
			line:     "build_all:",
			expected: "build_all",
		},
		{
			name:     "target with dots",
			line:     "build.all:",
			expected: "build.all",
		},
		{
			name:     "target with slash",
			line:     "src/build:",
			expected: "src/build",
		},
		{
			name:     "very long target name",
			line:     "this-is-a-very-long-target-name-that-should-still-work:",
			expected: "this-is-a-very-long-target-name-that-should-still-work",
		},
		{
			name:     "target with equals in prerequisites",
			line:     "build: VAR=value",
			expected: "build",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ExtractTargetName(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractTargetName_RealWorldExamples(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "go build target",
			line:     "build: clean",
			expected: "build",
		},
		{
			name:     "docker compose target",
			line:     "docker-up:",
			expected: "docker-up",
		},
		{
			name:     "install target with path",
			line:     "install: /usr/local/bin",
			expected: "install",
		},
		{
			name:     "test with coverage",
			line:     "test-coverage: test",
			expected: "test-coverage",
		},
		{
			name:     "clean with force",
			line:     "clean-all:",
			expected: "clean-all",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ExtractTargetName(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}
