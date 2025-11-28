package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestErrorsImplementErrorInterface verifies all error types implement error interface.
func TestErrorsImplementErrorInterface(t *testing.T) {
	// This test verifies at compile time that all error types implement error interface.
	var _ error = &MixedCategorizationError{}
	var _ error = &UnknownCategoryError{}
	var _ error = &MakefileNotFoundError{}
	var _ error = &MakeExecutionError{}
	var _ error = &DuplicateHelpTargetError{}
	var _ error = &ValidationError{}
}

func TestMixedCategorizationError(t *testing.T) {
	err := NewMixedCategorizationError("found both categorized and uncategorized targets")
	assert.Contains(t, err.Error(), "mixed categorization")
	assert.Contains(t, err.Error(), "--default-category")
}

func TestUnknownCategoryError(t *testing.T) {
	err := NewUnknownCategoryError("Unknown", []string{"Build", "Test", "Deploy"})
	assert.Contains(t, err.Error(), `unknown category "Unknown"`)
	assert.Contains(t, err.Error(), "Build, Test, Deploy")
}

func TestMakefileNotFoundError(t *testing.T) {
	err := NewMakefileNotFoundError("/path/to/Makefile")
	assert.Contains(t, err.Error(), "Makefile not found")
	assert.Contains(t, err.Error(), "/path/to/Makefile")
	assert.Contains(t, err.Error(), "--makefile-path")
}

func TestMakeExecutionError(t *testing.T) {
	err := NewMakeExecutionError("make -p", "make: *** No rule to make target")
	assert.Contains(t, err.Error(), "make command failed")
	assert.Contains(t, err.Error(), "make -p")
	assert.Contains(t, err.Error(), "No rule to make target")

	// Test with empty stderr
	err2 := NewMakeExecutionError("make -p", "")
	assert.Contains(t, err2.Error(), "make command failed")
	assert.NotContains(t, err2.Error(), "\n")
}

func TestDuplicateHelpTargetError(t *testing.T) {
	err := NewDuplicateHelpTargetError("Makefile:15")
	assert.Contains(t, err.Error(), "help target already exists")
	assert.Contains(t, err.Error(), "Makefile:15")
	assert.Contains(t, err.Error(), "remove-target")
}

func TestValidationError(t *testing.T) {
	err := NewValidationError("syntax error", "line 10: unexpected token")
	assert.Contains(t, err.Error(), "validation failed")
	assert.Contains(t, err.Error(), "syntax error")
	assert.Contains(t, err.Error(), "line 10")

	// Test with empty details
	err2 := NewValidationError("invalid target", "")
	assert.Contains(t, err2.Error(), "validation failed")
	assert.NotContains(t, err2.Error(), "\n")
}
