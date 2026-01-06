package format

import (
	"fmt"
	"io"
	"strings"

	"github.com/sdlcforge/make-help/internal/model"
)

// TextFormatter generates plain text output suitable for terminal display or text files.
// The output uses ANSI color codes when color is enabled.
type TextFormatter struct {
	config *FormatterConfig
	colors *ColorScheme
}

// NewTextFormatter creates a new TextFormatter with the given configuration.
func NewTextFormatter(config *FormatterConfig) *TextFormatter {
	config = normalizeConfig(config)

	return &TextFormatter{
		config: config,
		colors: initColorScheme(config),
	}
}

// RenderHelp generates the complete help output from a HelpModel.
// The output includes:
//   - Usage line
//   - Entry point file documentation (if any)
//   - Included files section (if any non-entry files have docs)
//   - Targets section with categories (if applicable)
func (f *TextFormatter) RenderHelp(helpModel *model.HelpModel, w io.Writer) error {
	if helpModel == nil {
		return errNilHelpModel("text")
	}

	var buf strings.Builder

	// Usage line
	buf.WriteString("Usage: make [<target>...] [<ENV_VAR>=<value>...]\n")

	// File documentation
	if len(helpModel.FileDocs) > 0 {
		// Render entry point file docs first
		entryPointDocs := extractEntryPointDocs(helpModel.FileDocs)
		if entryPointDocs != nil {
			buf.WriteString("\n")
			for _, line := range entryPointDocs {
				buf.WriteString(line)
				buf.WriteString("\n")
			}
		}

		// Render included files section
		includedFiles := extractIncludedFiles(helpModel.FileDocs)
		if len(includedFiles) > 0 {
			buf.WriteString("\nIncluded files:\n")
			for _, fileDoc := range includedFiles {
				// File path
				buf.WriteString("  ")
				buf.WriteString(fileDoc.SourceFile)
				buf.WriteString("\n")

				// Documentation (indented)
				for _, line := range fileDoc.Documentation {
					if line == "" {
						buf.WriteString("\n")
					} else {
						buf.WriteString("    ")
						buf.WriteString(line)
						buf.WriteString("\n")
					}
				}
				buf.WriteString("\n") // Blank line after each file
			}
		}
	}

	// Targets section
	if len(helpModel.Categories) > 0 {
		buf.WriteString("\nTargets:\n")

		for _, category := range helpModel.Categories {
			f.renderCategory(&buf, &category)
		}
	}

	_, err := w.Write([]byte(buf.String()))
	return err
}

// renderCategory renders a single category with its targets.
// If the category has a name, it's displayed as a colored header.
// Each target is rendered with proper indentation.
func (f *TextFormatter) renderCategory(buf *strings.Builder, category *model.Category) {
	// Render category name (if present)
	if category.Name != model.UncategorizedCategoryName {
		buf.WriteString("\n")
		buf.WriteString(f.colors.CategoryName)
		buf.WriteString(category.Name)
		buf.WriteString(":")
		buf.WriteString(f.colors.Reset)
		buf.WriteString("\n")
	}

	// Render each target in the category
	for _, target := range category.Targets {
		f.renderTarget(buf, &target)
	}
}

// renderTarget renders a single target with its name, aliases, summary, and variables.
// Format:
//   - <target>[ <alias1>, ...]: <summary>
//     [Vars: <VAR1>, <VAR2>...]
func (f *TextFormatter) renderTarget(buf *strings.Builder, target *model.Target) {
	// Indentation for target line
	buf.WriteString("  - ")

	// Target name (colored)
	buf.WriteString(f.colors.TargetName)
	buf.WriteString(target.Name)
	buf.WriteString(f.colors.Reset)

	// Aliases (if any)
	if len(target.Aliases) > 0 {
		buf.WriteString(" ")
		buf.WriteString(f.colors.Alias)
		buf.WriteString(strings.Join(target.Aliases, ", "))
		buf.WriteString(f.colors.Reset)
	}

	// Summary: Use plain text for terminal output (strips markdown formatting)
	if len(target.Summary) > 0 && target.Summary[0] != "" {
		buf.WriteString(": ")
		buf.WriteString(f.colors.Documentation)
		buf.WriteString(target.Summary[0])
		buf.WriteString(f.colors.Reset)
	}

	buf.WriteString("\n")

	// Variables (if any)
	if len(target.Variables) > 0 {
		buf.WriteString("    Vars: ")
		varNames := make([]string, len(target.Variables))
		for i, v := range target.Variables {
			varNames[i] = v.Name
		}
		buf.WriteString(f.colors.Variable)
		buf.WriteString(strings.Join(varNames, ", "))
		buf.WriteString(f.colors.Reset)
		buf.WriteString("\n")
	}
}

// RenderDetailedTarget renders a detailed view of a single target.
// This is used for the help-<target> functionality.
// It includes the full documentation, not just the summary.
func (f *TextFormatter) RenderDetailedTarget(target *model.Target, w io.Writer) error {
	if target == nil {
		return errNilTarget("text")
	}

	var buf strings.Builder

	// Target name
	buf.WriteString(f.colors.TargetName)
	buf.WriteString("Target: ")
	buf.WriteString(target.Name)
	buf.WriteString(f.colors.Reset)
	buf.WriteString("\n")

	// Aliases
	if len(target.Aliases) > 0 {
		buf.WriteString(f.colors.Alias)
		buf.WriteString("Aliases: ")
		buf.WriteString(strings.Join(target.Aliases, ", "))
		buf.WriteString(f.colors.Reset)
		buf.WriteString("\n")
	}

	// Variables
	if len(target.Variables) > 0 {
		buf.WriteString(f.colors.Variable)
		buf.WriteString("Variables:\n")
		buf.WriteString(f.colors.Reset)
		for _, v := range target.Variables {
			buf.WriteString("  - ")
			buf.WriteString(f.colors.Variable)
			buf.WriteString(v.Name)
			buf.WriteString(f.colors.Reset)
			if v.Description != "" {
				buf.WriteString(": ")
				buf.WriteString(f.colors.Documentation)
				buf.WriteString(v.Description)
				buf.WriteString(f.colors.Reset)
			}
			buf.WriteString("\n")
		}
	}

	// Full documentation (blank line only after Variables section)
	if len(target.Documentation) > 0 {
		if len(target.Variables) > 0 {
			buf.WriteString("\n")
		}
		for _, line := range target.Documentation {
			buf.WriteString(f.colors.Documentation)
			buf.WriteString(line)
			buf.WriteString(f.colors.Reset)
			buf.WriteString("\n")
		}
	}

	// Source information
	if target.SourceFile != "" {
		buf.WriteString(fmt.Sprintf("\nSource: %s:%d\n", target.SourceFile, target.LineNumber))
	}

	_, err := w.Write([]byte(buf.String()))
	return err
}

// RenderBasicTarget renders minimal info for a target without documentation.
// This is used when a target exists but has no associated documentation.
// Shows target name and source location if available.
func (f *TextFormatter) RenderBasicTarget(name string, sourceFile string, lineNumber int, w io.Writer) error {
	var buf strings.Builder

	// Target name
	buf.WriteString(f.colors.TargetName)
	buf.WriteString("Target: ")
	buf.WriteString(name)
	buf.WriteString(f.colors.Reset)
	buf.WriteString("\n")

	// No documentation message
	buf.WriteString("\n")
	buf.WriteString(f.colors.Documentation)
	buf.WriteString("No documentation available.\n")
	buf.WriteString(f.colors.Reset)

	// Source information (if available)
	if sourceFile != "" {
		buf.WriteString(fmt.Sprintf("\nSource: %s:%d\n", sourceFile, lineNumber))
	}

	_, err := w.Write([]byte(buf.String()))
	return err
}

// ContentType returns the MIME type for text format.
func (f *TextFormatter) ContentType() string {
	return "text/plain"
}

// DefaultExtension returns the default file extension for text format.
func (f *TextFormatter) DefaultExtension() string {
	return ".txt"
}
