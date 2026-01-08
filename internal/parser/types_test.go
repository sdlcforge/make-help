package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirectiveType_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		dt       DirectiveType
		expected string
	}{
		{
			name:     "file directive",
			dt:       DirectiveFile,
			expected: "file",
		},
		{
			name:     "category directive",
			dt:       DirectiveCategory,
			expected: "category",
		},
		{
			name:     "var directive",
			dt:       DirectiveVar,
			expected: "var",
		},
		{
			name:     "alias directive",
			dt:       DirectiveAlias,
			expected: "alias",
		},
		{
			name:     "doc directive",
			dt:       DirectiveDoc,
			expected: "doc",
		},
		{
			name:     "unknown directive",
			dt:       DirectiveType(999),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.dt.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}
