package format

import (
	"fmt"
	"strings"

	"github.com/sdlcforge/make-help/internal/model"
)

// Renderer handles the formatting and rendering of help output.
type Renderer struct {
	colors *ColorScheme
}

// NewRenderer creates a new Renderer with the given color mode.
// The color scheme is determined by the useColor setting.
func NewRenderer(useColor bool) *Renderer {
	return &Renderer{
		colors: NewColorScheme(useColor),
	}
}

// Render generates the complete help output from a HelpModel.
// The output includes:
//   - Usage line
//   - Entry point file documentation (if any)
//   - Included files section (if any non-entry files have docs)
//   - Targets section with categories (if applicable)
//
// Returns the formatted help string and any error encountered.
func (r *Renderer) Render(helpModel *model.HelpModel) (string, error) {
	var buf strings.Builder

	// Usage line
	buf.WriteString("Usage: make [<target>...] [<ENV_VAR>=<value>...]\n")

	// File documentation
	if len(helpModel.FileDocs) > 0 {
		// Render entry point file docs first
		for _, fileDoc := range helpModel.FileDocs {
			if fileDoc.IsEntryPoint && len(fileDoc.Documentation) > 0 {
				buf.WriteString("\n")
				for _, line := range fileDoc.Documentation {
					buf.WriteString(line)
					buf.WriteString("\n")
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
			buf.WriteString("\nIncluded Files:\n")
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
	if category.Name != model.UncategorizedCategoryName {
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

	// Summary (pre-computed during model building)
	if target.Summary != "" {
		buf.WriteString(": ")
		buf.WriteString(r.colors.Documentation)
		buf.WriteString(target.Summary)
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

	// Full documentation (blank line only after Variables section)
	if len(target.Documentation) > 0 {
		if len(target.Variables) > 0 {
			buf.WriteString("\n")
		}
		for _, line := range target.Documentation {
			buf.WriteString(r.colors.Documentation)
			buf.WriteString(line)
			buf.WriteString(r.colors.Reset)
			buf.WriteString("\n")
		}
	}

	// Source information
	if target.SourceFile != "" {
		buf.WriteString(fmt.Sprintf("\nSource: %s:%d\n", target.SourceFile, target.LineNumber))
	}

	return buf.String()
}

// RenderBasicTarget renders minimal info for a target without documentation.
// This is used when a target exists but has no associated documentation.
// Shows target name and source location if available.
func (r *Renderer) RenderBasicTarget(name string, sourceFile string, lineNumber int) string {
	var buf strings.Builder

	// Target name
	buf.WriteString(r.colors.TargetName)
	buf.WriteString("Target: ")
	buf.WriteString(name)
	buf.WriteString(r.colors.Reset)
	buf.WriteString("\n")

	// No documentation message
	buf.WriteString("\n")
	buf.WriteString(r.colors.Documentation)
	buf.WriteString("No documentation available.\n")
	buf.WriteString(r.colors.Reset)

	// Source information (if available)
	if sourceFile != "" {
		buf.WriteString(fmt.Sprintf("\nSource: %s:%d\n", sourceFile, lineNumber))
	}

	return buf.String()
}

// RenderForMakefile generates help output suitable for embedding in Makefile @echo statements.
// Returns a slice of strings, each representing one line to be echoed.
// Each line is properly escaped for shell/Makefile context.
// ANSI color codes are embedded as literal escape sequences (e.g., \033[36m).
func (r *Renderer) RenderForMakefile(helpModel *model.HelpModel) ([]string, error) {
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
			lines = append(lines, escapeForMakefileEcho("Included Files:"))
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
			categoryLines := r.renderCategoryForMakefile(&category)
			lines = append(lines, categoryLines...)
		}
	}

	return lines, nil
}

// renderCategoryForMakefile renders a single category for Makefile output.
func (r *Renderer) renderCategoryForMakefile(category *model.Category) []string {
	var lines []string

	// Category name (if present)
	if category.Name != model.UncategorizedCategoryName {
		lines = append(lines, escapeForMakefileEcho(""))
		categoryLine := r.colors.CategoryName + category.Name + ":" + r.colors.Reset
		lines = append(lines, escapeForMakefileEcho(categoryLine))
	}

	// Each target in the category
	for _, target := range category.Targets {
		targetLines := r.renderTargetForMakefile(&target)
		lines = append(lines, targetLines...)
	}

	return lines
}

// renderTargetForMakefile renders a single target for Makefile output.
func (r *Renderer) renderTargetForMakefile(target *model.Target) []string {
	var lines []string
	var buf strings.Builder

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

	// Summary (pre-computed during model building)
	if target.Summary != "" {
		buf.WriteString(": ")
		buf.WriteString(r.colors.Documentation)
		buf.WriteString(target.Summary)
		buf.WriteString(r.colors.Reset)
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
		buf.WriteString(r.colors.Variable)
		buf.WriteString(strings.Join(varNames, ", "))
		buf.WriteString(r.colors.Reset)
		lines = append(lines, escapeForMakefileEcho(buf.String()))
	}

	return lines
}

// RenderDetailedForMakefile renders detailed help for a single target suitable for Makefile @echo.
// Returns a slice of strings, each representing one line to be echoed.
// Each line is properly escaped for shell/Makefile context.
func (r *Renderer) RenderDetailedForMakefile(target *model.Target) []string {
	var lines []string

	// Target name
	targetLine := r.colors.TargetName + "Target: " + target.Name + r.colors.Reset
	lines = append(lines, escapeForMakefileEcho(targetLine))

	// Aliases
	if len(target.Aliases) > 0 {
		aliasLine := r.colors.Alias + "Aliases: " + strings.Join(target.Aliases, ", ") + r.colors.Reset
		lines = append(lines, escapeForMakefileEcho(aliasLine))
	}

	// Variables
	if len(target.Variables) > 0 {
		varHeader := r.colors.Variable + "Variables:" + r.colors.Reset
		lines = append(lines, escapeForMakefileEcho(varHeader))
		for _, v := range target.Variables {
			var varBuf strings.Builder
			varBuf.WriteString("  - ")
			varBuf.WriteString(r.colors.Variable)
			varBuf.WriteString(v.Name)
			varBuf.WriteString(r.colors.Reset)
			if v.Description != "" {
				varBuf.WriteString(": ")
				varBuf.WriteString(r.colors.Documentation)
				varBuf.WriteString(v.Description)
				varBuf.WriteString(r.colors.Reset)
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
			docLine := r.colors.Documentation + line + r.colors.Reset
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

// escapeForMakefileEcho escapes a string for use in Makefile @echo statements.
// Special characters that need escaping:
//   - $ → $$ (Makefile variable escape)
//   - " → \" (shell quote escape)
//   - \ → \\ (shell backslash escape, except for ANSI codes)
//   - ` → \` (shell backtick escape to prevent command substitution)
//   - \x1b (ANSI escape) → \033 (literal form for echo)
//
// ANSI color codes (e.g., \x1b[36m) are converted to literal form (\033[36m) so they work in echo.
func escapeForMakefileEcho(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch ch {
		case '$':
			// Escape $ as $$ for Makefile
			result.WriteString("$$")
		case '"':
			// Escape " as \" for shell
			result.WriteString("\\\"")
		case '\\':
			// Escape \ as \\ for shell
			result.WriteString("\\\\")
		case '`':
			// Escape backtick to prevent command substitution
			result.WriteString("\\`")
		case '\x1b':
			// Convert ANSI escape character to literal \033 for echo
			result.WriteString("\\033")
		default:
			result.WriteByte(ch)
		}
	}
	return result.String()
}
