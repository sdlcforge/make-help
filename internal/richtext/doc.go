// Package richtext provides parsing and representation of markdown-formatted text.
//
// This package parses inline markdown formatting (bold, italic, code, links) into
// a structured representation that preserves semantic information about emphasis
// and importance. This allows formatters to render the same content differently
// based on output context (terminal, file, etc.) while maintaining the original
// intent.
//
// # Basic Usage
//
//	parser := richtext.NewParser()
//	text := parser.Parse("This is **bold** and *italic* text")
//
//	// Get plain text without formatting
//	plain := text.PlainText() // "This is bold and italic text"
//
//	// Get markdown with formatting preserved
//	markdown := text.Markdown() // "This is **bold** and *italic* text"
//
// # Supported Formatting
//
// The parser recognizes the following markdown inline formatting:
//
//   - Bold: **text** or __text__
//   - Italic: *text* or _text_
//   - Code: `text`
//   - Links: [text](url)
//
// # Precedence Rules
//
// When formatting markers overlap or nest, the parser applies precedence rules
// to handle them correctly:
//
//  1. Links (highest precedence)
//  2. Code
//  3. Bold
//  4. Italic (lowest precedence)
//
// For example, `**bold `code` bold**` will parse as bold text containing the
// literal string "`code`", because code has higher precedence and would have
// already been matched if the backticks were balanced.
//
// # Security
//
// The parser includes several security features:
//
//   - Input length limit: Maximum 10KB per input text
//   - Segment length limit: Maximum 2000 characters per formatted segment
//   - Bounded regex patterns: Prevents ReDoS (Regular Expression Denial of Service)
//   - ANSI stripping: Removes ANSI escape codes to prevent terminal injection
//   - Error recovery: On parse error, returns plain text instead of failing
//
// These limits are enforced automatically and do not require caller action.
//
// # Design Philosophy
//
// This package maintains a separation between:
//   - Content: The actual text being displayed
//   - Formatting: How that text should be emphasized
//   - Rendering: How formatting is presented (handled by other packages)
//
// This allows the same rich text to be rendered with ANSI colors in a terminal,
// as markdown in a file, or as plain text in a context where formatting is not
// supported.
//
// # Rendering Methods
//
// RichText provides multiple methods for rendering the same content in different
// contexts:
//
//   - PlainText(): Strips all markdown formatting, returning plain text only.
//     Used by terminal and file formatters where formatting would be distracting.
//   - Markdown(): Preserves the original markdown formatting (e.g., **bold**).
//     Used by markdown and HTML formatters that can interpret the formatting.
//
// Formatters choose which method to call based on their output context. For
// example, the TextFormatter uses PlainText() for clean terminal output, while
// the MarkdownFormatter uses Markdown() to preserve formatting. The HTMLFormatter
// uses a custom renderRichText() method that converts markdown segments directly
// to HTML elements.
package richtext
