# Implementation Plan: make-help

This document provides a detailed, step-by-step implementation plan for the `make-help` CLI tool based on the design specification in `docs/design.md`.

## Overview

The implementation is organized into 8 phases, following the recommended order from the design document. Each phase builds upon previous work and includes specific deliverables with acceptance criteria.

---

## Phase 1: Project Setup & Core Data Structures

**Goal:** Establish project foundation and define all core types.

### 1.1 Initialize Go Module

```bash
go mod init github.com/sdlcforge/make-help
```

**Files to create:**
- `go.mod`
- `go.sum` (after adding dependencies)

### 1.2 Add Dependencies

```bash
go get github.com/spf13/cobra@latest
go get github.com/stretchr/testify@latest  # For testing
```

### 1.3 Create Directory Structure

```
make-help/
├── cmd/
│   └── make-help/
│       └── main.go
├── internal/
│   ├── cli/
│   ├── discovery/
│   ├── parser/
│   ├── model/
│   ├── ordering/
│   ├── summary/
│   ├── format/
│   ├── target/
│   └── errors/
├── test/
│   ├── fixtures/
│   │   ├── makefiles/
│   │   └── expected/
│   └── integration/
├── go.mod
├── go.sum
└── README.md
```

### 1.4 Implement Core Data Structures

**File:** `internal/model/types.go`

Implement the following types from design section 3.1:
- `HelpModel` - Complete parsed help documentation
- `Category` - Documentation category with targets
- `Target` - Documented target with aliases, variables, docs
- `Variable` - Environment variable documentation

**File:** `internal/model/doc.go`

Package-level documentation as specified in design section 2.2.

**File:** `internal/parser/types.go`

Implement:
- `Directive` - Parsed documentation directive
- `DirectiveType` - Enum for directive types (File, Category, Var, Alias, Doc)
- `ParsedFile` - Single parsed Makefile result

**File:** `internal/parser/doc.go`

Package-level documentation.

**File:** `internal/cli/config.go`

Implement from design section 3.2:
- `Config` - All CLI configuration options
- `ColorMode` - Enum for Auto, Always, Never

**File:** `internal/cli/doc.go`

Package-level documentation.

### 1.5 Implement Custom Error Types

**File:** `internal/errors/errors.go`

Implement from design section 7.2:
- `MixedCategorizationError`
- `UnknownCategoryError`
- `MakefileNotFoundError`
- `MakeExecutionError`

**File:** `internal/errors/doc.go`

Package-level documentation.

### Phase 1 Acceptance Criteria

- [x] `go build ./...` succeeds
- [x] All types compile without errors
- [x] Package documentation present in all doc.go files
- [x] Error types implement `error` interface

**Status: COMPLETE** (2024-11-28)

---

## Phase 2: Parser Service

**Goal:** Implement Makefile scanning and directive extraction.

### 2.1 Implement Scanner

**File:** `internal/parser/scanner.go`

Implement `Scanner` struct with methods:
- `NewScanner() *Scanner`
- `ScanFile(path string) (*ParsedFile, error)`
- `ScanContent(content string, path string) (*ParsedFile, error)` (for testing)
- `parseDirective(line string, lineNum int) Directive`
- `parseTarget(line string) string`

Key behaviors from design section 4.3:
- Stateful scanning tracking current category
- Pending documentation queue for target association
- Handle `## ` prefix for documentation lines
- Detect `@file`, `@category`, `@var`, `@alias` directives
- Extract target names from lines with `:` or `&:`
- Skip recipe lines (tab/space-prefixed)
- Clear pending docs on non-doc, non-target lines

### 2.2 Implement Directive Parsing

**File:** `internal/parser/directive.go`

Helper functions:
- `isDocumentationLine(line string) bool`
- `isTargetLine(line string) bool`
- `extractTargetName(line string) string`

Edge cases to handle:
- Grouped targets: `foo bar baz:` → extract "foo"
- Variable targets: `$(VAR):` → extract "$(VAR)"
- Grouped target operator: `&:`

### 2.3 Write Parser Tests

**File:** `internal/parser/scanner_test.go`

Test cases:
- `@file` directive parsing
- `@category` directive parsing
- `@var` directive parsing (with and without description)
- `@alias` directive parsing (single and multiple)
- Regular documentation lines
- Target detection (simple, grouped, variable)
- Documentation-target association
- Non-doc line clears pending docs
- Recipe lines ignored
- Multiple files processing

**File:** `internal/parser/directive_test.go`

Test cases for helper functions.

### Phase 2 Acceptance Criteria

- [ ] Parser correctly extracts all directive types
- [ ] Target names extracted from various formats
- [ ] Documentation associated with correct targets
- [ ] All unit tests pass
- [ ] Test coverage ≥ 95%

---

## Phase 3: Summary Extractor

**Goal:** Port extract-topic algorithm for first sentence extraction.

### 3.1 Implement Extractor

**File:** `internal/summary/extractor.go`

Implement `Extractor` struct with pre-compiled regex patterns:
- `sentenceRegex` - First sentence extraction
- `headerRegex` - Markdown header removal
- `boldRegex`, `italicRegex` - Bold/italic removal
- `boldUnderRegex`, `italicUnderRegex` - Underscore variants
- `codeRegex` - Inline code removal
- `linkRegex` - Markdown link removal
- `htmlTagRegex` - HTML tag removal
- `whitespaceRegex` - Whitespace normalization

Methods:
- `NewExtractor() *Extractor`
- `Extract(documentation []string) string`
- `stripMarkdownHeaders(text string) string`
- `stripMarkdownFormatting(text string) string`
- `stripHTMLTags(text string) string`
- `normalizeWhitespace(text string) string`
- `extractFirstSentence(text string) string`

### 3.2 Implement Markdown Utilities

**File:** `internal/summary/markdown.go`

Additional helper functions if needed for complex markdown handling.

**File:** `internal/summary/doc.go`

Package-level documentation.

### 3.3 Write Extractor Tests

**File:** `internal/summary/extractor_test.go`

Critical test cases from design section 6.4:
- Simple sentence: "Build the project. Run tests." → "Build the project."
- Ellipsis handling: "Wait for it... then proceed." → "Wait for it... then proceed."
- IP address handling: "Connect to 127.0.0.1. Then test." → "Connect to 127.0.0.1."
- Markdown stripping: `**Bold** and *italic*` → "Bold and italic"
- Link removal: `[text](url)` → "text"
- HTML tag removal: `<b>bold</b>` → "bold"
- No sentence terminator: "No terminator" → "No terminator"
- Multiple sentences: only first extracted
- Empty input: returns empty string
- Whitespace normalization

### Phase 3 Acceptance Criteria

- [ ] All regex patterns compile correctly
- [ ] First sentence extracted accurately
- [ ] Edge cases (ellipsis, IPs) handled correctly
- [ ] Markdown formatting stripped
- [ ] All unit tests pass
- [ ] Test coverage = 100%

---

## Phase 4: Model Builder

**Goal:** Aggregate parsed directives into structured HelpModel.

### 4.1 Implement Builder

**File:** `internal/model/builder.go`

Implement `Builder` struct with methods:
- `NewBuilder(config *cli.Config) *Builder`
- `Build(parsedFiles []*parser.ParsedFile) (*HelpModel, error)`
- `parseVarDirective(value string) Variable`
- `parseAliasDirective(value string) []string`

Key behaviors:
- Aggregate `@file` documentation sections
- Group targets by category
- Track discovery order for categories and targets
- Merge targets into existing categories (split category support)
- Associate aliases and variables with targets

### 4.2 Implement Validator

**File:** `internal/model/validator.go`

Implement validation logic:
- `validateCategorization(model *HelpModel, targetMap map[string]*Target) error`
- `applyDefaultCategory(model *HelpModel, targetMap map[string]*Target, categoryMap map[string]*Category)`

Validation rules:
- If categories exist, all targets must be categorized (unless `--default-category`)
- Mixed categorization returns `MixedCategorizationError`

### 4.3 Write Builder Tests

**File:** `internal/model/builder_test.go`

Test cases:
- Basic model building from parsed files
- File documentation aggregation
- Category creation and grouping
- Target-category association
- Split categories (same name in multiple files)
- Discovery order tracking
- `@var` directive parsing (with/without description)
- `@alias` directive parsing (single/multiple)
- Mixed categorization error
- Default category application
- Empty parsed files

**File:** `internal/model/validator_test.go`

Test cases for validation logic.

### Phase 4 Acceptance Criteria

- [ ] HelpModel correctly built from parsed files
- [ ] Categories properly grouped
- [ ] Discovery order tracked correctly
- [ ] Mixed categorization detected and error thrown
- [ ] Default category applied when specified
- [ ] All unit tests pass
- [ ] Test coverage ≥ 90%

---

## Phase 5: Ordering Service

**Goal:** Implement category and target sorting strategies.

### 5.1 Implement Ordering Service

**File:** `internal/ordering/service.go`

Implement `Service` struct with methods:
- `NewService(config *cli.Config) *Service`
- `ApplyOrdering(model *HelpModel) error`
- `orderCategories(model *HelpModel) error`
- `orderTargets(category *Category)`

### 5.2 Implement Ordering Strategies

**File:** `internal/ordering/strategy.go`

Implement sorting functions:
- `sortCategoriesAlphabetically(categories []Category)`
- `sortCategoriesByDiscoveryOrder(categories []Category)`
- `applyExplicitCategoryOrder(model *HelpModel, order []string) error`
- `sortTargetsAlphabetically(targets []Target)`
- `sortTargetsByDiscoveryOrder(targets []Target)`

Ordering logic from design section 4.5:
- Default: alphabetical for both categories and targets
- `--keep-order-categories`: preserve category discovery order
- `--keep-order-targets`: preserve target discovery order within categories
- `--category-order`: explicit order, remaining categories appended alphabetically
- Validate all categories in `--category-order` exist

**File:** `internal/ordering/doc.go`

Package-level documentation.

### 5.3 Write Ordering Tests

**File:** `internal/ordering/service_test.go`

Test cases:
- Alphabetical category ordering (default)
- Alphabetical target ordering (default)
- Discovery order preservation for categories
- Discovery order preservation for targets
- Explicit category order
- Explicit order with remaining categories appended
- Unknown category in `--category-order` returns error
- Combined flags (`--keep-order-all`)

### Phase 5 Acceptance Criteria

- [ ] Alphabetical ordering works correctly
- [ ] Discovery order preserved when flags set
- [ ] Explicit category order applied correctly
- [ ] Unknown category error with available categories listed
- [ ] All unit tests pass
- [ ] Test coverage ≥ 95%

---

## Phase 6: Formatter Service

**Goal:** Implement help output rendering with color support.

### 6.1 Implement Color Scheme

**File:** `internal/format/color.go`

Implement from design section 3.4:
- `ColorScheme` struct with ANSI codes
- `NewColorScheme(useColor bool) *ColorScheme`

Color assignments:
- CategoryName: Bold Cyan (`\033[1;36m`)
- TargetName: Bold Green (`\033[1;32m`)
- Alias: Yellow (`\033[0;33m`)
- Variable: Magenta (`\033[0;35m`)
- Documentation: White (`\033[0;37m`)
- Reset: `\033[0m`

### 6.2 Implement Renderer

**File:** `internal/format/renderer.go`

Implement `Renderer` struct with methods:
- `NewRenderer(config *cli.Config) *Renderer`
- `Render(model *HelpModel) (string, error)`
- `renderCategory(buf *strings.Builder, category *Category)`
- `renderTarget(buf *strings.Builder, target *Target)`
- `RenderDetailedTarget(target *Target) string` (for `help-<target>`)

Output format from design:
```
Usage: make [<target>...] [<ENV_VAR>=<value>...]

<@file documentation>

Targets:

[Category Name:]
  - <target>[ <alias1>, ...]: <summary>
    [Vars: <VAR1>, <VAR2>...]
```

**File:** `internal/format/doc.go`

Package-level documentation.

### 6.3 Write Formatter Tests

**File:** `internal/format/renderer_test.go`

Test cases:
- Basic rendering without categories
- Rendering with categories
- Target with aliases
- Target with variables
- Target with summary
- File documentation rendering
- Color output enabled
- Color output disabled
- Detailed target rendering
- Empty model handling

**File:** `internal/format/color_test.go`

Test cases for color scheme creation.

### Phase 6 Acceptance Criteria

- [ ] Help output matches expected format
- [ ] Colors applied correctly when enabled
- [ ] Colors omitted when disabled
- [ ] All elements rendered (categories, targets, aliases, vars)
- [ ] All unit tests pass
- [ ] Test coverage ≥ 85%

---

## Phase 7: Discovery Service & CLI Layer

**Goal:** Implement Makefile discovery and command-line interface.

### 7.1 Implement Command Executor Interface

**File:** `internal/discovery/executor.go`

Implement:
- `CommandExecutor` interface
- `DefaultExecutor` implementation using `exec.Command`
- Context-aware execution with timeout support

### 7.2 Implement Discovery Service

**File:** `internal/discovery/service.go`

Implement `Service` struct with methods:
- `NewService(executor CommandExecutor, verbose bool) *Service`
- `DiscoverMakefiles(mainPath string) ([]string, error)`
- `DiscoverTargets(makefilePath string) ([]string, error)`
- `resolveAbsolutePaths(files []string, baseDir string) ([]string, error)`

**File:** `internal/discovery/makefile.go`

Implement:
- `ResolveMakefilePath(path string) (string, error)` - Resolve to absolute path
- `ValidateMakefileExists(path string) error`

**File:** `internal/discovery/filelist.go`

Implement MAKEFILE_LIST discovery:
- Create temporary file with appended `_list_makefiles` target
- Execute with 30-second timeout
- Parse space-separated output
- Clean up temporary file

Security: Use temporary physical files, NOT bash process substitution.

**File:** `internal/discovery/targets.go`

Implement target discovery via `make -p -r`:
- `parseTargetsFromDatabase(output string) []string`
- Filter comments and whitespace-prefixed lines
- Extract target names using regex

**File:** `internal/discovery/doc.go`

Package-level documentation.

### 7.3 Implement CLI Layer

**File:** `internal/cli/root.go`

Implement root command using Cobra:
- `NewRootCmd() *cobra.Command`
- Global flags: `--makefile-path`, `--no-color`, `--color`, `--verbose`
- Help generation flags: `--keep-order-all`, `--keep-order-categories`, `--keep-order-targets`, `--category-order`, `--default-category`
- Color mode resolution (terminal detection)

**File:** `internal/cli/help.go`

Implement help command (default):
- `runHelp(config *Config) error`
- Orchestrate: discovery → parsing → building → ordering → summary → formatting
- Write output to stdout

**File:** `internal/cli/terminal.go`

Implement terminal detection:
- `IsTerminal(fd uintptr) bool`
- `ResolveColorMode(config *Config) bool`

### 7.4 Implement Entry Point

**File:** `cmd/make-help/main.go`

```go
package main

import (
    "os"
    "github.com/sdlcforge/make-help/internal/cli"
)

func main() {
    if err := cli.NewRootCmd().Execute(); err != nil {
        os.Exit(1)
    }
}
```

### 7.5 Write Discovery & CLI Tests

**File:** `internal/discovery/service_test.go`

Test cases using mock executor:
- Makefile list discovery
- Target discovery
- Absolute path resolution
- Timeout handling
- Error scenarios (make fails, file not found)

**File:** `internal/cli/root_test.go`

Test cases:
- Flag parsing
- Color mode resolution
- Invalid flag combinations

### Phase 7 Acceptance Criteria

- [ ] `make-help` command runs successfully
- [ ] Makefile discovery works with includes
- [ ] Target discovery extracts all targets
- [ ] Flags parsed correctly
- [ ] Color auto-detection works
- [ ] Verbose output shows discovery info
- [ ] 30-second timeout on make commands
- [ ] All unit tests pass
- [ ] Test coverage ≥ 80%

---

## Phase 8: Target Manipulation (add-target, remove-target)

**Goal:** Implement help target injection and removal.

### 8.1 Implement Atomic File Operations

**File:** `internal/target/file.go`

Implement:
- `atomicWriteFile(filename string, data []byte, perm os.FileMode) error`
  - Write to temp file in same directory
  - Sync to disk
  - Atomic rename
  - Clean up on error

### 8.2 Implement Add-Target Service

**File:** `internal/target/add.go`

Implement `AddService` struct with methods:
- `NewAddService(config *cli.Config, executor CommandExecutor, verbose bool) *AddService`
- `AddTarget() error`
- `validateMakefile(makefilePath string) error` - Run `make -n`
- `determineTargetFile(makefilePath string) (string, bool, error)`
- `generateHelpTarget() string`
- `addIncludeDirective(makefilePath, targetFile string) error`

Target file location strategy:
1. If `--target-file` specified: use it, add include directive
2. If `include make/*.mk` pattern found: create `make/01-help.mk`
3. Else: append to main Makefile

Generated target includes all relevant flags from config.

### 8.3 Implement Remove-Target Service

**File:** `internal/target/remove.go`

Implement `RemoveService` struct with methods:
- `NewRemoveService(config *cli.Config, executor CommandExecutor, verbose bool) *RemoveService`
- `RemoveTarget() error`
- `validateMakefile(makefilePath string) error`
- `removeIncludeDirectives(makefilePath string) error`
- `removeInlineHelpTarget(makefilePath string) error`
- `removeHelpTargetFiles(makefilePath string) error`

Removal strategy:
1. Remove `include ...help...\.mk` directives
2. Remove inline `help:` target and `.PHONY: help`
3. Delete `make/01-help.mk` if exists

**File:** `internal/target/generator.go`

Implement help target content generation with flag pass-through.

**File:** `internal/target/doc.go`

Package-level documentation.

### 8.4 Implement CLI Subcommands

**File:** `internal/cli/add_target.go`

Implement add-target command:
- `newAddTargetCmd(config *Config) *cobra.Command`
- Flag: `--target-file`
- Inherits all help generation flags

**File:** `internal/cli/remove_target.go`

Implement remove-target command:
- `newRemoveTargetCmd(config *Config) *cobra.Command`

### 8.5 Write Target Manipulation Tests

**File:** `internal/target/add_test.go`

Test cases:
- Append to Makefile (no include pattern)
- Create `make/01-help.mk` (include pattern found)
- Explicit `--target-file`
- Directory creation for make/
- Include directive injection
- Flag pass-through in generated target
- Makefile validation failure
- Duplicate help target detection

**File:** `internal/target/remove_test.go`

Test cases:
- Remove include directives
- Remove inline help target
- Remove recipe lines
- Delete help target files
- Multiple help targets removed
- No changes needed (already clean)

**File:** `internal/target/file_test.go`

Test cases for atomic write operations.

### Phase 8 Acceptance Criteria

- [ ] `add-target` creates help target correctly
- [ ] Three-tier target file strategy works
- [ ] Flags passed through to generated target
- [ ] Include directive added when needed
- [ ] `remove-target` removes all artifacts
- [ ] Atomic writes prevent corruption
- [ ] Makefile validated before modification
- [ ] All unit tests pass
- [ ] Test coverage ≥ 85%

---

## Phase 9: Integration Testing & Documentation

**Goal:** End-to-end testing and final documentation.

### 9.1 Create Test Fixtures

**Directory:** `test/fixtures/makefiles/`

Create test Makefiles:
- `basic.mk` - Simple targets with documentation
- `categorized.mk` - Targets with `@category`
- `with_includes.mk` - Makefile with include directives
- `mixed_categorization.mk` - Error case
- `empty.mk` - Empty Makefile
- `with_make_include.mk` - Has `include make/*.mk`
- `complex.mk` - All features combined

**Directory:** `test/fixtures/expected/`

Create expected outputs:
- `basic_help.txt`
- `categorized_help.txt`
- `categorized_ordered_help.txt`
- `with_includes_help.txt`
- `empty_with_help.mk`
- `01-help.mk`

### 9.2 Implement Integration Tests

**File:** `test/integration/cli_test.go`

Test cases:
- Help generation for each fixture
- `--category-order` flag
- `--keep-order-*` flags
- `--default-category` flag
- `--no-color` flag
- Add-target scenarios
- Remove-target scenarios
- Error cases (mixed categorization, unknown category)

### 9.3 Create README

**File:** `README.md`

Contents:
- Project description
- Installation instructions
- Usage examples
- Documentation syntax reference
- CLI options reference
- Contributing guidelines

### Phase 9 Acceptance Criteria

- [ ] All integration tests pass
- [ ] Test fixtures cover all scenarios
- [ ] README complete and accurate
- [ ] `go test ./...` passes
- [ ] Overall test coverage ≥ 90%

---

## Implementation Checklist Summary

### Phase 1: Project Setup & Core Data Structures (COMPLETE)
- [x] Initialize Go module
- [x] Add dependencies (cobra, testify)
- [x] Create directory structure
- [x] Implement `internal/model/types.go`
- [x] Implement `internal/parser/types.go`
- [x] Implement `internal/cli/config.go`
- [x] Implement `internal/errors/errors.go`
- [x] Create all `doc.go` files

### Phase 2: Parser Service
- [ ] Implement `internal/parser/scanner.go`
- [ ] Implement `internal/parser/directive.go`
- [ ] Write `internal/parser/scanner_test.go`
- [ ] Write `internal/parser/directive_test.go`

### Phase 3: Summary Extractor
- [ ] Implement `internal/summary/extractor.go`
- [ ] Implement `internal/summary/markdown.go`
- [ ] Write `internal/summary/extractor_test.go`

### Phase 4: Model Builder
- [ ] Implement `internal/model/builder.go`
- [ ] Implement `internal/model/validator.go`
- [ ] Write `internal/model/builder_test.go`
- [ ] Write `internal/model/validator_test.go`

### Phase 5: Ordering Service
- [ ] Implement `internal/ordering/service.go`
- [ ] Implement `internal/ordering/strategy.go`
- [ ] Write `internal/ordering/service_test.go`

### Phase 6: Formatter Service
- [ ] Implement `internal/format/color.go`
- [ ] Implement `internal/format/renderer.go`
- [ ] Write `internal/format/renderer_test.go`
- [ ] Write `internal/format/color_test.go`

### Phase 7: Discovery Service & CLI Layer
- [ ] Implement `internal/discovery/executor.go`
- [ ] Implement `internal/discovery/service.go`
- [ ] Implement `internal/discovery/makefile.go`
- [ ] Implement `internal/discovery/filelist.go`
- [ ] Implement `internal/discovery/targets.go`
- [ ] Implement `internal/cli/root.go`
- [ ] Implement `internal/cli/help.go`
- [ ] Implement `internal/cli/terminal.go`
- [ ] Implement `cmd/make-help/main.go`
- [ ] Write `internal/discovery/service_test.go`
- [ ] Write `internal/cli/root_test.go`

### Phase 8: Target Manipulation
- [ ] Implement `internal/target/file.go`
- [ ] Implement `internal/target/add.go`
- [ ] Implement `internal/target/remove.go`
- [ ] Implement `internal/target/generator.go`
- [ ] Implement `internal/cli/add_target.go`
- [ ] Implement `internal/cli/remove_target.go`
- [ ] Write `internal/target/add_test.go`
- [ ] Write `internal/target/remove_test.go`
- [ ] Write `internal/target/file_test.go`

### Phase 9: Integration Testing & Documentation
- [ ] Create test fixtures (makefiles)
- [ ] Create expected outputs
- [ ] Write `test/integration/cli_test.go`
- [ ] Create `README.md`
- [ ] Verify overall test coverage ≥ 90%

---

## Dependencies Between Phases

```
Phase 1 (Setup)
    │
    ├──> Phase 2 (Parser)
    │        │
    │        └──> Phase 4 (Model Builder)
    │                 │
    │                 ├──> Phase 5 (Ordering)
    │                 │        │
    │                 │        └──> Phase 7 (CLI) ──> Phase 8 (Target)
    │                 │                                    │
    │                 └──> Phase 6 (Formatter) ────────────┤
    │                                                      │
    └──> Phase 3 (Summary) ────────────────────────────────┤
                                                           │
                                                           └──> Phase 9 (Integration)
```

**Parallel work opportunities:**
- Phase 2 (Parser) and Phase 3 (Summary) can proceed in parallel after Phase 1
- Phase 5 (Ordering) and Phase 6 (Formatter) can proceed in parallel after Phase 4

---

## Risk Mitigation

### High-Risk Areas

1. **MAKEFILE_LIST Discovery**
   - Risk: Complex Makefiles may have unexpected include behavior
   - Mitigation: Test with variety of real-world Makefiles; use 30s timeout

2. **Summary Extraction Regex**
   - Risk: Edge cases in sentence detection
   - Mitigation: Port exact regex from extract-topic; comprehensive test suite

3. **Atomic File Writes**
   - Risk: File corruption on crashes
   - Mitigation: Write to temp file first, then rename; test interrupt scenarios

4. **Target Detection**
   - Risk: Complex Makefile syntax (variables, grouped targets)
   - Mitigation: Focus on documented targets only; test edge cases

### Testing Strategy

- Unit tests for each package with mocks for external dependencies
- Integration tests with real `make` command execution
- Fixture-based testing for deterministic output verification
- Test coverage targets per package as specified in design

---

## Success Metrics

1. **Functionality:** All CLI commands work as specified
2. **Reliability:** No crashes on malformed input; graceful error handling
3. **Performance:** Help generation completes in < 2 seconds for typical projects
4. **Test Coverage:** Overall ≥ 90%, critical packages (parser, summary) at 95-100%
5. **Documentation:** Complete README with usage examples
