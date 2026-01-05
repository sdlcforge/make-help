package format

import "strings"

// escapeForMakefileEcho escapes a string for use in Makefile @echo statements.
// This function is shared between renderer.go and make_formatter.go to ensure
// consistent escaping behavior across both the legacy renderer and the new formatter.
//
// Special characters that need escaping:
//   - $ → $$ (Makefile variable escape)
//   - " → \" (shell quote escape)
//   - \ → \\ (shell backslash escape, except for ANSI codes)
//   - ` → \` (shell backtick escape to prevent command substitution)
//   - \n → \\n (newline escape to prevent breaking out of quotes)
//   - \r → \\r (carriage return escape to prevent breaking out of quotes)
//   - \t → \\t (tab escape to prevent breaking out of quotes)
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
		case '\n':
			// Escape newline to prevent breaking out of quotes
			result.WriteString("\\n")
		case '\r':
			// Escape carriage return to prevent breaking out of quotes
			result.WriteString("\\r")
		case '\t':
			// Escape tab to prevent breaking out of quotes
			result.WriteString("\\t")
		case '\x1b':
			// Convert ANSI escape character to literal \033 for echo
			result.WriteString("\\033")
		default:
			result.WriteByte(ch)
		}
	}
	return result.String()
}
