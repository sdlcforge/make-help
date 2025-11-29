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
	assert.Len(t, model.FileDocs, 2)
	assert.Contains(t, model.FileDocs, "Main project Makefile")
	assert.Contains(t, model.FileDocs, "Build tools and utilities")
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
		if model.Categories[i].Name == "" {
			uncategorized = &model.Categories[i]
			break
		}
	}
	require.NotNil(t, uncategorized)
	assert.Len(t, uncategorized.Targets, 1)
	assert.Equal(t, "build", uncategorized.Targets[0].Name)
	assert.Len(t, uncategorized.Targets[0].Documentation, 2)
	assert.Equal(t, "Build the project.", uncategorized.Targets[0].Summary)
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
	assert.Equal(t, "Build the project.", model.Categories[0].Targets[0].Summary)
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
	assert.Contains(t, model.FileDocs, "Main Makefile")
	assert.Contains(t, model.FileDocs, "Include file")
}

func TestParseVarDirective(t *testing.T) {
	builder := NewBuilder(&BuilderConfig{DefaultCategory: ""})

	tests := []struct {
		name        string
		input       string
		wantName    string
		wantDesc    string
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
	// Empty file doc values should be skipped
	assert.Len(t, model.FileDocs, 1)
	assert.Equal(t, "Has content", model.FileDocs[0])
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
		if target.Name == "build" {
			assert.False(t, target.IsPhony, "build should not be phony")
		} else if target.Name == "clean" {
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
