// Package errors defines custom error types for make-help.
//
// All error types implement the standard error interface and provide
// actionable suggestions in their error messages to help users resolve issues.
//
// # Error Types
//
//   - MixedCategorizationError: Returned when categorized and uncategorized
//     targets are mixed without --default-category
//
//   - UnknownCategoryError: Returned when --category-order references a
//     category that doesn't exist; includes list of available categories
//
//   - MakefileNotFoundError: Returned when the specified Makefile doesn't
//     exist at the given path
//
//   - MakeExecutionError: Returned when a make command fails; includes
//     the stderr output for debugging
//
//   - DuplicateHelpTargetError: Returned when add-target finds an existing
//     help target
//
//   - ValidationError: Returned when Makefile validation fails (e.g.,
//     syntax errors detected by make -n)
//
// # Usage
//
// All error types have constructor functions (NewXxxError) that create
// properly initialized error instances.
package errors
