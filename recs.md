# Pre-Release Recommendations

Consolidated findings from Product Manager, Architect, and Go Developer reviews.

**Overall Verdict: APPROVE FOR v1.0 RELEASE** (9.5/10 code quality)

---

## Must-Fix Before Release

### Code Quality (Go Developer)

| # | Issue | File | Fix |
|---|-------|------|-----|
| ~~1~~ | ~~Dead Scanner field~~ | ~~`internal/parser/scanner.go:14`~~ | ~~Remove unused `currentCategory` field from Scanner struct~~ ✅ |
| ~~2~~ | ~~Magic timeout constants~~ | ~~`internal/discovery/targets.go:30`, `internal/target/add.go:93`~~ | ~~Define named constants: `makeDiscoveryTimeout = 30 * time.Second`, `makeValidationTimeout = 10 * time.Second`~~ ✅ |
| ~~3~~ | ~~Missing bounds check~~ | ~~`internal/summary/extractor.go:192`~~ | ~~Add defensive guard: `if len(strippedSentence) == 0 { return "" }`~~ ✅ |

---

## High Priority

### Product Manager

| # | Issue | Action |
|---|-------|--------|
| ~~4~~ | ~~`--show-help` documentation inconsistency~~ | ~~CLAUDE.md mentions `--show-help` but flag doesn't exist; remove all references~~ ✅ |
| ~~5~~ | ~~Absolute paths in generated files~~ | ~~Generated `help.mk` files contain absolute Source: paths; use relative paths or make configurable~~ ✅ |

---

## Medium Priority

### Documentation (Product Manager)

| # | Issue | Action |
|---|-------|--------|
| ~~6~~ | ~~README lacks quick example~~ | ~~Add "30-Second Example" section before "Why make-help?"~~ ✅ |
| 7 | CLI help is wall of text | Add "Quick start: Run `make-help` in a directory with a Makefile" to `--help` output |
| ~~8~~ | ~~Inconsistent `!category Help`~~ | ~~Ensure consistency between `help` and `update-help` targets in generated files~~ ✅ |
| ~~9~~ | ~~`!file` behavior unclear~~ | ~~Add callout box explaining entry point vs included file behavior~~ ✅ |

### Testing (Go Developer)

| # | Issue | Action |
|---|-------|--------|
| 10 | Missing error path tests | Add integration tests for invalid syntax, timeouts, permissions |
| 11 | No concurrency tests | Test `AtomicWriteFile` concurrent writes don't corrupt |
| 12 | Parser edge cases | Test Unicode, large files, mixed line endings, malformed directives |
| 13 | Summary extraction edge cases | Test "v1.0.0.", "U.S.", URLs with periods, backtick code with periods |

---

## Low Priority (Future Enhancements)

| # | Issue | Action |
|---|-------|--------|
| 14 | First-time setup friction | Add `make-help init` command |
| 15 | No shell completion | Add bash/zsh/fish completion (Cobra has built-in support) |
| 16 | Regex compilation in functions | Move to package-level vars in `target/add.go` |
| 17 | Map allocations | Pre-size maps when target count is known |
| 18 | No performance benchmarks | Add `*_bench_test.go` files for regression detection |

---

## Already Completed (Earlier in Session)

1. ✅ `!noalias` → `!notalias` (README.md)
2. ✅ Coverage targets updated (testing-strategy.md)
3. ✅ `--help-file-rel-path` default clarified (README.md)
4. ✅ Config struct fields updated (data-models.md)
5. ✅ DirectiveNotAlias added to enum (data-models.md)
6. ✅ SegmentType names fixed (data-models.md)
7. ✅ `--makefile-path` clarified (README.md)
8. ✅ AtomicWriteFile casing fixed (components.md)

---

## Test Coverage Status

| Package | Current | Target | Status |
|---------|---------|--------|--------|
| errors | 100% | - | ✓ |
| ordering | 100% | 95% | ✓ |
| model | 98.6% | 90% | ✓ |
| format | 97.8% | 85% | ✓ |
| parser | 95.1% | 95% | ✓ |
| richtext | 95.0% | - | ✓ |
| lint | 90.8% | - | ✓ |
| discovery | 87.9% | 80% | ✓ |
| cli | 87.1% | - | ✓ |
| target | 79.1% | 85% | ⚠️ Below target |
| summary | 77.1% | 90% | ⚠️ Below target |

---

## Security Verification (All Pass)

- ✓ No shell injection (uses `exec.CommandContext` with args)
- ✓ Command timeouts (30s) prevent DoS
- ✓ Atomic file writes (temp + rename)
- ✓ Recursion detection (`MAKE_HELP_GENERATING` env)
- ✓ Physical temp files (no process substitution)
- ✓ No `unsafe` package usage
- ✓ Zero TODO/FIXME/HACK markers in codebase
