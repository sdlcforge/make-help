package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewScanner(t *testing.T) {
	scanner := NewScanner()
	assert.NotNil(t, scanner)
	assert.Empty(t, scanner.pendingDocs)
	assert.Empty(t, scanner.currentFile)
	assert.Empty(t, scanner.currentCategory)
}

func TestScanContent_FileDirective(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []Directive
	}{
		{
			name: "file directive without description",
			content: `## @file
build:
	echo "building"`,
			expected: []Directive{
				{Type: DirectiveFile, Value: "", SourceFile: "test.mk", LineNumber: 1},
			},
		},
		{
			name: "file directive with description",
			content: `## @file
## This is the main build file
## with multiple lines of documentation
build:
	echo "building"`,
			expected: []Directive{
				{Type: DirectiveFile, Value: "", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "This is the main build file", SourceFile: "test.mk", LineNumber: 2},
				{Type: DirectiveDoc, Value: "with multiple lines of documentation", SourceFile: "test.mk", LineNumber: 3},
			},
		},
		{
			name: "multiple file directives",
			content: `## @file
## First section
## @file
## Second section`,
			expected: []Directive{
				{Type: DirectiveFile, Value: "", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveFile, Value: "", SourceFile: "test.mk", LineNumber: 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner()
			result, err := scanner.ScanContent(tt.content, "test.mk")
			require.NoError(t, err)
			assert.Equal(t, len(tt.expected), len(result.Directives))
			for i, expected := range tt.expected {
				assert.Equal(t, expected.Type, result.Directives[i].Type)
				assert.Equal(t, expected.Value, result.Directives[i].Value)
				assert.Equal(t, expected.SourceFile, result.Directives[i].SourceFile)
				assert.Equal(t, expected.LineNumber, result.Directives[i].LineNumber)
			}
		})
	}
}

func TestScanContent_CategoryDirective(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []Directive
	}{
		{
			name: "single category with target",
			content: `## @category Build
## Build the project
build:
	go build`,
			expected: []Directive{
				{Type: DirectiveCategory, Value: "Build", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "Build the project", SourceFile: "test.mk", LineNumber: 2},
			},
		},
		{
			name: "multiple categories",
			content: `## @category Build
## Build the project
build:
	go build

## @category Test
## Run tests
test:
	go test`,
			expected: []Directive{
				{Type: DirectiveCategory, Value: "Build", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "Build the project", SourceFile: "test.mk", LineNumber: 2},
				{Type: DirectiveCategory, Value: "Test", SourceFile: "test.mk", LineNumber: 6},
				{Type: DirectiveDoc, Value: "Run tests", SourceFile: "test.mk", LineNumber: 7},
			},
		},
		{
			name: "category with multi-word name",
			content: `## @category Build and Deploy
## Build and deploy the project
deploy:
	./deploy.sh`,
			expected: []Directive{
				{Type: DirectiveCategory, Value: "Build and Deploy", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "Build and deploy the project", SourceFile: "test.mk", LineNumber: 2},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner()
			result, err := scanner.ScanContent(tt.content, "test.mk")
			require.NoError(t, err)
			assert.Equal(t, len(tt.expected), len(result.Directives))
			for i, expected := range tt.expected {
				assert.Equal(t, expected.Type, result.Directives[i].Type)
				assert.Equal(t, expected.Value, result.Directives[i].Value)
				assert.Equal(t, expected.LineNumber, result.Directives[i].LineNumber)
			}
		})
	}
}

func TestScanContent_VarDirective(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []Directive
	}{
		{
			name: "var directive with description",
			content: `## @var PORT - The port to listen on
## Start the server
serve:
	./server`,
			expected: []Directive{
				{Type: DirectiveVar, Value: "PORT - The port to listen on", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "Start the server", SourceFile: "test.mk", LineNumber: 2},
			},
		},
		{
			name: "multiple var directives",
			content: `## @var HOST - The hostname
## @var PORT - The port number
## Start the server
serve:
	./server`,
			expected: []Directive{
				{Type: DirectiveVar, Value: "HOST - The hostname", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveVar, Value: "PORT - The port number", SourceFile: "test.mk", LineNumber: 2},
				{Type: DirectiveDoc, Value: "Start the server", SourceFile: "test.mk", LineNumber: 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner()
			result, err := scanner.ScanContent(tt.content, "test.mk")
			require.NoError(t, err)
			assert.Equal(t, len(tt.expected), len(result.Directives))
			for i, expected := range tt.expected {
				assert.Equal(t, expected.Type, result.Directives[i].Type)
				assert.Equal(t, expected.Value, result.Directives[i].Value)
			}
		})
	}
}

func TestScanContent_AliasDirective(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []Directive
	}{
		{
			name: "single alias",
			content: `## @alias b
## Build the project
build:
	go build`,
			expected: []Directive{
				{Type: DirectiveAlias, Value: "b", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "Build the project", SourceFile: "test.mk", LineNumber: 2},
			},
		},
		{
			name: "multiple aliases",
			content: `## @alias b, compile
## Build the project
build:
	go build`,
			expected: []Directive{
				{Type: DirectiveAlias, Value: "b, compile", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "Build the project", SourceFile: "test.mk", LineNumber: 2},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner()
			result, err := scanner.ScanContent(tt.content, "test.mk")
			require.NoError(t, err)
			assert.Equal(t, len(tt.expected), len(result.Directives))
			for i, expected := range tt.expected {
				assert.Equal(t, expected.Type, result.Directives[i].Type)
				assert.Equal(t, expected.Value, result.Directives[i].Value)
			}
		})
	}
}

func TestScanContent_RegularDocumentation(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []Directive
	}{
		{
			name: "single line documentation",
			content: `## Build the project
build:
	go build`,
			expected: []Directive{
				{Type: DirectiveDoc, Value: "Build the project", SourceFile: "test.mk", LineNumber: 1},
			},
		},
		{
			name: "multi-line documentation",
			content: `## Build the project
## This compiles all Go files
## and creates a binary
build:
	go build`,
			expected: []Directive{
				{Type: DirectiveDoc, Value: "Build the project", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "This compiles all Go files", SourceFile: "test.mk", LineNumber: 2},
				{Type: DirectiveDoc, Value: "and creates a binary", SourceFile: "test.mk", LineNumber: 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner()
			result, err := scanner.ScanContent(tt.content, "test.mk")
			require.NoError(t, err)
			assert.Equal(t, len(tt.expected), len(result.Directives))
			for i, expected := range tt.expected {
				assert.Equal(t, expected.Type, result.Directives[i].Type)
				assert.Equal(t, expected.Value, result.Directives[i].Value)
			}
		})
	}
}

func TestScanContent_TargetDetection(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		targetMap map[string]int
	}{
		{
			name: "simple target",
			content: `## Build the project
build:
	go build`,
			targetMap: map[string]int{"build": 2},
		},
		{
			name: "grouped target operator",
			content: `## Build the project
build&:
	go build`,
			targetMap: map[string]int{"build": 2},
		},
		{
			name: "multiple targets on one line",
			content: `## Build all
all build compile:
	go build`,
			targetMap: map[string]int{"all": 2},
		},
		{
			name: "variable target",
			content: `## Variable target
$(VAR):
	echo "var"`,
			targetMap: map[string]int{"$(VAR)": 2},
		},
		{
			name: "target with prerequisites",
			content: `## Build with deps
build: deps lint
	go build`,
			targetMap: map[string]int{"build": 2},
		},
		{
			name: "multiple targets",
			content: `## Build
build:
	go build

## Test
test:
	go test`,
			targetMap: map[string]int{"build": 2, "test": 6},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner()
			result, err := scanner.ScanContent(tt.content, "test.mk")
			require.NoError(t, err)
			assert.Equal(t, tt.targetMap, result.TargetMap)
		})
	}
}

func TestScanContent_PendingDocsAssociation(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		wantDirectives int
	}{
		{
			name: "docs associated with target",
			content: `## Build the project
build:
	go build`,
			wantDirectives: 1, // 1 doc directive
		},
		{
			name: "docs cleared by blank line",
			content: `## This documentation is orphaned

build:
	go build`,
			wantDirectives: 0, // Blank line clears pending docs
		},
		{
			name: "docs cleared by non-doc line",
			content: `## This documentation is orphaned
VARIABLE := value
build:
	go build`,
			wantDirectives: 0, // Variable assignment clears pending docs
		},
		{
			name: "category and docs associated",
			content: `## @category Build
## Build the project
build:
	go build`,
			wantDirectives: 2, // Category + doc
		},
		{
			name: "recipe lines do not clear docs",
			content: `## Build the project
build:
	go build
	go test`,
			wantDirectives: 1, // Doc is associated with build target
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner()
			result, err := scanner.ScanContent(tt.content, "test.mk")
			require.NoError(t, err)
			assert.Equal(t, tt.wantDirectives, len(result.Directives))
		})
	}
}

func TestScanContent_RecipeLinesSkipped(t *testing.T) {
	content := `## Build the project
build:
	go build
	go test
	@echo "Done"

## Clean build artifacts
clean:
	rm -rf build/`

	scanner := NewScanner()
	result, err := scanner.ScanContent(content, "test.mk")
	require.NoError(t, err)

	// Should have 2 targets
	assert.Equal(t, 2, len(result.TargetMap))
	assert.Contains(t, result.TargetMap, "build")
	assert.Contains(t, result.TargetMap, "clean")

	// Should have 2 documentation directives
	assert.Equal(t, 2, len(result.Directives))
}

func TestScanContent_ComplexMakefile(t *testing.T) {
	content := `## @file
## Main build file for the project

## @category Build
## Build the application
## @var GO_VERSION - Go version to use
## @alias b, compile
build:
	go build -o app

## @category Test
## Run unit tests
## @var COVERAGE - Enable coverage report
test:
	go test ./...

## @category Deploy
## Deploy to production
## @var ENV - Target environment
deploy:
	./deploy.sh`

	scanner := NewScanner()
	result, err := scanner.ScanContent(content, "test.mk")
	require.NoError(t, err)

	// Verify targets
	assert.Equal(t, 3, len(result.TargetMap))
	assert.Contains(t, result.TargetMap, "build")
	assert.Contains(t, result.TargetMap, "test")
	assert.Contains(t, result.TargetMap, "deploy")

	// Count directives by type
	var fileCount, catCount, varCount, aliasCount, docCount int
	for _, d := range result.Directives {
		switch d.Type {
		case DirectiveFile:
			fileCount++
		case DirectiveCategory:
			catCount++
		case DirectiveVar:
			varCount++
		case DirectiveAlias:
			aliasCount++
		case DirectiveDoc:
			docCount++
		}
	}

	assert.Equal(t, 1, fileCount)    // 1 @file
	assert.Equal(t, 3, catCount)     // 3 @category
	assert.Equal(t, 3, varCount)     // 3 @var
	assert.Equal(t, 1, aliasCount)   // 1 @alias
	assert.Equal(t, 3, docCount)     // 3 regular doc lines (one per target)
}

func TestScanContent_EmptyContent(t *testing.T) {
	scanner := NewScanner()
	result, err := scanner.ScanContent("", "test.mk")
	require.NoError(t, err)
	assert.Empty(t, result.Directives)
	assert.Empty(t, result.TargetMap)
}

func TestScanContent_NoDocumentation(t *testing.T) {
	content := `build:
	go build

test:
	go test`

	scanner := NewScanner()
	result, err := scanner.ScanContent(content, "test.mk")
	require.NoError(t, err)

	// Should have targets but no directives
	assert.Equal(t, 2, len(result.TargetMap))
	assert.Empty(t, result.Directives)
}

func TestScanContent_OnlyDocumentation(t *testing.T) {
	content := `## @file
## This is a documentation-only file
## With multiple lines`

	scanner := NewScanner()
	result, err := scanner.ScanContent(content, "test.mk")
	require.NoError(t, err)

	// Should have directives but no targets
	assert.Equal(t, 1, len(result.Directives))
	assert.Empty(t, result.TargetMap)
}

func TestScanFile_FileNotFound(t *testing.T) {
	scanner := NewScanner()
	_, err := scanner.ScanFile("/nonexistent/file.mk")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read")
}

func TestScanContent_SourceFileTracking(t *testing.T) {
	content := `## @category Build
## Build the project
build:
	go build`

	scanner := NewScanner()
	result, err := scanner.ScanContent(content, "/path/to/Makefile")
	require.NoError(t, err)

	// All directives should have correct source file
	for _, d := range result.Directives {
		assert.Equal(t, "/path/to/Makefile", d.SourceFile)
	}

	// Result should have correct path
	assert.Equal(t, "/path/to/Makefile", result.Path)
}

func TestScanContent_LineNumberTracking(t *testing.T) {
	content := `## Line 1
## Line 2
build:
	go build
## Line 5
test:
	go test`

	scanner := NewScanner()
	result, err := scanner.ScanContent(content, "test.mk")
	require.NoError(t, err)

	// Check line numbers
	assert.Equal(t, 1, result.Directives[0].LineNumber)
	assert.Equal(t, 2, result.Directives[1].LineNumber)
	assert.Equal(t, 5, result.Directives[2].LineNumber)

	// Check target line numbers
	assert.Equal(t, 3, result.TargetMap["build"])
	assert.Equal(t, 6, result.TargetMap["test"])
}

func TestScanContent_StatePersistenceAcrossScans(t *testing.T) {
	scanner := NewScanner()

	// First scan
	content1 := `## @category Build
## Build
build:
	go build`
	result1, err := scanner.ScanContent(content1, "file1.mk")
	require.NoError(t, err)
	assert.Equal(t, "file1.mk", result1.Path)

	// Second scan - should reset state
	content2 := `## Test
test:
	go test`
	result2, err := scanner.ScanContent(content2, "file2.mk")
	require.NoError(t, err)
	assert.Equal(t, "file2.mk", result2.Path)

	// Verify state was reset
	assert.Empty(t, scanner.pendingDocs)
	assert.Equal(t, "file2.mk", scanner.currentFile)
	assert.Equal(t, "", scanner.currentCategory) // Should be reset
}
