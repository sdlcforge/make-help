# Comprehensive Review TODO

Generated: 2025-11-30

This document consolidates outstanding findings from three parallel reviews:
- **Product Manager**: Developer experience perspective
- **Architect**: Structure and documentation consistency
- **Go Developer**: Code clarity for new developers

---

## Priority 3: Nice to Have

### Features

- [ ] **Add `--show-version` flag**
  - Issue: No way to display tool's own version (existing `--version` flag is for pinning in generated files)

- [ ] **Add `--check` / `--lint` validation mode**
  - Use case: CI pipelines to validate documentation without generating output
  - Would validate: all targets documented, categories consistent, no orphaned directives

- [x] **Add dry-run mode for `--create-help-target`** ✓ DONE
  - Issue: Cannot preview what will be generated without actually creating files
  - Implementation: Added `--dry-run` flag that shows what files would be created/modified
  - Tests: Comprehensive coverage in `create_help_target_test.go`
  - Docs: Updated README.md, components.md, developer-brief.md

- [ ] **Move 'help.mk' to 'make/00-help.mk' by default**

### Documentation

- [ ] **Add migration guide**
  - Issue: No documentation for users with existing `help` target
  - Action: Document resolution path when conflicts detected

- [ ] **Add "Uninstalling" section**
  - Document: What `--remove-help-target` removes, whether `.bin/` is cleaned, version control considerations

- [ ] **Add comparison to alternatives**
  - Compare to: AWK-based `make help` solutions, `just`, other Makefile documentation tools

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
