# make-help Architecture

Comprehensive architecture documentation for the make-help CLI tool.

## Table of Contents

- [Overview](#overview)
- [Design Principles](#design-principles)
- [System Architecture](#system-architecture)
- [Component Responsibilities](#component-responsibilities)
- [Package Structure](#package-structure)
- [Key Design Decisions](#key-design-decisions)
- [Detailed Documentation](#detailed-documentation)

---

## Overview

make-help is a CLI tool that generates formatted help output from specially-formatted Makefile comments. The system processes Makefiles through a pipeline of discovery, parsing, model building, ordering, and formatting stages.

## Design Principles

| Principle | Implementation |
|-----------|---------------|
| **Simplicity** | Clear separation of concerns; minimal dependencies (Cobra only) |
| **Security** | No shell injection; atomic file writes; command timeouts |
| **Testability** | Interfaces for external commands; fixture-based integration tests |
| **Usability** | Clear error messages; actionable suggestions; verbose mode |
| **Performance** | Single-pass parsing; pre-compiled regex; minimal allocations |

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                          CLI Layer                               │
│    (Cobra-based, flag validation, mode detection, routing)      │
└───────────────┬─────────────────────────────────────────────────┘
                │
                ├──> Default Mode (no mode flags)
                │    ┌────────────────────────────────────────────┐
                │    │  Discovery Service                         │
                │    │  - Makefile resolution                     │
                │    │  - File list discovery (MAKEFILE_LIST)     │
                │    │  - Target discovery (make -p)              │
                │    └──────┬─────────────────────────────────────┘
                │           │
                │    ┌──────▼─────────────────────────────────────┐
                │    │  Parser Service                            │
                │    │  - Line-by-line scanning                   │
                │    │  - Directive detection                     │
                │    │  - Target association                      │
                │    └──────┬─────────────────────────────────────┘
                │           │
                │    ┌──────▼─────────────────────────────────────┐
                │    │  Model Builder                             │
                │    │  - Aggregate file docs                     │
                │    │  - Group targets by category               │
                │    │  - Validate mixed categorization           │
                │    │  - Associate aliases and vars              │
                │    │  - Apply target filtering                  │
                │    └──────┬─────────────────────────────────────┘
                │           │
                │    ┌──────▼─────────────────────────────────────┐
                │    │  Ordering Service                          │
                │    │  - Apply discovery order preservation      │
                │    │  - Apply explicit category ordering        │
                │    │  - Sort alphabetically where configured    │
                │    └──────┬─────────────────────────────────────┘
                │           │
                │    ┌──────▼─────────────────────────────────────┐
                │    │  Summary Extractor                         │
                │    │  - Port of extract-topic algorithm         │
                │    │  - Markdown stripping                      │
                │    │  - First sentence extraction               │
                │    └──────┬─────────────────────────────────────┘
                │           │
                │    ┌──────▼─────────────────────────────────────┐
                │    │  Formatter Service                         │
                │    │  - Template rendering (summary view)       │
                │    │  - Color application                       │
                │    │  - Layout formatting                       │
                │    └──────┬─────────────────────────────────────┘
                │           │
                │           ▼
                │       [STDOUT]
                │
                ├──> --target <name> Mode
                │    ┌────────────────────────────────────────────┐
                │    │  Discovery → Parser → Model Builder        │
                │    │  Extract single target's full docs         │
                │    │  Format detailed view (full documentation) │
                │    └──────┬─────────────────────────────────────┘
                │           ▼
                │       [STDOUT]
                │
                ├──> --create-help-target Mode
                │    ┌────────────────────────────────────────────┐
                │    │  Help Target Generator                     │
                │    │  - Detect include pattern                  │
                │    │  - Generate help target file               │
                │    │  - Create help-<target> targets            │
                │    │  - Inject include directive                │
                │    └────────────────────────────────────────────┘
                │
                └──> --remove-help-target Mode
                     ┌────────────────────────────────────────────┐
                     │  Help Target Remover                       │
                     │  - Identify help target artifacts          │
                     │  - Remove include directives               │
                     │  - Clean up generated files                │
                     └────────────────────────────────────────────┘
```

## Component Responsibilities

| Component | Responsibility | Key Inputs | Key Outputs |
|-----------|---------------|------------|-------------|
| CLI Layer | Parse commands/flags, validate options, detect terminal capabilities | Args, env vars | Config object |
| Discovery Service | Find Makefile paths and extract available targets | Makefile path | File list, target list |
| Parser Service | Extract documentation directives from Makefile content | File content | Raw directives |
| Model Builder | Build internal documentation model from directives | Raw directives | HelpModel |
| Ordering Service | Apply ordering rules to categories and targets | HelpModel + flags | Ordered HelpModel |
| Summary Extractor | Generate concise summaries from full documentation | Markdown text | Summary text |
| Formatter Service | Render help output with colors and layout | HelpModel + config | Formatted string |
| Add-Target Service | Generate and inject help target into Makefile | Makefile, options | Modified files |
| Remove-Target Service | Remove help target artifacts | Makefile | Modified files |

## Package Structure

```
make-help/
├── cmd/make-help/           # CLI entry point
├── internal/
│   ├── cli/                 # Command-line interface (Cobra-based)
│   ├── discovery/           # Makefile and target discovery
│   ├── parser/              # Documentation parsing
│   ├── model/               # Data structures and builder
│   ├── ordering/            # Sorting strategies
│   ├── summary/             # Summary extraction (extract-topic port)
│   ├── format/              # Output rendering with colors
│   ├── target/              # Help file generation/removal
│   └── errors/              # Custom error types
├── examples/                # Working example projects
├── scripts/                 # Helper scripts
├── test/                    # Test fixtures and integration tests
└── docs/                    # Architecture documentation
```

### Package Design Rationale

- **`internal/`**: All code is internal (not intended as library) to prevent API commitment
- **`internal/cli/`**: Thin layer, delegates to services; uses Cobra for consistency with Go ecosystem
- **`internal/discovery/`**: Isolates external `make` command execution for testability
- **`internal/parser/`**: Pure functions for parsing; no side effects
- **`internal/model/`**: Central data structures; builder pattern for construction
- **`internal/ordering/`**: Strategy pattern for flexible ordering algorithms
- **`internal/summary/`**: Port of extract-topic; isolated for unit testing
- **`internal/format/`**: Template-based rendering for flexibility and testability
- **`internal/target/`**: Help target generation and removal; file manipulation with atomic writes
- **`internal/errors/`**: Centralized error definitions for consistent handling

## Key Design Decisions

1. **Security - No Shell Injection:** MAKEFILE_LIST discovery uses temporary physical files instead of bash process substitution `<(...)` to eliminate command injection risk.

2. **Robustness - Command Timeouts:** All `make` command executions use `context.WithTimeout` (30 seconds) to prevent indefinite hangs on malformed Makefiles.

3. **Robustness - Atomic File Writes:** All file modifications (add-target, remove-target) write to a temporary file first, then rename. This prevents file corruption if the process crashes mid-write.

4. **Robustness - Makefile Validation:** Before modifying any Makefile, run `make -n` (dry-run) to validate syntax and catch errors early.

5. **Performance - Pre-compiled Regex:** All regex patterns in the summary extractor are compiled once at construction time, avoiding repeated compilation when processing many targets.

6. **Maintainability - Package Documentation:** Each package includes a doc.go file with comprehensive godoc comments explaining purpose, design decisions, and security considerations.

7. **Usability - Verbose Mode:** The `--verbose` flag enables detailed output about file discovery, target parsing, and file modifications for debugging.

---

## Detailed Documentation

For in-depth information on specific aspects of the architecture:

| Document | Description |
|----------|-------------|
| [Components](architecture/components.md) | Detailed specifications for each major component |
| [Data Models](architecture/data-models.md) | Core data structures and type definitions |
| [Algorithms](architecture/algorithms.md) | Key algorithms and their implementations |
| [Program Flow](architecture/program-flow.md) | Step-by-step processing pipelines |
| [Error Handling](architecture/error-handling.md) | Error classification and handling strategies |
| [Testing Strategy](architecture/testing-strategy.md) | Comprehensive testing approach |

---

## Architecture Review Summary

This design follows Go idioms and emphasizes:

### Strengths

- **Simplicity:** Clear separation of concerns; minimal dependencies
- **Standard Approaches:** Cobra for CLI; builder pattern; strategy pattern
- **Usability:** Clear error messages; well-documented packages; verbose mode
- **Security:** No shell injection; file path validation; input sanitization
- **Maintainability:** Testable design; clear boundaries; comprehensive error handling
- **Performance:** Single-pass parsing; pre-compiled regex; minimal allocations
- **Robustness:** Command timeouts; atomic file writes; syntax validation
