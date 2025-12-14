package format_test

import (
	"fmt"

	"github.com/sdlcforge/make-help/internal/format"
	"github.com/sdlcforge/make-help/internal/model"
)

// ExampleRenderer demonstrates the basic usage of the Renderer
// to format help output from a HelpModel.
func ExampleRenderer() {
	// Create a config without colors for reproducible output
	useColor := false
	renderer := format.NewRenderer(useColor)

	// Create a sample HelpModel
	helpModel := &model.HelpModel{
		FileDocs: []string{"Example project Makefile"},
		Categories: []model.Category{
			{
				Name: "Build",
				Targets: []model.Target{
					{
						Name:          "build",
						Aliases:       []string{"b"},
						Documentation: []string{"Build the project."},
						Variables: []model.Variable{
							{Name: "GOOS", Description: "Target OS"},
						},
					},
				},
			},
		},
	}

	// Render the help output
	output, _ := renderer.Render(helpModel)
	fmt.Print(output)

	// Output:
	// Usage: make [<target>...] [<ENV_VAR>=<value>...]
	//
	// Example project Makefile
	//
	// Targets:
	//
	// Build:
	//   - build b: Build the project.
	//     Vars: GOOS
}

// ExampleRenderer_RenderDetailedTarget demonstrates rendering
// a detailed view of a single target.
func ExampleRenderer_RenderDetailedTarget() {
	useColor := false
	renderer := format.NewRenderer(useColor)

	target := &model.Target{
		Name:    "test",
		Aliases: []string{"t"},
		Documentation: []string{
			"Run all tests with verbose output.",
		},
		Variables: []model.Variable{
			{Name: "VERBOSE", Description: "Enable verbose output"},
		},
		SourceFile: "Makefile",
		LineNumber: 15,
	}

	output := renderer.RenderDetailedTarget(target)
	fmt.Print(output)

	// Output:
	// Target: test
	// Aliases: t
	// Variables:
	//   - VERBOSE: Enable verbose output
	//
	// Run all tests with verbose output.
	//
	// Source: Makefile:15
}
