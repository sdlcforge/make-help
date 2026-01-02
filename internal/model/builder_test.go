package model

import (
	"testing"

	"github.com/sdlcforge/make-help/internal/errors"
	"github.com/sdlcforge/make-help/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuilder(t *testing.T) {
	config := &BuilderConfig{DefaultCategory: ""}
	builder := NewBuilder(config)

	assert.NotNil(t, builder)
	assert.NotNil(t, builder.config)
	assert.NotNil(t, builder.extractor)
}

func TestBuild_EmptyParsedFiles(t *testing.T) {
	config := &BuilderConfig{DefaultCategory: ""}
	builder := NewBuilder(config)

	model, err := builder.Build([]*parser.ParsedFile{})

	require.NoError(t, err)
	assert.NotNil(t, model)
	assert.Empty(t, model.FileDocs)
	assert.Empty(t, model.Categories)
	assert.False(t, model.HasCategories)
}

func TestBuild_FileDocumentation(t *testing.T) {
	config := &BuilderConfig{DefaultCategory: ""}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveFile, Value: "Main project Makefile", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveFile, Value: "Build tools and utilities", SourceFile: "Makefile", LineNumber: 2},
			},
			TargetMap: map[string]int{},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	assert.Len(t, model.FileDocs, 1)
	assert.Equal(t, "Makefile", model.FileDocs[0].SourceFile)
	assert.True(t, model.FileDocs[0].IsEntryPoint)
	assert.Equal(t, 0, model.FileDocs[0].DiscoveryOrder)
	// Multiple !file blocks in same file are concatenated with blank line separation
	assert.Equal(t, []string{"Main project Makefile", "", "Build tools and utilities"}, model.FileDocs[0].Documentation)
}

func TestBuild_BasicTargetWithDocs(t *testing.T) {
	config := &BuilderConfig{DefaultCategory: ""}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveDoc, Value: "Build the project.", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveDoc, Value: "Compiles all source files.", SourceFile: "Makefile", LineNumber: 2},
			},
			TargetMap: map[string]int{
				"build": 3,
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	assert.Len(t, model.Categories, 1)

	// Find the uncategorized category
	var uncategorized *Category
	for i := range model.Categories {
		if model.Categories[i].Name == UncategorizedCategoryName {
			uncategorized = &model.Categories[i]
			break
		}
	}
	require.NotNil(t, uncategorized)
	assert.Len(t, uncategorized.Targets, 1)
	assert.Equal(t, "build", uncategorized.Targets[0].Name)
	assert.Len(t, uncategorized.Targets[0].Documentation, 2)
	assert.Equal(t, "Build the project.", uncategorized.Targets[0].Summary.PlainText())
}

func TestBuild_TargetWithCategory(t *testing.T) {
	config := &BuilderConfig{DefaultCategory: ""}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveCategory, Value: "Build", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveDoc, Value: "Build the project.", SourceFile: "Makefile", LineNumber: 2},
			},
			TargetMap: map[string]int{
				"build": 3,
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	assert.True(t, model.HasCategories)
	assert.Len(t, model.Categories, 1)
	assert.Equal(t, "Build", model.Categories[0].Name)
	assert.Len(t, model.Categories[0].Targets, 1)
	assert.Equal(t, "build", model.Categories[0].Targets[0].Name)
}

func TestBuild_MultipleCategories(t *testing.T) {
	config := &BuilderConfig{DefaultCategory: ""}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveCategory, Value: "Build", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveDoc, Value: "Build the project.", SourceFile: "Makefile", LineNumber: 2},
				{Type: parser.DirectiveCategory, Value: "Test", SourceFile: "Makefile", LineNumber: 5},
				{Type: parser.DirectiveDoc, Value: "Run tests.", SourceFile: "Makefile", LineNumber: 6},
			},
			TargetMap: map[string]int{
				"build": 3,
				"test":  7,
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	assert.True(t, model.HasCategories)
	assert.Len(t, model.Categories, 2)

	categoryNames := make(map[string]bool)
	for _, cat := range model.Categories {
		categoryNames[cat.Name] = true
	}
	assert.True(t, categoryNames["Build"])
	assert.True(t, categoryNames["Test"])
}

func TestBuild_CategorySwitchBehavior(t *testing.T) {
	// Test that !category acts as a switch - subsequent targets
	// inherit the category until another !category is encountered
	config := &BuilderConfig{DefaultCategory: ""}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveCategory, Value: "Build", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveDoc, Value: "Compile the project.", SourceFile: "Makefile", LineNumber: 2},
				{Type: parser.DirectiveDoc, Value: "Build with debug.", SourceFile: "Makefile", LineNumber: 5},
				{Type: parser.DirectiveCategory, Value: "Test", SourceFile: "Makefile", LineNumber: 8},
				{Type: parser.DirectiveDoc, Value: "Run unit tests.", SourceFile: "Makefile", LineNumber: 9},
				{Type: parser.DirectiveDoc, Value: "Run integration tests.", SourceFile: "Makefile", LineNumber: 12},
			},
			TargetMap: map[string]int{
				"build":            3,
				"build-debug":      6,
				"test":             10,
				"test-integration": 13,
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	assert.True(t, model.HasCategories)
	assert.Len(t, model.Categories, 2)

	// Find Build category - should have 2 targets
	buildCat := findCategory(model, "Build")
	require.NotNil(t, buildCat, "Build category should exist")
	assert.Len(t, buildCat.Targets, 2, "Build category should have 2 targets")
	targetNames := make(map[string]bool)
	for _, target := range buildCat.Targets {
		targetNames[target.Name] = true
	}
	assert.True(t, targetNames["build"], "Build category should contain 'build' target")
	assert.True(t, targetNames["build-debug"], "Build category should contain 'build-debug' target")

	// Find Test category - should have 2 targets
	testCat := findCategory(model, "Test")
	require.NotNil(t, testCat, "Test category should exist")
	assert.Len(t, testCat.Targets, 2, "Test category should have 2 targets")
	targetNames = make(map[string]bool)
	for _, target := range testCat.Targets {
		targetNames[target.Name] = true
	}
	assert.True(t, targetNames["test"], "Test category should contain 'test' target")
	assert.True(t, targetNames["test-integration"], "Test category should contain 'test-integration' target")
}

func TestBuild_TargetWithVariables(t *testing.T) {
	config := &BuilderConfig{DefaultCategory: ""}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveDoc, Value: "Build the project.", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveVar, Value: "DEBUG - Enable debug mode", SourceFile: "Makefile", LineNumber: 2},
				{Type: parser.DirectiveVar, Value: "PORT - Port number", SourceFile: "Makefile", LineNumber: 3},
			},
			TargetMap: map[string]int{
				"build": 4,
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	require.Len(t, model.Categories, 1)
	require.Len(t, model.Categories[0].Targets, 1)

	target := model.Categories[0].Targets[0]
	assert.Len(t, target.Variables, 2)
	assert.Equal(t, "DEBUG", target.Variables[0].Name)
	assert.Equal(t, "Enable debug mode", target.Variables[0].Description)
	assert.Equal(t, "PORT", target.Variables[1].Name)
	assert.Equal(t, "Port number", target.Variables[1].Description)
}

func TestBuild_TargetWithAliases(t *testing.T) {
	config := &BuilderConfig{DefaultCategory: ""}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveDoc, Value: "Build the project.", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveAlias, Value: "b, compile", SourceFile: "Makefile", LineNumber: 2},
			},
			TargetMap: map[string]int{
				"build": 3,
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	require.Len(t, model.Categories, 1)
	require.Len(t, model.Categories[0].Targets, 1)

	target := model.Categories[0].Targets[0]
	assert.Len(t, target.Aliases, 2)
	assert.Contains(t, target.Aliases, "b")
	assert.Contains(t, target.Aliases, "compile")
}

func TestBuild_MixedCategorizationError(t *testing.T) {
	config := &BuilderConfig{DefaultCategory: ""}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveCategory, Value: "Build", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveDoc, Value: "Build the project.", SourceFile: "Makefile", LineNumber: 2},
			},
			TargetMap: map[string]int{
				"build": 3,
			},
		},
		{
			Path: "include.mk",
			Directives: []parser.Directive{
				// No category directive - target will be uncategorized
				{Type: parser.DirectiveDoc, Value: "Run tests.", SourceFile: "include.mk", LineNumber: 1},
			},
			TargetMap: map[string]int{
				"test": 2,
			},
		},
	}

	_, err := builder.Build(parsedFiles)

	require.Error(t, err)
	assert.IsType(t, &errors.MixedCategorizationError{}, err)
}

func TestBuild_MixedCategorizationWithDefaultCategory(t *testing.T) {
	defaultCategory := "Other"
	config := &BuilderConfig{DefaultCategory: defaultCategory}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveCategory, Value: "Build", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveDoc, Value: "Build the project.", SourceFile: "Makefile", LineNumber: 2},
			},
			TargetMap: map[string]int{
				"build": 3,
			},
		},
		{
			Path: "include.mk",
			Directives: []parser.Directive{
				{Type: parser.DirectiveDoc, Value: "Run tests.", SourceFile: "include.mk", LineNumber: 1},
			},
			TargetMap: map[string]int{
				"test": 2,
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)

	// Should have Build and Other categories
	categoryNames := make(map[string]bool)
	for _, cat := range model.Categories {
		categoryNames[cat.Name] = true
	}
	assert.True(t, categoryNames["Build"])
	assert.True(t, categoryNames["Other"])
}

func TestBuild_SplitCategories(t *testing.T) {
	config := &BuilderConfig{DefaultCategory: ""}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveCategory, Value: "Build", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveDoc, Value: "Build the project.", SourceFile: "Makefile", LineNumber: 2},
			},
			TargetMap: map[string]int{
				"build": 3,
			},
		},
		{
			Path: "include.mk",
			Directives: []parser.Directive{
				// Same category name in different file
				{Type: parser.DirectiveCategory, Value: "Build", SourceFile: "include.mk", LineNumber: 1},
				{Type: parser.DirectiveDoc, Value: "Build with debug.", SourceFile: "include.mk", LineNumber: 2},
			},
			TargetMap: map[string]int{
				"build-debug": 3,
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)

	// Should merge into single Build category
	buildCategory := findCategory(model, "Build")
	require.NotNil(t, buildCategory)
	assert.Len(t, buildCategory.Targets, 2)
}

func TestBuild_DiscoveryOrder(t *testing.T) {
	config := &BuilderConfig{DefaultCategory: ""}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveCategory, Value: "Zebra", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveDoc, Value: "Zebra target.", SourceFile: "Makefile", LineNumber: 2},
				{Type: parser.DirectiveCategory, Value: "Alpha", SourceFile: "Makefile", LineNumber: 5},
				{Type: parser.DirectiveDoc, Value: "Alpha target.", SourceFile: "Makefile", LineNumber: 6},
			},
			TargetMap: map[string]int{
				"zebra": 3,
				"alpha": 7,
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)

	// Find categories and check discovery order
	zebra := findCategory(model, "Zebra")
	alpha := findCategory(model, "Alpha")
	require.NotNil(t, zebra)
	require.NotNil(t, alpha)

	// Zebra was discovered first
	assert.Less(t, zebra.DiscoveryOrder, alpha.DiscoveryOrder)
}

func TestBuild_SummaryExtraction(t *testing.T) {
	config := &BuilderConfig{DefaultCategory: ""}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveDoc, Value: "Build the project. This compiles all source files.", SourceFile: "Makefile", LineNumber: 1},
			},
			TargetMap: map[string]int{
				"build": 2,
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	require.Len(t, model.Categories, 1)
	require.Len(t, model.Categories[0].Targets, 1)

	// Summary should be first sentence only
	assert.Equal(t, "Build the project.", model.Categories[0].Targets[0].Summary.PlainText())
}

func TestBuild_TargetSourceTracking(t *testing.T) {
	config := &BuilderConfig{DefaultCategory: ""}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "/path/to/Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveDoc, Value: "Build target.", SourceFile: "/path/to/Makefile", LineNumber: 1},
			},
			TargetMap: map[string]int{
				"build": 2,
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	require.Len(t, model.Categories, 1)
	require.Len(t, model.Categories[0].Targets, 1)

	target := model.Categories[0].Targets[0]
	assert.Equal(t, "/path/to/Makefile", target.SourceFile)
	assert.Equal(t, 2, target.LineNumber)
}

func TestBuild_MultipleFilesAggregation(t *testing.T) {
	config := &BuilderConfig{DefaultCategory: ""}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveFile, Value: "Main Makefile", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveDoc, Value: "Build target.", SourceFile: "Makefile", LineNumber: 2},
			},
			TargetMap: map[string]int{
				"build": 3,
			},
		},
		{
			Path: "include.mk",
			Directives: []parser.Directive{
				{Type: parser.DirectiveFile, Value: "Include file", SourceFile: "include.mk", LineNumber: 1},
				{Type: parser.DirectiveDoc, Value: "Test target.", SourceFile: "include.mk", LineNumber: 2},
			},
			TargetMap: map[string]int{
				"test": 3,
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	assert.Len(t, model.FileDocs, 2)
	// First file should be entry point
	assert.Equal(t, "Makefile", model.FileDocs[0].SourceFile)
	assert.Equal(t, []string{"Main Makefile"}, model.FileDocs[0].Documentation)
	assert.True(t, model.FileDocs[0].IsEntryPoint)
	assert.Equal(t, 0, model.FileDocs[0].DiscoveryOrder)
	// Second file should not be entry point
	assert.Equal(t, "include.mk", model.FileDocs[1].SourceFile)
	assert.Equal(t, []string{"Include file"}, model.FileDocs[1].Documentation)
	assert.False(t, model.FileDocs[1].IsEntryPoint)
	assert.Equal(t, 1, model.FileDocs[1].DiscoveryOrder)
}

func TestParseVarDirective(t *testing.T) {
	builder := NewBuilder(&BuilderConfig{DefaultCategory: ""})

	tests := []struct {
		name     string
		input    string
		wantName string
		wantDesc string
	}{
		{
			name:     "with description",
			input:    "DEBUG - Enable debug mode",
			wantName: "DEBUG",
			wantDesc: "Enable debug mode",
		},
		{
			name:     "without description",
			input:    "DEBUG",
			wantName: "DEBUG",
			wantDesc: "",
		},
		{
			name:     "with extra spaces",
			input:    "  DEBUG  -  Enable debug mode  ",
			wantName: "DEBUG",
			wantDesc: "Enable debug mode",
		},
		{
			name:     "description with hyphen",
			input:    "PORT - The port number - defaults to 8080",
			wantName: "PORT",
			wantDesc: "The port number - defaults to 8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.parseVarDirective(tt.input)
			assert.Equal(t, tt.wantName, result.Name)
			assert.Equal(t, tt.wantDesc, result.Description)
		})
	}
}

func TestParseAliasDirective(t *testing.T) {
	builder := NewBuilder(&BuilderConfig{DefaultCategory: ""})

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "single alias",
			input: "b",
			want:  []string{"b"},
		},
		{
			name:  "multiple aliases",
			input: "b, compile, c",
			want:  []string{"b", "compile", "c"},
		},
		{
			name:  "with extra spaces",
			input: "  b  ,  compile  ,  c  ",
			want:  []string{"b", "compile", "c"},
		},
		{
			name:  "empty entries filtered",
			input: "b, , c",
			want:  []string{"b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.parseAliasDirective(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestBuild_NoDocTargetsFiltered(t *testing.T) {
	// Test that targets without documentation are filtered by default
	config := &BuilderConfig{DefaultCategory: ""}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path:       "Makefile",
			Directives: []parser.Directive{},
			TargetMap: map[string]int{
				"build": 1,
				"test":  2,
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	// With no documentation, targets should be filtered out
	if len(model.Categories) > 0 {
		assert.Len(t, model.Categories[0].Targets, 0)
	}
}

func TestBuild_EmptyFileDocValue(t *testing.T) {
	config := &BuilderConfig{DefaultCategory: ""}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveFile, Value: "", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveFile, Value: "Has content", SourceFile: "Makefile", LineNumber: 2},
			},
			TargetMap: map[string]int{},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	// Empty file doc values should be skipped, but the file should still appear with only non-empty content
	assert.Len(t, model.FileDocs, 1)
	assert.Equal(t, "Makefile", model.FileDocs[0].SourceFile)
	assert.Equal(t, []string{"Has content"}, model.FileDocs[0].Documentation)
	assert.True(t, model.FileDocs[0].IsEntryPoint)
}

// Helper function to find a category by name
func findCategory(model *HelpModel, name string) *Category {
	for i := range model.Categories {
		if model.Categories[i].Name == name {
			return &model.Categories[i]
		}
	}
	return nil
}

func TestBuild_FilterUndocumented(t *testing.T) {
	// Test that undocumented targets are excluded by default
	config := &BuilderConfig{
		DefaultCategory: "",
		PhonyTargets:    map[string]bool{},
	}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveDoc, Value: "Documented target.", SourceFile: "Makefile", LineNumber: 1},
			},
			TargetMap: map[string]int{
				"build":          2, // documented
				"undocumented":   3, // undocumented - should be filtered out
				"another-hidden": 4, // undocumented - should be filtered out
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	assert.Len(t, model.Categories, 1)
	assert.Len(t, model.Categories[0].Targets, 1)
	assert.Equal(t, "build", model.Categories[0].Targets[0].Name)
}

func TestBuild_IncludeTargets(t *testing.T) {
	// Test --include-target includes specific undocumented targets
	config := &BuilderConfig{
		DefaultCategory: "",
		IncludeTargets:  []string{"undocumented"},
		PhonyTargets:    map[string]bool{},
	}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveDoc, Value: "Documented target.", SourceFile: "Makefile", LineNumber: 1},
			},
			TargetMap: map[string]int{
				"build":        2, // documented
				"undocumented": 3, // undocumented but explicitly included
				"hidden":       4, // undocumented and not included
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	assert.Len(t, model.Categories, 1)
	assert.Len(t, model.Categories[0].Targets, 2)

	// Check both targets are present
	targetNames := []string{model.Categories[0].Targets[0].Name, model.Categories[0].Targets[1].Name}
	assert.Contains(t, targetNames, "build")
	assert.Contains(t, targetNames, "undocumented")
}

func TestBuild_IncludeAllPhony(t *testing.T) {
	// Test --include-all-phony includes all .PHONY targets
	config := &BuilderConfig{
		DefaultCategory: "",
		IncludeAllPhony: true,
		PhonyTargets: map[string]bool{
			"clean": true,
			"test":  true,
		},
	}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveDoc, Value: "Documented target.", SourceFile: "Makefile", LineNumber: 1},
			},
			TargetMap: map[string]int{
				"build":  2, // documented
				"clean":  3, // undocumented but .PHONY
				"test":   4, // undocumented but .PHONY
				"hidden": 5, // undocumented and not .PHONY - should be filtered out
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	assert.Len(t, model.Categories, 1)
	assert.Len(t, model.Categories[0].Targets, 3)

	// Check expected targets are present
	targetNames := make(map[string]bool)
	for _, target := range model.Categories[0].Targets {
		targetNames[target.Name] = true
	}
	assert.True(t, targetNames["build"])
	assert.True(t, targetNames["clean"])
	assert.True(t, targetNames["test"])
	assert.False(t, targetNames["hidden"])
}

func TestBuild_PhonyStatusSet(t *testing.T) {
	// Test that IsPhony field is correctly set on targets
	config := &BuilderConfig{
		DefaultCategory: "",
		PhonyTargets: map[string]bool{
			"clean": true,
			"test":  true,
		},
	}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveDoc, Value: "Build target.", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveDoc, Value: "Clean target.", SourceFile: "Makefile", LineNumber: 3},
			},
			TargetMap: map[string]int{
				"build": 2, // not phony
				"clean": 4, // phony
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	assert.Len(t, model.Categories, 1)
	assert.Len(t, model.Categories[0].Targets, 2)

	// Find each target and check IsPhony
	for _, target := range model.Categories[0].Targets {
		switch target.Name {
		case "build":
			assert.False(t, target.IsPhony, "build should not be phony")
		case "clean":
			assert.True(t, target.IsPhony, "clean should be phony")
		}
	}
}

func TestBuild_CombinedFiltering(t *testing.T) {
	// Test combination of documented, included, and phony targets
	config := &BuilderConfig{
		DefaultCategory: "",
		IncludeTargets:  []string{"special"},
		IncludeAllPhony: true,
		PhonyTargets: map[string]bool{
			"clean": true,
		},
	}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveDoc, Value: "Documented target.", SourceFile: "Makefile", LineNumber: 1},
			},
			TargetMap: map[string]int{
				"build":   2, // documented
				"clean":   3, // undocumented but phony
				"special": 4, // undocumented but explicitly included
				"hidden":  5, // undocumented, not phony, not included - should be filtered
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	assert.Len(t, model.Categories, 1)
	assert.Len(t, model.Categories[0].Targets, 3)

	// Check expected targets are present
	targetNames := make(map[string]bool)
	for _, target := range model.Categories[0].Targets {
		targetNames[target.Name] = true
	}
	assert.True(t, targetNames["build"])
	assert.True(t, targetNames["clean"])
	assert.True(t, targetNames["special"])
	assert.False(t, targetNames["hidden"])
}

func TestBuild_CategoryReset(t *testing.T) {
	// Test that !category _ resets the category to empty/uncategorized
	// This should create a mixed categorization error without a default category
	config := &BuilderConfig{DefaultCategory: "Misc"}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveCategory, Value: "Build", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveDoc, Value: "Build the project.", SourceFile: "Makefile", LineNumber: 2},
				{Type: parser.DirectiveCategory, Value: "_", SourceFile: "Makefile", LineNumber: 5},
				{Type: parser.DirectiveDoc, Value: "Clean artifacts.", SourceFile: "Makefile", LineNumber: 6},
			},
			TargetMap: map[string]int{
				"build": 3,
				"clean": 7,
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	assert.True(t, model.HasCategories, "HasCategories should be true")
	assert.Len(t, model.Categories, 2, "Should have 2 categories: Build and Misc")

	// Find Build category - should have 1 target
	buildCat := findCategory(model, "Build")
	require.NotNil(t, buildCat, "Build category should exist")
	assert.Len(t, buildCat.Targets, 1, "Build category should have 1 target")
	assert.Equal(t, "build", buildCat.Targets[0].Name)

	// Find Misc category (default applied to uncategorized)
	miscCat := findCategory(model, "Misc")
	require.NotNil(t, miscCat, "Misc category should exist")
	assert.Len(t, miscCat.Targets, 1, "Misc category should have 1 target")
	assert.Equal(t, "clean", miscCat.Targets[0].Name)
}

func TestBuild_CategoryResetMixedError(t *testing.T) {
	// Test that !category _ creates mixed categorization error
	// when there are both categorized and uncategorized targets
	config := &BuilderConfig{DefaultCategory: ""}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveCategory, Value: "Build", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveDoc, Value: "Build the project.", SourceFile: "Makefile", LineNumber: 2},
				{Type: parser.DirectiveCategory, Value: "_", SourceFile: "Makefile", LineNumber: 5},
				{Type: parser.DirectiveDoc, Value: "Clean artifacts.", SourceFile: "Makefile", LineNumber: 6},
			},
			TargetMap: map[string]int{
				"build": 3,
				"clean": 7,
			},
		},
	}

	_, err := builder.Build(parsedFiles)

	require.Error(t, err)
	assert.IsType(t, &errors.MixedCategorizationError{}, err)
}

func TestBuild_CategoryResetWithDefaultCategory(t *testing.T) {
	// Test that !category _ with --default-category resolves the mixed categorization
	defaultCategory := "Other"
	config := &BuilderConfig{DefaultCategory: defaultCategory}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveCategory, Value: "Build", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveDoc, Value: "Build the project.", SourceFile: "Makefile", LineNumber: 2},
				{Type: parser.DirectiveCategory, Value: "_", SourceFile: "Makefile", LineNumber: 5},
				{Type: parser.DirectiveDoc, Value: "Clean artifacts.", SourceFile: "Makefile", LineNumber: 6},
			},
			TargetMap: map[string]int{
				"build": 3,
				"clean": 7,
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	assert.True(t, model.HasCategories)
	assert.Len(t, model.Categories, 2, "Should have 2 categories: Build and Other")

	// Find Build category
	buildCat := findCategory(model, "Build")
	require.NotNil(t, buildCat)
	assert.Len(t, buildCat.Targets, 1)
	assert.Equal(t, "build", buildCat.Targets[0].Name)

	// Find Other category (default category applied to uncategorized)
	otherCat := findCategory(model, "Other")
	require.NotNil(t, otherCat, "Other category should exist")
	assert.Len(t, otherCat.Targets, 1)
	assert.Equal(t, "clean", otherCat.Targets[0].Name)
}

func TestBuild_CategoryResetMultipleTimes(t *testing.T) {
	// Test that !category _ can be used multiple times
	config := &BuilderConfig{DefaultCategory: "Misc"}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveCategory, Value: "Build", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveDoc, Value: "Build the project.", SourceFile: "Makefile", LineNumber: 2},
				{Type: parser.DirectiveCategory, Value: "_", SourceFile: "Makefile", LineNumber: 5},
				{Type: parser.DirectiveDoc, Value: "Clean artifacts.", SourceFile: "Makefile", LineNumber: 6},
				{Type: parser.DirectiveCategory, Value: "Test", SourceFile: "Makefile", LineNumber: 9},
				{Type: parser.DirectiveDoc, Value: "Run tests.", SourceFile: "Makefile", LineNumber: 10},
				{Type: parser.DirectiveCategory, Value: "_", SourceFile: "Makefile", LineNumber: 13},
				{Type: parser.DirectiveDoc, Value: "Show help.", SourceFile: "Makefile", LineNumber: 14},
			},
			TargetMap: map[string]int{
				"build": 3,
				"clean": 7,
				"test":  11,
				"help":  15,
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	assert.Len(t, model.Categories, 3, "Should have 3 categories: Build, Test, and Misc")

	// Find Build category
	buildCat := findCategory(model, "Build")
	require.NotNil(t, buildCat)
	assert.Len(t, buildCat.Targets, 1)
	assert.Equal(t, "build", buildCat.Targets[0].Name)

	// Find Test category
	testCat := findCategory(model, "Test")
	require.NotNil(t, testCat)
	assert.Len(t, testCat.Targets, 1)
	assert.Equal(t, "test", testCat.Targets[0].Name)

	// Find Misc category (default applied to both uncategorized targets)
	miscCat := findCategory(model, "Misc")
	require.NotNil(t, miscCat)
	assert.Len(t, miscCat.Targets, 2, "Misc should have 2 targets from reset")
	targetNames := make(map[string]bool)
	for _, target := range miscCat.Targets {
		targetNames[target.Name] = true
	}
	assert.True(t, targetNames["clean"])
	assert.True(t, targetNames["help"])
}

func TestBuild_CategoryResetNoTargetsAfter(t *testing.T) {
	// Test !category _ with no targets following it (edge case)
	config := &BuilderConfig{DefaultCategory: ""}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveCategory, Value: "Build", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveDoc, Value: "Build the project.", SourceFile: "Makefile", LineNumber: 2},
				{Type: parser.DirectiveCategory, Value: "_", SourceFile: "Makefile", LineNumber: 5},
			},
			TargetMap: map[string]int{
				"build": 3,
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	assert.Len(t, model.Categories, 1, "Should only have Build category")
	assert.Equal(t, "Build", model.Categories[0].Name)
	assert.Len(t, model.Categories[0].Targets, 1)
}

func TestBuild_CategoryResetAtStart(t *testing.T) {
	// Test !category _ at the start of a file (before any other category)
	// All targets should be uncategorized
	config := &BuilderConfig{DefaultCategory: "Misc"}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveCategory, Value: "_", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveDoc, Value: "Build the project.", SourceFile: "Makefile", LineNumber: 2},
				{Type: parser.DirectiveDoc, Value: "Run tests.", SourceFile: "Makefile", LineNumber: 5},
			},
			TargetMap: map[string]int{
				"build": 3,
				"test":  6,
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	// With !category _ at start, HasCategories is true but all targets are uncategorized
	// DefaultCategory "Misc" should collect them
	assert.Len(t, model.Categories, 1)
	assert.Equal(t, "Misc", model.Categories[0].Name)
	assert.Len(t, model.Categories[0].Targets, 2)
}

func TestBuild_CategoryResetAtStartNoDefault(t *testing.T) {
	// Test !category _ at start without default category
	// This should NOT error - there's no mixing since no real categories exist.
	// All targets are simply uncategorized.
	config := &BuilderConfig{DefaultCategory: ""}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveCategory, Value: "_", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveDoc, Value: "Build the project.", SourceFile: "Makefile", LineNumber: 2},
			},
			TargetMap: map[string]int{
				"build": 3,
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	// Should NOT error - no actual categories exist, just uncategorized targets
	require.NoError(t, err)
	// The targets end up in the uncategorized (empty name) bucket
	assert.Len(t, model.Categories, 1)
	assert.Equal(t, "", model.Categories[0].Name)
	assert.Len(t, model.Categories[0].Targets, 1)
}

func TestBuild_DuplicateTargetInMultipleFiles(t *testing.T) {
	// Test that when the same target is defined in multiple files,
	// only the first occurrence is used (first file wins).
	config := &BuilderConfig{DefaultCategory: ""}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveDoc, Value: "First documentation for build target.", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveVar, Value: "DEBUG - Debug mode from first file", SourceFile: "Makefile", LineNumber: 2},
			},
			TargetMap: map[string]int{
				"build": 3,
			},
		},
		{
			Path: "include.mk",
			Directives: []parser.Directive{
				{Type: parser.DirectiveDoc, Value: "Second documentation for build target.", SourceFile: "include.mk", LineNumber: 1},
				{Type: parser.DirectiveDoc, Value: "This should be ignored.", SourceFile: "include.mk", LineNumber: 2},
				{Type: parser.DirectiveVar, Value: "PORT - Port from second file", SourceFile: "include.mk", LineNumber: 3},
			},
			TargetMap: map[string]int{
				"build": 4, // Same target name - should be skipped
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	assert.Len(t, model.Categories, 1)
	assert.Len(t, model.Categories[0].Targets, 1)

	// Verify that the first file's documentation is used
	target := model.Categories[0].Targets[0]
	assert.Equal(t, "build", target.Name)
	assert.Equal(t, "Makefile", target.SourceFile)
	assert.Len(t, target.Documentation, 1)
	assert.Equal(t, "First documentation for build target.", target.Documentation[0])
	assert.Equal(t, "First documentation for build target.", target.Summary.PlainText())

	// Verify variables are from the first file only
	assert.Len(t, target.Variables, 1)
	assert.Equal(t, "DEBUG", target.Variables[0].Name)
	assert.Equal(t, "Debug mode from first file", target.Variables[0].Description)
}

func TestBuild_DocumentationWithoutTargets(t *testing.T) {
	// Test that documentation comments without any following target
	// are handled gracefully. Documentation accumulates until a target is found,
	// so all directives before a target get attached to it.
	// Only directives AFTER all targets (orphaned at end of file) are discarded.
	config := &BuilderConfig{DefaultCategory: ""}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveDoc, Value: "Build the project.", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveDoc, Value: "Orphaned doc after target.", SourceFile: "Makefile", LineNumber: 5},
				{Type: parser.DirectiveVar, Value: "ORPHAN - This variable has no target", SourceFile: "Makefile", LineNumber: 6},
				{Type: parser.DirectiveAlias, Value: "orphan1, orphan2", SourceFile: "Makefile", LineNumber: 7},
			},
			TargetMap: map[string]int{
				"build": 2, // Target appears early
				// No more targets after line 2, so directives at lines 5-7 are orphaned
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	assert.Len(t, model.Categories, 1)
	assert.Len(t, model.Categories[0].Targets, 1)

	// Verify the target has only the documentation that preceded it
	target := model.Categories[0].Targets[0]
	assert.Equal(t, "build", target.Name)
	assert.Len(t, target.Documentation, 1)
	assert.Equal(t, "Build the project.", target.Documentation[0])
	// Orphaned directives after the target should be discarded
	assert.Len(t, target.Variables, 0, "Orphaned variables after all targets should not be attached")
	assert.Len(t, target.Aliases, 0, "Orphaned aliases after all targets should not be attached")
}

func TestBuild_UndocumentedPhonyWithIncludeAllPhony(t *testing.T) {
	// Test edge case: undocumented .PHONY targets are included with IncludeAllPhony,
	// and they should have empty documentation, no summary, but still appear in output.
	config := &BuilderConfig{
		DefaultCategory: "",
		IncludeAllPhony: true,
		PhonyTargets: map[string]bool{
			"clean":   true,
			"test":    true,
			"install": true,
		},
	}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				// Only one target has documentation
				{Type: parser.DirectiveDoc, Value: "Remove build artifacts.", SourceFile: "Makefile", LineNumber: 1},
			},
			TargetMap: map[string]int{
				"clean":   2, // documented
				"test":    3, // undocumented but .PHONY
				"install": 4, // undocumented but .PHONY
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	assert.Len(t, model.Categories, 1)
	assert.Len(t, model.Categories[0].Targets, 3)

	// Find each target and verify their properties
	targetMap := make(map[string]*Target)
	for i := range model.Categories[0].Targets {
		target := &model.Categories[0].Targets[i]
		targetMap[target.Name] = target
	}

	// Check clean target (documented)
	require.NotNil(t, targetMap["clean"])
	assert.Len(t, targetMap["clean"].Documentation, 1)
	assert.Equal(t, "Remove build artifacts.", targetMap["clean"].Summary.PlainText())
	assert.True(t, targetMap["clean"].IsPhony)

	// Check test target (undocumented but included because .PHONY)
	require.NotNil(t, targetMap["test"])
	assert.Len(t, targetMap["test"].Documentation, 0, "Undocumented target should have no documentation")
	assert.Equal(t, "", targetMap["test"].Summary.PlainText(), "Undocumented target should have empty summary")
	assert.True(t, targetMap["test"].IsPhony)

	// Check install target (undocumented but included because .PHONY)
	require.NotNil(t, targetMap["install"])
	assert.Len(t, targetMap["install"].Documentation, 0, "Undocumented target should have no documentation")
	assert.Equal(t, "", targetMap["install"].Summary.PlainText(), "Undocumented target should have empty summary")
	assert.True(t, targetMap["install"].IsPhony)
}

func TestBuild_DocumentedTargetNotImplicitAlias(t *testing.T) {
	// Test that a documented target with a single phony dependency is NOT an implicit alias.
	// Documented targets are semantically distinct, not just shortcuts.
	config := &BuilderConfig{
		DefaultCategory: "",
		PhonyTargets: map[string]bool{
			"test-unit": true,
			"test-all":  true,
			"test":      true,
		},
		Dependencies: map[string][]string{
			"test-all": {"test-unit"}, // single dep, but documented
			"test":     {"test-unit"}, // single dep, undocumented (should be alias)
		},
		HasRecipe: map[string]bool{
			"test-unit": true,
			"test-all":  false, // no recipe
			"test":      false, // no recipe
		},
	}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveDoc, Value: "Runs unit tests.", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveDoc, Value: "Runs all tests. Currently only unit tests exist.", SourceFile: "Makefile", LineNumber: 4},
				// test has no documentation - should be implicit alias
			},
			TargetMap: map[string]int{
				"test-unit": 2, // documented
				"test-all":  5, // documented - NOT an alias despite single dep
				"test":      6, // undocumented - IS an alias
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	assert.Len(t, model.Categories, 1)
	// test-unit and test-all should be targets; test should be an alias of test-unit
	assert.Len(t, model.Categories[0].Targets, 2)

	// Find each target
	targetMap := make(map[string]*Target)
	for i := range model.Categories[0].Targets {
		target := &model.Categories[0].Targets[i]
		targetMap[target.Name] = target
	}

	// test-unit should exist and have "test" as an alias
	require.NotNil(t, targetMap["test-unit"], "test-unit should exist as a target")
	assert.Contains(t, targetMap["test-unit"].Aliases, "test", "test should be an alias of test-unit")

	// test-all should exist as a separate target (not an alias) because it's documented
	require.NotNil(t, targetMap["test-all"], "test-all should exist as a separate target (documented)")
	assert.Equal(t, "Runs all tests.", targetMap["test-all"].Summary.PlainText())

	// test should NOT exist as a separate target (it's an alias)
	assert.Nil(t, targetMap["test"], "test should not be a separate target (it's an implicit alias)")
}

func TestBuild_NotAliasDirective(t *testing.T) {
	// Test that !notalias directive prevents a target from being treated as an implicit alias.
	config := &BuilderConfig{
		DefaultCategory: "",
		IncludeAllPhony: true, // Include undocumented phony targets
		PhonyTargets: map[string]bool{
			"test-unit": true,
			"test":      true,
			"t":         true,
		},
		Dependencies: map[string][]string{
			"test": {"test-unit"}, // single dep, marked with !notalias
			"t":    {"test-unit"}, // single dep, no !notalias (should be alias)
		},
		HasRecipe: map[string]bool{
			"test-unit": true,
			"test":      false,
			"t":         false,
		},
	}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveDoc, Value: "Runs unit tests.", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveNotAlias, Value: "", SourceFile: "Makefile", LineNumber: 4},
				// "test" target follows !notalias - should NOT be alias
				// "t" has no !notalias - should be alias
			},
			TargetMap: map[string]int{
				"test-unit": 2,
				"test":      5, // has !notalias above it
				"t":         6, // no !notalias
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	assert.Len(t, model.Categories, 1)
	// test-unit and test should be targets; t should be an alias
	assert.Len(t, model.Categories[0].Targets, 2, "Should have 2 targets: test-unit and test")

	// Find each target
	targetMap := make(map[string]*Target)
	for i := range model.Categories[0].Targets {
		target := &model.Categories[0].Targets[i]
		targetMap[target.Name] = target
	}

	// test-unit should exist and have "t" as an alias (not "test")
	require.NotNil(t, targetMap["test-unit"], "test-unit should exist as a target")
	assert.Contains(t, targetMap["test-unit"].Aliases, "t", "t should be an alias of test-unit")
	assert.NotContains(t, targetMap["test-unit"].Aliases, "test", "test should NOT be an alias due to !notalias")

	// test should exist as a separate target because of !notalias
	require.NotNil(t, targetMap["test"], "test should exist as a separate target due to !notalias")

	// Verify !notalias is tracked in builder
	assert.True(t, builder.NotAliasTargets()["test"], "test should be tracked as !notalias target")
	assert.False(t, builder.NotAliasTargets()["t"], "t should not be tracked as !notalias target")
}

func TestBuild_NotAliasWithDocumentation(t *testing.T) {
	// Test edge case: !notalias combined with documentation.
	// The !notalias is redundant (documented targets are never implicit aliases),
	// but should not cause errors.
	config := &BuilderConfig{
		DefaultCategory: "",
		PhonyTargets: map[string]bool{
			"build": true,
			"b":     true,
		},
		Dependencies: map[string][]string{
			"b": {"build"},
		},
		HasRecipe: map[string]bool{
			"build": true,
			"b":     false,
		},
	}
	builder := NewBuilder(config)

	parsedFiles := []*parser.ParsedFile{
		{
			Path: "Makefile",
			Directives: []parser.Directive{
				{Type: parser.DirectiveDoc, Value: "Build the project.", SourceFile: "Makefile", LineNumber: 1},
				{Type: parser.DirectiveNotAlias, Value: "", SourceFile: "Makefile", LineNumber: 4},
				{Type: parser.DirectiveDoc, Value: "Short alias for build.", SourceFile: "Makefile", LineNumber: 5},
			},
			TargetMap: map[string]int{
				"build": 2,
				"b":     6, // has both !notalias and documentation
			},
		},
	}

	model, err := builder.Build(parsedFiles)

	require.NoError(t, err)
	assert.Len(t, model.Categories, 1)
	assert.Len(t, model.Categories[0].Targets, 2)

	// Find each target
	targetMap := make(map[string]*Target)
	for i := range model.Categories[0].Targets {
		target := &model.Categories[0].Targets[i]
		targetMap[target.Name] = target
	}

	// Both should exist as separate targets
	require.NotNil(t, targetMap["build"])
	require.NotNil(t, targetMap["b"])
	assert.Equal(t, "Short alias for build.", targetMap["b"].Summary.PlainText())

	// b should be tracked as !notalias (even though redundant)
	assert.True(t, builder.NotAliasTargets()["b"])
}
