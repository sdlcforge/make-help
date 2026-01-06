# Formatter Architecture Design

**Version:** 1.0
**Date:** 2025-12-19
**Status:** Design Document

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Architecture Overview](#architecture-overview)
3. [Rich Text Handling](#rich-text-handling)
4. [Interface Definitions](#interface-definitions)
5. [Format Implementations](#format-implementations)
6. [CLI Changes](#cli-changes)
7. [File Structure](#file-structure)
8. [Migration Path](#migration-path)
9. [Implementation Priorities](#implementation-priorities)

---

## Executive Summary

This design introduces a common formatter architecture to support multiple output formats (Make, Text, HTML, Markdown) for the `make-help` CLI tool. Key changes include:

1. **New `Formatter` interface** - Common abstraction for all output formats
2. **Rich text preservation** - Markdown formatting (bold, italic) preserved/converted instead of stripped
3. **Format selection via `--format` flag** - Replaces mode detection logic
4. **Output destination via `--output` flag** - Replaces `--show-help` flag
5. **Extensible design** - Easy to add new formats in the future

### Key Benefits

- **Consistency**: All formats share common rendering logic
- **Flexibility**: Users can choose output format and destination independently
- **Rich formatting**: Documentation preserves formatting appropriate to each format
- **Extensibility**: New formats can be added without changing core logic
- **Clarity**: Clear separation between format selection and output destination

---

## Architecture Overview

### Current Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLI Layer                                │
│  (root.go, help.go, create_help_target.go)                      │
└──────────────────────┬──────────────────────────────────────────┘
                       │
                       v
┌─────────────────────────────────────────────────────────────────┐
│                  Processing Pipeline                             │
│  Discovery → Parser → Model Builder → Ordering                   │
└──────────────────────┬──────────────────────────────────────────┘
                       │
                       v
┌─────────────────────────────────────────────────────────────────┐
│                  Summary Extraction                              │
│  (Strips markdown formatting for summaries)                      │
└──────────────────────┬──────────────────────────────────────────┘
                       │
                       v
┌─────────────────────────────────────────────────────────────────┐
│                    Single Renderer                               │
│  format.Renderer (ANSI-colored text or plain text)              │
│    - Render() → stdout display                                   │
│    - RenderForMakefile() → @echo statements                      │
│    - RenderDetailedTarget() → detailed view                      │
└─────────────────────────────────────────────────────────────────┘
```

### Proposed Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLI Layer                                │
│  Flags: --format {make|text|html|markdown}                      │
│         --output {path|-} (- means stdout)                       │
└──────────────────────┬──────────────────────────────────────────┘
                       │
                       v
┌─────────────────────────────────────────────────────────────────┐
│                  Processing Pipeline                             │
│  Discovery → Parser → Model Builder → Ordering                   │
│  (No changes to existing pipeline)                               │
└──────────────────────┬──────────────────────────────────────────┘
                       │
                       v
┌─────────────────────────────────────────────────────────────────┐
│                  Rich Text Processing                            │
│  (Preserves markdown, converts based on format)                 │
│    - Parse markdown inline formatting                            │
│    - Summary extraction WITHOUT stripping formatting             │
│    - Format-specific conversion (bold/italic)                    │
└──────────────────────┬──────────────────────────────────────────┘
                       │
                       v
┌─────────────────────────────────────────────────────────────────┐
│                    Formatter Factory                             │
│  Creates appropriate formatter based on --format flag            │
└──────────────────────┬──────────────────────────────────────────┘
                       │
           ┌───────────┴───────────┬──────────────┬──────────────┐
           │                       │              │              │
           v                       v              v              v
    ┌──────────┐          ┌──────────┐    ┌──────────┐   ┌──────────┐
    │   Make   │          │   Text   │    │   HTML   │   │ Markdown │
    │Formatter │          │Formatter │    │Formatter │   │Formatter │
    └─────┬────┘          └─────┬────┘    └─────┬────┘   └─────┬────┘
          │                     │               │              │
          └─────────────────────┴───────────────┴──────────────┘
                                │
                                v
                      ┌──────────────────┐
                      │   Output Writer  │
                      │  (stdout or file)│
                      └──────────────────┘
```

### Data Flow

```
Documentation Input:
  "**Build** the application using *Go compiler*."

    ↓

Rich Text Parsing:
  Segments: [
    {type: "text", content: ""},
    {type: "bold", content: "Build"},
    {type: "text", content: " the application using "},
    {type: "italic", content: "Go compiler"},
    {type: "text", content: "."}
  ]

    ↓

Summary Extraction (preserves formatting):
  "**Build** the application using *Go compiler*."

    ↓

Format-Specific Rendering:

  Make (with color):
    "\033[1mBuild\033[0m the application using \033[4mGo compiler\033[0m."
    (wrapped in @printf '%b\n' "..." statements)

  Make (no color):
    "Build the application using Go compiler."

  Text (with color):
    "\033[1mBuild\033[0m the application using \033[4mGo compiler\033[0m."

  Text (no color):
    "Build the application using Go compiler."

  HTML:
    "<strong>Build</strong> the application using <em>Go compiler</em>."

  Markdown:
    "**Build** the application using *Go compiler*."
```

---

## Rich Text Handling

### Current Behavior (Problem)

The `summary.Extractor` strips all markdown formatting:

```go
// internal/summary/extractor.go (lines 76-90)
func (e *Extractor) stripMarkdownFormatting(text string) string {
    text = e.boldRegex.ReplaceAllString(text, "$1")       // **bold** → bold
    text = e.italicRegex.ReplaceAllString(text, "$1")     // *italic* → italic
    text = e.boldUnderRegex.ReplaceAllString(text, "$1")  // __bold__ → bold
    text = e.italicUnderRegex.ReplaceAllString(text, "$1") // _italic_ → italic
    // ...
}
```

This loses semantic information about emphasis and importance.

### Proposed Solution: Rich Text Segment Model

#### 1. Rich Text Data Model

```go
// internal/richtext/types.go

package richtext

// SegmentType represents the type of text segment
type SegmentType int

const (
    SegmentPlain SegmentType = iota  // Plain text
    SegmentBold                       // **text** or __text__
    SegmentItalic                     // *text* or _text_
    SegmentCode                       // `code` (future: might render as monospace)
    SegmentLink                       // [text](url)
)

// Segment represents a piece of text with optional formatting
type Segment struct {
    Type    SegmentType
    Content string      // The text content (without markdown markers)
    URL     string      // For links only (empty for other types)
}

// RichText represents formatted text as a sequence of segments
type RichText []Segment

// PlainText returns the text with all formatting stripped
func (rt RichText) PlainText() string {
    var buf strings.Builder
    for _, seg := range rt {
        buf.WriteString(seg.Content)
    }
    return buf.String()
}

// Markdown returns the text with markdown formatting preserved
func (rt RichText) Markdown() string {
    var buf strings.Builder
    for _, seg := range rt {
        switch seg.Type {
        case SegmentBold:
            buf.WriteString("**")
            buf.WriteString(seg.Content)
            buf.WriteString("**")
        case SegmentItalic:
            buf.WriteString("*")
            buf.WriteString(seg.Content)
            buf.WriteString("*")
        case SegmentCode:
            buf.WriteString("`")
            buf.WriteString(seg.Content)
            buf.WriteString("`")
        case SegmentLink:
            buf.WriteString("[")
            buf.WriteString(seg.Content)
            buf.WriteString("](")
            buf.WriteString(seg.URL)
            buf.WriteString(")")
        default:
            buf.WriteString(seg.Content)
        }
    }
    return buf.String()
}
```

#### 2. Rich Text Parser

```go
// internal/richtext/parser.go

package richtext

// Parser parses markdown inline formatting into RichText segments
type Parser struct {
    boldRegex        *regexp.Regexp
    italicRegex      *regexp.Regexp
    boldUnderRegex   *regexp.Regexp
    italicUnderRegex *regexp.Regexp
    codeRegex        *regexp.Regexp
    linkRegex        *regexp.Regexp
}

func NewParser() *Parser {
    return &Parser{
        boldRegex:        regexp.MustCompile(`\*\*([^*]+)\*\*`),
        italicRegex:      regexp.MustCompile(`\*([^*]+)\*`),
        boldUnderRegex:   regexp.MustCompile(`__([^_]+)__`),
        italicUnderRegex: regexp.MustCompile(`_([^_]+)_`),
        codeRegex:        regexp.MustCompile("`([^`]+)`"),
        linkRegex:        regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`),
    }
}

// Parse converts a markdown string into RichText segments
// Process order: links → code → bold → italic
func (p *Parser) Parse(text string) RichText {
    // Implementation would:
    // 1. Find all formatting patterns with their positions
    // 2. Sort by position
    // 3. Handle overlapping/nested patterns (bold takes precedence over italic)
    // 4. Build segments in order
    // 5. Return RichText
}
```

#### 3. Summary Extraction with Rich Text

```go
// internal/summary/extractor.go (modified)

// Extract now returns RichText instead of string
func (e *Extractor) Extract(documentation []string) richtext.RichText {
    if len(documentation) == 0 {
        return nil
    }

    // Join all documentation lines
    fullText := strings.Join(documentation, " ")

    // Strip markdown headers (unchanged)
    fullText = e.stripMarkdownHeaders(fullText)

    // Parse into rich text BEFORE stripping formatting
    richParser := richtext.NewParser()
    richText := richParser.Parse(fullText)

    // Strip HTML tags from plain content
    plainText := richText.PlainText()
    plainText = e.stripHTMLTags(plainText)

    // Normalize whitespace
    plainText = e.normalizeWhitespace(plainText)

    // Extract first sentence
    firstSentence := e.extractFirstSentence(plainText)

    // Parse the first sentence to preserve formatting
    return richParser.Parse(firstSentence)
}
```

#### 4. Model Changes

```go
// internal/model/types.go (modified)

import "github.com/sdlcforge/make-help/internal/richtext"

type Target struct {
    Name          string
    Aliases       []string
    Documentation []string
    Summary       richtext.RichText  // CHANGED: was string
    Variables     []Variable
    // ... rest unchanged
}
```

### Format-Specific Rendering

Each formatter converts `richtext.RichText` to its native format:

| Format   | Bold                  | Italic                | Plain Text |
|----------|-----------------------|----------------------|------------|
| Make (color)  | `\033[1m...\033[0m`   | `\033[4m...\033[0m`  | as-is      |
| Make (no color) | strip                | strip                | as-is      |
| Text (color)  | `\033[1m...\033[0m`   | `\033[4m...\033[0m`  | as-is      |
| Text (no color) | strip                | strip                | as-is      |
| HTML     | `<strong>...</strong>` | `<em>...</em>`       | HTML escape |
| Markdown | `**...**`             | `*...*`              | as-is      |

---

## Interface Definitions

### Core Formatter Interface

```go
// internal/format/formatter.go

package format

import (
    "io"
    "github.com/sdlcforge/make-help/internal/model"
)

// Formatter is the interface that all output format implementations must satisfy.
// Each formatter knows how to render a HelpModel in its specific format.
type Formatter interface {
    // RenderHelp generates the complete help output from a HelpModel.
    // This is the summary view showing all categories and targets.
    RenderHelp(model *model.HelpModel, w io.Writer) error

    // RenderDetailedTarget generates detailed help for a single target.
    // Shows full documentation, variables, aliases, and source location.
    RenderDetailedTarget(target *model.Target, w io.Writer) error

    // RenderBasicTarget generates minimal help for an undocumented target.
    // Shows target name and source location if available.
    RenderBasicTarget(name string, sourceFile string, lineNumber int, w io.Writer) error

    // ContentType returns the MIME type for this format (for HTTP responses, future use).
    ContentType() string

    // DefaultExtension returns the default file extension for this format.
    DefaultExtension() string
}

// FormatterConfig holds configuration options common to all formatters.
type FormatterConfig struct {
    // UseColor enables colored/styled output (where applicable).
    // For terminal formats: ANSI escape codes
    // For HTML: CSS classes
    // For Markdown/Make: no effect (or minimal effect)
    UseColor bool

    // ColorScheme defines colors for different elements (terminal formats only).
    // When UseColor is false, this is nil.
    ColorScheme *ColorScheme
}

// NewFormatter creates a formatter for the specified format type.
// This is the factory function that replaces direct renderer construction.
func NewFormatter(formatType string, config *FormatterConfig) (Formatter, error) {
    switch formatType {
    case "make", "mk":
        return NewMakeFormatter(config), nil
    case "text", "txt":
        return NewTextFormatter(config), nil
    case "html":
        return NewHTMLFormatter(config), nil
    case "markdown", "md":
        return NewMarkdownFormatter(config), nil
    default:
        return nil, fmt.Errorf("unknown format type: %s (supported: make, text, html, markdown)", formatType)
    }
}
```

### Rich Text Renderer Interface

```go
// internal/format/richtext_renderer.go

package format

import "github.com/sdlcforge/make-help/internal/richtext"

// RichTextRenderer converts RichText segments to format-specific output.
// Each formatter embeds or implements this interface.
type RichTextRenderer interface {
    // RenderRichText converts a RichText to a string in this format.
    RenderRichText(rt richtext.RichText) string
}

// BaseRichTextRenderer provides common rich text rendering logic.
// Formatters can embed this and override specific methods.
type BaseRichTextRenderer struct {
    useColor bool
    colors   *ColorScheme
}

func (r *BaseRichTextRenderer) RenderRichText(rt richtext.RichText) string {
    // Default implementation for terminal/text formats
}
```

---

## Format Implementations

### 1. Make Formatter

**Purpose:** Generate Makefile content with help targets using `@printf` statements.

**Location:** `internal/format/make_formatter.go`

```go
package format

type MakeFormatter struct {
    config          *FormatterConfig
    richTextRenderer *MakeRichTextRenderer
}

func NewMakeFormatter(config *FormatterConfig) *MakeFormatter {
    return &MakeFormatter{
        config:          config,
        richTextRenderer: NewMakeRichTextRenderer(config),
    }
}

// RenderHelp generates a complete Makefile with help targets
func (f *MakeFormatter) RenderHelp(model *model.HelpModel, w io.Writer) error {
    // Generate header comments (generated-by, command, date)
    // Generate variables (MAKE_HELP_DIR, MAKE_HELP_MAKEFILES)
    // Generate main help target with @printf statements
    // Generate help-<target> targets for detailed help
    // Generate update-help target for regeneration
}

func (f *MakeFormatter) ContentType() string {
    return "text/x-makefile"
}

func (f *MakeFormatter) DefaultExtension() string {
    return ".mk"
}

// MakeRichTextRenderer handles rich text conversion for Makefile @printf statements
type MakeRichTextRenderer struct {
    useColor bool
    colors   *ColorScheme
}

func (r *MakeRichTextRenderer) RenderRichText(rt richtext.RichText) string {
    var buf strings.Builder
    for _, seg := range rt {
        switch seg.Type {
        case richtext.SegmentBold:
            if r.useColor {
                buf.WriteString(r.colors.BoldCode)
                buf.WriteString(seg.Content)
                buf.WriteString(r.colors.Reset)
            } else {
                buf.WriteString(seg.Content)
            }
        case richtext.SegmentItalic:
            if r.useColor {
                buf.WriteString(r.colors.UnderlineCode) // Italic → underline in terminal
                buf.WriteString(seg.Content)
                buf.WriteString(r.colors.Reset)
            } else {
                buf.WriteString(seg.Content)
            }
        default:
            buf.WriteString(seg.Content)
        }
    }
    return escapeForMakefileEcho(buf.String())
}
```

**Key Features:**
- Embeds ANSI color codes in `@printf` statements when `UseColor` is true
- Escapes special characters for shell/Makefile context (`$`, `"`, `\`, backticks)
- Generates complete Makefile structure (variables, .PHONY, help targets, update-help target)

### 2. Text Formatter

**Purpose:** Plain text output suitable for terminal display or text files.

**Location:** `internal/format/text_formatter.go`

```go
package format

type TextFormatter struct {
    config          *FormatterConfig
    richTextRenderer *TextRichTextRenderer
}

func NewTextFormatter(config *FormatterConfig) *TextFormatter {
    return &TextFormatter{
        config:          config,
        richTextRenderer: NewTextRichTextRenderer(config),
    }
}

func (f *TextFormatter) RenderHelp(model *model.HelpModel, w io.Writer) error {
    // Render: Usage line
    // Render: File documentation
    // Render: Included files
    // Render: Targets by category
    // Direct output to writer (no @echo wrapping)
}

func (f *TextFormatter) ContentType() string {
    return "text/plain"
}

func (f *TextFormatter) DefaultExtension() string {
    return ".txt"
}

type TextRichTextRenderer struct {
    useColor bool
    colors   *ColorScheme
}

func (r *TextRichTextRenderer) RenderRichText(rt richtext.RichText) string {
    // Same as MakeRichTextRenderer but without escaping for Makefile
    var buf strings.Builder
    for _, seg := range rt {
        switch seg.Type {
        case richtext.SegmentBold:
            if r.useColor {
                buf.WriteString("\033[1m")
                buf.WriteString(seg.Content)
                buf.WriteString("\033[0m")
            } else {
                buf.WriteString(seg.Content)
            }
        case richtext.SegmentItalic:
            if r.useColor {
                buf.WriteString("\033[4m")
                buf.WriteString(seg.Content)
                buf.WriteString("\033[0m")
            } else {
                buf.WriteString(seg.Content)
            }
        default:
            buf.WriteString(seg.Content)
        }
    }
    return buf.String()
}
```

**Key Features:**
- ANSI color codes for terminal display (when `UseColor` is true)
- No escaping needed (unlike Make format)
- Identical structure to current `Renderer.Render()` output

### 3. HTML Formatter

**Purpose:** HTML output for web display or documentation sites.

**Location:** `internal/format/html_formatter.go`

```go
package format

type HTMLFormatter struct {
    config          *FormatterConfig
    richTextRenderer *HTMLRichTextRenderer
}

func NewHTMLFormatter(config *FormatterConfig) *HTMLFormatter {
    return &HTMLFormatter{
        config:          config,
        richTextRenderer: NewHTMLRichTextRenderer(config),
    }
}

func (f *HTMLFormatter) RenderHelp(model *model.HelpModel, w io.Writer) error {
    // Write HTML structure:
    // <!DOCTYPE html>
    // <html>
    // <head><style>...</style></head>
    // <body>
    //   <h1>Makefile Help</h1>
    //   <section class="usage">...</section>
    //   <section class="file-docs">...</section>
    //   <section class="targets">
    //     <div class="category">
    //       <h2>Category Name</h2>
    //       <ul>
    //         <li class="target">
    //           <span class="target-name">build</span>
    //           <span class="summary">...</span>
    //         </li>
    //       </ul>
    //     </div>
    //   </section>
    // </body>
    // </html>
}

func (f *HTMLFormatter) ContentType() string {
    return "text/html"
}

func (f *HTMLFormatter) DefaultExtension() string {
    return ".html"
}

type HTMLRichTextRenderer struct {
    useColor bool // Controls whether to include CSS classes
}

func (r *HTMLRichTextRenderer) RenderRichText(rt richtext.RichText) string {
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
            buf.WriteString("<a href=\"")
            buf.WriteString(html.EscapeString(seg.URL))
            buf.WriteString("\">")
            buf.WriteString(html.EscapeString(seg.Content))
            buf.WriteString("</a>")
        default:
            buf.WriteString(html.EscapeString(seg.Content))
        }
    }
    return buf.String()
}
```

**Key Features:**
- Semantic HTML5 structure
- Embedded CSS for styling (when `UseColor` is true)
- Proper HTML escaping for all text content
- CSS classes for styling flexibility

**Sample CSS** (embedded in `<style>` tag):
```css
body { font-family: sans-serif; max-width: 1000px; margin: 2em auto; padding: 0 1em; }
h1 { color: #2c3e50; border-bottom: 2px solid #3498db; padding-bottom: 0.5em; }
h2 { color: #34495e; margin-top: 1.5em; }
.target { margin: 0.5em 0; }
.target-name { font-weight: bold; color: #27ae60; }
.alias { color: #f39c12; font-style: italic; }
.summary { color: #7f8c8d; }
.variable { color: #9b59b6; font-family: monospace; }
```

### 4. Markdown Formatter

**Purpose:** Markdown output for GitHub/GitLab/documentation sites.

**Location:** `internal/format/markdown_formatter.go`

```go
package format

type MarkdownFormatter struct {
    config          *FormatterConfig
    richTextRenderer *MarkdownRichTextRenderer
}

func NewMarkdownFormatter(config *FormatterConfig) *MarkdownFormatter {
    return &MarkdownFormatter{
        config:          config,
        richTextRenderer: NewMarkdownRichTextRenderer(config),
    }
}

func (f *MarkdownFormatter) RenderHelp(model *model.HelpModel, w io.Writer) error {
    // # Makefile Help
    //
    // ## Usage
    // ```
    // make [<target>...] [<ENV_VAR>=<value>...]
    // ```
    //
    // ## File Documentation
    // ...
    //
    // ## Targets
    //
    // ### Category Name
    // - **build**: Summary text
    //   - Vars: `DEBUG`, `VERBOSE`
}

func (f *MarkdownFormatter) ContentType() string {
    return "text/markdown"
}

func (f *MarkdownFormatter) DefaultExtension() string {
    return ".md"
}

type MarkdownRichTextRenderer struct{}

func (r *MarkdownRichTextRenderer) RenderRichText(rt richtext.RichText) string {
    // Simply return the markdown form (preserve original formatting)
    return rt.Markdown()
}
```

**Key Features:**
- Clean, readable markdown structure
- Preserves original markdown formatting for bold/italic
- Code blocks for usage examples
- Nested lists for targets and variables

---

## CLI Changes

### New Flags

#### `--format` Flag

**Type:** String
**Valid values:** `make`, `mk`, `text`, `txt`, `html`, `markdown`, `md`
**Default:** `make`
**Aliases:** `mk` → `make`, `txt` → `text`, `md` → `markdown`

```go
// internal/cli/root.go (setupFlags function)

rootCmd.Flags().StringVar(
    &config.Format,
    "format",
    "make",
    "Output format (make, text, html, markdown)",
)
```

#### `--output` Flag

**Type:** String
**Valid values:** File path or `-` (stdout)
**Default:** Format-dependent (see table below)

| Format   | Default Output            |
|----------|---------------------------|
| make     | `./make/help.mk`          |
| text     | `./make-help.txt`         |
| html     | `./make-help.html`        |
| markdown | `./make-help.md`          |

**Special value:** `-` means stdout (replaces `--show-help` functionality)

```go
// internal/cli/root.go (setupFlags function)

rootCmd.Flags().StringVar(
    &config.Output,
    "output",
    "",
    "Output destination (file path or - for stdout). Default depends on format.",
)
```

### Removed Flags

#### `--show-help` (REMOVED)

**Replacement:** `--output -`

**Migration:**
- Old: `make-help --show-help`
- New: `make-help --format text --output -`

For backward compatibility during transition:
```go
// internal/cli/root.go (PreRunE validation)

if config.ShowHelp {
    // Show deprecation warning
    fmt.Fprintln(os.Stderr, "Warning: --show-help is deprecated. Use --output - instead.")
    config.Output = "-"
}
```

### Updated Config Struct

```go
// internal/cli/config.go

type Config struct {
    // ... existing fields ...

    // Format specifies the output format type.
    // Valid values: "make", "text", "html", "markdown" (and aliases)
    Format string

    // Output specifies the output destination.
    // "-" means stdout, otherwise it's a file path.
    // Empty string means use format-specific default.
    Output string

    // ShowHelp is DEPRECATED (kept for backward compatibility)
    // Use Output = "-" instead
    ShowHelp bool
}
```

### Flag Validation Logic

```go
// internal/cli/root.go (PreRunE)

PreRunE: func(cmd *cobra.Command, args []string) error {
    // Validate format
    validFormats := map[string]string{
        "make": "make", "mk": "make",
        "text": "text", "txt": "text",
        "html": "html",
        "markdown": "markdown", "md": "markdown",
    }
    normalizedFormat, ok := validFormats[config.Format]
    if !ok {
        return fmt.Errorf("invalid format: %s (valid: make, text, html, markdown)", config.Format)
    }
    config.Format = normalizedFormat

    // Resolve output destination
    if config.Output == "" {
        // Use format-specific default
        config.Output = getDefaultOutput(config.Format)
    }

    // Handle deprecated --show-help
    if config.ShowHelp {
        fmt.Fprintln(os.Stderr, "Warning: --show-help is deprecated. Use --output - instead.")
        config.Output = "-"
    }

    // --target requires --output - (stdout mode)
    if config.Target != "" && config.Output != "-" {
        return fmt.Errorf("--target requires --output - (stdout mode)")
    }

    // --remove-help validation (unchanged, but check format is 'make')
    if config.RemoveHelpTarget && config.Format != "make" {
        return fmt.Errorf("--remove-help only applies to 'make' format")
    }

    return nil
}

func getDefaultOutput(format string) string {
    switch format {
    case "make":
        return "./make/help.mk"
    case "text":
        return "./make-help.txt"
    case "html":
        return "./make-help.html"
    case "markdown":
        return "./make-help.md"
    default:
        return "./make-help.txt"
    }
}
```

### Updated Command Dispatch

```go
// internal/cli/root.go (RunE)

RunE: func(cmd *cobra.Command, args []string) error {
    // Resolve color mode
    config.UseColor = ResolveColorMode(config)

    // Dispatch based on mode
    if config.Lint {
        return runLint(config)
    } else if config.RemoveHelpTarget {
        return runRemoveHelpTarget(config)
    } else if config.Target != "" {
        return runDetailedHelp(config)
    } else if config.Output == "-" {
        // Stdout mode (replaces --show-help)
        return runShowHelp(config)
    } else {
        // File generation mode
        return runCreateHelpFile(config)
    }
}
```

### Example Usage

```bash
# Default behavior (unchanged): generate ./make/help.mk
make-help

# Generate HTML documentation
make-help --format html --output docs/make-help.html

# Display help in terminal (replaces --show-help)
make-help --format text --output -

# Generate markdown for README
make-help --format markdown --output README-targets.md

# Show detailed target help (unchanged behavior, but uses new flags)
make-help --target build --output -

# Preserve colors in text output to file
make-help --format text --color always --output help.txt
```

---

## File Structure

### New Files

```
internal/
├── richtext/
│   ├── types.go           # RichText, Segment, SegmentType definitions
│   ├── parser.go          # Parser for markdown inline formatting
│   ├── parser_test.go     # Parser tests
│   └── doc.go             # Package documentation
│
├── format/
│   ├── formatter.go       # Formatter interface, factory function
│   ├── make_formatter.go  # Make format implementation
│   ├── text_formatter.go  # Text format implementation
│   ├── html_formatter.go  # HTML format implementation
│   ├── markdown_formatter.go # Markdown format implementation
│   ├── richtext_renderer.go  # RichTextRenderer helpers
│   ├── color.go           # ColorScheme (existing, may need updates)
│   ├── make_formatter_test.go
│   ├── text_formatter_test.go
│   ├── html_formatter_test.go
│   ├── markdown_formatter_test.go
│   └── doc.go             # Package documentation
```

### Modified Files

```
internal/
├── cli/
│   ├── config.go          # Add Format, Output fields
│   ├── root.go            # Add --format, --output flags; update validation
│   ├── help.go            # Update to use Formatter interface
│   └── create_help_target.go # Update to use Formatter interface
│
├── summary/
│   └── extractor.go       # Return richtext.RichText instead of string
│
├── model/
│   └── types.go           # Change Target.Summary to richtext.RichText
│
└── target/
    └── generator.go       # Use MakeFormatter instead of Renderer
```

### Deprecated Files (to be removed in future)

```
internal/format/
└── renderer.go            # Will be removed after migration
                           # Functionality split into formatters
```

---

## Migration Path

### Phase 1: Foundation (Minimal Breaking Changes)

**Goal:** Introduce rich text model and formatters without breaking existing behavior.

**Steps:**

1. **Create `internal/richtext` package**
   - Implement `types.go` (RichText, Segment, SegmentType)
   - Implement `parser.go` (markdown parsing)
   - Write comprehensive tests

2. **Create new formatters** (alongside existing Renderer)
   - Implement `formatter.go` (interface + factory)
   - Implement `make_formatter.go` (port logic from renderer.go)
   - Implement `text_formatter.go` (port logic from renderer.go)
   - Write tests to ensure output matches existing Renderer

3. **Update model.Target.Summary type**
   - Change from `string` to `richtext.RichText`
   - Update summary.Extractor to return RichText
   - This is a breaking change to internal types but not to CLI

4. **Test compatibility**
   - Ensure `MakeFormatter` produces identical output to existing `Renderer.RenderForMakefile()`
   - Ensure `TextFormatter` produces identical output to existing `Renderer.Render()`

**Deliverables:**
- `internal/richtext` package (complete)
- `internal/format/formatter.go` (interface)
- `internal/format/make_formatter.go` (complete)
- `internal/format/text_formatter.go` (complete)
- All tests passing
- No CLI changes yet (existing flags still work)

### Phase 2: CLI Integration (Additive Changes)

**Goal:** Add new flags while maintaining backward compatibility.

**Steps:**

1. **Add new flags to `internal/cli/config.go`**
   - Add `Format` field (default: "make")
   - Add `Output` field (default: "")

2. **Add flag registration in `internal/cli/root.go`**
   - Add `--format` flag
   - Add `--output` flag
   - Keep `--show-help` flag (mark as deprecated in help text)

3. **Update flag validation**
   - Validate `--format` values
   - Resolve `--output` default based on format
   - Map `--show-help` to `--output -` (with deprecation warning)

4. **Update command dispatch in `internal/cli/root.go`**
   - Detect mode based on `--output` value (- = stdout mode)
   - Create formatter using factory
   - Pass formatter to runHelp/runCreateHelpFile

5. **Update orchestration functions**
   - `runHelp()` → use TextFormatter
   - `runCreateHelpFile()` → use formatter from factory
   - `runDetailedHelp()` → use TextFormatter

**Deliverables:**
- `--format` and `--output` flags working
- Backward compatibility maintained (`--show-help` still works with warning)
- Default behavior unchanged (generates ./make/help.mk)
- All existing tests passing

### Phase 3: Additional Formats (Non-Breaking Extensions)

**Goal:** Add HTML and Markdown formatters.

**Steps:**

1. **Implement `internal/format/html_formatter.go`**
   - RenderHelp with HTML structure
   - RenderDetailedTarget with HTML
   - Embedded CSS styling
   - Tests

2. **Implement `internal/format/markdown_formatter.go`**
   - RenderHelp with Markdown structure
   - RenderDetailedTarget with Markdown
   - Tests

3. **Update documentation**
   - README examples for new formats
   - Architecture docs

**Deliverables:**
- HTML format working
- Markdown format working
- Documentation updated

### Phase 4: Deprecation (Breaking Change Window)

**Goal:** Remove deprecated code and flags.

**Timeline:** Major version bump (e.g., v2.0.0)

**Steps:**

1. **Remove `--show-help` flag**
   - Remove from config.go
   - Remove from root.go
   - Update tests
   - Update documentation (migration guide)

2. **Remove `internal/format/renderer.go`**
   - Ensure all references updated to use formatters
   - Remove file
   - Update imports

3. **Update error messages**
   - If user tries to use `--show-help`, suggest `--output -`

**Deliverables:**
- Clean codebase without deprecated code
- Migration guide for v1 → v2 users

### Phase 5: Polish & Optimization

**Goal:** Refine and optimize based on real-world usage.

**Steps:**

1. **Performance testing**
   - Benchmark formatters with large Makefiles
   - Optimize hotspots

2. **User feedback integration**
   - Gather feedback on new formats
   - Refine HTML/CSS styling
   - Improve markdown structure

3. **Additional features**
   - JSON format (if requested)
   - Custom HTML templates (if requested)
   - Syntax highlighting in HTML (if requested)

---

## Implementation Priorities

### Must-Have (MVP)

1. **Rich text model** (`internal/richtext`)
   - Essential for preserving formatting across formats

2. **Formatter interface** (`internal/format/formatter.go`)
   - Core abstraction

3. **Make formatter** (`internal/format/make_formatter.go`)
   - Must maintain current default behavior

4. **Text formatter** (`internal/format/text_formatter.go`)
   - Needed for `--output -` (replaces `--show-help`)

5. **CLI flags** (`--format`, `--output`)
   - User-facing changes

### Should-Have (Next Priority)

6. **HTML formatter** (`internal/format/html_formatter.go`)
   - High value for documentation sites

7. **Markdown formatter** (`internal/format/markdown_formatter.go`)
   - High value for GitHub/GitLab

8. **Deprecation warnings**
   - Smooth migration path for users

### Nice-to-Have (Future)

9. **JSON format**
   - For programmatic consumption

10. **Custom HTML templates**
    - For branding/styling flexibility

11. **Syntax highlighting in HTML**
    - For code blocks in documentation

---

## Architecture Review

Now applying the seven-dimensional analysis:

### 1. SIMPLICITY

**Strengths:**
- Clear separation of concerns (parsing → model → formatting → output)
- Single responsibility for each formatter
- Factory pattern keeps creation logic centralized
- RichText model is simple (just segments with types)

**Concerns:**
- Introducing RichText adds complexity to the model
- Four formatters increases codebase size
- Migration path has multiple phases

**Recommendations:**
- Consider lazy initialization for formatters (only create when needed)
- Share common rendering logic in base classes to avoid duplication
- Document the "why" for RichText model to justify complexity

### 2. STANDARD APPROACHES

**Strengths:**
- Factory pattern for formatter creation (standard GoF pattern)
- Strategy pattern for format-specific rendering (standard GoF pattern)
- Interface-based design (Go best practice)
- io.Writer for output (standard Go idiom)

**Concerns:**
- Could use `encoding` package interfaces (e.g., `encoding.TextMarshaler`)

**Recommendations:**
- Consider implementing `encoding.TextMarshaler` interface for RichText
- Consider using `text/template` for HTML/Markdown formatters (more flexible)

### 3. USABILITY

**Strengths:**
- Clear flag names (`--format`, `--output`)
- Intuitive defaults (format-specific output paths)
- `-` for stdout is common Unix convention
- Backward compatibility with `--show-help` (with warning)

**Concerns:**
- Users may not discover new formats easily
- No `--list-formats` flag

**Recommendations:**
- Add `--list-formats` flag (lists available formats)
- Show deprecation warning prominently
- Document migration in README prominently

### 4. SECURITY

**Strengths:**
- HTML escaping in HTMLFormatter prevents XSS
- Makefile escaping prevents injection
- No user-controlled template execution
- io.Writer prevents path traversal in output

**Concerns:**
- If future features add template support, need sandboxing
- Rich text parsing could be vulnerable to ReDoS (regex denial of service)

**Recommendations:**
- Add input size limits for rich text parsing
- Use bounded regex patterns (avoid unbounded `*` and `+`)
- Review regex patterns for ReDoS vulnerabilities
- Add tests for malicious input (nested formatting, very long strings)

### 5. MAINTAINABILITY

**Strengths:**
- Clear package boundaries (`richtext`, `format`)
- Each formatter is self-contained
- Factory centralizes creation logic
- Tests for each formatter separately

**Concerns:**
- Four formatters to maintain
- Rich text parsing adds complexity
- Migration path spans multiple releases

**Recommendations:**
- Use table-driven tests for formatters (reduce duplication)
- Document formatter responsibilities clearly
- Create integration tests that cover all formats
- Use linters to enforce consistency across formatters

### 6. PERFORMANCE

**Strengths:**
- io.Writer allows streaming output (no large string buffers)
- Rich text parsing is one-pass
- Formatters don't buffer entire output

**Concerns:**
- Rich text parsing on every summary extraction
- Multiple regex matches per segment
- String concatenation in rich text rendering

**Recommendations:**
- Use `strings.Builder` consistently (avoid `+` concatenation)
- Pre-compile all regex patterns (done in parser constructor)
- Consider caching parsed rich text if summaries are reused
- Benchmark with large Makefiles (100+ targets)

### 7. ROBUSTNESS

**Strengths:**
- Error handling for unknown formats
- Validation of flag combinations
- Fallback to plain text if formatting fails
- Atomic file writes (existing behavior)

**Concerns:**
- What if rich text parsing fails?
- What if formatter returns partial output?
- What if output file is not writable?

**Recommendations:**
- Add error recovery in rich text parser (return plain text on error)
- Add write verification after file creation
- Add dry-run support for all formats (show what would be generated)
- Add logging for format conversion failures

---

## Conclusion

This design provides a robust, extensible architecture for multiple output formats in `make-help`. The key innovations are:

1. **Rich text preservation** - Formatting is preserved/converted instead of stripped
2. **Clean abstraction** - Formatter interface allows easy addition of new formats
3. **Backward compatibility** - Existing behavior maintained through migration phases
4. **User-friendly** - Clear flags and intuitive defaults

The migration path is conservative, introducing changes incrementally to minimize risk. The architecture follows Go best practices and standard design patterns, making it maintainable and extensible.

**Next Steps:**
1. Review this design with stakeholders
2. Proceed with Phase 1 implementation (rich text model + core formatters)
3. Gather early feedback before committing to full migration
