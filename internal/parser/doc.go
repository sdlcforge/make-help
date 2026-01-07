// Package parser scans Makefile content and extracts documentation directives.
//
// It recognizes the following directive types:
//   - !file: File-level documentation
//   - !category: Category grouping for targets
//   - !var: Environment variable documentation
//   - !alias: Target aliases
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
//
// # Scanner State Machine
//
// The Scanner maintains internal state to associate documentation with targets.
// It operates as a simple state machine with the following states and transitions:
//
//	┌─────────────────────────────────────────────────────────────────────┐
//	│                                                                     │
//	│  State: IDLE (pendingDocs empty)                                   │
//	│  - No documentation queued                                          │
//	│  - Waiting for doc comments or target                               │
//	│                                                                     │
//	└──────┬────────────────────────────────┬─────────────────────────────┘
//	       │                                │
//	       │ Doc line (##)                  │ Target line (foo:)
//	       │ - Queue to pendingDocs         │ - No docs to attach
//	       │ - !file goes to output         │ - Continue
//	       │                                │
//	       v                                │
//	┌─────────────────────────────────────┐ │
//	│                                     │ │
//	│  State: ACCUMULATING                │ │
//	│  (pendingDocs not empty)            │ │
//	│  - Doc lines queued                 │ │
//	│  - Waiting for target or break      │ │
//	│                                     │ │
//	└──┬────────────┬──────────────────┬──┘ │
//	   │            │                  │    │
//	   │ More       │ Target line      │    │
//	   │ doc lines  │ (foo:)           │    │
//	   │ - Append   │ - Attach all     │    │
//	   │   to queue │   pendingDocs    │    │
//	   │            │ - Clear queue    │    │
//	   │            │ - Return to IDLE │    │
//	   │            │                  │    │
//	   │            v                  │    │
//	   └────────► IDLE ◄───────────────┘    │
//	                ▲                        │
//	                │                        │
//	                └────────────────────────┘
//	   Other line (non-doc, non-target)
//	   - Clear pendingDocs (orphaned docs)
//	   - Return to IDLE
//
// State Variables:
//   - pendingDocs: Queue of directives awaiting target association
//   - currentFile: Source file path for error reporting
//
// Key Behaviors:
//   - !file directives bypass pendingDocs and go directly to output
//   - !category directives queue to pendingDocs for association with next target
//   - Orphaned docs (followed by non-target line) are discarded
//   - Multiple doc blocks can exist; each associates with the next target
package parser
