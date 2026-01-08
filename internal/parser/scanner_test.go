package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewScanner(t *testing.T) {
	t.Parallel()
	scanner := NewScanner()
	assert.NotNil(t, scanner)
	assert.Empty(t, scanner.pendingDocs)
	assert.Empty(t, scanner.currentFile)
}

func TestScanContent_FileDirective(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		content  string
		expected []Directive
	}{
		{
			name: "file directive without description",
			content: `## !file
build:
	echo "building"`,
			expected: []Directive{
				{Type: DirectiveFile, Value: "", SourceFile: "test.mk", LineNumber: 1},
			},
		},
		{
			name: "file directive with description",
			content: `## !file
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
			content: `## !file
## First section
## !file
## Second section`,
			expected: []Directive{
				{Type: DirectiveFile, Value: "", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveFile, Value: "", SourceFile: "test.mk", LineNumber: 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
	t.Parallel()
	tests := []struct {
		name     string
		content  string
		expected []Directive
	}{
		{
			name: "single category with target",
			content: `## !category Build
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
			content: `## !category Build
## Build the project
build:
	go build

## !category Test
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
			content: `## !category Build and Deploy
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
			t.Parallel()
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
	t.Parallel()
	tests := []struct {
		name     string
		content  string
		expected []Directive
	}{
		{
			name: "var directive with description",
			content: `## !var PORT - The port to listen on
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
			content: `## !var HOST - The hostname
## !var PORT - The port number
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
			t.Parallel()
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
	t.Parallel()
	tests := []struct {
		name     string
		content  string
		expected []Directive
	}{
		{
			name: "single alias",
			content: `## !alias b
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
			content: `## !alias b, compile
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
			t.Parallel()
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

func TestScanContent_NotAliasDirective(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		content  string
		expected []Directive
	}{
		{
			name: "notalias directive",
			content: `## !notalias
build:
	go build`,
			expected: []Directive{
				{Type: DirectiveNotAlias, Value: "", SourceFile: "test.mk", LineNumber: 1},
			},
		},
		{
			name: "notalias with trailing content ignored",
			content: `## !notalias some extra text
build:
	go build`,
			expected: []Directive{
				{Type: DirectiveNotAlias, Value: "", SourceFile: "test.mk", LineNumber: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
	t.Parallel()
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
			t.Parallel()
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
	t.Parallel()
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
			t.Parallel()
			scanner := NewScanner()
			result, err := scanner.ScanContent(tt.content, "test.mk")
			require.NoError(t, err)
			assert.Equal(t, tt.targetMap, result.TargetMap)
		})
	}
}

func TestScanContent_PendingDocsAssociation(t *testing.T) {
	t.Parallel()
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
			content: `## !category Build
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
			t.Parallel()
			scanner := NewScanner()
			result, err := scanner.ScanContent(tt.content, "test.mk")
			require.NoError(t, err)
			assert.Equal(t, tt.wantDirectives, len(result.Directives))
		})
	}
}

func TestScanContent_RecipeLinesSkipped(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	content := `## !file
## Main build file for the project

## !category Build
## Build the application
## !var GO_VERSION - Go version to use
## !alias b, compile
build:
	go build -o app

## !category Test
## Run unit tests
## !var COVERAGE - Enable coverage report
test:
	go test ./...

## !category Deploy
## Deploy to production
## !var ENV - Target environment
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

	assert.Equal(t, 1, fileCount)    // 1 !file
	assert.Equal(t, 3, catCount)     // 3 !category
	assert.Equal(t, 3, varCount)     // 3 !var
	assert.Equal(t, 1, aliasCount)   // 1 !alias
	assert.Equal(t, 3, docCount)     // 3 regular doc lines (one per target)
}

func TestScanContent_EmptyContent(t *testing.T) {
	t.Parallel()
	scanner := NewScanner()
	result, err := scanner.ScanContent("", "test.mk")
	require.NoError(t, err)
	assert.Empty(t, result.Directives)
	assert.Empty(t, result.TargetMap)
}

func TestScanContent_NoDocumentation(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	content := `## !file
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
	t.Parallel()
	scanner := NewScanner()
	_, err := scanner.ScanFile("/nonexistent/file.mk")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read")
}

func TestScanContent_SourceFileTracking(t *testing.T) {
	t.Parallel()
	content := `## !category Build
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
	t.Parallel()
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
	t.Parallel()
	scanner := NewScanner()

	// First scan
	content1 := `## !category Build
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
}

// TestScanContent_UnicodeHandling tests that the parser handles non-ASCII characters correctly
// in target names, documentation, and directives.
func TestScanContent_UnicodeHandling(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		content  string
		expected []Directive
		targets  map[string]int
	}{
		{
			name: "unicode in documentation",
			content: `## Compilar el proyecto 🚀
## Español: construir la aplicación
build:
	go build`,
			expected: []Directive{
				{Type: DirectiveDoc, Value: "Compilar el proyecto 🚀", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "Español: construir la aplicación", SourceFile: "test.mk", LineNumber: 2},
			},
			targets: map[string]int{"build": 3},
		},
		{
			name: "unicode in category names",
			content: `## !category 构建工具
## 中文文档
build:
	go build`,
			expected: []Directive{
				{Type: DirectiveCategory, Value: "构建工具", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "中文文档", SourceFile: "test.mk", LineNumber: 2},
			},
			targets: map[string]int{"build": 3},
		},
		{
			name: "unicode in variable descriptions",
			content: `## !var PORT - Порт сервера (русский)
## Start server
serve:
	./server`,
			expected: []Directive{
				{Type: DirectiveVar, Value: "PORT - Порт сервера (русский)", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "Start server", SourceFile: "test.mk", LineNumber: 2},
			},
			targets: map[string]int{"serve": 3},
		},
		{
			name: "unicode in target names",
			content: `## テスト実行
テスト:
	go test`,
			expected: []Directive{
				{Type: DirectiveDoc, Value: "テスト実行", SourceFile: "test.mk", LineNumber: 1},
			},
			targets: map[string]int{"テスト": 2},
		},
		{
			name: "mixed unicode scripts",
			content: `## !category Διάφορα
## تثبيت التبعيات (Arabic)
## Install dependencies 한국어
install:
	npm install`,
			expected: []Directive{
				{Type: DirectiveCategory, Value: "Διάφορα", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "تثبيت التبعيات (Arabic)", SourceFile: "test.mk", LineNumber: 2},
				{Type: DirectiveDoc, Value: "Install dependencies 한국어", SourceFile: "test.mk", LineNumber: 3},
			},
			targets: map[string]int{"install": 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			scanner := NewScanner()
			result, err := scanner.ScanContent(tt.content, "test.mk")
			require.NoError(t, err)
			assert.Equal(t, len(tt.expected), len(result.Directives))
			for i, expected := range tt.expected {
				assert.Equal(t, expected.Type, result.Directives[i].Type, "directive type mismatch at index %d", i)
				assert.Equal(t, expected.Value, result.Directives[i].Value, "directive value mismatch at index %d", i)
			}
			assert.Equal(t, tt.targets, result.TargetMap)
		})
	}
}

// TestScanContent_LargeFiles tests the parser's ability to handle files with many targets.
func TestScanContent_LargeFiles(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		numTargets     int
		wantTargets    int
		wantDirectives int
	}{
		{
			name:           "100 targets",
			numTargets:     100,
			wantTargets:    100,
			wantDirectives: 100, // One doc per target
		},
		{
			name:           "500 targets",
			numTargets:     500,
			wantTargets:    500,
			wantDirectives: 500,
		},
		{
			name:           "1000 targets",
			numTargets:     1000,
			wantTargets:    1000,
			wantDirectives: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Generate content with many targets
			var builder strings.Builder
			for i := 1; i <= tt.numTargets; i++ {
				builder.WriteString(fmt.Sprintf("## Build target %d\n", i))
				builder.WriteString(fmt.Sprintf("target%d:\n", i))
				builder.WriteString("\t@echo \"building\"\n\n")
			}

			scanner := NewScanner()
			result, err := scanner.ScanContent(builder.String(), "test.mk")
			require.NoError(t, err)
			assert.Equal(t, tt.wantTargets, len(result.TargetMap))
			assert.Equal(t, tt.wantDirectives, len(result.Directives))

			// Verify all targets are present
			for i := 1; i <= tt.numTargets; i++ {
				targetName := fmt.Sprintf("target%d", i)
				assert.Contains(t, result.TargetMap, targetName)
			}
		})
	}
}

// TestScanContent_LargeFileWithCategories tests large files with category organization.
func TestScanContent_LargeFileWithCategories(t *testing.T) {
	t.Parallel()
	// Generate 200 targets across 10 categories
	var builder strings.Builder
	numCategories := 10
	targetsPerCategory := 20

	for catNum := 1; catNum <= numCategories; catNum++ {
		builder.WriteString(fmt.Sprintf("## !category Category%d\n", catNum))
		for targetNum := 1; targetNum <= targetsPerCategory; targetNum++ {
			builder.WriteString(fmt.Sprintf("## Target %d in category %d\n", targetNum, catNum))
			builder.WriteString(fmt.Sprintf("cat%d_target%d:\n", catNum, targetNum))
			builder.WriteString("\t@echo \"building\"\n\n")
		}
	}

	scanner := NewScanner()
	result, err := scanner.ScanContent(builder.String(), "test.mk")
	require.NoError(t, err)

	// Count category and doc directives
	var catCount, docCount int
	for _, d := range result.Directives {
		switch d.Type {
		case DirectiveCategory:
			catCount++
		case DirectiveDoc:
			docCount++
		}
	}

	assert.Equal(t, numCategories, catCount)
	assert.Equal(t, numCategories*targetsPerCategory, docCount)
	assert.Equal(t, numCategories*targetsPerCategory, len(result.TargetMap))
}

// TestScanContent_MixedLineEndings tests parsing of files with different line ending styles.
// Note: The scanner splits on \n only. Directive values use TrimSpace (removes \r),
// but regular doc lines preserve \r characters. This tests the current behavior.
func TestScanContent_MixedLineEndings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		content  string
		expected []Directive
		targets  map[string]int
	}{
		{
			name: "unix line endings (LF) - clean handling",
			content: "## Build\nbuild:\n\tgo build",
			expected: []Directive{
				{Type: DirectiveDoc, Value: "Build", SourceFile: "test.mk", LineNumber: 1},
			},
			targets: map[string]int{"build": 2},
		},
		{
			name: "windows line endings (CRLF) - CR remains in doc content",
			content: "## Build\r\nbuild:\r\n\tgo build",
			expected: []Directive{
				{Type: DirectiveDoc, Value: "Build\r", SourceFile: "test.mk", LineNumber: 1},
			},
			targets: map[string]int{"build": 2},
		},
		{
			name: "category directive with CRLF - TrimSpace removes CR",
			content: "## !category Build\r\n## Build the project\nbuild:\r\n\tgo build\n\n## Test\r\ntest:\n\tgo test",
			expected: []Directive{
				{Type: DirectiveCategory, Value: "Build", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "Build the project", SourceFile: "test.mk", LineNumber: 2},
				{Type: DirectiveDoc, Value: "Test\r", SourceFile: "test.mk", LineNumber: 6},
			},
			targets: map[string]int{"build": 3, "test": 7},
		},
		{
			name: "old mac line endings (CR only) - treated as single line",
			content: "## Build\rbuild:\r\tgo build",
			expected: []Directive{
				// CR-only: no \n means no line split. "## Build\rbuild:\r\tgo build" is one line.
				// This line doesn't match "## " pattern (has "\r" not " " after ##)
				// so it's not recognized as documentation
			},
			targets: map[string]int{}, // No valid target detected
		},
		{
			name: "file directive with CRLF - TrimSpace removes CR",
			content: "## !file\r\n## Main build file\r\n## !category Build\r\n## Build\nbuild:\r\n\tgo build",
			expected: []Directive{
				{Type: DirectiveFile, Value: "", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "Main build file\r", SourceFile: "test.mk", LineNumber: 2},
				{Type: DirectiveCategory, Value: "Build", SourceFile: "test.mk", LineNumber: 3},
				{Type: DirectiveDoc, Value: "Build", SourceFile: "test.mk", LineNumber: 4},
			},
			targets: map[string]int{"build": 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			scanner := NewScanner()
			result, err := scanner.ScanContent(tt.content, "test.mk")
			require.NoError(t, err)
			assert.Equal(t, len(tt.expected), len(result.Directives), "directive count mismatch")
			for i, expected := range tt.expected {
				if i < len(result.Directives) {
					assert.Equal(t, expected.Type, result.Directives[i].Type, "directive type mismatch at index %d", i)
					assert.Equal(t, expected.Value, result.Directives[i].Value, "directive value mismatch at index %d", i)
				}
			}
			assert.Equal(t, tt.targets, result.TargetMap)
		})
	}
}

// TestScanContent_MalformedDirectives tests behavior with incomplete or malformed directives.
func TestScanContent_MalformedDirectives(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		content  string
		expected []Directive
		targets  map[string]int
	}{
		{
			name: "directive without space after !category",
			content: `## !category
## Build
build:
	go build`,
			expected: []Directive{
				{Type: DirectiveDoc, Value: "!category", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "Build", SourceFile: "test.mk", LineNumber: 2},
			},
			targets: map[string]int{"build": 3},
		},
		{
			name: "directive with only prefix no content",
			content: `## !var
## Build
build:
	go build`,
			expected: []Directive{
				{Type: DirectiveDoc, Value: "!var", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "Build", SourceFile: "test.mk", LineNumber: 2},
			},
			targets: map[string]int{"build": 3},
		},
		{
			name: "directive with only prefix no content for alias",
			content: `## !alias
## Build
build:
	go build`,
			expected: []Directive{
				{Type: DirectiveDoc, Value: "!alias", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "Build", SourceFile: "test.mk", LineNumber: 2},
			},
			targets: map[string]int{"build": 3},
		},
		{
			name: "unknown directive type",
			content: `## !unknown directive type
## Build
build:
	go build`,
			expected: []Directive{
				{Type: DirectiveDoc, Value: "!unknown directive type", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "Build", SourceFile: "test.mk", LineNumber: 2},
			},
			targets: map[string]int{"build": 3},
		},
		{
			name: "directive with extra whitespace",
			content: `## !category   Build Tools
## Build
build:
	go build`,
			expected: []Directive{
				{Type: DirectiveCategory, Value: "Build Tools", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "Build", SourceFile: "test.mk", LineNumber: 2},
			},
			targets: map[string]int{"build": 3},
		},
		{
			name: "directive with tabs instead of spaces - not recognized as doc",
			content: "##\t!category\tBuild\n## Build\nbuild:\n\tgo build",
			expected: []Directive{
				// "##\t" is not recognized as documentation line (needs "## " with space)
				{Type: DirectiveDoc, Value: "Build", SourceFile: "test.mk", LineNumber: 2},
			},
			targets: map[string]int{"build": 3},
		},
		{
			name: "file directive with inline description",
			content: `## !file This is the main file
## Build
build:
	go build`,
			expected: []Directive{
				{Type: DirectiveFile, Value: "This is the main file", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "Build", SourceFile: "test.mk", LineNumber: 2},
			},
			targets: map[string]int{"build": 3},
		},
		{
			name: "category with special characters",
			content: `## !category Build & Deploy
## Build
build:
	go build`,
			expected: []Directive{
				{Type: DirectiveCategory, Value: "Build & Deploy", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "Build", SourceFile: "test.mk", LineNumber: 2},
			},
			targets: map[string]int{"build": 3},
		},
		{
			name: "empty category value - treated as regular doc",
			content: `## !category
## Build
build:
	go build`,
			expected: []Directive{
				// !category without space after is treated as regular documentation
				{Type: DirectiveDoc, Value: "!category", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "Build", SourceFile: "test.mk", LineNumber: 2},
			},
			targets: map[string]int{"build": 3},
		},
		{
			name: "directive case sensitivity",
			content: `## !CATEGORY Build
## !Category Test
## !CaTeGoRy Mixed
## Build
build:
	go build`,
			expected: []Directive{
				{Type: DirectiveDoc, Value: "!CATEGORY Build", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "!Category Test", SourceFile: "test.mk", LineNumber: 2},
				{Type: DirectiveDoc, Value: "!CaTeGoRy Mixed", SourceFile: "test.mk", LineNumber: 3},
				{Type: DirectiveDoc, Value: "Build", SourceFile: "test.mk", LineNumber: 4},
			},
			targets: map[string]int{"build": 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			scanner := NewScanner()
			result, err := scanner.ScanContent(tt.content, "test.mk")
			require.NoError(t, err)
			assert.Equal(t, len(tt.expected), len(result.Directives), "directive count mismatch")
			for i, expected := range tt.expected {
				assert.Equal(t, expected.Type, result.Directives[i].Type, "directive type mismatch at index %d", i)
				assert.Equal(t, expected.Value, result.Directives[i].Value, "directive value mismatch at index %d", i)
			}
			assert.Equal(t, tt.targets, result.TargetMap)
		})
	}
}

// TestScanContent_EdgeCaseTargetNames tests unusual but valid target name formats.
func TestScanContent_EdgeCaseTargetNames(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		content string
		targets map[string]int
	}{
		{
			name: "target with dots",
			content: `## Build binary
build.linux.amd64:
	go build`,
			targets: map[string]int{"build.linux.amd64": 2},
		},
		{
			name: "target with hyphens",
			content: `## Build
build-all-targets:
	go build`,
			targets: map[string]int{"build-all-targets": 2},
		},
		{
			name: "target with underscores",
			content: `## Build
build_go_binary:
	go build`,
			targets: map[string]int{"build_go_binary": 2},
		},
		{
			name: "target with path-like name",
			content: `## Build
./bin/app:
	go build`,
			targets: map[string]int{"./bin/app": 2},
		},
		{
			name: "target with percent sign (pattern rule)",
			content: `## Build
%.o:
	gcc -c $<`,
			targets: map[string]int{"%.o": 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			scanner := NewScanner()
			result, err := scanner.ScanContent(tt.content, "test.mk")
			require.NoError(t, err)
			assert.Equal(t, tt.targets, result.TargetMap)
		})
	}
}

// TestScanContent_EdgeCaseWhitespace tests handling of unusual whitespace patterns.
func TestScanContent_EdgeCaseWhitespace(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		content  string
		expected []Directive
	}{
		{
			name:    "documentation with trailing spaces",
			content: "## Build the project   \nbuild:\n\tgo build",
			expected: []Directive{
				{Type: DirectiveDoc, Value: "Build the project   ", SourceFile: "test.mk", LineNumber: 1},
			},
		},
		{
			name:    "documentation with only spaces after ## - one space trimmed",
			content: "##     \nbuild:\n\tgo build",
			expected: []Directive{
				// TrimPrefix("## ") removes first space, leaving 4 spaces
				{Type: DirectiveDoc, Value: "    ", SourceFile: "test.mk", LineNumber: 1},
			},
		},
		{
			name:    "multiple consecutive blank documentation lines",
			content: "##\n##\n## Build\nbuild:\n\tgo build",
			expected: []Directive{
				{Type: DirectiveDoc, Value: "", SourceFile: "test.mk", LineNumber: 1},
				{Type: DirectiveDoc, Value: "", SourceFile: "test.mk", LineNumber: 2},
				{Type: DirectiveDoc, Value: "Build", SourceFile: "test.mk", LineNumber: 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
