package format

import (
	"github.com/sdlcforge/make-help/internal/richtext"
	"strings"
	"testing"

	"github.com/sdlcforge/make-help/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRenderer(t *testing.T) {
	renderer := NewRenderer(true)

	assert.NotNil(t, renderer.colors)
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
		FileDocs: []model.FileDoc{
			{
				SourceFile:     "Makefile",
				Documentation:  []string{"This is the main Makefile for the project.", "It provides common development tasks."},
				DiscoveryOrder: 0,
				IsEntryPoint:   true,
			},
		},
	}
	output, err := renderer.Render(helpModel)

	require.NoError(t, err)
	assert.Contains(t, output, "This is the main Makefile for the project.")
	assert.Contains(t, output, "It provides common development tasks.")
}

func TestRender_WithIncludedFiles(t *testing.T) {
	renderer := NewRenderer(false)

	helpModel := &model.HelpModel{
		FileDocs: []model.FileDoc{
			{
				SourceFile:     "Makefile",
				Documentation:  []string{"Main project Makefile.", "Entry point for all tasks."},
				DiscoveryOrder: 0,
				IsEntryPoint:   true,
			},
			{
				SourceFile:     "make/build.mk",
				Documentation:  []string{"Build tasks and compilation.", "Handles Go build process."},
				DiscoveryOrder: 1,
				IsEntryPoint:   false,
			},
			{
				SourceFile:     "make/test.mk",
				Documentation:  []string{"Testing utilities."},
				DiscoveryOrder: 2,
				IsEntryPoint:   false,
			},
		},
	}

	output, err := renderer.Render(helpModel)

	require.NoError(t, err)

	// Entry point docs should appear first (not under "Included Files")
	assert.Contains(t, output, "Main project Makefile.")
	assert.Contains(t, output, "Entry point for all tasks.")

	// Should have "Included Files:" section
	assert.Contains(t, output, "Included Files:")

	// Should contain file paths
	assert.Contains(t, output, "make/build.mk")
	assert.Contains(t, output, "make/test.mk")

	// Should contain indented documentation for included files
	assert.Contains(t, output, "    Build tasks and compilation.")
	assert.Contains(t, output, "    Handles Go build process.")
	assert.Contains(t, output, "    Testing utilities.")

	// Verify structure: entry point docs before "Included Files:"
	entryPointIdx := strings.Index(output, "Main project Makefile.")
	includedFilesIdx := strings.Index(output, "Included Files:")
	assert.Less(t, entryPointIdx, includedFilesIdx, "Entry point docs should appear before 'Included Files:' section")
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
						Summary: richtext.FromPlainText("Build the project."),
					},
					{
						Name:          "test",
						Documentation: []string{"Run all tests."},
						Summary: richtext.FromPlainText("Run all tests."),
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
						Summary: richtext.FromPlainText("Build the project."),
					},
					{
						Name:          "clean",
						Documentation: []string{"Clean build artifacts."},
						Summary: richtext.FromPlainText("Clean build artifacts."),
					},
				},
			},
			{
				Name: "Test",
				Targets: []model.Target{
					{
						Name:          "test",
						Documentation: []string{"Run all tests."},
						Summary: richtext.FromPlainText("Run all tests."),
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
						Summary: richtext.FromPlainText("Build the project."),
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
						Summary: richtext.FromPlainText("Start the development server."),
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
						Summary: richtext.FromPlainText("Start the development server."),
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
						Summary: richtext.FromPlainText("Build the project."),
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
		FileDocs: []model.FileDoc{
			{
				SourceFile:     "Makefile",
				Documentation:  []string{"Project Makefile", "Common development tasks"},
				DiscoveryOrder: 0,
				IsEntryPoint:   true,
			},
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
						Summary: richtext.FromPlainText("Build the project."),
						Variables: []model.Variable{
							{Name: "GOOS", Description: "Target OS"},
							{Name: "GOARCH", Description: "Target architecture"},
						},
					},
					{
						Name:          "clean",
						Documentation: []string{"Remove build artifacts."},
						Summary: richtext.FromPlainText("Remove build artifacts."),
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
						Summary: richtext.FromPlainText("Run all tests."),
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
						Summary: richtext.FromPlainText("Run all tests."), // Summary is pre-computed by builder
					},
				},
			},
		},
	}

	output, err := renderer.Render(helpModel)

	require.NoError(t, err)
	// Summary should be shown (pre-computed during model building)
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
						Summary: richtext.FromPlainText("Build the project with make."), // Markdown stripped by builder
					},
				},
			},
		},
	}

	output, err := renderer.Render(helpModel)

	require.NoError(t, err)
	// Markdown should already be stripped in pre-computed summary
	assert.Contains(t, output, "- build: Build the project with make.")
	assert.NotContains(t, output, "**")
	assert.NotContains(t, output, "`")
	assert.NotContains(t, output, "[docs]")
}

func TestRenderBasicTarget_Complete(t *testing.T) {
	useColor := false
	renderer := NewRenderer(useColor)

	output := renderer.RenderBasicTarget("undocumented", "/path/to/Makefile", 15)

	assert.Contains(t, output, "Target: undocumented")
	assert.Contains(t, output, "No documentation available.")
	assert.Contains(t, output, "Source: /path/to/Makefile:15")
}

func TestRenderBasicTarget_NoSourceInfo(t *testing.T) {
	useColor := false
	renderer := NewRenderer(useColor)

	output := renderer.RenderBasicTarget("undocumented", "", 0)

	assert.Contains(t, output, "Target: undocumented")
	assert.Contains(t, output, "No documentation available.")
	assert.NotContains(t, output, "Source:")
}

func TestRenderBasicTarget_WithColors(t *testing.T) {
	useColor := true
	renderer := NewRenderer(useColor)

	output := renderer.RenderBasicTarget("undocumented", "/path/to/Makefile", 20)

	// Should contain color codes
	assert.Contains(t, output, "\033[1;32m", "Should contain green for target")
	assert.Contains(t, output, "\033[0;37m", "Should contain white for documentation message")
	assert.Contains(t, output, "\033[0m", "Should contain reset code")
	assert.Contains(t, output, "Target: undocumented")
	assert.Contains(t, output, "No documentation available.")
}

func TestEscapeForMakefileEcho_DollarSign(t *testing.T) {
	input := "Use $VAR in your command"
	expected := "Use $$VAR in your command" // $ becomes $$
	result := escapeForMakefileEcho(input)
	assert.Equal(t, expected, result)
}

func TestEscapeForMakefileEcho_DoubleQuote(t *testing.T) {
	input := `Say "hello" to the world`
	expected := `Say \"hello\" to the world` // " becomes \"
	result := escapeForMakefileEcho(input)
	assert.Equal(t, expected, result)
}

func TestEscapeForMakefileEcho_ANSICode(t *testing.T) {
	// Input contains actual ANSI escape character (\x1b)
	input := "\x1b[36mCyan text\x1b[0m"
	expected := "\\033[36mCyan text\\033[0m" // \x1b becomes literal \033
	result := escapeForMakefileEcho(input)
	assert.Equal(t, expected, result)
}

func TestEscapeForMakefileEcho_MixedSpecialChars(t *testing.T) {
	// Input contains $ and " and actual ANSI escape characters
	input := "Use $VAR with \"quotes\" and \x1b[32mcolor\x1b[0m"
	expected := "Use $$VAR with \\\"quotes\\\" and \\033[32mcolor\\033[0m"
	result := escapeForMakefileEcho(input)
	assert.Equal(t, expected, result)
}

func TestEscapeForMakefileEcho_PlainText(t *testing.T) {
	input := "Plain text with no special characters"
	expected := "Plain text with no special characters"
	result := escapeForMakefileEcho(input)
	assert.Equal(t, expected, result)
}

func TestEscapeForMakefileEcho_Backslash(t *testing.T) {
	input := "path\\to\\file"
	expected := "path\\\\to\\\\file"
	result := escapeForMakefileEcho(input)
	assert.Equal(t, expected, result)
}

func TestEscapeForMakefileEcho_Backtick(t *testing.T) {
	input := "Use `make build` to compile"
	expected := "Use \\`make build\\` to compile"
	result := escapeForMakefileEcho(input)
	assert.Equal(t, expected, result)
}

func TestRenderForMakefile_EmptyModel(t *testing.T) {
	renderer := NewRenderer(false)
	helpModel := &model.HelpModel{}

	lines, err := renderer.RenderForMakefile(helpModel)

	require.NoError(t, err)
	require.Len(t, lines, 1)
	assert.Equal(t, "Usage: make [<target>...] [<ENV_VAR>=<value>...]", lines[0])
}

func TestRenderForMakefile_WithFileDocumentation(t *testing.T) {
	renderer := NewRenderer(false)
	helpModel := &model.HelpModel{
		FileDocs: []model.FileDoc{
			{
				SourceFile:     "Makefile",
				Documentation:  []string{"This is the main Makefile for the project.", "It provides common development tasks."},
				DiscoveryOrder: 0,
				IsEntryPoint:   true,
			},
		},
	}

	lines, err := renderer.RenderForMakefile(helpModel)

	require.NoError(t, err)
	assert.Contains(t, lines, "Usage: make [<target>...] [<ENV_VAR>=<value>...]")
	assert.Contains(t, lines, "")
	assert.Contains(t, lines, "This is the main Makefile for the project.")
	assert.Contains(t, lines, "It provides common development tasks.")
}

func TestRenderForMakefile_BasicTargets(t *testing.T) {
	renderer := NewRenderer(false)
	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: "",
				Targets: []model.Target{
					{
						Name:          "build",
						Documentation: []string{"Build the project."},
						Summary: richtext.FromPlainText("Build the project."),
					},
					{
						Name:          "test",
						Documentation: []string{"Run all tests."},
						Summary: richtext.FromPlainText("Run all tests."),
					},
				},
			},
		},
	}

	lines, err := renderer.RenderForMakefile(helpModel)

	require.NoError(t, err)
	assert.Contains(t, lines, "Targets:")
	assert.Contains(t, lines, "  - build: Build the project.")
	assert.Contains(t, lines, "  - test: Run all tests.")
}

func TestRenderForMakefile_WithCategories(t *testing.T) {
	renderer := NewRenderer(false)
	helpModel := &model.HelpModel{
		HasCategories: true,
		Categories: []model.Category{
			{
				Name: "Build",
				Targets: []model.Target{
					{
						Name:          "build",
						Documentation: []string{"Build the project."},
						Summary: richtext.FromPlainText("Build the project."),
					},
				},
			},
			{
				Name: "Test",
				Targets: []model.Target{
					{
						Name:          "test",
						Documentation: []string{"Run all tests."},
						Summary: richtext.FromPlainText("Run all tests."),
					},
				},
			},
		},
	}

	lines, err := renderer.RenderForMakefile(helpModel)

	require.NoError(t, err)
	assert.Contains(t, lines, "Build:")
	assert.Contains(t, lines, "Test:")
	assert.Contains(t, lines, "  - build: Build the project.")
	assert.Contains(t, lines, "  - test: Run all tests.")
}

func TestRenderForMakefile_WithColors(t *testing.T) {
	renderer := NewRenderer(true)
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

	lines, err := renderer.RenderForMakefile(helpModel)

	require.NoError(t, err)

	// Find the category line
	var categoryLine string
	for _, line := range lines {
		if strings.Contains(line, "Build:") {
			categoryLine = line
			break
		}
	}
	assert.NotEmpty(t, categoryLine)
	assert.Contains(t, categoryLine, "\\033[1;36m", "Should contain escaped cyan for category")
	assert.Contains(t, categoryLine, "\\033[0m", "Should contain escaped reset")

	// Find the target line
	var targetLine string
	for _, line := range lines {
		if strings.Contains(line, "build") && strings.Contains(line, "  - ") {
			targetLine = line
			break
		}
	}
	assert.NotEmpty(t, targetLine)
	assert.Contains(t, targetLine, "\\033[1;32m", "Should contain escaped green for target")
	assert.Contains(t, targetLine, "\\033[0;33m", "Should contain escaped yellow for alias")
}

func TestRenderForMakefile_WithColorsDisabled(t *testing.T) {
	renderer := NewRenderer(false)
	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: "Build",
				Targets: []model.Target{
					{
						Name:          "build",
						Documentation: []string{"Build the project."},
					},
				},
			},
		},
	}

	lines, err := renderer.RenderForMakefile(helpModel)

	require.NoError(t, err)

	// Should not contain any ANSI escape codes
	for _, line := range lines {
		assert.NotContains(t, line, "\\033[", "Line should not contain escaped ANSI codes: %s", line)
		assert.NotContains(t, line, "\033[", "Line should not contain raw ANSI codes: %s", line)
	}
}

func TestRenderForMakefile_WithAliasesAndVariables(t *testing.T) {
	renderer := NewRenderer(false)
	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: "",
				Targets: []model.Target{
					{
						Name:          "serve",
						Aliases:       []string{"s", "start"},
						Documentation: []string{"Start the development server."},
						Summary: richtext.FromPlainText("Start the development server."),
						Variables: []model.Variable{
							{Name: "PORT"},
							{Name: "DEBUG"},
						},
					},
				},
			},
		},
	}

	lines, err := renderer.RenderForMakefile(helpModel)

	require.NoError(t, err)
	assert.Contains(t, lines, "  - serve s, start: Start the development server.")
	assert.Contains(t, lines, "    Vars: PORT, DEBUG")
}

func TestRenderForMakefile_SpecialCharactersEscaped(t *testing.T) {
	renderer := NewRenderer(false)
	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: "",
				Targets: []model.Target{
					{
						Name:          "deploy",
						Documentation: []string{`Use $VAR and "quotes" in command.`},
						Summary: richtext.FromPlainText(`Use $VAR and "quotes" in command.`),
					},
				},
			},
		},
	}

	lines, err := renderer.RenderForMakefile(helpModel)

	require.NoError(t, err)

	// Find the target line
	var targetLine string
	for _, line := range lines {
		if strings.Contains(line, "deploy") {
			targetLine = line
			break
		}
	}
	assert.NotEmpty(t, targetLine)
	assert.Contains(t, targetLine, "$$VAR", "Should escape $ as $$")
	assert.Contains(t, targetLine, "\\\"quotes\\\"", "Should escape quotes")
}

func TestRenderDetailedForMakefile_Complete(t *testing.T) {
	renderer := NewRenderer(false)
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

	lines := renderer.RenderDetailedForMakefile(target)

	assert.Contains(t, lines, "Target: build")
	assert.Contains(t, lines, "Aliases: b, compile")
	assert.Contains(t, lines, "Variables:")
	assert.Contains(t, lines, "  - GOOS: Target OS")
	assert.Contains(t, lines, "  - GOARCH: Target architecture")
	assert.Contains(t, lines, "Build the project.")
	assert.Contains(t, lines, "This compiles all source files.")
	assert.Contains(t, lines, "Source: /path/to/Makefile:42")
}

func TestRenderDetailedForMakefile_MinimalTarget(t *testing.T) {
	renderer := NewRenderer(false)
	target := &model.Target{
		Name: "simple",
	}

	lines := renderer.RenderDetailedForMakefile(target)

	assert.Equal(t, []string{"Target: simple"}, lines)
}

func TestRenderDetailedForMakefile_WithColors(t *testing.T) {
	renderer := NewRenderer(true)
	target := &model.Target{
		Name:          "build",
		Aliases:       []string{"b"},
		Documentation: []string{"Build the project."},
		Variables: []model.Variable{
			{Name: "DEBUG", Description: "Enable debug mode"},
		},
	}

	lines := renderer.RenderDetailedForMakefile(target)

	// Check for escaped ANSI codes in output
	targetLine := lines[0]
	assert.Contains(t, targetLine, "\\033[1;32m", "Should contain escaped green for target")
	assert.Contains(t, targetLine, "\\033[0m", "Should contain escaped reset")

	// Find alias line
	var aliasLine string
	for _, line := range lines {
		if strings.Contains(line, "Aliases:") {
			aliasLine = line
			break
		}
	}
	assert.Contains(t, aliasLine, "\\033[0;33m", "Should contain escaped yellow for alias")
}

func TestRenderDetailedForMakefile_VariableWithoutDescription(t *testing.T) {
	renderer := NewRenderer(false)
	target := &model.Target{
		Name: "build",
		Variables: []model.Variable{
			{Name: "PORT"},
			{Name: "DEBUG", Description: "Enable debug mode"},
		},
	}

	lines := renderer.RenderDetailedForMakefile(target)

	assert.Contains(t, lines, "  - PORT")
	assert.Contains(t, lines, "  - DEBUG: Enable debug mode")
}

func TestRenderDetailedForMakefile_SpecialCharactersEscaped(t *testing.T) {
	renderer := NewRenderer(false)
	target := &model.Target{
		Name:          "deploy",
		Documentation: []string{`Set $DEPLOY_ENV to "production" before running.`},
		Variables: []model.Variable{
			{Name: "DEPLOY_ENV", Description: `Use "production" or "staging"`},
		},
	}

	lines := renderer.RenderDetailedForMakefile(target)

	// Check that special characters are escaped
	docLine := ""
	for _, line := range lines {
		if strings.Contains(line, "$$DEPLOY_ENV") {
			docLine = line
			break
		}
	}
	assert.NotEmpty(t, docLine)
	assert.Contains(t, docLine, "$$DEPLOY_ENV", "Should escape $ as $$")
	assert.Contains(t, docLine, "\\\"production\\\"", "Should escape quotes")

	// Check variable description
	varLine := ""
	for _, line := range lines {
		if strings.Contains(line, "DEPLOY_ENV:") {
			varLine = line
			break
		}
	}
	assert.NotEmpty(t, varLine)
	assert.Contains(t, varLine, "\\\"production\\\"", "Should escape quotes in variable description")
}
