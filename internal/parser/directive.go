package parser

import (
	"strings"
)

// IsDocumentationLine checks if a line is a documentation line.
// Matches lines starting with "## " or exactly "##" (empty doc line for blank paragraphs).
func IsDocumentationLine(line string) bool {
	return strings.HasPrefix(line, "## ") || line == "##"
}

// IsTargetLine checks if a line is a target definition.
// A target line contains ":" and is not indented (not a recipe line).
func IsTargetLine(line string) bool {
	// Recipe lines start with tab or space - skip those
	if strings.HasPrefix(line, "\t") || strings.HasPrefix(line, " ") {
		return false
	}

	// Comment lines start with # - skip those
	if strings.HasPrefix(line, "#") {
		return false
	}

	// Must contain a colon to be a target definition
	return strings.Contains(line, ":")
}

// ExtractTargetName extracts the target name from a target definition line.
//
// Handles the following cases:
//   - Simple targets: "build:" -> "build"
//   - Grouped target operator: "build&:" -> "build"
//   - Multiple targets: "foo bar baz:" -> "foo"
//   - Variable targets: "$(VAR):" -> "$(VAR)"
//
// Returns empty string if the line is not a valid target definition.
// Variable assignments (":=", "?=", "+=", "!=") are NOT targets and return empty string.
func ExtractTargetName(line string) string {
	// Skip if line starts with whitespace (recipe line)
	if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
		return ""
	}

	// Find the colon
	colonIdx := strings.Index(line, ":")
	if colonIdx == -1 {
		return ""
	}

	// Check if this is a variable assignment (:=, ?=, +=, !=)
	// These are NOT target definitions
	if colonIdx+1 < len(line) {
		nextChar := line[colonIdx+1]
		if nextChar == '=' || nextChar == ':' {
			return "" // Variable assignment, not a target
		}
	}

	// Extract everything before the colon
	beforeColon := line[:colonIdx]

	// Handle grouped target operator (&:)
	if strings.HasSuffix(beforeColon, "&") {
		beforeColon = strings.TrimSuffix(beforeColon, "&")
	}

	// Trim whitespace
	targetPart := strings.TrimSpace(beforeColon)
	if targetPart == "" {
		return ""
	}

	// Extract first token (handles multiple targets on same line)
	fields := strings.Fields(targetPart)
	if len(fields) > 0 {
		return fields[0]
	}

	return ""
}
