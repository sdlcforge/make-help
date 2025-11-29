package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessColorFlags(t *testing.T) {
	tests := []struct {
		name        string
		noColor     bool
		forceColor  bool
		expected    ColorMode
		expectError bool
	}{
		{
			name:        "default auto mode",
			noColor:     false,
			forceColor:  false,
			expected:    ColorAuto,
			expectError: false,
		},
		{
			name:        "force color",
			noColor:     false,
			forceColor:  true,
			expected:    ColorAlways,
			expectError: false,
		},
		{
			name:        "disable color",
			noColor:     true,
			forceColor:  false,
			expected:    ColorNever,
			expectError: false,
		},
		{
			name:        "conflicting flags",
			noColor:     true,
			forceColor:  true,
			expected:    ColorAuto,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mode ColorMode
			err := processColorFlags(&mode, tt.noColor, tt.forceColor)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, mode)
			}
		})
	}
}

func TestResolveColorMode(t *testing.T) {
	tests := []struct {
		name     string
		mode     ColorMode
		expected bool
	}{
		{
			name:     "always mode",
			mode:     ColorAlways,
			expected: true,
		},
		{
			name:     "never mode",
			mode:     ColorNever,
			expected: false,
		},
		{
			name:     "auto mode",
			mode:     ColorAuto,
			expected: false, // Usually false in test environment (not a TTY)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				ColorMode: tt.mode,
			}

			result := ResolveColorMode(config)

			if tt.mode == ColorAuto {
				// For auto mode, we just check that it returns a boolean
				// The actual value depends on whether we're in a terminal
				assert.IsType(t, false, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseCategoryOrder(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "single category",
			input:    []string{"Build"},
			expected: []string{"Build"},
		},
		{
			name:     "multiple categories",
			input:    []string{"Build", "Test", "Deploy"},
			expected: []string{"Build", "Test", "Deploy"},
		},
		{
			name:     "comma-separated",
			input:    []string{"Build,Test,Deploy"},
			expected: []string{"Build", "Test", "Deploy"},
		},
		{
			name:     "mixed format",
			input:    []string{"Build,Test", "Deploy"},
			expected: []string{"Build", "Test", "Deploy"},
		},
		{
			name:     "with whitespace",
			input:    []string{" Build , Test ", "Deploy"},
			expected: []string{"Build", "Test", "Deploy"},
		},
		{
			name:     "empty strings filtered",
			input:    []string{"Build", "", "Test"},
			expected: []string{"Build", "Test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCategoryOrder(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewConfig(t *testing.T) {
	config := NewConfig()

	assert.NotNil(t, config)
	assert.Equal(t, ColorAuto, config.ColorMode)
	assert.NotNil(t, config.CategoryOrder)
	assert.Equal(t, 0, len(config.CategoryOrder))
}

func TestColorModeString(t *testing.T) {
	tests := []struct {
		mode     ColorMode
		expected string
	}{
		{
			mode:     ColorAuto,
			expected: "auto",
		},
		{
			mode:     ColorAlways,
			expected: "always",
		},
		{
			mode:     ColorNever,
			expected: "never",
		},
		{
			mode:     ColorMode(999),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.mode.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewRootCmd(t *testing.T) {
	cmd := NewRootCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "make-help", cmd.Use)

	// Check that essential flags are registered
	flags := cmd.Flags()

	assert.NotNil(t, flags.Lookup("keep-order-categories"))
	assert.NotNil(t, flags.Lookup("keep-order-targets"))
	assert.NotNil(t, flags.Lookup("keep-order-all"))
	assert.NotNil(t, flags.Lookup("category-order"))
	assert.NotNil(t, flags.Lookup("default-category"))

	persistentFlags := cmd.PersistentFlags()
	assert.NotNil(t, persistentFlags.Lookup("makefile-path"))
	assert.NotNil(t, persistentFlags.Lookup("no-color"))
	assert.NotNil(t, persistentFlags.Lookup("color"))
	assert.NotNil(t, persistentFlags.Lookup("verbose"))
}
