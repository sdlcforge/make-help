package format

import (
	"github.com/sdlcforge/make-help/internal/richtext"
	"bytes"
	"strings"
	"testing"

	"github.com/sdlcforge/make-help/internal/model"
)

// TestTextFormatter_RenderHelp_EmptyModel tests rendering an empty help model
func TestTextFormatter_RenderHelp_EmptyModel(t *testing.T) {
	formatter := NewTextFormatter(&FormatterConfig{UseColor: false})
	helpModel := &model.HelpModel{}

	var buf bytes.Buffer
	err := formatter.RenderHelp(helpModel, &buf)

	if err != nil {
		t.Fatalf("RenderHelp() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Usage: make [<target>...] [<ENV_VAR>=<value>...]") {
		t.Error("Output should contain usage line")
	}
	if strings.Contains(output, "Targets:") {
		t.Error("Output should not contain 'Targets:' for empty model")
	}
}

// TestTextFormatter_RenderHelp_WithTargets tests rendering with basic targets
func TestTextFormatter_RenderHelp_WithTargets(t *testing.T) {
	formatter := NewTextFormatter(&FormatterConfig{UseColor: false})
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
	if !strings.Contains(output, "Targets:") {
		t.Error("Output should contain 'Targets:' header")
	}
	if !strings.Contains(output, "- build: Build the project.") {
		t.Error("Output should contain build target")
	}
	if !strings.Contains(output, "- test: Run all tests.") {
		t.Error("Output should contain test target")
	}
}

// TestTextFormatter_RenderHelp_WithCategories tests rendering with categories
func TestTextFormatter_RenderHelp_WithCategories(t *testing.T) {
	formatter := NewTextFormatter(&FormatterConfig{UseColor: false})
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
	if !strings.Contains(output, "Build:") {
		t.Error("Output should contain 'Build:' category")
	}
	if !strings.Contains(output, "Test:") {
		t.Error("Output should contain 'Test:' category")
	}
}

// TestTextFormatter_RenderHelp_WithColors tests ANSI color codes
func TestTextFormatter_RenderHelp_WithColors(t *testing.T) {
	formatter := NewTextFormatter(&FormatterConfig{UseColor: true})
	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: "Build",
				Targets: []model.Target{
					{
						Name:    "build",
						Aliases: []string{"b"},
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
	// Should contain raw ANSI codes (not escaped)
	if !strings.Contains(output, "\033[1;36m") {
		t.Error("Output should contain cyan for category")
	}
	if !strings.Contains(output, "\033[1;32m") {
		t.Error("Output should contain green for target")
	}
	if !strings.Contains(output, "\033[0;33m") {
		t.Error("Output should contain yellow for alias")
	}
	if !strings.Contains(output, "\033[0m") {
		t.Error("Output should contain reset code")
	}
}

// TestTextFormatter_RenderHelp_NoColors tests plain text output
func TestTextFormatter_RenderHelp_NoColors(t *testing.T) {
	formatter := NewTextFormatter(&FormatterConfig{UseColor: false})
	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: "Build",
				Targets: []model.Target{
					{
						Name:    "build",
						Aliases: []string{"b"},
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
	// Should NOT contain any ANSI codes
	if strings.Contains(output, "\033[") {
		t.Error("Output should not contain ANSI codes when color is disabled")
	}
}

// TestTextFormatter_RenderDetailedTarget tests detailed target rendering
func TestTextFormatter_RenderDetailedTarget(t *testing.T) {
	formatter := NewTextFormatter(&FormatterConfig{UseColor: false})
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
	if !strings.Contains(output, "Target: build") {
		t.Error("Output should contain target name")
	}
	if !strings.Contains(output, "Aliases: b, compile") {
		t.Error("Output should contain aliases")
	}
	if !strings.Contains(output, "Variables:") {
		t.Error("Output should contain variables header")
	}
	if !strings.Contains(output, "- GOOS: Target OS") {
		t.Error("Output should contain variable GOOS")
	}
	if !strings.Contains(output, "Build the project.") {
		t.Error("Output should contain documentation")
	}
	if !strings.Contains(output, "Source: /path/to/Makefile:42") {
		t.Error("Output should contain source location")
	}
}

// TestTextFormatter_RenderBasicTarget tests basic target rendering
func TestTextFormatter_RenderBasicTarget(t *testing.T) {
	formatter := NewTextFormatter(&FormatterConfig{UseColor: false})

	var buf bytes.Buffer
	err := formatter.RenderBasicTarget("undocumented", "/path/to/Makefile", 15, &buf)

	if err != nil {
		t.Fatalf("RenderBasicTarget() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Target: undocumented") {
		t.Error("Output should contain target name")
	}
	if !strings.Contains(output, "No documentation available.") {
		t.Error("Output should contain 'no documentation' message")
	}
	if !strings.Contains(output, "Source: /path/to/Makefile:15") {
		t.Error("Output should contain source location")
	}
}

// TestTextFormatter_ContentType tests content type
func TestTextFormatter_ContentType(t *testing.T) {
	formatter := NewTextFormatter(&FormatterConfig{})

	contentType := formatter.ContentType()
	if contentType != "text/plain" {
		t.Errorf("ContentType() = %q, want %q", contentType, "text/plain")
	}
}

// TestTextFormatter_DefaultExtension tests default extension
func TestTextFormatter_DefaultExtension(t *testing.T) {
	formatter := NewTextFormatter(&FormatterConfig{})

	ext := formatter.DefaultExtension()
	if ext != ".txt" {
		t.Errorf("DefaultExtension() = %q, want %q", ext, ".txt")
	}
}

// TestTextFormatter_WithFileDocumentation tests file documentation rendering
func TestTextFormatter_WithFileDocumentation(t *testing.T) {
	formatter := NewTextFormatter(&FormatterConfig{UseColor: false})
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
	if !strings.Contains(output, "This is the main Makefile.") {
		t.Error("Output should contain file documentation")
	}
	if !strings.Contains(output, "It provides tasks.") {
		t.Error("Output should contain second line of file documentation")
	}
}

// TestTextFormatter_WithIncludedFiles tests included files rendering
func TestTextFormatter_WithIncludedFiles(t *testing.T) {
	formatter := NewTextFormatter(&FormatterConfig{UseColor: false})
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
	if !strings.Contains(output, "Included Files:") {
		t.Error("Output should contain 'Included Files:' header")
	}
	if !strings.Contains(output, "make/build.mk") {
		t.Error("Output should contain included file path")
	}
	// Check for indented documentation
	if !strings.Contains(output, "    Build tasks.") {
		t.Error("Output should contain indented file documentation")
	}
}

// TestTextFormatter_WithVariables tests target variables rendering
func TestTextFormatter_WithVariables(t *testing.T) {
	formatter := NewTextFormatter(&FormatterConfig{UseColor: false})
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
	if !strings.Contains(output, "Vars: PORT, DEBUG") {
		t.Error("Output should contain variables list")
	}
}

// TestTextFormatter_WithAliases tests target aliases rendering
func TestTextFormatter_WithAliases(t *testing.T) {
	formatter := NewTextFormatter(&FormatterConfig{UseColor: false})
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
	if !strings.Contains(output, "- build b, compile: Build the project.") {
		t.Error("Output should contain target with aliases")
	}
}

// TestTextFormatter_ComplexModel tests a complex help model
func TestTextFormatter_ComplexModel(t *testing.T) {
	formatter := NewTextFormatter(&FormatterConfig{UseColor: false})
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
	if !strings.Contains(output, "Project Makefile") {
		t.Error("Output should contain file documentation")
	}
	// Categories
	if !strings.Contains(output, "Build:") {
		t.Error("Output should contain Build category")
	}
	if !strings.Contains(output, "Test:") {
		t.Error("Output should contain Test category")
	}
	// Targets
	if !strings.Contains(output, "- build b: Build the project.") {
		t.Error("Output should contain build target with alias")
	}
	if !strings.Contains(output, "Vars: GOOS, GOARCH") {
		t.Error("Output should contain variables")
	}
}
