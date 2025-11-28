// Package parser scans Makefile content and extracts documentation directives.
//
// It recognizes the following directive types:
//   - @file: File-level documentation
//   - @category: Category grouping for targets
//   - @var: Environment variable documentation
//   - @alias: Target aliases
//
// # Documentation Syntax
//
// Documentation lines must start with "## " (hash-hash-space).
// The parser maintains state to track the current category and pending
// documentation lines that will be associated with the next target definition.
//
// # Target Detection
//
// The parser detects target definitions by looking for lines containing ":"
// that are not indented (recipe lines are indented with tab or spaces).
// It handles grouped targets (foo bar baz:) and the grouped target
// operator (&:).
//
// # Association Rules
//
// Documentation lines immediately preceding a target definition are
// associated with that target. A non-documentation, non-target line
// breaks this association and clears pending documentation.
package parser
