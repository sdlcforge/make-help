# Comprehensive Review TODO

Generated: 2025-11-30

This document consolidates outstanding findings from three parallel reviews:
- **Product Manager**: Developer experience perspective
- **Architect**: Structure and documentation consistency
- **Go Developer**: Code clarity for new developers

---

## Priority 3: Nice to Have

### Features

- [x] **Implement '--fix'**
  - same behavior and valid with our without '--lint'

- [x] **Add `-version` flag**
  - Issue: No way to display tool's own version

- [x] **Add `--check` / `--lint` validation mode**
  - Use case: CI pipelines to validate documentation without generating output
  - Would validate: all targets documented, categories consistent, no orphaned directives

- [x] **Add dry-run mode for `--create-help-target`** ✓ DONE
  - Issue: Cannot preview what will be generated without actually creating files
  - Implementation: Added `--dry-run` flag that shows what files would be created/modified
  - Tests: Comprehensive coverage in `create_help_target_test.go`
  - Docs: Updated README.md, components.md, developer-brief.md

- [x] **Move 'help.mk' to 'make/00-help.mk' by default**

- [x] **move the 'make-help' binary into ./bin.**

### Documentation

- [x] **Add "Uninstalling" section**
  - Document: What `--remove-help-target` removes, whether `.bin/` is cleaned, version control considerations

- [ ] **Add comparison to alternatives**
  - Compare to: AWK-based `make help` solutions, `just`, other Makefile documentation tools

- [x] **Add visual diagrams** ✓ DONE
  - Location: `docs/architecture/diagrams/`
  - Created 4 Mermaid diagrams: pipeline-overview.mmd, cli-mode-routing.mmd, parser-state-machine.mmd, makefile-discovery.mmd
  - Added Makefile target `make diagrams` to generate SVG files via mmdc
  - Referenced SVG files in architecture.md, program-flow.md, and algorithms.md

- [ ] **Add glossary**
  - Location: `docs/glossary.md`
  - Terms: directive, discovery order, categorization, etc.

### Code Quality

- [x] **Remove unused extractor from Renderer**
  - Location: `internal/format/renderer.go:14, 22`
  - Issue: Renderer embeds `summary.Extractor` but summaries are extracted during model building
  - Action: Remove field (minor memory overhead)

- [x] **Define max int constant**
  - Location: `internal/model/builder.go:182`
  - Issue: `int(^uint(0) >> 1)` is obscure
  - Action: Define `const maxInt = int(^uint(0) >> 1)`

- [x] **Define uncategorized category constant** ✓ DONE
  - Location: `internal/model/types.go`
  - Added `const UncategorizedCategoryName = ""`
  - Updated all usages in `validator.go`, `renderer.go`, and `builder_test.go`

- [x] **Add edge case tests for model builder** ✓ DONE
  - Test: Target defined in multiple files (`TestBuild_DuplicateTargetInMultipleFiles`)
  - Test: Documentation without targets (`TestBuild_DocumentationWithoutTargets`)
  - Test: Targets without documentation but with `--include-all-phony` (`TestBuild_UndocumentedPhonyWithIncludeAllPhony`)

- [x] **Add integration tests for error scenarios** ✓ DONE
  - Test: Invalid Makefile syntax (`TestErrorScenario_InvalidMakefileSyntax`)
  - Test: Missing Makefile (`TestErrorScenario_MissingMakefile`)
  - Test: Mixed categorization without `--default-category` (`TestErrorScenario_MixedCategorizationWithoutDefault`)

- [x] **Add review dates to architecture docs** ✓ DONE
  - Added "Last reviewed: 2025-12-25T16:43Z" to all .md files in the repo

- [x] **Note that components.md code is illustrative** ✓ DONE
  - Replaced verbose code examples with concise pseudocode
  - Added GitHub permalink references to actual source code
  - Document reduced from ~1,683 lines to ~595 lines

### Infrastructure

- [x] **Consider prebuilt binaries**
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
5. **Empty string as category** - Now clarified via `UncategorizedCategoryName` constant in `model/types.go`

Last reviewed: 2025-12-25T16:43Z
