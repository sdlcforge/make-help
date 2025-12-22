package richtext

import (
	"strings"
	"testing"
)

func TestParser_Parse_PlainText(t *testing.T) {
	parser := NewParser()
	tests := []struct {
		name     string
		input    string
		expected RichText
	}{
		{
			name:  "empty string",
			input: "",
			expected: RichText{},
		},
		{
			name:  "plain text",
			input: "This is plain text",
			expected: RichText{
				{Type: SegmentPlain, Content: "This is plain text"},
			},
		},
		{
			name:  "text with spaces",
			input: "  spaces  around  ",
			expected: RichText{
				{Type: SegmentPlain, Content: "  spaces  around  "},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)
			if !richTextEqual(result, tt.expected) {
				t.Errorf("Parse() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestParser_Parse_Bold(t *testing.T) {
	parser := NewParser()
	tests := []struct {
		name     string
		input    string
		expected RichText
	}{
		{
			name:  "bold with asterisks",
			input: "**bold**",
			expected: RichText{
				{Type: SegmentBold, Content: "bold"},
			},
		},
		{
			name:  "bold with underscores",
			input: "__bold__",
			expected: RichText{
				{Type: SegmentBold, Content: "bold"},
			},
		},
		{
			name:  "bold in middle",
			input: "text **bold** text",
			expected: RichText{
				{Type: SegmentPlain, Content: "text "},
				{Type: SegmentBold, Content: "bold"},
				{Type: SegmentPlain, Content: " text"},
			},
		},
		{
			name:  "multiple bold",
			input: "**first** and **second**",
			expected: RichText{
				{Type: SegmentBold, Content: "first"},
				{Type: SegmentPlain, Content: " and "},
				{Type: SegmentBold, Content: "second"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)
			if !richTextEqual(result, tt.expected) {
				t.Errorf("Parse() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestParser_Parse_Italic(t *testing.T) {
	parser := NewParser()
	tests := []struct {
		name     string
		input    string
		expected RichText
	}{
		{
			name:  "italic with asterisks",
			input: "*italic*",
			expected: RichText{
				{Type: SegmentItalic, Content: "italic"},
			},
		},
		{
			name:  "italic with underscores",
			input: "_italic_",
			expected: RichText{
				{Type: SegmentItalic, Content: "italic"},
			},
		},
		{
			name:  "italic in middle",
			input: "text *italic* text",
			expected: RichText{
				{Type: SegmentPlain, Content: "text "},
				{Type: SegmentItalic, Content: "italic"},
				{Type: SegmentPlain, Content: " text"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)
			if !richTextEqual(result, tt.expected) {
				t.Errorf("Parse() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestParser_Parse_Mixed(t *testing.T) {
	parser := NewParser()
	tests := []struct {
		name     string
		input    string
		expected RichText
	}{
		{
			name:  "bold and italic",
			input: "**bold** and *italic*",
			expected: RichText{
				{Type: SegmentBold, Content: "bold"},
				{Type: SegmentPlain, Content: " and "},
				{Type: SegmentItalic, Content: "italic"},
			},
		},
		{
			name:  "italic and bold",
			input: "*italic* and **bold**",
			expected: RichText{
				{Type: SegmentItalic, Content: "italic"},
				{Type: SegmentPlain, Content: " and "},
				{Type: SegmentBold, Content: "bold"},
			},
		},
		{
			name:  "multiple formatting types",
			input: "Start **bold** middle *italic* end",
			expected: RichText{
				{Type: SegmentPlain, Content: "Start "},
				{Type: SegmentBold, Content: "bold"},
				{Type: SegmentPlain, Content: " middle "},
				{Type: SegmentItalic, Content: "italic"},
				{Type: SegmentPlain, Content: " end"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)
			if !richTextEqual(result, tt.expected) {
				t.Errorf("Parse() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestParser_Parse_Code(t *testing.T) {
	parser := NewParser()
	tests := []struct {
		name     string
		input    string
		expected RichText
	}{
		{
			name:  "code inline",
			input: "`code`",
			expected: RichText{
				{Type: SegmentCode, Content: "code"},
			},
		},
		{
			name:  "code in sentence",
			input: "Use `make help` to see options",
			expected: RichText{
				{Type: SegmentPlain, Content: "Use "},
				{Type: SegmentCode, Content: "make help"},
				{Type: SegmentPlain, Content: " to see options"},
			},
		},
		{
			name:  "multiple code segments",
			input: "`first` and `second`",
			expected: RichText{
				{Type: SegmentCode, Content: "first"},
				{Type: SegmentPlain, Content: " and "},
				{Type: SegmentCode, Content: "second"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)
			if !richTextEqual(result, tt.expected) {
				t.Errorf("Parse() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestParser_Parse_Links(t *testing.T) {
	parser := NewParser()
	tests := []struct {
		name     string
		input    string
		expected RichText
	}{
		{
			name:  "simple link",
			input: "[text](url)",
			expected: RichText{
				{Type: SegmentLink, Content: "text", URL: "url"},
			},
		},
		{
			name:  "link with full URL",
			input: "[GitHub](https://github.com)",
			expected: RichText{
				{Type: SegmentLink, Content: "GitHub", URL: "https://github.com"},
			},
		},
		{
			name:  "link in sentence",
			input: "Visit [our site](https://example.com) for more",
			expected: RichText{
				{Type: SegmentPlain, Content: "Visit "},
				{Type: SegmentLink, Content: "our site", URL: "https://example.com"},
				{Type: SegmentPlain, Content: " for more"},
			},
		},
		{
			name:  "multiple links",
			input: "[first](url1) and [second](url2)",
			expected: RichText{
				{Type: SegmentLink, Content: "first", URL: "url1"},
				{Type: SegmentPlain, Content: " and "},
				{Type: SegmentLink, Content: "second", URL: "url2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)
			if !richTextEqual(result, tt.expected) {
				t.Errorf("Parse() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestParser_Parse_Nested(t *testing.T) {
	parser := NewParser()
	tests := []struct {
		name     string
		input    string
		expected RichText
	}{
		{
			name:  "bold takes precedence over italic",
			input: "**bold *nested italic* bold**",
			expected: RichText{
				{Type: SegmentBold, Content: "bold *nested italic* bold"},
			},
		},
		{
			name:  "code takes precedence over bold",
			input: "`code **not bold** code`",
			expected: RichText{
				{Type: SegmentCode, Content: "code **not bold** code"},
			},
		},
		{
			name:  "link takes precedence over all",
			input: "[**bold** `code` *italic*](url)",
			expected: RichText{
				{Type: SegmentLink, Content: "**bold** `code` *italic*", URL: "url"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)
			if !richTextEqual(result, tt.expected) {
				t.Errorf("Parse() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestParser_Parse_EdgeCases(t *testing.T) {
	parser := NewParser()
	tests := []struct {
		name     string
		input    string
		expected RichText
	}{
		{
			name:  "unmatched bold start",
			input: "**bold without end",
			expected: RichText{
				{Type: SegmentPlain, Content: "**bold without end"},
			},
		},
		{
			name:  "unmatched italic start",
			input: "*italic without end",
			expected: RichText{
				{Type: SegmentPlain, Content: "*italic without end"},
			},
		},
		{
			name:  "unmatched code start",
			input: "`code without end",
			expected: RichText{
				{Type: SegmentPlain, Content: "`code without end"},
			},
		},
		{
			name:  "unmatched link start",
			input: "[link without url",
			expected: RichText{
				{Type: SegmentPlain, Content: "[link without url"},
			},
		},
		{
			name:  "adjacent formatting",
			input: "**bold***italic*",
			expected: RichText{
				{Type: SegmentBold, Content: "bold"},
				{Type: SegmentItalic, Content: "italic"},
			},
		},
		{
			name:  "empty bold",
			input: "****",
			expected: RichText{
				{Type: SegmentPlain, Content: "****"},
			},
		},
		{
			name:  "empty code",
			input: "``",
			expected: RichText{
				{Type: SegmentPlain, Content: "``"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.Parse(tt.input)
			if !richTextEqual(result, tt.expected) {
				t.Errorf("Parse() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestParser_Parse_Security(t *testing.T) {
	parser := NewParser()

	t.Run("strips ANSI escape codes", func(t *testing.T) {
		input := "\x1b[31mred text\x1b[0m"
		result := parser.Parse(input)
		expected := RichText{
			{Type: SegmentPlain, Content: "red text"},
		}
		if !richTextEqual(result, expected) {
			t.Errorf("Parse() = %+v, want %+v", result, expected)
		}
	})

	t.Run("handles max input length", func(t *testing.T) {
		// Create input larger than MaxInputLength
		input := strings.Repeat("a", MaxInputLength+1000)
		result := parser.Parse(input)

		// Should truncate to MaxInputLength
		if len(result) != 1 || result[0].Type != SegmentPlain {
			t.Errorf("Expected single plain text segment")
		}
		if len(result[0].Content) > MaxInputLength {
			t.Errorf("Content length %d exceeds max %d", len(result[0].Content), MaxInputLength)
		}
	})

	t.Run("bounded segment length prevents ReDoS", func(t *testing.T) {
		// Create input with very long bold segment (over MaxSegmentLength)
		input := "**" + strings.Repeat("a", MaxSegmentLength+100) + "**"
		result := parser.Parse(input)

		// Should not match as bold (over segment limit)
		if len(result) != 1 || result[0].Type != SegmentPlain {
			t.Errorf("Expected plain text for oversized segment, got %+v", result)
		}
	})
}

func TestRichText_PlainText(t *testing.T) {
	tests := []struct {
		name     string
		richText RichText
		expected string
	}{
		{
			name:     "empty",
			richText: RichText{},
			expected: "",
		},
		{
			name: "plain only",
			richText: RichText{
				{Type: SegmentPlain, Content: "plain"},
			},
			expected: "plain",
		},
		{
			name: "bold strips formatting",
			richText: RichText{
				{Type: SegmentBold, Content: "bold"},
			},
			expected: "bold",
		},
		{
			name: "mixed formatting strips all",
			richText: RichText{
				{Type: SegmentPlain, Content: "text "},
				{Type: SegmentBold, Content: "bold"},
				{Type: SegmentPlain, Content: " "},
				{Type: SegmentItalic, Content: "italic"},
			},
			expected: "text bold italic",
		},
		{
			name: "link strips formatting but keeps content",
			richText: RichText{
				{Type: SegmentLink, Content: "link text", URL: "https://example.com"},
			},
			expected: "link text",
		},
		{
			name: "code strips backticks",
			richText: RichText{
				{Type: SegmentCode, Content: "code"},
			},
			expected: "code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.richText.PlainText()
			if result != tt.expected {
				t.Errorf("PlainText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestRichText_Markdown(t *testing.T) {
	tests := []struct {
		name     string
		richText RichText
		expected string
	}{
		{
			name:     "empty",
			richText: RichText{},
			expected: "",
		},
		{
			name: "plain only",
			richText: RichText{
				{Type: SegmentPlain, Content: "plain"},
			},
			expected: "plain",
		},
		{
			name: "bold preserves markers",
			richText: RichText{
				{Type: SegmentBold, Content: "bold"},
			},
			expected: "**bold**",
		},
		{
			name: "italic preserves markers",
			richText: RichText{
				{Type: SegmentItalic, Content: "italic"},
			},
			expected: "*italic*",
		},
		{
			name: "code preserves backticks",
			richText: RichText{
				{Type: SegmentCode, Content: "code"},
			},
			expected: "`code`",
		},
		{
			name: "link preserves format",
			richText: RichText{
				{Type: SegmentLink, Content: "text", URL: "url"},
			},
			expected: "[text](url)",
		},
		{
			name: "mixed preserves all",
			richText: RichText{
				{Type: SegmentPlain, Content: "text "},
				{Type: SegmentBold, Content: "bold"},
				{Type: SegmentPlain, Content: " "},
				{Type: SegmentItalic, Content: "italic"},
				{Type: SegmentPlain, Content: " "},
				{Type: SegmentCode, Content: "code"},
			},
			expected: "text **bold** *italic* `code`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.richText.Markdown()
			if result != tt.expected {
				t.Errorf("Markdown() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestRichText_String(t *testing.T) {
	richText := RichText{
		{Type: SegmentBold, Content: "bold"},
	}

	// String() should be alias for Markdown()
	if richText.String() != richText.Markdown() {
		t.Errorf("String() != Markdown()")
	}
}

func TestParser_RoundTrip(t *testing.T) {
	parser := NewParser()
	tests := []string{
		"plain text",
		"**bold**",
		"*italic*",
		"`code`",
		"[link](url)",
		"**bold** and *italic*",
		"text `code` more",
		"[link](url) with **bold**",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			parsed := parser.Parse(input)
			markdown := parsed.Markdown()

			// Re-parse the markdown output
			reparsed := parser.Parse(markdown)

			// Should get the same segments
			if !richTextEqual(parsed, reparsed) {
				t.Errorf("Round trip failed:\nOriginal: %+v\nReparsed: %+v", parsed, reparsed)
			}
		})
	}
}

// Helper function to compare RichText slices
func richTextEqual(a, b RichText) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Type != b[i].Type || a[i].Content != b[i].Content || a[i].URL != b[i].URL {
			return false
		}
	}
	return true
}
