// Package ordering applies sorting rules to categories and targets.
//
// It supports three ordering strategies:
//   - Alphabetical (default)
//   - Discovery order (--keep-order-* flags)
//   - Explicit order (--category-order flag)
//
// # Ordering Strategies
//
// Alphabetical ordering sorts categories and targets by name using
// standard string comparison.
//
// Discovery order preserves the order in which categories and targets
// were first encountered during Makefile parsing. This is useful when
// the Makefile author has intentionally ordered their documentation.
//
// Explicit ordering allows specifying exact category order via
// --category-order. Categories not in the list are appended alphabetically.
//
// # Strategy Selection
//
// The package uses the strategy pattern to select the appropriate sorting
// algorithm based on configuration flags:
//   - --keep-order-categories: Use discovery order for categories
//   - --keep-order-targets: Use discovery order for targets
//   - --keep-order-all: Both of the above
//   - --category-order: Explicit category order (overrides keep-order-categories)
package ordering
