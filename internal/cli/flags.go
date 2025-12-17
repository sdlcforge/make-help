package cli

import "strings"

// parseIncludeTargets normalizes the --include-target flag values.
// Handles both comma-separated ("foo,bar") and repeated flags.
// Input: ["foo,bar", "baz"] -> Output: ["foo", "bar", "baz"]
func parseIncludeTargets(input []string) []string {
	var result []string
	for _, item := range input {
		parts := strings.Split(item, ",")
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				result = append(result, trimmed)
			}
		}
	}
	return result
}
