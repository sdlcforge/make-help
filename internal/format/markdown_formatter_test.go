package format

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sdlcforge/make-help/internal/model"
	"github.com/sdlcforge/make-help/internal/richtext"
)

// TestMarkdownFormatter_RenderHelp_EmptyModel tests rendering an empty help model
func TestMarkdownFormatter_RenderHelp_EmptyModel(t *testing.T) {
	formatter := NewMarkdownFormatter(&FormatterConfig{UseColor: false})
	helpModel := &model.HelpModel{}

	var buf bytes.Buffer
	err := formatter.RenderHelp(helpModel, &buf)

	if err != nil {
		t.Fatalf("RenderHelp() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "# Makefile Help") {
		t.Error("Output should contain main heading")
	}
	if !strings.Contains(output, "## Usage") {
		t.Error("Output should contain usage section")
	}
	if !strings.Contains(output, "make [<target>...] [<ENV_VAR>=<value>...]") {
		t.Error("Output should contain usage line")
	}
	if strings.Contains(output, "## Targets") {
		t.Error("Output should not contain targets section for empty model")
	}
}

// TestMarkdownFormatter_RenderHelp_WithTargets tests rendering with basic targets
func TestMarkdownFormatter_RenderHelp_WithTargets(t *testing.T) {
	formatter := NewMarkdownFormatter(&FormatterConfig{UseColor: false})
	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: model.UncategorizedCategoryName,
				Targets: []model.Target{
					{
						Name:    "build",
						Summary: richtext.FromPlainText("Build the project."),
					},
					{
						Name:    "test",
						Summary: richtext.FromPlainText("Run all tests."),
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := formatter.RenderHelp(helpModel, &buf)

	if err != nil {
		t.Fatalf("RenderHelp() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "## Targets") {
		t.Error("Output should contain targets section")
	}
	if !strings.Contains(output, "- **build**: Build the project.") {
		t.Error("Output should contain build target")
	}
	if !strings.Contains(output, "- **test**: Run all tests.") {
		t.Error("Output should contain test target")
	}
}

// TestMarkdownFormatter_RenderHelp_WithCategories tests rendering with categories
func TestMarkdownFormatter_RenderHelp_WithCategories(t *testing.T) {
	formatter := NewMarkdownFormatter(&FormatterConfig{UseColor: false})
	helpModel := &model.HelpModel{
		HasCategories: true,
		Categories: []model.Category{
			{
				Name: "Build",
				Targets: []model.Target{
					{
						Name:    "build",
						Summary: richtext.FromPlainText("Build the project."),
					},
				},
			},
			{
				Name: "Test",
				Targets: []model.Target{
					{
						Name:    "test",
						Summary: richtext.FromPlainText("Run all tests."),
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := formatter.RenderHelp(helpModel, &buf)

	if err != nil {
		t.Fatalf("RenderHelp() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "### Build") {
		t.Error("Output should contain Build category")
	}
	if !strings.Contains(output, "### Test") {
		t.Error("Output should contain Test category")
	}
}

// TestMarkdownFormatter_RenderHelp_WithFileDocumentation tests file documentation rendering
func TestMarkdownFormatter_RenderHelp_WithFileDocumentation(t *testing.T) {
	formatter := NewMarkdownFormatter(&FormatterConfig{UseColor: false})
	helpModel := &model.HelpModel{
		FileDocs: []model.FileDoc{
			{
				SourceFile:     "Makefile",
				Documentation:  []string{"This is the main Makefile.", "It provides tasks."},
				IsEntryPoint:   true,
				DiscoveryOrder: 0,
			},
		},
	}

	var buf bytes.Buffer
	err := formatter.RenderHelp(helpModel, &buf)

	if err != nil {
		t.Fatalf("RenderHelp() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "## Description") {
		t.Error("Output should contain Description heading")
	}
	if !strings.Contains(output, "This is the main Makefile.") {
		t.Error("Output should contain file documentation")
	}
	if !strings.Contains(output, "It provides tasks.") {
		t.Error("Output should contain second line of file documentation")
	}
}

// TestMarkdownFormatter_RenderHelp_WithIncludedFiles tests included files rendering
func TestMarkdownFormatter_RenderHelp_WithIncludedFiles(t *testing.T) {
	formatter := NewMarkdownFormatter(&FormatterConfig{UseColor: false})
	helpModel := &model.HelpModel{
		FileDocs: []model.FileDoc{
			{
				SourceFile:     "Makefile",
				Documentation:  []string{"Main Makefile."},
				IsEntryPoint:   true,
				DiscoveryOrder: 0,
			},
			{
				SourceFile:     "make/build.mk",
				Documentation:  []string{"Build tasks."},
				IsEntryPoint:   false,
				DiscoveryOrder: 1,
			},
		},
	}

	var buf bytes.Buffer
	err := formatter.RenderHelp(helpModel, &buf)

	if err != nil {
		t.Fatalf("RenderHelp() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "## Included files") {
		t.Error("Output should contain Included files heading")
	}
	if !strings.Contains(output, "### make/build.mk") {
		t.Error("Output should contain included file path as h3")
	}
	if !strings.Contains(output, "Build tasks.") {
		t.Error("Output should contain included file documentation")
	}
}

// TestMarkdownFormatter_RenderHelp_WithVariables tests target variables rendering
func TestMarkdownFormatter_RenderHelp_WithVariables(t *testing.T) {
	formatter := NewMarkdownFormatter(&FormatterConfig{UseColor: false})
	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: model.UncategorizedCategoryName,
				Targets: []model.Target{
					{
						Name:    "serve",
						Summary: richtext.FromPlainText("Start server."),
						Variables: []model.Variable{
							{Name: "PORT"},
							{Name: "DEBUG"},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := formatter.RenderHelp(helpModel, &buf)

	if err != nil {
		t.Fatalf("RenderHelp() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Variables: `PORT`, `DEBUG`") {
		t.Error("Output should contain variables list")
	}
}

// TestMarkdownFormatter_RenderHelp_WithAliases tests target aliases rendering
func TestMarkdownFormatter_RenderHelp_WithAliases(t *testing.T) {
	formatter := NewMarkdownFormatter(&FormatterConfig{UseColor: false})
	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: model.UncategorizedCategoryName,
				Targets: []model.Target{
					{
						Name:    "build",
						Aliases: []string{"b", "compile"},
						Summary: richtext.FromPlainText("Build the project."),
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := formatter.RenderHelp(helpModel, &buf)

	if err != nil {
		t.Fatalf("RenderHelp() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "- **build** _(b, compile)_: Build the project.") {
		t.Error("Output should contain target with aliases in italics")
	}
}

// TestMarkdownFormatter_RenderHelp_WithRichText tests RichText rendering
func TestMarkdownFormatter_RenderHelp_WithRichText(t *testing.T) {
	formatter := NewMarkdownFormatter(&FormatterConfig{UseColor: false})
	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: model.UncategorizedCategoryName,
				Targets: []model.Target{
					{
						Name: "build",
						Summary: richtext.RichText{
							{Type: richtext.SegmentPlain, Content: "Build with "},
							{Type: richtext.SegmentBold, Content: "bold"},
							{Type: richtext.SegmentPlain, Content: ", "},
							{Type: richtext.SegmentItalic, Content: "italic"},
							{Type: richtext.SegmentPlain, Content: ", "},
							{Type: richtext.SegmentCode, Content: "code"},
							{Type: richtext.SegmentPlain, Content: ", and "},
							{Type: richtext.SegmentLink, Content: "link", URL: "https://example.com"},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := formatter.RenderHelp(helpModel, &buf)

	if err != nil {
		t.Fatalf("RenderHelp() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "**bold**") {
		t.Error("Output should contain bold markdown")
	}
	if !strings.Contains(output, "*italic*") {
		t.Error("Output should contain italic markdown")
	}
	if !strings.Contains(output, "`code`") {
		t.Error("Output should contain code markdown")
	}
	if !strings.Contains(output, "[link](https://example.com)") {
		t.Error("Output should contain link markdown")
	}
}

// TestMarkdownFormatter_RenderDetailedTarget tests detailed target rendering
func TestMarkdownFormatter_RenderDetailedTarget(t *testing.T) {
	formatter := NewMarkdownFormatter(&FormatterConfig{UseColor: false})
	target := &model.Target{
		Name:    "build",
		Aliases: []string{"b", "compile"},
		Documentation: []string{
			"Build the project.",
			"",
			"This compiles all source files.",
		},
		Variables: []model.Variable{
			{Name: "GOOS", Description: "Target OS"},
			{Name: "GOARCH", Description: "Target architecture"},
		},
		SourceFile: "/path/to/Makefile",
		LineNumber: 42,
	}

	var buf bytes.Buffer
	err := formatter.RenderDetailedTarget(target, &buf)

	if err != nil {
		t.Fatalf("RenderDetailedTarget() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "# Target: build") {
		t.Error("Output should contain target name as h1")
	}
	if !strings.Contains(output, "**Aliases:** b, compile") {
		t.Error("Output should contain aliases")
	}
	if !strings.Contains(output, "**Variables:**") {
		t.Error("Output should contain variables header")
	}
	if !strings.Contains(output, "- `GOOS`: Target OS") {
		t.Error("Output should contain variable GOOS with description")
	}
	if !strings.Contains(output, "## Description") {
		t.Error("Output should contain description section")
	}
	if !strings.Contains(output, "Build the project.") {
		t.Error("Output should contain documentation")
	}
	if !strings.Contains(output, "**Source:** `/path/to/Makefile:42`") {
		t.Error("Output should contain source location")
	}
}

// TestMarkdownFormatter_RenderBasicTarget tests basic target rendering
func TestMarkdownFormatter_RenderBasicTarget(t *testing.T) {
	formatter := NewMarkdownFormatter(&FormatterConfig{UseColor: false})

	var buf bytes.Buffer
	err := formatter.RenderBasicTarget("undocumented", "/path/to/Makefile", 15, &buf)

	if err != nil {
		t.Fatalf("RenderBasicTarget() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "# Target: undocumented") {
		t.Error("Output should contain target name as h1")
	}
	if !strings.Contains(output, "_No documentation available._") {
		t.Error("Output should contain 'no documentation' message in italics")
	}
	if !strings.Contains(output, "**Source:** `/path/to/Makefile:15`") {
		t.Error("Output should contain source location")
	}
}

// TestMarkdownFormatter_ContentType tests content type
func TestMarkdownFormatter_ContentType(t *testing.T) {
	formatter := NewMarkdownFormatter(&FormatterConfig{})

	contentType := formatter.ContentType()
	if contentType != "text/markdown" {
		t.Errorf("ContentType() = %q, want %q", contentType, "text/markdown")
	}
}

// TestMarkdownFormatter_DefaultExtension tests default extension
func TestMarkdownFormatter_DefaultExtension(t *testing.T) {
	formatter := NewMarkdownFormatter(&FormatterConfig{})

	ext := formatter.DefaultExtension()
	if ext != ".md" {
		t.Errorf("DefaultExtension() = %q, want %q", ext, ".md")
	}
}

// TestMarkdownFormatter_NilModel tests error handling for nil model
func TestMarkdownFormatter_NilModel(t *testing.T) {
	formatter := NewMarkdownFormatter(&FormatterConfig{})

	var buf bytes.Buffer
	err := formatter.RenderHelp(nil, &buf)

	if err == nil {
		t.Error("RenderHelp(nil) should return an error")
	}
	if !strings.Contains(err.Error(), "cannot be nil") {
		t.Errorf("Error should mention nil, got: %v", err)
	}
}

// TestMarkdownFormatter_NilTarget tests error handling for nil target
func TestMarkdownFormatter_NilTarget(t *testing.T) {
	formatter := NewMarkdownFormatter(&FormatterConfig{})

	var buf bytes.Buffer
	err := formatter.RenderDetailedTarget(nil, &buf)

	if err == nil {
		t.Error("RenderDetailedTarget(nil) should return an error")
	}
	if !strings.Contains(err.Error(), "cannot be nil") {
		t.Errorf("Error should mention nil, got: %v", err)
	}
}

// TestMarkdownFormatter_RenderHelp_SpecialCharacters tests Markdown escaping
func TestMarkdownFormatter_RenderHelp_SpecialCharacters(t *testing.T) {
	formatter := NewMarkdownFormatter(&FormatterConfig{UseColor: false})
	helpModel := &model.HelpModel{
		FileDocs: []model.FileDoc{
			{
				SourceFile:     "path/[brackets]/file.mk",
				Documentation:  []string{"File with brackets."},
				IsEntryPoint:   false,
				DiscoveryOrder: 1,
			},
		},
		Categories: []model.Category{
			{
				Name: model.UncategorizedCategoryName,
				Targets: []model.Target{
					{
						Name:    "build*test",
						Aliases: []string{"b_uild", "test**break"},
						Summary: richtext.FromPlainText("Build with asterisks."),
						Variables: []model.Variable{
							{Name: "VAR_WITH_`BACKTICKS`"},
							{Name: "VAR#WITH#HASH"},
						},
					},
					{
						Name:    "test_underscore",
						Summary: richtext.FromPlainText("Test with underscores."),
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := formatter.RenderHelp(helpModel, &buf)

	if err != nil {
		t.Fatalf("RenderHelp() error = %v", err)
	}

	output := buf.String()

	// Test target names are escaped
	if !strings.Contains(output, `- **build\*test**`) {
		t.Error("Target name with asterisks should be escaped")
	}
	if !strings.Contains(output, `- **test\_underscore**`) {
		t.Error("Target name with underscores should be escaped")
	}

	// Test aliases are escaped
	if !strings.Contains(output, `b\_uild`) {
		t.Error("Alias with underscore should be escaped")
	}
	if !strings.Contains(output, `test\*\*break`) {
		t.Error("Alias with asterisks should be escaped")
	}

	// Test variable names are escaped
	if !strings.Contains(output, "`VAR\\_WITH\\_\\`BACKTICKS\\``") {
		t.Errorf("Variable name with backticks should be escaped. Output:\n%s", output)
	}
	if !strings.Contains(output, "`VAR\\#WITH\\#HASH`") {
		t.Error("Variable name with hash should be escaped")
	}

	// Test file paths are escaped
	if !strings.Contains(output, `### path/\[brackets\]/file.mk`) {
		t.Error("File path with brackets should be escaped")
	}
}

// TestMarkdownFormatter_ComplexModel tests a complex help model
func TestMarkdownFormatter_ComplexModel(t *testing.T) {
	formatter := NewMarkdownFormatter(&FormatterConfig{UseColor: false})
	helpModel := &model.HelpModel{
		FileDocs: []model.FileDoc{
			{
				SourceFile:     "Makefile",
				Documentation:  []string{"Project Makefile", "Common tasks"},
				IsEntryPoint:   true,
				DiscoveryOrder: 0,
			},
		},
		HasCategories: true,
		Categories: []model.Category{
			{
				Name: "Build",
				Targets: []model.Target{
					{
						Name:    "build",
						Aliases: []string{"b"},
						Summary: richtext.FromPlainText("Build the project."),
						Variables: []model.Variable{
							{Name: "GOOS"},
							{Name: "GOARCH"},
						},
					},
				},
			},
			{
				Name: "Test",
				Targets: []model.Target{
					{
						Name:    "test",
						Aliases: []string{"t"},
						Summary: richtext.FromPlainText("Run all tests."),
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := formatter.RenderHelp(helpModel, &buf)

	if err != nil {
		t.Fatalf("RenderHelp() error = %v", err)
	}

	output := buf.String()
	// File docs
	if !strings.Contains(output, "## Description") {
		t.Error("Output should contain description section")
	}
	if !strings.Contains(output, "Project Makefile") {
		t.Error("Output should contain file documentation")
	}
	// Categories
	if !strings.Contains(output, "### Build") {
		t.Error("Output should contain Build category")
	}
	if !strings.Contains(output, "### Test") {
		t.Error("Output should contain Test category")
	}
	// Targets
	if !strings.Contains(output, "- **build** _(b)_: Build the project.") {
		t.Error("Output should contain build target with alias")
	}
	if !strings.Contains(output, "Variables: `GOOS`, `GOARCH`") {
		t.Error("Output should contain variables")
	}
}
