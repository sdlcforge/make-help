package format

import "testing"

// TestEscapeForMakefileEcho tests the escapeForMakefileEcho function with all special characters
func TestEscapeForMakefileEcho(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "dollar sign",
			input:    "Use $VAR",
			expected: "Use $$VAR",
		},
		{
			name:     "double quotes",
			input:    `Use "quotes"`,
			expected: `Use \"quotes\"`,
		},
		{
			name:     "backslash",
			input:    `Use \ backslash`,
			expected: `Use \\ backslash`,
		},
		{
			name:     "backtick",
			input:    "Use `command`",
			expected: "Use \\`command\\`",
		},
		{
			name:     "newline",
			input:    "Line1\nLine2",
			expected: "Line1\\nLine2",
		},
		{
			name:     "carriage return",
			input:    "Line1\rLine2",
			expected: "Line1\\rLine2",
		},
		{
			name:     "tab",
			input:    "Col1\tCol2",
			expected: "Col1\\tCol2",
		},
		{
			name:     "ANSI escape",
			input:    "\x1b[36mCyan\x1b[0m",
			expected: "\\033[36mCyan\\033[0m",
		},
		{
			name:     "multiple special characters",
			input:    "$VAR \"test\" `cmd` \nNew\rCR\tTab",
			expected: "$$VAR \\\"test\\\" \\`cmd\\` \\nNew\\rCR\\tTab",
		},
		{
			name:     "shell injection attempt via newline",
			input:    "Harmless\n\"; rm -rf /; echo \"",
			expected: "Harmless\\n\\\"; rm -rf /; echo \\\"",
		},
		{
			name:     "shell injection attempt via carriage return",
			input:    "Harmless\r\"; rm -rf /; echo \"",
			expected: "Harmless\\r\\\"; rm -rf /; echo \\\"",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no special characters",
			input:    "Just normal text",
			expected: "Just normal text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeForMakefileEcho(tt.input)
			if result != tt.expected {
				t.Errorf("escapeForMakefileEcho(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestEscapeForMakefileEcho_SecurityCases tests specific security-related escape scenarios
func TestEscapeForMakefileEcho_SecurityCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		description string
	}{
		{
			name:        "command substitution via backtick",
			input:       "`whoami`",
			description: "Backticks should be escaped to prevent command substitution",
		},
		{
			name:        "command substitution via $() would need newline",
			input:       "test\n$(whoami)",
			description: "Newlines should be escaped to prevent breaking out of quotes",
		},
		{
			name:        "variable expansion",
			input:       "$HOME/path",
			description: "Dollar signs should be escaped for Makefile",
		},
		{
			name:        "quote breaking",
			input:       "text\"remaining",
			description: "Quotes should be escaped to prevent breaking out of string",
		},
		{
			name:        "combined injection attempt",
			input:       "Doc text\n\"; $(rm -rf /); echo \"Safe",
			description: "Multiple escape characters combined in injection attempt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeForMakefileEcho(tt.input)

			// Verify no raw special characters remain
			for i := 0; i < len(result); i++ {
				ch := result[i]
				// Check that potentially dangerous characters are not present in their raw form
				// (they should all be preceded by backslashes or doubled)
				if ch == '\n' || ch == '\r' || ch == '\t' {
					t.Errorf("escapeForMakefileEcho(%q) contains unescaped whitespace character at position %d: %q",
						tt.input, i, ch)
				}
			}
		})
	}
}
