package format

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/sdlcforge/make-help/internal/model"
)

// TestJSONFormatter_RenderHelp_EmptyModel tests rendering an empty help model
func TestJSONFormatter_RenderHelp_EmptyModel(t *testing.T) {
	formatter := NewJSONFormatter(&FormatterConfig{UseColor: false})
	helpModel := &model.HelpModel{}

	var buf bytes.Buffer
	err := formatter.RenderHelp(helpModel, &buf)

	if err != nil {
		t.Fatalf("RenderHelp() error = %v", err)
	}

	// Parse JSON to verify it's valid
	var output jsonHelpOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Verify fields
	if output.Usage != "make [<target>...] [<ENV_VAR>=<value>...]" {
		t.Errorf("Usage = %q, want %q", output.Usage, "make [<target>...] [<ENV_VAR>=<value>...]")
	}
	if output.Description != "" {
		t.Errorf("Description should be empty, got %q", output.Description)
	}
	if len(output.Categories) != 0 {
		t.Errorf("Categories should be empty, got %d items", len(output.Categories))
	}
}

// TestJSONFormatter_RenderHelp_WithTargets tests rendering with basic targets
func TestJSONFormatter_RenderHelp_WithTargets(t *testing.T) {
	formatter := NewJSONFormatter(&FormatterConfig{UseColor: false})
	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: model.UncategorizedCategoryName,
				Targets: []model.Target{
					{
						Name:       "build",
						Summary:    []string{"Build the project."},
						SourceFile: "Makefile",
						LineNumber: 10,
					},
					{
						Name:       "test",
						Summary:    []string{"Run all tests."},
						SourceFile: "Makefile",
						LineNumber: 15,
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

	// Parse JSON
	var output jsonHelpOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Verify structure
	if len(output.Categories) != 1 {
		t.Fatalf("Expected 1 category, got %d", len(output.Categories))
	}

	category := output.Categories[0]
	if category.Name != model.UncategorizedCategoryName {
		t.Errorf("Category name = %q, want %q", category.Name, model.UncategorizedCategoryName)
	}

	if len(category.Targets) != 2 {
		t.Fatalf("Expected 2 targets, got %d", len(category.Targets))
	}

	// Verify build target
	build := category.Targets[0]
	if build.Name != "build" {
		t.Errorf("Target name = %q, want %q", build.Name, "build")
	}
	if build.Summary != "Build the project." {
		t.Errorf("Target summary = %q, want %q", build.Summary, "Build the project.")
	}
	if build.SourceFile != "Makefile" {
		t.Errorf("Target source = %q, want %q", build.SourceFile, "Makefile")
	}
	if build.LineNumber != 10 {
		t.Errorf("Target line = %d, want %d", build.LineNumber, 10)
	}
}

// TestJSONFormatter_RenderHelp_WithCategories tests rendering with categories
func TestJSONFormatter_RenderHelp_WithCategories(t *testing.T) {
	formatter := NewJSONFormatter(&FormatterConfig{UseColor: false})
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

	// Parse JSON
	var output jsonHelpOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Verify categories
	if len(output.Categories) != 2 {
		t.Fatalf("Expected 2 categories, got %d", len(output.Categories))
	}

	if output.Categories[0].Name != "Build" {
		t.Errorf("First category = %q, want %q", output.Categories[0].Name, "Build")
	}
	if output.Categories[1].Name != "Test" {
		t.Errorf("Second category = %q, want %q", output.Categories[1].Name, "Test")
	}
}

// TestJSONFormatter_RenderHelp_WithVariablesAndAliases tests variables and aliases
func TestJSONFormatter_RenderHelp_WithVariablesAndAliases(t *testing.T) {
	formatter := NewJSONFormatter(&FormatterConfig{UseColor: false})
	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: model.UncategorizedCategoryName,
				Targets: []model.Target{
					{
						Name:    "build",
						Aliases: []string{"b", "compile"},
						Summary: []string{"Build the project."},
						Variables: []model.Variable{
							{Name: "GOOS", Description: "Target OS"},
							{Name: "GOARCH", Description: "Target architecture"},
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

	// Parse JSON
	var output jsonHelpOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	target := output.Categories[0].Targets[0]

	// Verify aliases
	if len(target.Aliases) != 2 {
		t.Fatalf("Expected 2 aliases, got %d", len(target.Aliases))
	}
	if target.Aliases[0] != "b" || target.Aliases[1] != "compile" {
		t.Errorf("Aliases = %v, want [b, compile]", target.Aliases)
	}

	// Verify variables
	if len(target.Variables) != 2 {
		t.Fatalf("Expected 2 variables, got %d", len(target.Variables))
	}
	if target.Variables[0].Name != "GOOS" || target.Variables[0].Description != "Target OS" {
		t.Errorf("Variable[0] = %+v, want {Name:GOOS, Description:Target OS}", target.Variables[0])
	}
	if target.Variables[1].Name != "GOARCH" || target.Variables[1].Description != "Target architecture" {
		t.Errorf("Variable[1] = %+v, want {Name:GOARCH, Description:Target architecture}", target.Variables[1])
	}
}

// TestJSONFormatter_RenderHelp_WithFileDocumentation tests file documentation
func TestJSONFormatter_RenderHelp_WithFileDocumentation(t *testing.T) {
	formatter := NewJSONFormatter(&FormatterConfig{UseColor: false})
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

	// Parse JSON
	var output jsonHelpOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	expected := "This is the main Makefile.\nIt provides tasks."
	if output.Description != expected {
		t.Errorf("Description = %q, want %q", output.Description, expected)
	}
}

// TestJSONFormatter_RenderHelp_WithIncludedFiles tests included files
func TestJSONFormatter_RenderHelp_WithIncludedFiles(t *testing.T) {
	formatter := NewJSONFormatter(&FormatterConfig{UseColor: false})
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
			{
				SourceFile:     "make/test.mk",
				Documentation:  []string{"Test tasks.", "", "Includes unit and integration tests."},
				IsEntryPoint:   false,
				DiscoveryOrder: 2,
			},
		},
	}

	var buf bytes.Buffer
	err := formatter.RenderHelp(helpModel, &buf)

	if err != nil {
		t.Fatalf("RenderHelp() error = %v", err)
	}

	// Parse JSON
	var output jsonHelpOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Verify included files
	if len(output.IncludedFiles) != 2 {
		t.Fatalf("Expected 2 included files, got %d", len(output.IncludedFiles))
	}

	// First included file
	if output.IncludedFiles[0].Path != "make/build.mk" {
		t.Errorf("IncludedFiles[0].Path = %q, want %q", output.IncludedFiles[0].Path, "make/build.mk")
	}
	if output.IncludedFiles[0].Description != "Build tasks." {
		t.Errorf("IncludedFiles[0].Description = %q, want %q", output.IncludedFiles[0].Description, "Build tasks.")
	}

	// Second included file
	if output.IncludedFiles[1].Path != "make/test.mk" {
		t.Errorf("IncludedFiles[1].Path = %q, want %q", output.IncludedFiles[1].Path, "make/test.mk")
	}
	expectedDesc := "Test tasks.\n\nIncludes unit and integration tests."
	if output.IncludedFiles[1].Description != expectedDesc {
		t.Errorf("IncludedFiles[1].Description = %q, want %q", output.IncludedFiles[1].Description, expectedDesc)
	}
}

// TestJSONFormatter_RenderDetailedTarget tests detailed target rendering
func TestJSONFormatter_RenderDetailedTarget(t *testing.T) {
	formatter := NewJSONFormatter(&FormatterConfig{UseColor: false})
	target := &model.Target{
		Name:    "build",
		Aliases: []string{"b", "compile"},
		Summary: []string{"Build the project."},
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

	// Parse JSON
	var output jsonDetailedTarget
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Verify fields
	if output.Name != "build" {
		t.Errorf("Name = %q, want %q", output.Name, "build")
	}
	if output.Summary != "Build the project." {
		t.Errorf("Summary = %q, want %q", output.Summary, "Build the project.")
	}
	if len(output.Documentation) != 3 {
		t.Fatalf("Expected 3 documentation lines, got %d", len(output.Documentation))
	}
	if output.Documentation[0] != "Build the project." {
		t.Errorf("Documentation[0] = %q, want %q", output.Documentation[0], "Build the project.")
	}
	if len(output.Aliases) != 2 {
		t.Fatalf("Expected 2 aliases, got %d", len(output.Aliases))
	}
	if len(output.Variables) != 2 {
		t.Fatalf("Expected 2 variables, got %d", len(output.Variables))
	}
	if output.SourceFile != "/path/to/Makefile" {
		t.Errorf("SourceFile = %q, want %q", output.SourceFile, "/path/to/Makefile")
	}
	if output.LineNumber != 42 {
		t.Errorf("LineNumber = %d, want %d", output.LineNumber, 42)
	}
}

// TestJSONFormatter_RenderBasicTarget tests basic target rendering
func TestJSONFormatter_RenderBasicTarget(t *testing.T) {
	formatter := NewJSONFormatter(&FormatterConfig{UseColor: false})

	var buf bytes.Buffer
	err := formatter.RenderBasicTarget("undocumented", "/path/to/Makefile", 15, &buf)

	if err != nil {
		t.Fatalf("RenderBasicTarget() error = %v", err)
	}

	// Parse JSON
	var output jsonBasicTarget
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Verify fields
	if output.Name != "undocumented" {
		t.Errorf("Name = %q, want %q", output.Name, "undocumented")
	}
	if output.SourceFile != "/path/to/Makefile" {
		t.Errorf("SourceFile = %q, want %q", output.SourceFile, "/path/to/Makefile")
	}
	if output.LineNumber != 15 {
		t.Errorf("LineNumber = %d, want %d", output.LineNumber, 15)
	}
}

// TestJSONFormatter_ContentType tests content type
func TestJSONFormatter_ContentType(t *testing.T) {
	formatter := NewJSONFormatter(&FormatterConfig{})

	contentType := formatter.ContentType()
	if contentType != "application/json" {
		t.Errorf("ContentType() = %q, want %q", contentType, "application/json")
	}
}

// TestJSONFormatter_DefaultExtension tests default extension
func TestJSONFormatter_DefaultExtension(t *testing.T) {
	formatter := NewJSONFormatter(&FormatterConfig{})

	ext := formatter.DefaultExtension()
	if ext != ".json" {
		t.Errorf("DefaultExtension() = %q, want %q", ext, ".json")
	}
}

// TestJSONFormatter_NilModel tests nil model handling
func TestJSONFormatter_NilModel(t *testing.T) {
	formatter := NewJSONFormatter(&FormatterConfig{})

	var buf bytes.Buffer
	err := formatter.RenderHelp(nil, &buf)

	if err == nil {
		t.Error("RenderHelp() should return error for nil model")
	}
	if !strings.Contains(err.Error(), "cannot be nil") {
		t.Errorf("Error message should mention nil, got: %v", err)
	}
}

// TestJSONFormatter_NilTarget tests nil target handling
func TestJSONFormatter_NilTarget(t *testing.T) {
	formatter := NewJSONFormatter(&FormatterConfig{})

	var buf bytes.Buffer
	err := formatter.RenderDetailedTarget(nil, &buf)

	if err == nil {
		t.Error("RenderDetailedTarget() should return error for nil target")
	}
	if !strings.Contains(err.Error(), "cannot be nil") {
		t.Errorf("Error message should mention nil, got: %v", err)
	}
}

// TestJSONFormatter_ComplexModel tests a complex help model
func TestJSONFormatter_ComplexModel(t *testing.T) {
	formatter := NewJSONFormatter(&FormatterConfig{UseColor: false})
	helpModel := &model.HelpModel{
		FileDocs: []model.FileDoc{
			{
				SourceFile:     "Makefile",
				Documentation:  []string{"Project Makefile", "Common tasks"},
				IsEntryPoint:   true,
				DiscoveryOrder: 0,
			},
			{
				SourceFile:     "make/build.mk",
				Documentation:  []string{"Build tasks"},
				IsEntryPoint:   false,
				DiscoveryOrder: 1,
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
						Summary: []string{"Build the project."},
						Variables: []model.Variable{
							{Name: "GOOS"},
							{Name: "GOARCH"},
						},
						SourceFile: "Makefile",
						LineNumber: 10,
					},
				},
			},
			{
				Name: "Test",
				Targets: []model.Target{
					{
						Name:       "test",
						Aliases:    []string{"t"},
						Summary:    []string{"Run all tests."},
						SourceFile: "Makefile",
						LineNumber: 20,
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

	// Parse JSON
	var output jsonHelpOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Verify description
	if !strings.Contains(output.Description, "Project Makefile") {
		t.Error("Description should contain file documentation")
	}

	// Verify included files
	if len(output.IncludedFiles) != 1 {
		t.Fatalf("Expected 1 included file, got %d", len(output.IncludedFiles))
	}
	if output.IncludedFiles[0].Path != "make/build.mk" {
		t.Errorf("Included file path = %q, want %q", output.IncludedFiles[0].Path, "make/build.mk")
	}

	// Verify categories
	if len(output.Categories) != 2 {
		t.Fatalf("Expected 2 categories, got %d", len(output.Categories))
	}
	if output.Categories[0].Name != "Build" {
		t.Errorf("Category[0] name = %q, want %q", output.Categories[0].Name, "Build")
	}
	if output.Categories[1].Name != "Test" {
		t.Errorf("Category[1] name = %q, want %q", output.Categories[1].Name, "Test")
	}

	// Verify build target
	build := output.Categories[0].Targets[0]
	if build.Name != "build" {
		t.Errorf("Build target name = %q, want %q", build.Name, "build")
	}
	if len(build.Aliases) != 1 || build.Aliases[0] != "b" {
		t.Errorf("Build aliases = %v, want [b]", build.Aliases)
	}
	if len(build.Variables) != 2 {
		t.Errorf("Build variables count = %d, want 2", len(build.Variables))
	}
}

// TestJSONFormatter_PlainTextSummary tests that summaries use plain text
func TestJSONFormatter_PlainTextSummary(t *testing.T) {
	formatter := NewJSONFormatter(&FormatterConfig{UseColor: false})

	// Create a summary (stored as plain string in []string)
	// Note: With the new design, summaries are stored as plain text in []string
	// and formatters parse them to RichText only when needed for display
	summary := []string{"Build with debug mode."}

	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: model.UncategorizedCategoryName,
				Targets: []model.Target{
					{
						Name:    "build",
						Summary: summary,
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

	// Parse JSON
	var output jsonHelpOutput
	if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Verify that the summary is plain text (no markdown)
	expected := "Build with debug mode."
	if output.Categories[0].Targets[0].Summary != expected {
		t.Errorf("Summary = %q, want %q", output.Categories[0].Targets[0].Summary, expected)
	}
}
