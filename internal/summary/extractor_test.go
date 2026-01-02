package summary

import (
	"testing"
)

func TestExtract(t *testing.T) {
	tests := []struct {
		name     string
		docs     []string
		expected string
	}{
		// Standard cases
		{
			name:     "simple sentence with period",
			docs:     []string{"Build the project. Run tests."},
			expected: "Build the project.",
		},
		{
			name:     "sentence with exclamation",
			docs:     []string{"This is a test!"},
			expected: "This is a test!",
		},
		{
			name:     "sentence with question mark",
			docs:     []string{"Is this working? Yes."},
			expected: "Is this working?",
		},
		{
			name:     "multiple sentences",
			docs:     []string{"First sentence. Second sentence. Third."},
			expected: "First sentence.",
		},

		// Ellipsis handling
		{
			name:     "ellipsis not sentence boundary",
			docs:     []string{"Wait for it... then proceed. Done."},
			expected: "Wait for it... then proceed.",
		},
		{
			name:     "multiple ellipsis",
			docs:     []string{"Loading... processing... complete. Next step."},
			expected: "Loading... processing... complete.",
		},

		// IP address handling
		{
			name:     "IP address with period",
			docs:     []string{"Connect to 127.0.0.1. Then test."},
			expected: "Connect to 127.0.0.1.",
		},
		{
			name:     "full IP address",
			docs:     []string{"Server at 192.168.1.1. Ready to deploy."},
			expected: "Server at 192.168.1.1.",
		},
		{
			name:     "version number with dots",
			docs:     []string{"Version 1.2.3. Released today."},
			expected: "Version 1.2.3.",
		},

		// Markdown header stripping
		{
			name:     "markdown header single hash",
			docs:     []string{"# Header", "Content."},
			expected: "Header Content.",
		},
		{
			name:     "markdown header multiple hashes",
			docs:     []string{"## Subsection", "This is content."},
			expected: "Subsection This is content.",
		},
		{
			name:     "markdown header deeply nested",
			docs:     []string{"#### Deep Header", "Text follows."},
			expected: "Deep Header Text follows.",
		},

		// Markdown formatting stripping
		{
			name:     "bold with asterisks",
			docs:     []string{"**Bold** text."},
			expected: "Bold text.",
		},
		{
			name:     "italic with asterisks",
			docs:     []string{"*Italic* text."},
			expected: "Italic text.",
		},
		{
			name:     "bold and italic combined",
			docs:     []string{"**Bold** and *italic* text."},
			expected: "Bold and italic text.",
		},
		{
			name:     "bold with underscores",
			docs:     []string{"__Bold__ text."},
			expected: "Bold text.",
		},
		{
			name:     "italic with underscores",
			docs:     []string{"_Italic_ text."},
			expected: "Italic text.",
		},
		{
			name:     "inline code",
			docs:     []string{"`code` formatting."},
			expected: "code formatting.",
		},
		{
			name:     "markdown link",
			docs:     []string{"[link text](http://example.com) here."},
			expected: "link text here.",
		},
		{
			name:     "complex markdown mix",
			docs:     []string{"**Bold**, *italic*, `code`, and [link](url)."},
			expected: "Bold, italic, code, and link.",
		},

		// HTML tag stripping
		{
			name:     "simple HTML tag",
			docs:     []string{"<b>bold</b> text."},
			expected: "bold text.",
		},
		{
			name:     "multiple HTML tags",
			docs:     []string{"<p>Paragraph with <em>emphasis</em>.</p>"},
			expected: "Paragraph with emphasis.",
		},
		{
			name:     "self-closing HTML tag",
			docs:     []string{"Line break<br/>here."},
			expected: "Line breakhere.",
		},

		// No sentence terminator
		{
			name:     "no terminator",
			docs:     []string{"No terminator here"},
			expected: "No terminator here",
		},
		{
			name:     "only ellipsis",
			docs:     []string{"Just dots..."},
			expected: "Just dots...",
		},

		// Empty input
		{
			name:     "empty slice",
			docs:     []string{},
			expected: "",
		},
		{
			name:     "empty string",
			docs:     []string{""},
			expected: "",
		},
		{
			name:     "multiple empty strings",
			docs:     []string{"", "", ""},
			expected: "",
		},

		// Whitespace normalization
		{
			name:     "multiple spaces",
			docs:     []string{"Multiple   spaces."},
			expected: "Multiple spaces.",
		},
		{
			name:     "line breaks",
			docs:     []string{"Line\nbreak."},
			expected: "Line break.",
		},
		{
			name:     "multiple line breaks",
			docs:     []string{"Line\n\n\nbreak."},
			expected: "Line break.",
		},
		{
			name:     "tabs and spaces",
			docs:     []string{"Tab\t  space."},
			expected: "Tab space.",
		},
		{
			name:     "leading and trailing whitespace",
			docs:     []string{"  Leading and trailing.  "},
			expected: "Leading and trailing.",
		},

		// Multiple documentation lines
		{
			name:     "multiple doc lines",
			docs:     []string{"First line.", "Second line."},
			expected: "First line.",
		},
		{
			name:     "joined doc lines forming sentence",
			docs:     []string{"This is ", "a sentence."},
			expected: "This is a sentence.",
		},

		// Real-world examples
		{
			name:     "makefile target description",
			docs:     []string{"## Build", "Builds the project using Go compiler. Outputs to bin/ directory."},
			expected: "Build Builds the project using Go compiler.",
		},
		{
			name:     "complex real-world example",
			docs:     []string{
				"**Deploys** the application to production server at `192.168.1.100`.",
				"See [docs](http://example.com) for details.",
			},
			expected: "Deploys the application to production server at 192.168.1.100.",
		},

		// Edge cases with mixed content
		// Note: Abbreviations followed by space are treated as sentence boundaries
		// This is correct behavior per the regex spec - periods followed by space end sentences
		{
			name:     "abbreviation like e.g.",
			docs:     []string{"Use e.g. this example. Another sentence."},
			expected: "Use e.g.",
		},
		{
			name:     "abbreviation like i.e.",
			docs:     []string{"This means i.e. that is. Continue."},
			expected: "This means i.e.",
		},
		{
			name:     "file extension",
			docs:     []string{"Edit the config.yaml. Then restart."},
			expected: "Edit the config.yaml.",
		},

		// Sentence at end of string
		{
			name:     "sentence at end with no trailing space",
			docs:     []string{"Complete sentence."},
			expected: "Complete sentence.",
		},
		{
			name:     "exclamation at end",
			docs:     []string{"Done!"},
			expected: "Done!",
		},
		{
			name:     "question at end",
			docs:     []string{"Ready?"},
			expected: "Ready?",
		},
	}

	extractor := NewExtractor()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.Extract(tt.docs)
			resultText := result.PlainText()
			if resultText != tt.expected {
				t.Errorf("Extract().PlainText() = %q, want %q", resultText, tt.expected)
			}
		})
	}
}

func TestStripMarkdownHeaders(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single hash",
			input:    "# Header",
			expected: "Header",
		},
		{
			name:     "multiple hashes",
			input:    "### Subsection",
			expected: "Subsection",
		},
		{
			name:     "header with text after",
			input:    "## Header\nContent",
			expected: "Header\nContent",
		},
		{
			name:     "multiple headers",
			input:    "# Title\n## Subtitle\nContent",
			expected: "Title\nSubtitle\nContent",
		},
		{
			name:     "no headers",
			input:    "Just plain text",
			expected: "Just plain text",
		},
	}

	extractor := NewExtractor()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.stripMarkdownHeaders(tt.input)
			if result != tt.expected {
				t.Errorf("stripMarkdownHeaders() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestStripMarkdownFormatting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "bold asterisks",
			input:    "**bold**",
			expected: "bold",
		},
		{
			name:     "italic asterisks",
			input:    "*italic*",
			expected: "italic",
		},
		{
			name:     "bold underscores",
			input:    "__bold__",
			expected: "bold",
		},
		{
			name:     "italic underscores",
			input:    "_italic_",
			expected: "italic",
		},
		{
			name:     "inline code",
			input:    "`code`",
			expected: "code",
		},
		{
			name:     "link",
			input:    "[text](url)",
			expected: "text",
		},
		{
			name:     "mixed formatting",
			input:    "**bold** and *italic* and `code`",
			expected: "bold and italic and code",
		},
		{
			name:     "nested formatting - regex limitations",
			input:    "**bold *italic* bold**",
			// Regex can't perfectly handle nested markdown - this is expected behavior
			// The ** pattern matches first, leaving the inner * characters
			expected: "*bold italic bold*",
		},
	}

	extractor := NewExtractor()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.stripMarkdownFormatting(tt.input)
			if result != tt.expected {
				t.Errorf("stripMarkdownFormatting() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestStripHTMLTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple tag",
			input:    "<b>bold</b>",
			expected: "bold",
		},
		{
			name:     "multiple tags",
			input:    "<p>Text with <em>emphasis</em></p>",
			expected: "Text with emphasis",
		},
		{
			name:     "self-closing tag",
			input:    "Text<br/>more",
			expected: "Textmore",
		},
		{
			name:     "no tags",
			input:    "Plain text",
			expected: "Plain text",
		},
	}

	extractor := NewExtractor()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.stripHTMLTags(tt.input)
			if result != tt.expected {
				t.Errorf("stripHTMLTags() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNormalizeWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "multiple spaces",
			input:    "Multiple   spaces",
			expected: "Multiple spaces",
		},
		{
			name:     "newlines",
			input:    "Line\nbreak",
			expected: "Line break",
		},
		{
			name:     "multiple newlines",
			input:    "Line\n\n\nbreak",
			expected: "Line break",
		},
		{
			name:     "tabs",
			input:    "Tab\there",
			expected: "Tab here",
		},
		{
			name:     "leading whitespace",
			input:    "  Leading",
			expected: "Leading",
		},
		{
			name:     "trailing whitespace",
			input:    "Trailing  ",
			expected: "Trailing",
		},
		{
			name:     "mixed whitespace",
			input:    "  Multiple\n\nlines\t  with   spaces  ",
			expected: "Multiple lines with spaces",
		},
	}

	extractor := NewExtractor()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.normalizeWhitespace(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeWhitespace() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractFirstSentence(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "period terminator",
			input:    "First sentence. Second sentence.",
			expected: "First sentence.",
		},
		{
			name:     "exclamation terminator",
			input:    "Excited! More text.",
			expected: "Excited!",
		},
		{
			name:     "question terminator",
			input:    "Question? Answer.",
			expected: "Question?",
		},
		{
			name:     "ellipsis not boundary",
			input:    "Wait... continue. Done.",
			expected: "Wait... continue.",
		},
		{
			name:     "IP address not boundary",
			input:    "Server 127.0.0.1. Next.",
			expected: "Server 127.0.0.1.",
		},
		{
			name:     "version number not boundary",
			input:    "Version 2.1.0. Released.",
			expected: "Version 2.1.0.",
		},
		{
			name:     "no terminator",
			input:    "No ending",
			expected: "No ending",
		},
		{
			name:     "sentence at end",
			input:    "Complete.",
			expected: "Complete.",
		},
		{
			name:     "sentence with space after",
			input:    "Complete. ",
			expected: "Complete.",
		},
	}

	extractor := NewExtractor()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.extractFirstSentence(tt.input)
			if result != tt.expected {
				t.Errorf("extractFirstSentence() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestNewExtractor verifies that all regex patterns compile successfully
func TestNewExtractor(t *testing.T) {
	extractor := NewExtractor()

	if extractor.sentenceRegex == nil {
		t.Error("sentenceRegex was not initialized")
	}
	if extractor.headerRegex == nil {
		t.Error("headerRegex was not initialized")
	}
	if extractor.boldRegex == nil {
		t.Error("boldRegex was not initialized")
	}
	if extractor.italicRegex == nil {
		t.Error("italicRegex was not initialized")
	}
	if extractor.boldUnderRegex == nil {
		t.Error("boldUnderRegex was not initialized")
	}
	if extractor.italicUnderRegex == nil {
		t.Error("italicUnderRegex was not initialized")
	}
	if extractor.codeRegex == nil {
		t.Error("codeRegex was not initialized")
	}
	if extractor.linkRegex == nil {
		t.Error("linkRegex was not initialized")
	}
	if extractor.htmlTagRegex == nil {
		t.Error("htmlTagRegex was not initialized")
	}
	if extractor.whitespaceRegex == nil {
		t.Error("whitespaceRegex was not initialized")
	}
}

// BenchmarkExtract measures performance of the Extract method
func BenchmarkExtract(b *testing.B) {
	extractor := NewExtractor()
	docs := []string{
		"**Deploys** the application to production server at `192.168.1.100`.",
		"See [documentation](http://example.com) for more details.",
		"This process typically takes 5-10 minutes.",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = extractor.Extract(docs)
	}
}

// BenchmarkNewExtractor measures the cost of creating a new extractor
func BenchmarkNewExtractor(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewExtractor()
	}
}
