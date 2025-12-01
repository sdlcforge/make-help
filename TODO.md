# Comprehensive Review TODO

Generated: 2025-11-30

This document consolidates findings from three parallel reviews:
- **Product Manager**: Developer experience perspective
- **Architect**: Structure and documentation consistency
- **Go Developer**: Code clarity for new developers

---

## Priority 1: Critical (Do First)

### Documentation Fixes

- [x] **Fix license placeholder in README.md** ✓ DONE
  - Location: `README.md:599`
  - Issue: `[Add your license here]` is unprofessional
  - Action: Add actual license
  - Resolution: Updated to Apache 2.0 with link to LICENSE.txt

- [x] **Fix flag name inconsistency: `--help-file-path` → `--help-file-rel-path`** ✓ DONE
  - Location: `.claude/CLAUDE.md:105` - not present (already correct)
  - Location: `docs/architecture/components.md:104` - fixed
  - Action: Update all references to use `--help-file-rel-path` and `HelpFileRelPath`
  - Note: References in `.claude/plans/` are historical planning docs, not updated

- [x] **Fix or remove `--force` flag references** ✓ DONE
  - Location: `internal/errors/errors.go:99` - fixed
  - Location: `docs/architecture/error-handling.md:110-116` - fixed
  - Action: Removed `--force` references, fixed `remove-target` → `--remove-help-target`
  - Also updated: `add-target` → `--create-help-target` throughout docs

### Code Documentation

- [ ] **Add algorithm documentation to builder's `processFile()`**
  - Location: `internal/model/builder.go:180-259`
  - Issue: Most complex function in codebase; interleaving logic is non-obvious
  - Action: Add header comment explaining the two-pointer merge algorithm with example

---

## Priority 2: Important (Do Soon)

### README Improvements

- [ ] **Reorder installation section**
  - Location: `README.md` Installation section
  - Issue: "Project-Local (Recommended)" comes first but assumes tool is already installed
  - Action: Put global `go install` first, then project-local setup

- [ ] **Add motivation/problem statement**
  - Location: `README.md` (top)
  - Issue: Jumps straight to features without explaining the problem
  - Suggestion: Add brief paragraph about Makefiles lacking built-in help

### Documentation Consistency

- [ ] **Update Config struct in data-models.md**
  - Location: `docs/architecture/data-models.md:72-100`
  - Issue: References `TargetFile` instead of actual field `HelpFileRelPath`
  - Action: Sync with actual `internal/cli/config.go` Config struct

- [ ] **Update algorithms.md discovery implementation**
  - Location: `docs/architecture/algorithms.md:19-44`
  - Issue: Shows shell process substitution but code uses temp files (more secure)
  - Action: Update to show temp file approach

- [ ] **Clarify "template-based rendering" claim**
  - Location: `docs/architecture/components.md:858, 1000`
  - Issue: Code uses `strings.Builder`, not templates
  - Action: Change to "string builder-based rendering"

### Error Handling

- [ ] **Improve `--remove-help-target` error specificity**
  - Location: `internal/cli/root.go:57-89`
  - Issue: All 8 checks return same generic error message
  - Action: Each check should specify which flag was problematic

### Examples

- [ ] **Add examples/README.md**
  - Location: `examples/README.md` (new file)
  - Action: Brief description of each example and when to use it

### Code Quality

- [ ] **Refactor repetitive flag validation**
  - Location: `internal/cli/root.go:57-88`
  - Issue: 8 repetitive if statements checking same condition
  - Action: Extract to helper function or use table-driven validation

- [ ] **Add parser state machine documentation**
  - Location: `internal/parser/doc.go`
  - Issue: Scanner state management is subtle and could confuse new developers
  - Action: Add ASCII art state machine diagram

### CLAUDE.md Fixes

- [ ] **Fix "Adding a new directive type" instructions**
  - Location: `.claude/CLAUDE.md` Common Development Tasks section
  - Issue: Step 2 says "add case in `Build()`" but should say `processFile()`
  - Correct steps:
    1. Update `internal/parser/directive.go` (add constant to DirectiveType)
    2. Add parsing logic in `internal/parser/scanner.go` `parseDirective()` method
    3. Handle in `internal/model/builder.go` `processFile()` method (around line 197)
    4. Update formatter if needed
    5. Add tests
    6. Update README.md and docs/architecture.md

- [ ] **Expand Quick Troubleshooting section**
  - Location: `.claude/CLAUDE.md`
  - Add:
    - Target not appearing in help: Check if it's .PHONY (use --include-all-phony or --include-target)
    - Color codes appearing in output: Use --no-color when piping to files
    - "make command timed out": Check for infinite recursion in Makefile includes

---

## Priority 3: Nice to Have

### Features

- [ ] **Add `--show-version` flag**
  - Issue: No way to display tool's own version (existing `--version` flag is for pinning in generated files)

- [ ] **Add `--check` / `--lint` validation mode**
  - Use case: CI pipelines to validate documentation without generating output
  - Would validate: all targets documented, categories consistent, no orphaned directives

- [ ] **Add dry-run mode for `--create-help-target`**
  - Issue: Cannot preview what will be generated without actually creating files

### Documentation

- [ ] **Add migration guide**
  - Issue: No documentation for users with existing `help` target
  - Action: Document resolution path when conflicts detected

- [ ] **Add "Uninstalling" section**
  - Document: What `--remove-help-target` removes, whether `.bin/` is cleaned, version control considerations

- [ ] **Add comparison to alternatives**
  - Compare to: AWK-based `make help` solutions, `just`, other Makefile documentation tools

- [ ] **Add ADRs (Architecture Decision Records)**
  - Location: `docs/adr/` (new directory)
  - Topics:
    - Why flags instead of subcommands?
    - Why maintain parser state instead of multi-pass parsing?
    - Why use discovery order instead of file position?

- [ ] **Add visual diagrams**
  - Location: `docs/architecture/diagrams/`
  - Examples: Parser state machine, data flow through pipeline, include file resolution

- [ ] **Add glossary**
  - Location: `docs/glossary.md`
  - Terms: directive, discovery order, categorization, etc.

### Code Quality

- [ ] **Remove unused extractor from Renderer**
  - Location: `internal/format/renderer.go:14, 22`
  - Issue: Renderer embeds `summary.Extractor` but summaries are extracted during model building
  - Action: Remove field (minor memory overhead)

- [ ] **Define max int constant**
  - Location: `internal/model/builder.go:182`
  - Issue: `int(^uint(0) >> 1)` is obscure
  - Action: Define `const maxInt = int(^uint(0) >> 1)`

- [ ] **Define uncategorized category constant**
  - Location: `internal/model/types.go`
  - Issue: Empty string `""` has special meaning as uncategorized
  - Action: Add `const UncategorizedCategoryName = ""`

- [ ] **Add edge case tests for model builder**
  - Test: Target defined in multiple files
  - Test: Documentation without targets
  - Test: Targets without documentation but with `--include-all-phony`

- [ ] **Add integration tests for error scenarios**
  - Test: Invalid Makefile syntax
  - Test: Missing Makefile
  - Test: Mixed categorization without `--default-category`

- [ ] **Add example-based tests (godoc examples)**
  - Location: Create `examples_test.go` files
  - Purpose: Runnable examples that appear in godoc

- [ ] **Add review dates to architecture docs**
  - Action: Add "Last Reviewed: YYYY-MM-DD" to each architecture doc

- [ ] **Note that components.md code is illustrative**
  - Location: `docs/architecture/components.md`
  - Action: Add disclaimer that code examples are simplified/illustrative

### Infrastructure

- [ ] **Consider prebuilt binaries**
  - Issue: Non-Go users cannot use `go install`
  - Options: GitHub Releases, Homebrew

- [ ] **Add code coverage badge to README**

- [ ] **Add performance benchmarks**
  - Location: `internal/parser/scanner_bench_test.go`
  - Purpose: Baseline for future optimization

---

## Review Assessments

| Reviewer | Grade | Summary |
|----------|-------|---------|
| Product Manager | Ready for beta | Minor polish needed before 1.0 |
| Architect | APPROVE | Minor documentation updates needed |
| Go Developer | A- | Excellent with room for improvement |

### Strengths Identified

- Clear package structure and separation of concerns
- Excellent error handling with custom types and actionable messages
- Strong security practices (no shell injection, atomic writes, timeouts)
- Good documentation at package and function levels
- Strong testability through interface-based design
- Minimal external dependencies (only Cobra)
- Single-pass parsing architecture
- Pre-compiled regex patterns for performance

### Main Gaps

- Complex algorithms (builder interleaving, parser state) need more explanation
- Some repetitive code in CLI validation
- Documentation inconsistencies with actual implementation
- Missing edge case documentation and tests
- Could benefit from visual diagrams

---

## Potential Confusion Points for New Developers

1. **Parser state management** - Scanner maintains mutable state across lines; non-doc lines clear pending docs
2. **Discovery order vs. rendering order** - `DiscoveryOrder` field is set during building but only used during ordering
3. **`shouldIncludeTarget` conditions** - Three conditions interact with categorization in non-obvious ways
4. **Max int calculation** - `int(^uint(0) >> 1)` is a Go idiom that's not immediately obvious
5. **Empty string as category** - `""` is used for uncategorized targets with special meaning
