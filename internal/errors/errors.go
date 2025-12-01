package errors

import (
	"fmt"
	"strings"
)

// MixedCategorizationError is returned when a Makefile contains both
// categorized and uncategorized targets without --default-category.
type MixedCategorizationError struct {
	// Message provides additional context about the error.
	Message string
}

// Error implements the error interface.
func (e *MixedCategorizationError) Error() string {
	return fmt.Sprintf("mixed categorization: %s\nUse --default-category to assign uncategorized targets to a default category", e.Message)
}

// NewMixedCategorizationError creates a new MixedCategorizationError.
func NewMixedCategorizationError(message string) *MixedCategorizationError {
	return &MixedCategorizationError{Message: message}
}

// UnknownCategoryError is returned when --category-order references
// a category that doesn't exist in the Makefile.
type UnknownCategoryError struct {
	// CategoryName is the unknown category specified by the user.
	CategoryName string

	// Available lists all valid category names found in the Makefile.
	Available []string
}

// Error implements the error interface.
func (e *UnknownCategoryError) Error() string {
	return fmt.Sprintf("unknown category %q in --category-order\nAvailable categories: %s",
		e.CategoryName, strings.Join(e.Available, ", "))
}

// NewUnknownCategoryError creates a new UnknownCategoryError.
func NewUnknownCategoryError(categoryName string, available []string) *UnknownCategoryError {
	return &UnknownCategoryError{
		CategoryName: categoryName,
		Available:    available,
	}
}

// MakefileNotFoundError is returned when the specified Makefile doesn't exist.
type MakefileNotFoundError struct {
	// Path is the path that was searched.
	Path string
}

// Error implements the error interface.
func (e *MakefileNotFoundError) Error() string {
	return fmt.Sprintf("Makefile not found: %s\nUse --makefile-path to specify location", e.Path)
}

// NewMakefileNotFoundError creates a new MakefileNotFoundError.
func NewMakefileNotFoundError(path string) *MakefileNotFoundError {
	return &MakefileNotFoundError{Path: path}
}

// MakeExecutionError is returned when a make command fails.
type MakeExecutionError struct {
	// Command is the make command that was executed.
	Command string

	// Stderr contains the error output from make.
	Stderr string
}

// Error implements the error interface.
func (e *MakeExecutionError) Error() string {
	if e.Stderr != "" {
		return fmt.Sprintf("make command failed: %s\n%s", e.Command, e.Stderr)
	}
	return fmt.Sprintf("make command failed: %s", e.Command)
}

// NewMakeExecutionError creates a new MakeExecutionError.
func NewMakeExecutionError(command, stderr string) *MakeExecutionError {
	return &MakeExecutionError{
		Command: command,
		Stderr:  stderr,
	}
}

// DuplicateHelpTargetError is returned when --create-help-target is run
// but a help target already exists.
type DuplicateHelpTargetError struct {
	// Location describes where the existing help target was found.
	Location string
}

// Error implements the error interface.
func (e *DuplicateHelpTargetError) Error() string {
	return fmt.Sprintf("help target already exists in %s\nUse 'make-help --remove-help-target' first to remove it", e.Location)
}

// NewDuplicateHelpTargetError creates a new DuplicateHelpTargetError.
func NewDuplicateHelpTargetError(location string) *DuplicateHelpTargetError {
	return &DuplicateHelpTargetError{Location: location}
}

// ValidationError is returned when Makefile validation fails.
type ValidationError struct {
	// Message describes what validation failed.
	Message string

	// Details provides additional context (e.g., make stderr output).
	Details string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("validation failed: %s\n%s", e.Message, e.Details)
	}
	return fmt.Sprintf("validation failed: %s", e.Message)
}

// NewValidationError creates a new ValidationError.
func NewValidationError(message, details string) *ValidationError {
	return &ValidationError{
		Message: message,
		Details: details,
	}
}
