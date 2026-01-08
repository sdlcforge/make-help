package format

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sdlcforge/make-help/internal/model"
)

// TestMakeFormatter_RenderHelp_EmptyModel tests rendering an empty help model
func TestMakeFormatter_RenderHelp_EmptyModel(t *testing.T) {
	t.Parallel()
	formatter := NewMakeFormatter(&FormatterConfig{UseColor: false})
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
	if !strings.Contains(output, "@printf") {
		t.Error("Output should contain @printf statements")
	}
}

// TestMakeFormatter_RenderHelp_WithTargets tests rendering with basic targets
func TestMakeFormatter_RenderHelp_WithTargets(t *testing.T) {
	t.Parallel()
	formatter := NewMakeFormatter(&FormatterConfig{UseColor: false})
	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: model.UncategorizedCategoryName,
				Targets: []model.Target{
					{
						Name:    "build",
						Summary: []string{"Build the project."},
					},
					{
						Name:    "test",
						Summary: []string{"Run all tests."},
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
	if !strings.Contains(output, "build: Build the project.") {
		t.Error("Output should contain build target")
	}
	if !strings.Contains(output, "test: Run all tests.") {
		t.Error("Output should contain test target")
	}
}

// TestMakeFormatter_RenderHelp_WithCategories tests rendering with categories
func TestMakeFormatter_RenderHelp_WithCategories(t *testing.T) {
	t.Parallel()
	formatter := NewMakeFormatter(&FormatterConfig{UseColor: false})
	helpModel := &model.HelpModel{
		HasCategories: true,
		Categories: []model.Category{
			{
				Name: "Build",
				Targets: []model.Target{
					{
						Name:    "build",
						Summary: []string{"Build the project."},
					},
				},
			},
			{
				Name: "Test",
				Targets: []model.Target{
					{
						Name:    "test",
						Summary: []string{"Run all tests."},
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

// TestMakeFormatter_RenderHelp_WithColors tests color code escaping
func TestMakeFormatter_RenderHelp_WithColors(t *testing.T) {
	t.Parallel()
	formatter := NewMakeFormatter(&FormatterConfig{UseColor: true})
	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: "Build",
				Targets: []model.Target{
					{
						Name:    "build",
						Aliases: []string{"b"},
						Summary: []string{"Build the project."},
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
	// ANSI codes should be escaped as \033 for Makefile
	if !strings.Contains(output, "\\033[") {
		t.Error("Output should contain escaped ANSI codes")
	}
	if !strings.Contains(output, "\\033[1;36m") {
		t.Error("Output should contain escaped cyan for category")
	}
	if !strings.Contains(output, "\\033[1;32m") {
		t.Error("Output should contain escaped green for target")
	}
	if !strings.Contains(output, "\\033[0;33m") {
		t.Error("Output should contain escaped yellow for alias")
	}
}

// TestMakeFormatter_RenderHelp_SpecialCharactersEscaped tests escaping of special characters
func TestMakeFormatter_RenderHelp_SpecialCharactersEscaped(t *testing.T) {
	t.Parallel()
	formatter := NewMakeFormatter(&FormatterConfig{UseColor: false})
	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: model.UncategorizedCategoryName,
				Targets: []model.Target{
					{
						Name:    "deploy",
						Summary: []string{`Use $VAR and "quotes" in command.`},
					},
					{
						Name:    "inject",
						Summary: []string{"Line1\nLine2\rLine3\tTabbed"},
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
	if !strings.Contains(output, "$$VAR") {
		t.Error("Output should escape $ as $$")
	}
	if !strings.Contains(output, "\\\"quotes\\\"") {
		t.Error("Output should escape double quotes")
	}
	if !strings.Contains(output, "\\n") {
		t.Error("Output should escape newline as \\n")
	}
	if !strings.Contains(output, "\\r") {
		t.Error("Output should escape carriage return as \\r")
	}
	if !strings.Contains(output, "\\t") {
		t.Error("Output should escape tab as \\t")
	}
}

// TestMakeFormatter_RenderDetailedTarget tests detailed target rendering
func TestMakeFormatter_RenderDetailedTarget(t *testing.T) {
	t.Parallel()
	formatter := NewMakeFormatter(&FormatterConfig{UseColor: false})
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
	if !strings.Contains(output, "GOOS: Target OS") {
		t.Error("Output should contain variable GOOS")
	}
	if !strings.Contains(output, "Build the project.") {
		t.Error("Output should contain documentation")
	}
	if !strings.Contains(output, "Source: /path/to/Makefile:42") {
		t.Error("Output should contain source location")
	}
}

// TestMakeFormatter_RenderBasicTarget tests basic target rendering
func TestMakeFormatter_RenderBasicTarget(t *testing.T) {
	t.Parallel()
	formatter := NewMakeFormatter(&FormatterConfig{UseColor: false})

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

// TestMakeFormatter_ContentType tests content type
func TestMakeFormatter_ContentType(t *testing.T) {
	t.Parallel()
	formatter := NewMakeFormatter(&FormatterConfig{})

	contentType := formatter.ContentType()
	if contentType != "text/x-makefile" {
		t.Errorf("ContentType() = %q, want %q", contentType, "text/x-makefile")
	}
}

// TestMakeFormatter_DefaultExtension tests default extension
func TestMakeFormatter_DefaultExtension(t *testing.T) {
	t.Parallel()
	formatter := NewMakeFormatter(&FormatterConfig{})

	ext := formatter.DefaultExtension()
	if ext != ".mk" {
		t.Errorf("DefaultExtension() = %q, want %q", ext, ".mk")
	}
}

// TestMakeFormatter_WithFileDocumentation tests file documentation rendering
func TestMakeFormatter_WithFileDocumentation(t *testing.T) {
	t.Parallel()
	formatter := NewMakeFormatter(&FormatterConfig{UseColor: false})
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

// TestMakeFormatter_WithIncludedFiles tests included files rendering
func TestMakeFormatter_WithIncludedFiles(t *testing.T) {
	t.Parallel()
	formatter := NewMakeFormatter(&FormatterConfig{UseColor: false})
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
	if !strings.Contains(output, "Included files:") {
		t.Error("Output should contain 'Included files:' header")
	}
	if !strings.Contains(output, "make/build.mk") {
		t.Error("Output should contain included file path")
	}
	if !strings.Contains(output, "Build tasks.") {
		t.Error("Output should contain included file documentation")
	}
}

// TestMakeFormatter_WithVariables tests target variables rendering
func TestMakeFormatter_WithVariables(t *testing.T) {
	t.Parallel()
	formatter := NewMakeFormatter(&FormatterConfig{UseColor: false})
	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: model.UncategorizedCategoryName,
				Targets: []model.Target{
					{
						Name:    "serve",
						Summary: []string{"Start server."},
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
