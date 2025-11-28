package format

import (
	"fmt"
	"strings"

	"github.com/sdlcforge/make-help/internal/cli"
	"github.com/sdlcforge/make-help/internal/model"
	"github.com/sdlcforge/make-help/internal/summary"
)

// Renderer handles the formatting and rendering of help output.
type Renderer struct {
	config    *cli.Config
	colors    *ColorScheme
	extractor *summary.Extractor
}

// NewRenderer creates a new Renderer with the given configuration.
// The color scheme is determined by the config's UseColor setting.
func NewRenderer(config *cli.Config) *Renderer {
	return &Renderer{
		config:    config,
		colors:    NewColorScheme(config.UseColor),
		extractor: summary.NewExtractor(),
	}
}

// Render generates the complete help output from a HelpModel.
// The output includes:
//   - Usage line
//   - File documentation (if any)
//   - Targets section with categories (if applicable)
//
// Returns the formatted help string and any error encountered.
func (r *Renderer) Render(model *model.HelpModel) (string, error) {
	var buf strings.Builder

	// Usage line
	buf.WriteString("Usage: make [<target>...] [<ENV_VAR>=<value>...]\n")

	// File documentation
	if len(model.FileDocs) > 0 {
		buf.WriteString("\n")
		for _, doc := range model.FileDocs {
			buf.WriteString(doc)
			buf.WriteString("\n")
		}
	}

	// Targets section
	if len(model.Categories) > 0 {
		buf.WriteString("\nTargets:\n")

		for _, category := range model.Categories {
			r.renderCategory(&buf, &category)
		}
	}

	return buf.String(), nil
}

// renderCategory renders a single category with its targets.
// If the category has a name, it's displayed as a colored header.
// Each target is rendered with proper indentation.
func (r *Renderer) renderCategory(buf *strings.Builder, category *model.Category) {
	// Render category name (if present)
	if category.Name != "" {
		buf.WriteString("\n")
		buf.WriteString(r.colors.CategoryName)
		buf.WriteString(category.Name)
		buf.WriteString(":")
		buf.WriteString(r.colors.Reset)
		buf.WriteString("\n")
	}

	// Render each target in the category
	for _, target := range category.Targets {
		r.renderTarget(buf, &target)
	}
}

// renderTarget renders a single target with its name, aliases, summary, and variables.
// Format:
//   - <target>[ <alias1>, ...]: <summary>
//     [Vars: <VAR1>, <VAR2>...]
func (r *Renderer) renderTarget(buf *strings.Builder, target *model.Target) {
	// Indentation for target line
	buf.WriteString("  - ")

	// Target name (colored)
	buf.WriteString(r.colors.TargetName)
	buf.WriteString(target.Name)
	buf.WriteString(r.colors.Reset)

	// Aliases (if any)
	if len(target.Aliases) > 0 {
		buf.WriteString(" ")
		buf.WriteString(r.colors.Alias)
		buf.WriteString(strings.Join(target.Aliases, ", "))
		buf.WriteString(r.colors.Reset)
	}

	// Summary (extract first sentence from documentation)
	summary := r.extractor.Extract(target.Documentation)
	if summary != "" {
		buf.WriteString(": ")
		buf.WriteString(r.colors.Documentation)
		buf.WriteString(summary)
		buf.WriteString(r.colors.Reset)
	}

	buf.WriteString("\n")

	// Variables (if any)
	if len(target.Variables) > 0 {
		buf.WriteString("    Vars: ")
		varNames := make([]string, len(target.Variables))
		for i, v := range target.Variables {
			varNames[i] = v.Name
		}
		buf.WriteString(r.colors.Variable)
		buf.WriteString(strings.Join(varNames, ", "))
		buf.WriteString(r.colors.Reset)
		buf.WriteString("\n")
	}
}

// RenderDetailedTarget renders a detailed view of a single target.
// This is used for the help-<target> functionality.
// It includes the full documentation, not just the summary.
func (r *Renderer) RenderDetailedTarget(target *model.Target) string {
	var buf strings.Builder

	// Target name
	buf.WriteString(r.colors.TargetName)
	buf.WriteString("Target: ")
	buf.WriteString(target.Name)
	buf.WriteString(r.colors.Reset)
	buf.WriteString("\n")

	// Aliases
	if len(target.Aliases) > 0 {
		buf.WriteString(r.colors.Alias)
		buf.WriteString("Aliases: ")
		buf.WriteString(strings.Join(target.Aliases, ", "))
		buf.WriteString(r.colors.Reset)
		buf.WriteString("\n")
	}

	// Variables
	if len(target.Variables) > 0 {
		buf.WriteString(r.colors.Variable)
		buf.WriteString("Variables:\n")
		buf.WriteString(r.colors.Reset)
		for _, v := range target.Variables {
			buf.WriteString("  - ")
			buf.WriteString(r.colors.Variable)
			buf.WriteString(v.Name)
			buf.WriteString(r.colors.Reset)
			if v.Description != "" {
				buf.WriteString(": ")
				buf.WriteString(r.colors.Documentation)
				buf.WriteString(v.Description)
				buf.WriteString(r.colors.Reset)
			}
			buf.WriteString("\n")
		}
	}

	// Full documentation
	if len(target.Documentation) > 0 {
		buf.WriteString("\n")
		buf.WriteString(r.colors.Documentation)
		buf.WriteString("Documentation:\n")
		for _, line := range target.Documentation {
			buf.WriteString("  ")
			buf.WriteString(line)
			buf.WriteString("\n")
		}
		buf.WriteString(r.colors.Reset)
	}

	// Source information
	if target.SourceFile != "" {
		buf.WriteString(fmt.Sprintf("\nSource: %s:%d\n", target.SourceFile, target.LineNumber))
	}

	return buf.String()
}
