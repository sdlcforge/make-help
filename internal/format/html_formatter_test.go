package format

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sdlcforge/make-help/internal/model"
	"github.com/sdlcforge/make-help/internal/richtext"
)

// TestHTMLFormatter_RenderHelp_EmptyModel tests rendering an empty help model
func TestHTMLFormatter_RenderHelp_EmptyModel(t *testing.T) {
	formatter := NewHTMLFormatter(&FormatterConfig{UseColor: false})
	helpModel := &model.HelpModel{}

	var buf bytes.Buffer
	err := formatter.RenderHelp(helpModel, &buf)

	if err != nil {
		t.Fatalf("RenderHelp() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "<!DOCTYPE html>") {
		t.Error("Output should contain DOCTYPE")
	}
	if !strings.Contains(output, "<h1>Makefile Help</h1>") {
		t.Error("Output should contain h1 title")
	}
	if !strings.Contains(output, "make [&lt;target&gt;...] [&lt;ENV_VAR&gt;=&lt;value&gt;...]") {
		t.Error("Output should contain usage line with HTML entities")
	}
	if strings.Contains(output, "<section class=\"targets\">") {
		t.Error("Output should not contain targets section for empty model")
	}
}

// TestHTMLFormatter_RenderHelp_WithTargets tests rendering with basic targets
func TestHTMLFormatter_RenderHelp_WithTargets(t *testing.T) {
	formatter := NewHTMLFormatter(&FormatterConfig{UseColor: false})
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
	if !strings.Contains(output, "<section class=\"targets\">") {
		t.Error("Output should contain targets section")
	}
	if !strings.Contains(output, "<span class=\"target-name\">build</span>") {
		t.Error("Output should contain build target")
	}
	if !strings.Contains(output, "<span class=\"target-name\">test</span>") {
		t.Error("Output should contain test target")
	}
	if !strings.Contains(output, "Build the project.") {
		t.Error("Output should contain build summary")
	}
}

// TestHTMLFormatter_RenderHelp_WithCategories tests rendering with categories
func TestHTMLFormatter_RenderHelp_WithCategories(t *testing.T) {
	formatter := NewHTMLFormatter(&FormatterConfig{UseColor: false})
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
	if !strings.Contains(output, "<h3>Build</h3>") {
		t.Error("Output should contain Build category")
	}
	if !strings.Contains(output, "<h3>Test</h3>") {
		t.Error("Output should contain Test category")
	}
}

// TestHTMLFormatter_RenderHelp_WithCSS tests CSS embedding
func TestHTMLFormatter_RenderHelp_WithCSS(t *testing.T) {
	formatter := NewHTMLFormatter(&FormatterConfig{UseColor: true})
	helpModel := &model.HelpModel{}

	var buf bytes.Buffer
	err := formatter.RenderHelp(helpModel, &buf)

	if err != nil {
		t.Fatalf("RenderHelp() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "<style>") {
		t.Error("Output should contain style tag when color is enabled")
	}
	if !strings.Contains(output, ".target-name") {
		t.Error("Output should contain CSS classes")
	}
}

// TestHTMLFormatter_RenderHelp_WithoutCSS tests no CSS embedding when color disabled
func TestHTMLFormatter_RenderHelp_WithoutCSS(t *testing.T) {
	formatter := NewHTMLFormatter(&FormatterConfig{UseColor: false})
	helpModel := &model.HelpModel{}

	var buf bytes.Buffer
	err := formatter.RenderHelp(helpModel, &buf)

	if err != nil {
		t.Fatalf("RenderHelp() error = %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "<style>") {
		t.Error("Output should not contain style tag when color is disabled")
	}
}

// TestHTMLFormatter_RenderHelp_HTMLEscaping tests proper HTML escaping
func TestHTMLFormatter_RenderHelp_HTMLEscaping(t *testing.T) {
	formatter := NewHTMLFormatter(&FormatterConfig{UseColor: false})
	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: "Build & Test",
				Targets: []model.Target{
					{
						Name:    "build<>",
						Summary: richtext.FromPlainText("Build with <tags> & \"quotes\"."),
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
	// Check that special characters are escaped
	if !strings.Contains(output, "Build &amp; Test") {
		t.Error("Category name should have escaped ampersand")
	}
	if !strings.Contains(output, "build&lt;&gt;") {
		t.Error("Target name should have escaped angle brackets")
	}
	if !strings.Contains(output, "&lt;tags&gt;") {
		t.Error("Summary should have escaped angle brackets")
	}
	if !strings.Contains(output, "&amp;") {
		t.Error("Summary should have escaped ampersand")
	}
	// Should not contain raw HTML characters
	if strings.Contains(output, "Build & Test") && !strings.Contains(output, "&amp;") {
		t.Error("Should not contain unescaped ampersand in content")
	}
}

// TestHTMLFormatter_RenderHelp_WithFileDocumentation tests file documentation rendering
func TestHTMLFormatter_RenderHelp_WithFileDocumentation(t *testing.T) {
	formatter := NewHTMLFormatter(&FormatterConfig{UseColor: false})
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
	if !strings.Contains(output, "<section class=\"file-docs\">") {
		t.Error("Output should contain file-docs section")
	}
	if !strings.Contains(output, "<h2>Description</h2>") {
		t.Error("Output should contain Description header")
	}
	if !strings.Contains(output, "This is the main Makefile.") {
		t.Error("Output should contain file documentation")
	}
}

// TestHTMLFormatter_RenderHelp_WithIncludedFiles tests included files rendering
func TestHTMLFormatter_RenderHelp_WithIncludedFiles(t *testing.T) {
	formatter := NewHTMLFormatter(&FormatterConfig{UseColor: false})
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
	if !strings.Contains(output, "<section class=\"included-files\">") {
		t.Error("Output should contain included-files section")
	}
	if !strings.Contains(output, "<h3>make/build.mk</h3>") {
		t.Error("Output should contain included file path as h3")
	}
	if !strings.Contains(output, "Build tasks.") {
		t.Error("Output should contain included file documentation")
	}
}

// TestHTMLFormatter_RenderHelp_WithVariables tests target variables rendering
func TestHTMLFormatter_RenderHelp_WithVariables(t *testing.T) {
	formatter := NewHTMLFormatter(&FormatterConfig{UseColor: false})
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
	if !strings.Contains(output, "<div class=\"variables\">") {
		t.Error("Output should contain variables div")
	}
	if !strings.Contains(output, "<code class=\"variable\">PORT</code>") {
		t.Error("Output should contain PORT variable")
	}
	if !strings.Contains(output, "<code class=\"variable\">DEBUG</code>") {
		t.Error("Output should contain DEBUG variable")
	}
}

// TestHTMLFormatter_RenderHelp_WithAliases tests target aliases rendering
func TestHTMLFormatter_RenderHelp_WithAliases(t *testing.T) {
	formatter := NewHTMLFormatter(&FormatterConfig{UseColor: false})
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
	if !strings.Contains(output, "<span class=\"alias\">") {
		t.Error("Output should contain alias span")
	}
	if !strings.Contains(output, "b, compile") {
		t.Error("Output should contain aliases")
	}
}

// TestHTMLFormatter_RenderHelp_WithRichText tests RichText rendering
func TestHTMLFormatter_RenderHelp_WithRichText(t *testing.T) {
	formatter := NewHTMLFormatter(&FormatterConfig{UseColor: false})
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
	if !strings.Contains(output, "<strong>bold</strong>") {
		t.Error("Output should contain bold formatting")
	}
	if !strings.Contains(output, "<em>italic</em>") {
		t.Error("Output should contain italic formatting")
	}
	if !strings.Contains(output, "<code>code</code>") {
		t.Error("Output should contain code formatting")
	}
	if !strings.Contains(output, "<a href=\"https://example.com\">link</a>") {
		t.Error("Output should contain link formatting")
	}
}

// TestHTMLFormatter_RenderDetailedTarget tests detailed target rendering
func TestHTMLFormatter_RenderDetailedTarget(t *testing.T) {
	formatter := NewHTMLFormatter(&FormatterConfig{UseColor: false})
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
	if !strings.Contains(output, "<h1>Target: build</h1>") {
		t.Error("Output should contain target name in h1")
	}
	if !strings.Contains(output, "Aliases:") {
		t.Error("Output should contain aliases label")
	}
	if !strings.Contains(output, "b, compile") {
		t.Error("Output should contain aliases")
	}
	if !strings.Contains(output, "Variables:") {
		t.Error("Output should contain variables header")
	}
	if !strings.Contains(output, "<code class=\"variable\">GOOS</code>") {
		t.Error("Output should contain variable GOOS")
	}
	if !strings.Contains(output, "Target OS") {
		t.Error("Output should contain variable description")
	}
	if !strings.Contains(output, "Build the project.") {
		t.Error("Output should contain documentation")
	}
	if !strings.Contains(output, "Source:") && !strings.Contains(output, "/path/to/Makefile:42") {
		t.Error("Output should contain source location")
	}
}

// TestHTMLFormatter_RenderBasicTarget tests basic target rendering
func TestHTMLFormatter_RenderBasicTarget(t *testing.T) {
	formatter := NewHTMLFormatter(&FormatterConfig{UseColor: false})

	var buf bytes.Buffer
	err := formatter.RenderBasicTarget("undocumented", "/path/to/Makefile", 15, &buf)

	if err != nil {
		t.Fatalf("RenderBasicTarget() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "<h1>Target: undocumented</h1>") {
		t.Error("Output should contain target name in h1")
	}
	if !strings.Contains(output, "No documentation available.") {
		t.Error("Output should contain 'no documentation' message")
	}
	if !strings.Contains(output, "/path/to/Makefile:15") {
		t.Error("Output should contain source location")
	}
}

// TestHTMLFormatter_ContentType tests content type
func TestHTMLFormatter_ContentType(t *testing.T) {
	formatter := NewHTMLFormatter(&FormatterConfig{})

	contentType := formatter.ContentType()
	if contentType != "text/html" {
		t.Errorf("ContentType() = %q, want %q", contentType, "text/html")
	}
}

// TestHTMLFormatter_DefaultExtension tests default extension
func TestHTMLFormatter_DefaultExtension(t *testing.T) {
	formatter := NewHTMLFormatter(&FormatterConfig{})

	ext := formatter.DefaultExtension()
	if ext != ".html" {
		t.Errorf("DefaultExtension() = %q, want %q", ext, ".html")
	}
}

// TestHTMLFormatter_NilModel tests error handling for nil model
func TestHTMLFormatter_NilModel(t *testing.T) {
	formatter := NewHTMLFormatter(&FormatterConfig{})

	var buf bytes.Buffer
	err := formatter.RenderHelp(nil, &buf)

	if err == nil {
		t.Error("RenderHelp(nil) should return an error")
	}
	if !strings.Contains(err.Error(), "cannot be nil") {
		t.Errorf("Error should mention nil, got: %v", err)
	}
}

// TestHTMLFormatter_NilTarget tests error handling for nil target
func TestHTMLFormatter_NilTarget(t *testing.T) {
	formatter := NewHTMLFormatter(&FormatterConfig{})

	var buf bytes.Buffer
	err := formatter.RenderDetailedTarget(nil, &buf)

	if err == nil {
		t.Error("RenderDetailedTarget(nil) should return an error")
	}
	if !strings.Contains(err.Error(), "cannot be nil") {
		t.Errorf("Error should mention nil, got: %v", err)
	}
}
