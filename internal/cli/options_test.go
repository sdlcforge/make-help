package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasAnyOptions(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "no arguments",
			args:     []string{"make-help"},
			expected: false,
		},
		{
			name:     "single flag",
			args:     []string{"make-help", "--help"},
			expected: true,
		},
		{
			name:     "short flag",
			args:     []string{"make-help", "-v"},
			expected: true,
		},
		{
			name:     "flag with value",
			args:     []string{"make-help", "--makefile-path", "Makefile"},
			expected: true,
		},
		{
			name:     "multiple flags",
			args:     []string{"make-help", "--verbose", "--no-color"},
			expected: true,
		},
		{
			name:     "non-flag argument",
			args:     []string{"make-help", "somefile"},
			expected: false,
		},
		{
			name:     "flag after non-flag",
			args:     []string{"make-help", "somefile", "--verbose"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore os.Args
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			os.Args = tt.args
			result := HasAnyOptions()
			assert.Equal(t, tt.expected, result)
		})
	}
}
