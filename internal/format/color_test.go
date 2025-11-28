package format

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewColorScheme_WithColors(t *testing.T) {
	scheme := NewColorScheme(true)

	assert.NotEmpty(t, scheme.CategoryName, "CategoryName should have color code")
	assert.NotEmpty(t, scheme.TargetName, "TargetName should have color code")
	assert.NotEmpty(t, scheme.Alias, "Alias should have color code")
	assert.NotEmpty(t, scheme.Variable, "Variable should have color code")
	assert.NotEmpty(t, scheme.Documentation, "Documentation should have color code")
	assert.NotEmpty(t, scheme.Reset, "Reset should have color code")

	// Verify specific color codes match design spec
	assert.Equal(t, "\033[1;36m", scheme.CategoryName, "CategoryName should be bold cyan")
	assert.Equal(t, "\033[1;32m", scheme.TargetName, "TargetName should be bold green")
	assert.Equal(t, "\033[0;33m", scheme.Alias, "Alias should be yellow")
	assert.Equal(t, "\033[0;35m", scheme.Variable, "Variable should be magenta")
	assert.Equal(t, "\033[0;37m", scheme.Documentation, "Documentation should be white")
	assert.Equal(t, "\033[0m", scheme.Reset, "Reset should be standard reset code")
}

func TestNewColorScheme_WithoutColors(t *testing.T) {
	scheme := NewColorScheme(false)

	assert.Empty(t, scheme.CategoryName, "CategoryName should be empty")
	assert.Empty(t, scheme.TargetName, "TargetName should be empty")
	assert.Empty(t, scheme.Alias, "Alias should be empty")
	assert.Empty(t, scheme.Variable, "Variable should be empty")
	assert.Empty(t, scheme.Documentation, "Documentation should be empty")
	assert.Empty(t, scheme.Reset, "Reset should be empty")
}
