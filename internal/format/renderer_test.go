package format

import (
	"strings"
	"testing"

	"github.com/sdlcforge/make-help/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRenderer(t *testing.T) {
	renderer := NewRenderer(true)

	assert.NotNil(t, renderer.colors)
	assert.NotNil(t, renderer.extractor)
}

func TestRender_EmptyModel(t *testing.T) {
	renderer := NewRenderer(false)

	helpModel := &model.HelpModel{}
	output, err := renderer.Render(helpModel)

	require.NoError(t, err)
	assert.Contains(t, output, "Usage: make [<target>...] [<ENV_VAR>=<value>...]")
	assert.NotContains(t, output, "Targets:")
}

func TestRender_WithFileDocumentation(t *testing.T) {
	renderer := NewRenderer(false)

	helpModel := &model.HelpModel{
		FileDocs: []string{
			"This is the main Makefile for the project.",
			"It provides common development tasks.",
		},
	}
	output, err := renderer.Render(helpModel)

	require.NoError(t, err)
	assert.Contains(t, output, "This is the main Makefile for the project.")
	assert.Contains(t, output, "It provides common development tasks.")
}

func TestRender_BasicTargetsNoCategories(t *testing.T) {
	useColor := false
	renderer := NewRenderer(useColor)

	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: "", // Uncategorized
				Targets: []model.Target{
					{
						Name:          "build",
						Documentation: []string{"Build the project. This compiles all source files."},
					},
					{
						Name:          "test",
						Documentation: []string{"Run all tests."},
					},
				},
			},
		},
	}

	output, err := renderer.Render(helpModel)

	require.NoError(t, err)
	assert.Contains(t, output, "Targets:")
	assert.Contains(t, output, "- build: Build the project.")
	assert.Contains(t, output, "- test: Run all tests.")
	// Should not have category name since it's empty
	assert.NotContains(t, output, ":\n\n")
}

func TestRender_WithCategories(t *testing.T) {
	useColor := false
	renderer := NewRenderer(useColor)

	helpModel := &model.HelpModel{
		HasCategories: true,
		Categories: []model.Category{
			{
				Name: "Build",
				Targets: []model.Target{
					{
						Name:          "build",
						Documentation: []string{"Build the project."},
					},
					{
						Name:          "clean",
						Documentation: []string{"Clean build artifacts."},
					},
				},
			},
			{
				Name: "Test",
				Targets: []model.Target{
					{
						Name:          "test",
						Documentation: []string{"Run all tests."},
					},
				},
			},
		},
	}

	output, err := renderer.Render(helpModel)

	require.NoError(t, err)
	assert.Contains(t, output, "Build:")
	assert.Contains(t, output, "Test:")
	assert.Contains(t, output, "- build: Build the project.")
	assert.Contains(t, output, "- clean: Clean build artifacts.")
	assert.Contains(t, output, "- test: Run all tests.")
}

func TestRender_TargetWithAliases(t *testing.T) {
	useColor := false
	renderer := NewRenderer(useColor)

	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: "",
				Targets: []model.Target{
					{
						Name:          "build",
						Aliases:       []string{"b", "compile"},
						Documentation: []string{"Build the project."},
					},
				},
			},
		},
	}

	output, err := renderer.Render(helpModel)

	require.NoError(t, err)
	assert.Contains(t, output, "- build b, compile: Build the project.")
}

func TestRender_TargetWithVariables(t *testing.T) {
	useColor := false
	renderer := NewRenderer(useColor)

	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: "",
				Targets: []model.Target{
					{
						Name:          "serve",
						Documentation: []string{"Start the development server."},
						Variables: []model.Variable{
							{Name: "PORT", Description: "Server port"},
							{Name: "DEBUG", Description: "Enable debug mode"},
						},
					},
				},
			},
		},
	}

	output, err := renderer.Render(helpModel)

	require.NoError(t, err)
	assert.Contains(t, output, "- serve: Start the development server.")
	assert.Contains(t, output, "Vars: PORT, DEBUG")
}

func TestRender_TargetWithAliasesAndVariables(t *testing.T) {
	useColor := false
	renderer := NewRenderer(useColor)

	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: "",
				Targets: []model.Target{
					{
						Name:          "serve",
						Aliases:       []string{"s", "start"},
						Documentation: []string{"Start the development server."},
						Variables: []model.Variable{
							{Name: "PORT"},
						},
					},
				},
			},
		},
	}

	output, err := renderer.Render(helpModel)

	require.NoError(t, err)
	assert.Contains(t, output, "- serve s, start: Start the development server.")
	assert.Contains(t, output, "Vars: PORT")
}

func TestRender_TargetWithNoDocumentation(t *testing.T) {
	useColor := false
	renderer := NewRenderer(useColor)

	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: "",
				Targets: []model.Target{
					{
						Name:          "undocumented",
						Documentation: []string{},
					},
				},
			},
		},
	}

	output, err := renderer.Render(helpModel)

	require.NoError(t, err)
	assert.Contains(t, output, "- undocumented\n")
	// Should not have a colon after the name since no summary
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "undocumented") {
			assert.NotContains(t, line, "undocumented:")
			break
		}
	}
}

func TestRender_WithColorsEnabled(t *testing.T) {
	useColor := true
	renderer := NewRenderer(useColor)

	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: "Build",
				Targets: []model.Target{
					{
						Name:          "build",
						Aliases:       []string{"b"},
						Documentation: []string{"Build the project."},
						Variables: []model.Variable{
							{Name: "DEBUG"},
						},
					},
				},
			},
		},
	}

	output, err := renderer.Render(helpModel)

	require.NoError(t, err)
	// Should contain ANSI color codes
	assert.Contains(t, output, "\033[1;36m", "Should contain cyan for category")
	assert.Contains(t, output, "\033[1;32m", "Should contain green for target")
	assert.Contains(t, output, "\033[0;33m", "Should contain yellow for alias")
	assert.Contains(t, output, "\033[0;35m", "Should contain magenta for variable")
	assert.Contains(t, output, "\033[0;37m", "Should contain white for documentation")
	assert.Contains(t, output, "\033[0m", "Should contain reset code")
}

func TestRender_WithColorsDisabled(t *testing.T) {
	useColor := false
	renderer := NewRenderer(useColor)

	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: "Build",
				Targets: []model.Target{
					{
						Name:          "build",
						Aliases:       []string{"b"},
						Documentation: []string{"Build the project."},
					},
				},
			},
		},
	}

	output, err := renderer.Render(helpModel)

	require.NoError(t, err)
	// Should NOT contain any ANSI color codes
	assert.NotContains(t, output, "\033[")
}

func TestRender_ComplexHelpModel(t *testing.T) {
	useColor := false
	renderer := NewRenderer(useColor)

	helpModel := &model.HelpModel{
		FileDocs: []string{
			"Project Makefile",
			"Common development tasks",
		},
		HasCategories: true,
		Categories: []model.Category{
			{
				Name: "Build",
				Targets: []model.Target{
					{
						Name:          "build",
						Aliases:       []string{"b"},
						Documentation: []string{"Build the project. Compiles all source files."},
						Variables: []model.Variable{
							{Name: "GOOS", Description: "Target OS"},
							{Name: "GOARCH", Description: "Target architecture"},
						},
					},
					{
						Name:          "clean",
						Documentation: []string{"Remove build artifacts."},
					},
				},
			},
			{
				Name: "Test",
				Targets: []model.Target{
					{
						Name:          "test",
						Aliases:       []string{"t"},
						Documentation: []string{"Run all tests. Uses go test with verbose output."},
					},
				},
			},
		},
	}

	output, err := renderer.Render(helpModel)

	require.NoError(t, err)

	// File docs
	assert.Contains(t, output, "Project Makefile")
	assert.Contains(t, output, "Common development tasks")

	// Categories
	assert.Contains(t, output, "Build:")
	assert.Contains(t, output, "Test:")

	// Targets with summaries
	assert.Contains(t, output, "- build b: Build the project.")
	assert.Contains(t, output, "- clean: Remove build artifacts.")
	assert.Contains(t, output, "- test t: Run all tests.")

	// Variables
	assert.Contains(t, output, "Vars: GOOS, GOARCH")
}

func TestRenderDetailedTarget_Complete(t *testing.T) {
	useColor := false
	renderer := NewRenderer(useColor)

	target := &model.Target{
		Name:    "build",
		Aliases: []string{"b", "compile"},
		Documentation: []string{
			"Build the project.",
			"",
			"This compiles all source files and generates the binary.",
			"The output is placed in the bin/ directory.",
		},
		Variables: []model.Variable{
			{Name: "GOOS", Description: "Target operating system"},
			{Name: "GOARCH", Description: "Target architecture"},
			{Name: "DEBUG", Description: "Enable debug symbols"},
		},
		SourceFile: "/path/to/Makefile",
		LineNumber: 42,
	}

	output := renderer.RenderDetailedTarget(target)

	assert.Contains(t, output, "Target: build")
	assert.Contains(t, output, "Aliases: b, compile")
	assert.Contains(t, output, "Variables:")
	assert.Contains(t, output, "- GOOS: Target operating system")
	assert.Contains(t, output, "- GOARCH: Target architecture")
	assert.Contains(t, output, "- DEBUG: Enable debug symbols")
	assert.Contains(t, output, "Documentation:")
	assert.Contains(t, output, "Build the project.")
	assert.Contains(t, output, "This compiles all source files and generates the binary.")
	assert.Contains(t, output, "The output is placed in the bin/ directory.")
	assert.Contains(t, output, "Source: /path/to/Makefile:42")
}

func TestRenderDetailedTarget_MinimalTarget(t *testing.T) {
	useColor := false
	renderer := NewRenderer(useColor)

	target := &model.Target{
		Name: "simple",
	}

	output := renderer.RenderDetailedTarget(target)

	assert.Contains(t, output, "Target: simple")
	assert.NotContains(t, output, "Aliases:")
	assert.NotContains(t, output, "Variables:")
	assert.NotContains(t, output, "Documentation:")
	assert.NotContains(t, output, "Source:")
}

func TestRenderDetailedTarget_WithColors(t *testing.T) {
	useColor := true
	renderer := NewRenderer(useColor)

	target := &model.Target{
		Name:          "build",
		Aliases:       []string{"b"},
		Documentation: []string{"Build the project."},
		Variables: []model.Variable{
			{Name: "DEBUG", Description: "Enable debug mode"},
		},
		SourceFile: "/path/to/Makefile",
		LineNumber: 10,
	}

	output := renderer.RenderDetailedTarget(target)

	// Should contain color codes
	assert.Contains(t, output, "\033[1;32m", "Should contain green for target")
	assert.Contains(t, output, "\033[0;33m", "Should contain yellow for alias")
	assert.Contains(t, output, "\033[0;35m", "Should contain magenta for variable")
	assert.Contains(t, output, "\033[0;37m", "Should contain white for documentation")
	assert.Contains(t, output, "\033[0m", "Should contain reset code")
}

func TestRenderDetailedTarget_VariableWithoutDescription(t *testing.T) {
	useColor := false
	renderer := NewRenderer(useColor)

	target := &model.Target{
		Name: "build",
		Variables: []model.Variable{
			{Name: "PORT"},
			{Name: "DEBUG", Description: "Enable debug mode"},
		},
	}

	output := renderer.RenderDetailedTarget(target)

	assert.Contains(t, output, "- PORT\n")
	assert.Contains(t, output, "- DEBUG: Enable debug mode")
}

func TestRender_SummaryExtraction(t *testing.T) {
	useColor := false
	renderer := NewRenderer(useColor)

	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: "",
				Targets: []model.Target{
					{
						Name: "test",
						Documentation: []string{
							"Run all tests. This includes unit tests, integration tests, and end-to-end tests.",
						},
					},
				},
			},
		},
	}

	output, err := renderer.Render(helpModel)

	require.NoError(t, err)
	// Should extract only the first sentence
	assert.Contains(t, output, "- test: Run all tests.")
	assert.NotContains(t, output, "This includes unit tests")
}

func TestRender_MarkdownInSummary(t *testing.T) {
	useColor := false
	renderer := NewRenderer(useColor)

	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: "",
				Targets: []model.Target{
					{
						Name: "build",
						Documentation: []string{
							"Build the **project** with `make`. See [docs](url) for details.",
						},
					},
				},
			},
		},
	}

	output, err := renderer.Render(helpModel)

	require.NoError(t, err)
	// Markdown should be stripped by summary extractor
	assert.Contains(t, output, "- build: Build the project with make.")
	assert.NotContains(t, output, "**")
	assert.NotContains(t, output, "`")
	assert.NotContains(t, output, "[docs]")
}
