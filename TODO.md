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

- [x] **Add algorithm documentation to builder's `processFile()`** ✓ DONE
  - Location: `internal/model/builder.go:143-191`
  - Issue: Most complex function in codebase; interleaving logic is non-obvious
  - Action: Added comprehensive godoc comment explaining:
    - Two-pointer line-order merge algorithm
    - Step-by-step example with 6 lines of Makefile
    - Special cases (@file, @category, duplicate targets)
    - Rationale for the approach
  - Also: Added `maxInt` constant with explanatory comment

---

## Priority 2: Important (Do Soon)

### README Improvements

- [x] **Reorder installation section** ✓ DONE
  - Location: `README.md` Installation section
  - Issue: "Project-Local (Recommended)" comes first but assumes tool is already installed
  - Action: Reordered to "Install the Binary" first (with Go version requirement), then "Add Help to Your Project (Recommended)"

- [x] **Add motivation/problem statement** ✓ DONE
  - Location: `README.md:5-9`
  - Issue: Jumps straight to features without explaining the problem
  - Action: Added "Why make-help?" section explaining the problem (Makefiles lack built-in help, `make --help` shows Make's options not project targets)

### Documentation Consistency

- [x] **Update Config struct in data-models.md** ✓ DONE
  - Location: `docs/architecture/data-models.md:73-107`
  - Issue: References `TargetFile` instead of actual field `HelpFileRelPath`
  - Action: Synced with actual `internal/cli/config.go` Config struct, added all missing fields (CreateHelpTarget, RemoveHelpTarget, Version, IncludeTargets, IncludeAllPhony, Target)

- [x] **Update algorithms.md discovery implementation** ✓ DONE
  - Location: `docs/architecture/algorithms.md:16-69`
  - Issue: Shows shell process substitution but code uses temp files (more secure)
  - Action: Rewrote File Discovery algorithm to reflect temp file approach with security note

- [x] **Clarify "template-based rendering" claim** ✓ DONE
  - Location: `docs/architecture/components.md:855, 1004`
  - Issue: Code uses `strings.Builder`, not templates
  - Action: Changed to "String builder-based rendering" and "Structured rendering methods"

### Error Handling

- [x] **Improve `--remove-help-target` error specificity** ✓ DONE
  - Location: `internal/cli/root.go`
  - Issue: All 8 checks return same generic error message
  - Action: Refactored to table-driven validation with specific error messages like "--remove-help-target cannot be used with --target"

### Examples

- [x] **Add examples/README.md** ✓ DONE
  - Location: `examples/README.md` (new file)
  - Action: Created comprehensive README with quick start, example descriptions, comparison table, and usage patterns

### Code Quality

- [x] **Refactor repetitive flag validation** ✓ DONE
  - Location: `internal/cli/root.go`
  - Issue: 8 repetitive if statements checking same condition
  - Action: Created `validateRemoveHelpTargetFlags()` helper with table-driven validation (reduced 33 lines to cleaner structure)

- [x] **Add parser state machine documentation** ✓ DONE
  - Location: `internal/parser/doc.go`
  - Issue: Scanner state management is subtle and could confuse new developers
  - Action: Added ASCII art state machine diagram showing IDLE/ACCUMULATING states and transitions

### CLAUDE.md Fixes

- [x] **Fix "Adding a new directive type" instructions** ✓ DONE
  - Location: `.claude/CLAUDE.md:81-87`
  - Issue: Step 2 says "add case in `Build()`" but should say `processFile()`
  - Action: Updated with correct 6-step process including specific file locations and line numbers

- [x] **Expand Quick Troubleshooting section** ✓ DONE
  - Location: `.claude/CLAUDE.md:123-130`
  - Action: Added 3 new troubleshooting items:
    - Target not appearing in help
    - Color codes appearing in output
    - "make command timed out"

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

- [x] **Add design decisions document** ✓ DONE
  - Location: `docs/architecture/design-decisions.md`
  - Topics documented:
    - Why flags instead of subcommands?
    - Why maintain parser state instead of multi-pass parsing?
    - Why use discovery order instead of file position?
    - Plus additional decisions identified during review

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
