package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseIncludeTargets(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "single value",
			input:    []string{"foo"},
			expected: []string{"foo"},
		},
		{
			name:     "comma separated",
			input:    []string{"foo,bar"},
			expected: []string{"foo", "bar"},
		},
		{
			name:     "repeated flags",
			input:    []string{"foo", "bar"},
			expected: []string{"foo", "bar"},
		},
		{
			name:     "mixed",
			input:    []string{"foo,bar", "baz"},
			expected: []string{"foo", "bar", "baz"},
		},
		{
			name:     "with spaces",
			input:    []string{"foo, bar"},
			expected: []string{"foo", "bar"},
		},
		{
			name:     "empty",
			input:    []string{},
			expected: nil,
		},
		{
			name:     "empty strings filtered",
			input:    []string{"foo", "", "bar"},
			expected: []string{"foo", "bar"},
		},
		{
			name:     "trailing comma",
			input:    []string{"foo,bar,"},
			expected: []string{"foo", "bar"},
		},
		{
			name:     "leading comma",
			input:    []string{",foo,bar"},
			expected: []string{"foo", "bar"},
		},
		{
			name:     "multiple commas",
			input:    []string{"foo,,bar"},
			expected: []string{"foo", "bar"},
		},
		{
			name:     "only whitespace",
			input:    []string{"  ", "  "},
			expected: nil,
		},
		{
			name:     "complex mixed",
			input:    []string{"foo,bar", "baz", "qux,quux"},
			expected: []string{"foo", "bar", "baz", "qux", "quux"},
		},
		{
			name:     "with extra whitespace",
			input:    []string{" foo , bar ", " baz "},
			expected: []string{"foo", "bar", "baz"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := parseIncludeTargets(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
