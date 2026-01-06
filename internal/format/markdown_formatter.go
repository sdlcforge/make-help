package format

import (
	"fmt"
	"io"
	"strings"

	"github.com/sdlcforge/make-help/internal/model"
)

// MarkdownFormatter generates Markdown output for GitHub/GitLab/documentation sites.
type MarkdownFormatter struct {
	config *FormatterConfig
}

// NewMarkdownFormatter creates a new MarkdownFormatter with the given configuration.
func NewMarkdownFormatter(config *FormatterConfig) *MarkdownFormatter {
	if config == nil {
		config = &FormatterConfig{UseColor: false}
	}

	return &MarkdownFormatter{
		config: config,
	}
}

// escapeMarkdown escapes special Markdown characters in structural elements
// to prevent accidental formatting. Does not escape in documentation content
// where formatting may be intentional.
func escapeMarkdown(s string) string {
	// Escape: * _ ` [ ] ( ) #
	replacer := strings.NewReplacer(
		`*`, `\*`,
		`_`, `\_`,
		"`", "\\`",
		`[`, `\[`,
		`]`, `\]`,
		`(`, `\(`,
		`)`, `\)`,
		`#`, `\#`,
	)
	return replacer.Replace(s)
}

// RenderHelp generates the complete help output from a HelpModel in Markdown format.
func (f *MarkdownFormatter) RenderHelp(helpModel *model.HelpModel, w io.Writer) error {
	if helpModel == nil {
		return fmt.Errorf("markdown formatter: help model cannot be nil")
	}

	var buf strings.Builder

	// Title
	buf.WriteString("# Makefile Help\n\n")

	// Usage section
	buf.WriteString("## Usage\n\n")
	buf.WriteString("```\n")
	buf.WriteString("make [<target>...] [<ENV_VAR>=<value>...]\n")
	buf.WriteString("```\n\n")

	// File documentation section
	if len(helpModel.FileDocs) > 0 {
		// Render entry point file docs first
		for _, fileDoc := range helpModel.FileDocs {
			if fileDoc.IsEntryPoint && len(fileDoc.Documentation) > 0 {
				buf.WriteString("## Description\n\n")
				for _, line := range fileDoc.Documentation {
					buf.WriteString(line)
					buf.WriteString("\n")
				}
				buf.WriteString("\n")
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
			buf.WriteString("## Included files\n\n")
			for _, fileDoc := range includedFiles {
				buf.WriteString("### ")
				buf.WriteString(escapeMarkdown(fileDoc.SourceFile))
				buf.WriteString("\n\n")
				for _, line := range fileDoc.Documentation {
					buf.WriteString(line)
					buf.WriteString("\n")
				}
				buf.WriteString("\n")
			}
		}
	}

	// Targets section
	if len(helpModel.Categories) > 0 {
		buf.WriteString("## Targets\n\n")

		for _, category := range helpModel.Categories {
			f.renderCategory(&buf, &category)
		}
	}

	_, err := w.Write([]byte(buf.String()))
	return err
}

// renderCategory renders a single category with its targets in Markdown.
func (f *MarkdownFormatter) renderCategory(buf *strings.Builder, category *model.Category) {
	// Render category name (if present)
	if category.Name != model.UncategorizedCategoryName {
		buf.WriteString("### ")
		buf.WriteString(escapeMarkdown(category.Name))
		buf.WriteString("\n\n")
	}

	// Render targets as a list
	for _, target := range category.Targets {
		f.renderTarget(buf, &target)
	}

	buf.WriteString("\n")
}

// renderTarget renders a single target in Markdown.
func (f *MarkdownFormatter) renderTarget(buf *strings.Builder, target *model.Target) {
	buf.WriteString("- **")
	buf.WriteString(escapeMarkdown(target.Name))
	buf.WriteString("**")

	// Aliases (if any)
	if len(target.Aliases) > 0 {
		buf.WriteString(" _(")
		escapedAliases := make([]string, len(target.Aliases))
		for i, alias := range target.Aliases {
			escapedAliases[i] = escapeMarkdown(alias)
		}
		buf.WriteString(strings.Join(escapedAliases, ", "))
		buf.WriteString(")_")
	}

	// Summary: Preserve markdown formatting for markdown output
	summaryMarkdown := target.Summary.Markdown()
	if summaryMarkdown != "" {
		buf.WriteString(": ")
		buf.WriteString(summaryMarkdown)
	}

	buf.WriteString("\n")

	// Variables (if any)
	if len(target.Variables) > 0 {
		buf.WriteString("  - Variables: ")
		for i, v := range target.Variables {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString("`")
			buf.WriteString(escapeMarkdown(v.Name))
			buf.WriteString("`")
		}
		buf.WriteString("\n")
	}
}

// RenderDetailedTarget renders a detailed view of a single target in Markdown.
func (f *MarkdownFormatter) RenderDetailedTarget(target *model.Target, w io.Writer) error {
	if target == nil {
		return fmt.Errorf("markdown formatter: target cannot be nil")
	}

	var buf strings.Builder

	// Target name
	buf.WriteString("# Target: ")
	buf.WriteString(escapeMarkdown(target.Name))
	buf.WriteString("\n\n")

	// Aliases
	if len(target.Aliases) > 0 {
		buf.WriteString("**Aliases:** ")
		escapedAliases := make([]string, len(target.Aliases))
		for i, alias := range target.Aliases {
			escapedAliases[i] = escapeMarkdown(alias)
		}
		buf.WriteString(strings.Join(escapedAliases, ", "))
		buf.WriteString("\n\n")
	}

	// Variables
	if len(target.Variables) > 0 {
		buf.WriteString("**Variables:**\n\n")
		for _, v := range target.Variables {
			buf.WriteString("- `")
			buf.WriteString(escapeMarkdown(v.Name))
			buf.WriteString("`")
			if v.Description != "" {
				buf.WriteString(": ")
				buf.WriteString(v.Description)
			}
			buf.WriteString("\n")
		}
		buf.WriteString("\n")
	}

	// Full documentation
	if len(target.Documentation) > 0 {
		buf.WriteString("## Description\n\n")
		for _, line := range target.Documentation {
			buf.WriteString(line)
			buf.WriteString("\n")
		}
		buf.WriteString("\n")
	}

	// Source information
	if target.SourceFile != "" {
		buf.WriteString("**Source:** `")
		buf.WriteString(fmt.Sprintf("%s:%d", target.SourceFile, target.LineNumber))
		buf.WriteString("`\n")
	}

	_, err := w.Write([]byte(buf.String()))
	return err
}

// RenderBasicTarget renders minimal info for a target without documentation in Markdown.
func (f *MarkdownFormatter) RenderBasicTarget(name string, sourceFile string, lineNumber int, w io.Writer) error {
	var buf strings.Builder

	// Target name
	buf.WriteString("# Target: ")
	buf.WriteString(escapeMarkdown(name))
	buf.WriteString("\n\n")

	// No documentation message
	buf.WriteString("_No documentation available._\n\n")

	// Source information (if available)
	if sourceFile != "" {
		buf.WriteString("**Source:** `")
		buf.WriteString(fmt.Sprintf("%s:%d", sourceFile, lineNumber))
		buf.WriteString("`\n")
	}

	_, err := w.Write([]byte(buf.String()))
	return err
}

// ContentType returns the MIME type for Markdown format.
func (f *MarkdownFormatter) ContentType() string {
	return "text/markdown"
}

// DefaultExtension returns the default file extension for Markdown format.
func (f *MarkdownFormatter) DefaultExtension() string {
	return ".md"
}
