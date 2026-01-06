package format

import (
	"fmt"
	"io"
	"strings"

	"github.com/sdlcforge/make-help/internal/model"
)

// MakeFormatter generates Makefile content with help targets using @printf statements.
// The output is designed to be included in a Makefile and provides embedded help text.
type MakeFormatter struct {
	config *FormatterConfig
	colors *ColorScheme
}

// Compile-time check to ensure MakeFormatter implements LineRenderer interface.
var _ LineRenderer = (*MakeFormatter)(nil)

// NewMakeFormatter creates a new MakeFormatter with the given configuration.
func NewMakeFormatter(config *FormatterConfig) *MakeFormatter {
	config = normalizeConfig(config)

	// Use provided color scheme if available, otherwise create one
	colors := config.ColorScheme
	if colors == nil {
		colors = NewColorScheme(config.UseColor)
	}

	return &MakeFormatter{
		config: config,
		colors: colors,
	}
}

// RenderHelp generates the complete help output from a HelpModel.
// For Make format, this generates @printf statements that echo the help text.
func (f *MakeFormatter) RenderHelp(helpModel *model.HelpModel, w io.Writer) error {
	if helpModel == nil {
		return fmt.Errorf("make formatter: help model cannot be nil")
	}

	lines, err := f.RenderHelpLines(helpModel)
	if err != nil {
		return err
	}

	for _, line := range lines {
		if _, err := fmt.Fprintf(w, "\t@printf '%%b\\n' \"%s\"\n", line); err != nil {
			return err
		}
	}

	return nil
}

// RenderDetailedTarget generates detailed help for a single target.
func (f *MakeFormatter) RenderDetailedTarget(target *model.Target, w io.Writer) error {
	if target == nil {
		return fmt.Errorf("make formatter: target cannot be nil")
	}

	lines := f.RenderDetailedTargetLines(target)

	for _, line := range lines {
		if _, err := fmt.Fprintf(w, "\t@printf '%%b\\n' \"%s\"\n", line); err != nil {
			return err
		}
	}

	return nil
}

// RenderBasicTarget generates minimal help for an undocumented target.
func (f *MakeFormatter) RenderBasicTarget(name string, sourceFile string, lineNumber int, w io.Writer) error {
	lines := f.renderBasicTargetLines(name, sourceFile, lineNumber)

	for _, line := range lines {
		if _, err := fmt.Fprintf(w, "\t@printf '%%b\\n' \"%s\"\n", line); err != nil {
			return err
		}
	}

	return nil
}

// ContentType returns the MIME type for Makefile format.
func (f *MakeFormatter) ContentType() string {
	return "text/x-makefile"
}

// DefaultExtension returns the default file extension for Makefile format.
func (f *MakeFormatter) DefaultExtension() string {
	return ".mk"
}

// RenderHelpLines generates help output lines suitable for Makefile @printf statements.
// Returns a slice of strings, each representing one line to be printed.
// Each line is properly escaped for shell/Makefile context.
// This method implements the LineRenderer interface, allowing the generator package
// to embed help text without depending on the concrete MakeFormatter type.
func (f *MakeFormatter) RenderHelpLines(helpModel *model.HelpModel) ([]string, error) {
	var lines []string

	// Usage line
	lines = append(lines, escapeForMakefileEcho("Usage: make [<target>...] [<ENV_VAR>=<value>...]"))

	// File documentation
	if len(helpModel.FileDocs) > 0 {
		// Render entry point file docs first
		for _, fileDoc := range helpModel.FileDocs {
			if fileDoc.IsEntryPoint && len(fileDoc.Documentation) > 0 {
				lines = append(lines, escapeForMakefileEcho(""))
				for _, line := range fileDoc.Documentation {
					lines = append(lines, escapeForMakefileEcho(line))
				}
				break
			}
		}

		// Render included files section
		var includedFiles []model.FileDoc
		for _, fileDoc := range helpModel.FileDocs {
			if !fileDoc.IsEntryPoint && len(fileDoc.Documentation) > 0 {
				includedFiles = append(includedFiles, fileDoc)
			}
		}

		if len(includedFiles) > 0 {
			lines = append(lines, escapeForMakefileEcho(""))
			lines = append(lines, escapeForMakefileEcho("Included files:"))
			for _, fileDoc := range includedFiles {
				// File path
				lines = append(lines, escapeForMakefileEcho("  "+fileDoc.SourceFile))

				// Documentation (indented)
				for _, line := range fileDoc.Documentation {
					if line == "" {
						lines = append(lines, escapeForMakefileEcho(""))
					} else {
						lines = append(lines, escapeForMakefileEcho("    "+line))
					}
				}
				lines = append(lines, escapeForMakefileEcho("")) // Blank line after each file
			}
		}
	}

	// Targets section
	if len(helpModel.Categories) > 0 {
		lines = append(lines, escapeForMakefileEcho(""))
		lines = append(lines, escapeForMakefileEcho("Targets:"))

		for _, category := range helpModel.Categories {
			categoryLines := f.renderCategoryLines(&category)
			lines = append(lines, categoryLines...)
		}
	}

	return lines, nil
}

// renderCategoryLines renders a single category for Makefile output.
func (f *MakeFormatter) renderCategoryLines(category *model.Category) []string {
	var lines []string

	// Category name (if present)
	if category.Name != model.UncategorizedCategoryName {
		lines = append(lines, escapeForMakefileEcho(""))
		categoryLine := f.colors.CategoryName + category.Name + ":" + f.colors.Reset
		lines = append(lines, escapeForMakefileEcho(categoryLine))
	}

	// Each target in the category
	for _, target := range category.Targets {
		targetLines := f.renderTargetLines(&target)
		lines = append(lines, targetLines...)
	}

	return lines
}

// renderTargetLines renders a single target for Makefile output.
func (f *MakeFormatter) renderTargetLines(target *model.Target) []string {
	var lines []string
	var buf strings.Builder

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

	// Summary: Use plain text for Makefile embedding (strips markdown formatting)
	summaryText := target.Summary.PlainText()
	if summaryText != "" {
		buf.WriteString(": ")
		buf.WriteString(f.colors.Documentation)
		buf.WriteString(summaryText)
		buf.WriteString(f.colors.Reset)
	}

	lines = append(lines, escapeForMakefileEcho(buf.String()))

	// Variables (if any)
	if len(target.Variables) > 0 {
		buf.Reset()
		buf.WriteString("    Vars: ")
		varNames := make([]string, len(target.Variables))
		for i, v := range target.Variables {
			varNames[i] = v.Name
		}
		buf.WriteString(f.colors.Variable)
		buf.WriteString(strings.Join(varNames, ", "))
		buf.WriteString(f.colors.Reset)
		lines = append(lines, escapeForMakefileEcho(buf.String()))
	}

	return lines
}

// RenderDetailedTargetLines renders detailed help for a single target suitable for Makefile @printf.
// This method implements the LineRenderer interface, allowing the generator package
// to embed help text without depending on the concrete MakeFormatter type.
func (f *MakeFormatter) RenderDetailedTargetLines(target *model.Target) []string {
	var lines []string

	// Target name
	targetLine := f.colors.TargetName + "Target: " + target.Name + f.colors.Reset
	lines = append(lines, escapeForMakefileEcho(targetLine))

	// Aliases
	if len(target.Aliases) > 0 {
		aliasLine := f.colors.Alias + "Aliases: " + strings.Join(target.Aliases, ", ") + f.colors.Reset
		lines = append(lines, escapeForMakefileEcho(aliasLine))
	}

	// Variables
	if len(target.Variables) > 0 {
		varHeader := f.colors.Variable + "Variables:" + f.colors.Reset
		lines = append(lines, escapeForMakefileEcho(varHeader))
		for _, v := range target.Variables {
			var varBuf strings.Builder
			varBuf.WriteString("  - ")
			varBuf.WriteString(f.colors.Variable)
			varBuf.WriteString(v.Name)
			varBuf.WriteString(f.colors.Reset)
			if v.Description != "" {
				varBuf.WriteString(": ")
				varBuf.WriteString(f.colors.Documentation)
				varBuf.WriteString(v.Description)
				varBuf.WriteString(f.colors.Reset)
			}
			lines = append(lines, escapeForMakefileEcho(varBuf.String()))
		}
	}

	// Full documentation (blank line only after Variables section)
	if len(target.Documentation) > 0 {
		if len(target.Variables) > 0 {
			lines = append(lines, escapeForMakefileEcho(""))
		}
		for _, line := range target.Documentation {
			docLine := f.colors.Documentation + line + f.colors.Reset
			lines = append(lines, escapeForMakefileEcho(docLine))
		}
	}

	// Source information
	if target.SourceFile != "" {
		lines = append(lines, escapeForMakefileEcho(""))
		sourceLine := fmt.Sprintf("Source: %s:%d", target.SourceFile, target.LineNumber)
		lines = append(lines, escapeForMakefileEcho(sourceLine))
	}

	return lines
}

// renderBasicTargetLines renders minimal info for a target without documentation.
func (f *MakeFormatter) renderBasicTargetLines(name string, sourceFile string, lineNumber int) []string {
	var lines []string

	// Target name
	targetLine := f.colors.TargetName + "Target: " + name + f.colors.Reset
	lines = append(lines, escapeForMakefileEcho(targetLine))

	// No documentation message
	lines = append(lines, escapeForMakefileEcho(""))
	noDocsLine := f.colors.Documentation + "No documentation available." + f.colors.Reset
	lines = append(lines, escapeForMakefileEcho(noDocsLine))

	// Source information (if available)
	if sourceFile != "" {
		lines = append(lines, escapeForMakefileEcho(""))
		sourceLine := fmt.Sprintf("Source: %s:%d", sourceFile, lineNumber)
		lines = append(lines, escapeForMakefileEcho(sourceLine))
	}

	return lines
}
