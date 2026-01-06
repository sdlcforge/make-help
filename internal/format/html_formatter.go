package format

import (
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/sdlcforge/make-help/internal/model"
	"github.com/sdlcforge/make-help/internal/richtext"
)

// HTMLFormatter generates HTML output for web display or documentation sites.
type HTMLFormatter struct {
	config *FormatterConfig
}

// NewHTMLFormatter creates a new HTMLFormatter with the given configuration.
func NewHTMLFormatter(config *FormatterConfig) *HTMLFormatter {
	config = normalizeConfig(config)

	return &HTMLFormatter{
		config: config,
	}
}

// RenderHelp generates the complete help output from a HelpModel in HTML format.
func (f *HTMLFormatter) RenderHelp(helpModel *model.HelpModel, w io.Writer) error {
	if helpModel == nil {
		return fmt.Errorf("html formatter: help model cannot be nil")
	}

	var buf strings.Builder

	// Write HTML structure
	buf.WriteString("<!DOCTYPE html>\n")
	buf.WriteString("<html>\n")
	buf.WriteString("<head>\n")
	buf.WriteString("  <meta charset=\"UTF-8\">\n")
	buf.WriteString("  <title>Makefile Help</title>\n")

	// Embed CSS (only if color is enabled)
	if f.config.UseColor {
		buf.WriteString("  <style>\n")
		buf.WriteString(f.getCSS())
		buf.WriteString("  </style>\n")
	}

	buf.WriteString("</head>\n")
	buf.WriteString("<body>\n")
	buf.WriteString("  <h1>Makefile Help</h1>\n")

	// Usage section
	buf.WriteString("  <section class=\"usage\">\n")
	buf.WriteString("    <h2>Usage</h2>\n")
	buf.WriteString("    <pre>make [&lt;target&gt;...] [&lt;ENV_VAR&gt;=&lt;value&gt;...]</pre>\n")
	buf.WriteString("  </section>\n")

	// File documentation section
	if len(helpModel.FileDocs) > 0 {
		// Render entry point file docs first
		for _, fileDoc := range helpModel.FileDocs {
			if fileDoc.IsEntryPoint && len(fileDoc.Documentation) > 0 {
				buf.WriteString("  <section class=\"file-docs\">\n")
				buf.WriteString("    <h2>Description</h2>\n")
				buf.WriteString("    <div class=\"description\">\n")
				for _, line := range fileDoc.Documentation {
					if line == "" {
						buf.WriteString("      <br>\n")
					} else {
						buf.WriteString("      <p>")
						buf.WriteString(html.EscapeString(line))
						buf.WriteString("</p>\n")
					}
				}
				buf.WriteString("    </div>\n")
				buf.WriteString("  </section>\n")
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
			buf.WriteString("  <section class=\"included-files\">\n")
			buf.WriteString("    <h2>Included files</h2>\n")
			for _, fileDoc := range includedFiles {
				buf.WriteString("    <div class=\"file\">\n")
				buf.WriteString("      <h3>")
				buf.WriteString(html.EscapeString(fileDoc.SourceFile))
				buf.WriteString("</h3>\n")
				for _, line := range fileDoc.Documentation {
					if line == "" {
						buf.WriteString("      <br>\n")
					} else {
						buf.WriteString("      <p>")
						buf.WriteString(html.EscapeString(line))
						buf.WriteString("</p>\n")
					}
				}
				buf.WriteString("    </div>\n")
			}
			buf.WriteString("  </section>\n")
		}
	}

	// Targets section
	if len(helpModel.Categories) > 0 {
		buf.WriteString("  <section class=\"targets\">\n")
		buf.WriteString("    <h2>Targets</h2>\n")

		for _, category := range helpModel.Categories {
			f.renderCategory(&buf, &category)
		}

		buf.WriteString("  </section>\n")
	}

	buf.WriteString("</body>\n")
	buf.WriteString("</html>\n")

	_, err := w.Write([]byte(buf.String()))
	return err
}

// renderCategory renders a single category with its targets in HTML.
func (f *HTMLFormatter) renderCategory(buf *strings.Builder, category *model.Category) {
	buf.WriteString("    <div class=\"category\">\n")

	// Render category name (if present)
	if category.Name != model.UncategorizedCategoryName {
		buf.WriteString("      <h3>")
		buf.WriteString(html.EscapeString(category.Name))
		buf.WriteString("</h3>\n")
	}

	// Render targets as a list
	buf.WriteString("      <ul>\n")
	for _, target := range category.Targets {
		f.renderTarget(buf, &target)
	}
	buf.WriteString("      </ul>\n")
	buf.WriteString("    </div>\n")
}

// renderTarget renders a single target in HTML.
func (f *HTMLFormatter) renderTarget(buf *strings.Builder, target *model.Target) {
	buf.WriteString("        <li class=\"target\">\n")

	// Target name
	buf.WriteString("          <span class=\"target-name\">")
	buf.WriteString(html.EscapeString(target.Name))
	buf.WriteString("</span>")

	// Aliases (if any)
	if len(target.Aliases) > 0 {
		buf.WriteString(" <span class=\"alias\">(")
		for i, alias := range target.Aliases {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(html.EscapeString(alias))
		}
		buf.WriteString(")</span>")
	}

	// Summary: Convert markdown formatting to HTML elements
	summaryHTML := f.renderRichText(target.Summary)
	if summaryHTML != "" {
		buf.WriteString(": <span class=\"summary\">")
		buf.WriteString(summaryHTML)
		buf.WriteString("</span>")
	}

	buf.WriteString("\n")

	// Variables (if any)
	if len(target.Variables) > 0 {
		buf.WriteString("          <div class=\"variables\">\n")
		buf.WriteString("            Variables: ")
		for i, v := range target.Variables {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString("<code class=\"variable\">")
			buf.WriteString(html.EscapeString(v.Name))
			buf.WriteString("</code>")
		}
		buf.WriteString("\n          </div>\n")
	}

	buf.WriteString("        </li>\n")
}

// RenderDetailedTarget renders a detailed view of a single target in HTML.
func (f *HTMLFormatter) RenderDetailedTarget(target *model.Target, w io.Writer) error {
	if target == nil {
		return fmt.Errorf("html formatter: target cannot be nil")
	}

	var buf strings.Builder

	buf.WriteString("<!DOCTYPE html>\n")
	buf.WriteString("<html>\n")
	buf.WriteString("<head>\n")
	buf.WriteString("  <meta charset=\"UTF-8\">\n")
	buf.WriteString(fmt.Sprintf("  <title>Target: %s</title>\n", html.EscapeString(target.Name)))

	if f.config.UseColor {
		buf.WriteString("  <style>\n")
		buf.WriteString(f.getCSS())
		buf.WriteString("  </style>\n")
	}

	buf.WriteString("</head>\n")
	buf.WriteString("<body>\n")
	buf.WriteString("  <h1>Target: ")
	buf.WriteString(html.EscapeString(target.Name))
	buf.WriteString("</h1>\n")

	// Aliases
	if len(target.Aliases) > 0 {
		buf.WriteString("  <div class=\"aliases\">\n")
		buf.WriteString("    <strong>Aliases:</strong> ")
		buf.WriteString(html.EscapeString(strings.Join(target.Aliases, ", ")))
		buf.WriteString("\n  </div>\n")
	}

	// Variables
	if len(target.Variables) > 0 {
		buf.WriteString("  <div class=\"variables\">\n")
		buf.WriteString("    <strong>Variables:</strong>\n")
		buf.WriteString("    <ul>\n")
		for _, v := range target.Variables {
			buf.WriteString("      <li><code class=\"variable\">")
			buf.WriteString(html.EscapeString(v.Name))
			buf.WriteString("</code>")
			if v.Description != "" {
				buf.WriteString(": ")
				buf.WriteString(html.EscapeString(v.Description))
			}
			buf.WriteString("</li>\n")
		}
		buf.WriteString("    </ul>\n")
		buf.WriteString("  </div>\n")
	}

	// Full documentation
	if len(target.Documentation) > 0 {
		buf.WriteString("  <div class=\"documentation\">\n")
		for _, line := range target.Documentation {
			if line == "" {
				buf.WriteString("    <br>\n")
			} else {
				buf.WriteString("    <p>")
				buf.WriteString(html.EscapeString(line))
				buf.WriteString("</p>\n")
			}
		}
		buf.WriteString("  </div>\n")
	}

	// Source information
	if target.SourceFile != "" {
		buf.WriteString("  <div class=\"source\">\n")
		buf.WriteString("    <strong>Source:</strong> ")
		buf.WriteString(html.EscapeString(fmt.Sprintf("%s:%d", target.SourceFile, target.LineNumber)))
		buf.WriteString("\n  </div>\n")
	}

	buf.WriteString("</body>\n")
	buf.WriteString("</html>\n")

	_, err := w.Write([]byte(buf.String()))
	return err
}

// RenderBasicTarget renders minimal info for a target without documentation in HTML.
func (f *HTMLFormatter) RenderBasicTarget(name string, sourceFile string, lineNumber int, w io.Writer) error {
	var buf strings.Builder

	buf.WriteString("<!DOCTYPE html>\n")
	buf.WriteString("<html>\n")
	buf.WriteString("<head>\n")
	buf.WriteString("  <meta charset=\"UTF-8\">\n")
	buf.WriteString(fmt.Sprintf("  <title>Target: %s</title>\n", html.EscapeString(name)))

	if f.config.UseColor {
		buf.WriteString("  <style>\n")
		buf.WriteString(f.getCSS())
		buf.WriteString("  </style>\n")
	}

	buf.WriteString("</head>\n")
	buf.WriteString("<body>\n")
	buf.WriteString("  <h1>Target: ")
	buf.WriteString(html.EscapeString(name))
	buf.WriteString("</h1>\n")

	// No documentation message
	buf.WriteString("  <p class=\"no-docs\">No documentation available.</p>\n")

	// Source information (if available)
	if sourceFile != "" {
		buf.WriteString("  <div class=\"source\">\n")
		buf.WriteString("    <strong>Source:</strong> ")
		buf.WriteString(html.EscapeString(fmt.Sprintf("%s:%d", sourceFile, lineNumber)))
		buf.WriteString("\n  </div>\n")
	}

	buf.WriteString("</body>\n")
	buf.WriteString("</html>\n")

	_, err := w.Write([]byte(buf.String()))
	return err
}

// ContentType returns the MIME type for HTML format.
func (f *HTMLFormatter) ContentType() string {
	return "text/html"
}

// DefaultExtension returns the default file extension for HTML format.
func (f *HTMLFormatter) DefaultExtension() string {
	return ".html"
}

// isValidURL validates that a URL uses a safe scheme.
// Only http://, https://, and relative URLs (starting with / or without a colon) are allowed.
// This prevents javascript: and other potentially dangerous URL schemes.
func isValidURL(url string) bool {
	if url == "" {
		return false
	}

	// Normalize to lowercase to prevent case-sensitivity bypass
	normalizedURL := strings.ToLower(url)

	// Check for common safe prefixes
	if strings.HasPrefix(normalizedURL, "http://") || strings.HasPrefix(normalizedURL, "https://") {
		return true
	}

	// Allow relative URLs starting with /
	if strings.HasPrefix(url, "/") {
		return true
	}

	// Allow relative URLs without a scheme (no colon before any slash or end of string)
	colonIndex := strings.Index(normalizedURL, ":")
	slashIndex := strings.Index(normalizedURL, "/")

	// No colon = relative path
	if colonIndex == -1 {
		return true
	}

	// Colon after slash = relative path with colon in filename
	if slashIndex != -1 && colonIndex > slashIndex {
		return true
	}

	// Colon before any slash = scheme present, and it's not http/https
	return false
}

// renderRichText converts RichText segments to HTML.
func (f *HTMLFormatter) renderRichText(rt richtext.RichText) string {
	var buf strings.Builder
	for _, seg := range rt {
		switch seg.Type {
		case richtext.SegmentBold:
			buf.WriteString("<strong>")
			buf.WriteString(html.EscapeString(seg.Content))
			buf.WriteString("</strong>")
		case richtext.SegmentItalic:
			buf.WriteString("<em>")
			buf.WriteString(html.EscapeString(seg.Content))
			buf.WriteString("</em>")
		case richtext.SegmentCode:
			buf.WriteString("<code>")
			buf.WriteString(html.EscapeString(seg.Content))
			buf.WriteString("</code>")
		case richtext.SegmentLink:
			// Only render as link if URL scheme is safe
			if isValidURL(seg.URL) {
				buf.WriteString("<a href=\"")
				buf.WriteString(html.EscapeString(seg.URL))
				buf.WriteString("\">")
				buf.WriteString(html.EscapeString(seg.Content))
				buf.WriteString("</a>")
			} else {
				// Render as plain text if URL is unsafe
				buf.WriteString(html.EscapeString(seg.Content))
			}
		default:
			buf.WriteString(html.EscapeString(seg.Content))
		}
	}
	return buf.String()
}

// cachedHTMLCSS contains the embedded CSS stylesheet (cached at package level for performance).
// The color scheme is inspired by the Flat UI Colors palette (https://flatuicolors.com)
// to provide consistent, accessible styling with good contrast ratios.
var cachedHTMLCSS = `    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
      max-width: 1000px;
      margin: 2em auto;
      padding: 0 1em;
      line-height: 1.6;
      color: #333;
    }
    h1 {
      color: #2c3e50;  /* Midnight Blue - main heading */
      border-bottom: 2px solid #3498db;  /* Peter River - accent for main heading */
      padding-bottom: 0.5em;
    }
    h2 {
      color: #34495e;  /* Wet Asphalt - section headings (categories, files) */
      margin-top: 1.5em;
      border-bottom: 1px solid #ecf0f1;  /* Clouds - subtle divider */
      padding-bottom: 0.3em;
    }
    h3 {
      color: #34495e;  /* Wet Asphalt - subsection headings */
      margin-top: 1em;
    }
    pre {
      background-color: #f8f8f8;
      border: 1px solid #ddd;
      border-radius: 3px;
      padding: 1em;
      overflow-x: auto;
    }
    code {
      background-color: #f8f8f8;
      border-radius: 3px;
      padding: 0.2em 0.4em;
      font-family: "Monaco", "Menlo", "Consolas", monospace;
      font-size: 0.9em;
    }
    .category {
      margin-bottom: 2em;
    }
    .target {
      margin: 0.5em 0;
      line-height: 1.8;
    }
    .target-name {
      font-weight: bold;
      color: #27ae60;  /* Nephritis - make target names (green indicates actionable) */
    }
    .alias {
      color: #f39c12;  /* Orange - target aliases (distinctive color for alternative names) */
      font-style: italic;
    }
    .summary {
      color: #555;  /* Dark gray - summary text */
    }
    .variables {
      color: #7f8c8d;  /* Asbestos - variable section labels (muted gray) */
      font-size: 0.9em;
      margin-left: 1.5em;
      margin-top: 0.2em;
    }
    .variable {
      color: #9b59b6;  /* Amethyst - environment variable names (purple for configurables) */
    }
    .description p {
      margin: 0.5em 0;
    }
    .documentation p {
      margin: 0.5em 0;
    }
    .file {
      margin-bottom: 1.5em;
    }
    .source {
      margin-top: 1em;
      color: #7f8c8d;  /* Asbestos - source file references (muted gray for metadata) */
      font-size: 0.9em;
    }
    .no-docs {
      color: #95a5a6;  /* Concrete - placeholder text for undocumented items (light gray) */
      font-style: italic;
    }
    ul {
      list-style-type: none;
      padding-left: 0;
    }
    .aliases, .variables {
      margin: 0.5em 0;
    }
`

// getCSS returns the cached CSS stylesheet.
func (f *HTMLFormatter) getCSS() string {
	return cachedHTMLCSS
}
