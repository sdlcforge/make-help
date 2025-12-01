# Design Decisions

This document records the key architectural and design decisions made in the make-help project. Each decision explains the context, rationale, alternatives considered, and consequences.

## Table of Contents

- [CLI Design](#cli-design)
  - [Flags Instead of Subcommands](#flags-instead-of-subcommands)
- [Parsing and Processing](#parsing-and-processing)
  - [Stateful Single-Pass Parsing](#stateful-single-pass-parsing)
  - [Discovery Order Tracking](#discovery-order-tracking)
  - [Line-Order Merging in Model Builder](#line-order-merging-in-model-builder)
- [Security](#security)
  - [Temporary Files for MAKEFILE_LIST Discovery](#temporary-files-for-makefile_list-discovery)
  - [Command Timeouts](#command-timeouts)
  - [Atomic File Writes](#atomic-file-writes)
  - [Makefile Validation Before Modification](#makefile-validation-before-modification)
- [Code Organization](#code-organization)
  - [All Packages in internal/](#all-packages-in-internal)
  - [Cobra as Only External Dependency](#cobra-as-only-external-dependency)
  - [Interface-Based Command Execution](#interface-based-command-execution)
- [User Experience](#user-experience)
  - [All-or-Nothing Categorization](#all-or-nothing-categorization)

---

## CLI Design

### Flags Instead of Subcommands

**Decision**: Use flags (`--create-help-target`, `--remove-help-target`, `--target`) instead of subcommands (`make-help add-target`, `make-help remove-target`).

**Context**: The tool has several operational modes:
- Default: Display help summary
- Detail mode: Show full docs for a specific target
- Creation mode: Generate help target file
- Removal mode: Remove help target artifacts

**Rationale**:
1. **Simplicity**: The default behavior (showing help) is the most common use case and should require the fewest keystrokes
2. **Natural workflow**: Users typically run `make-help` to see help, then occasionally use flags for other operations
3. **Consistency with Make**: Make itself uses flags (`make -f`, `make -n`) rather than subcommands
4. **Discoverability**: Running just `make-help` immediately shows useful output (the help), whereas subcommands would require `make-help help` or similar

**Alternatives Considered**:
- **Subcommands** (e.g., `make-help show`, `make-help add-target`): More explicit but requires extra typing for the common case and breaks the natural workflow
- **Positional arguments** (e.g., `make-help show`, `make-help build`): Ambiguous when target names conflict with commands

**Consequences**:
- ✅ Default behavior (help display) requires no arguments
- ✅ Aligns with Make's own command-line interface
- ✅ Simpler mental model for users
- ⚠️ Mode detection logic in root command is slightly more complex
- ⚠️ Flag validation must ensure mutual exclusivity

**Implementation**: See `internal/cli/root.go:70-88` where the `RunE` function dispatches based on flag combinations.

---

## Parsing and Processing

### Stateful Single-Pass Parsing

**Decision**: Maintain parser state (`currentCategory`, `pendingDocs`) across lines instead of using multiple passes over the file.

**Context**: The parser needs to associate documentation comments with the target that follows them, and track which category subsequent targets belong to.

**Rationale**:
1. **Performance**: Single-pass parsing is O(n) vs. multi-pass which would be O(n×m) where m is number of passes
2. **Natural fit**: Makefiles are written linearly - documentation precedes the target it documents
3. **Memory efficiency**: No need to store intermediate parsed structures between passes
4. **Simplicity**: State machine approach matches how humans read Makefiles

**Alternatives Considered**:
- **Multi-pass parsing**: First pass collects all directives, second pass associates them with targets. More complex and slower.
- **Post-processing**: Parse everything into a flat list, then group afterward. Loses line number context needed for accurate association.
- **Backtracking**: Parse targets first, then look backward for docs. Fragile and error-prone.

**Consequences**:
- ✅ Fast: O(n) time complexity
- ✅ Low memory footprint
- ✅ Accurate line-number-based association
- ⚠️ Parser must maintain state between lines
- ⚠️ Documentation that doesn't precede a target is lost (but this matches developer intent)

**Implementation**: See `internal/parser/scanner.go:12-43` where the Scanner struct maintains state fields.

---

### Discovery Order Tracking

**Decision**: Track the order in which targets and categories are discovered (via `DiscoveryOrder` field) rather than using file position or line numbers for ordering.

**Context**: Users need control over the order of targets and categories in help output. The order should be configurable (alphabetical or discovery order) but needs a stable reference point.

**Rationale**:
1. **Separation of concerns**: Discovery order is independent of file location, allowing targets from multiple files to be sensibly ordered
2. **User control**: Provides a natural "discovery order" option that respects how Make would encounter targets when executing
3. **Stable across files**: Line numbers reset per file, but discovery order is global
4. **Predictable**: Discovery order matches how `make -p` lists targets

**Alternatives Considered**:
- **File position only**: Breaks when targets come from multiple included files
- **Line numbers**: File-specific and doesn't provide a global ordering
- **Alphabetical only**: Some users want to control order explicitly via Makefile structure
- **Explicit priority field**: Requires users to manually number everything

**Consequences**:
- ✅ Supports both discovery order and alphabetical sorting
- ✅ Works seamlessly across multiple included files
- ✅ Predictable and matches Make's own behavior
- ✅ Enables `--keep-order-all` flag for preserving Makefile structure
- ⚠️ Requires tracking order during model building

**Implementation**: See `internal/model/builder.go:62-63` where `categoryOrder` and `targetOrder` counters are incremented, and `internal/ordering/service.go` which uses these values for sorting.

---

### Line-Order Merging in Model Builder

**Decision**: Use a two-pointer algorithm in `processFile()` to merge directives and targets in line-number order rather than processing them separately.

**Context**: The parser produces two separate collections: a list of directives (with line numbers) and a map of target names to line numbers. These need to be reunited to associate documentation with targets.

**Rationale**:
1. **Correctness**: Line-order merging ensures directives are associated with the correct target
2. **Handles edge cases**: Properly handles targets without docs, docs followed by non-target lines, and category inheritance
3. **Single pass**: O(n + m) where n is directives and m is targets
4. **Natural model**: Mirrors how developers write Makefiles - documentation immediately precedes targets

**Alternatives Considered**:
- **Map-based association**: Parser stores target name in directive. Fragile - what if docs precede target definition?
- **Post-processing scan**: Build targets first, then scan backwards for docs. Requires multiple passes and is error-prone.
- **Greedy association**: Associate each directive with nearest target. Doesn't handle category state or gaps correctly.

**Consequences**:
- ✅ Correct association even with complex Makefile structures
- ✅ Handles state (category inheritance) naturally
- ✅ Single-pass efficiency
- ✅ Clear algorithm (documented in code with examples)
- ⚠️ Requires sorting targets by line number before merging
- ⚠️ More complex than naive approaches

**Implementation**: See `internal/model/builder.go:192-308` with detailed algorithm documentation and example walkthrough.

---

## Security

### Temporary Files for MAKEFILE_LIST Discovery

**Decision**: Use temporary physical files instead of bash process substitution (`<(...)`) when discovering included Makefiles.

**Context**: To discover all Makefiles via `MAKEFILE_LIST`, we need to append a discovery target to the main Makefile and run `make`. Originally considered using `make -f <(cat Makefile && echo discovery_target)`.

**Rationale**:
1. **Security**: Process substitution opens the door to command injection if Makefile path is user-controlled
2. **Portability**: Process substitution is bash-specific and doesn't work in all shells
3. **Robustness**: Temporary files are more reliable across different systems
4. **Auditability**: Security reviewers can easily verify no shell injection is possible

**Alternatives Considered**:
- **Process substitution**: `make -f <(cat Makefile && echo target)` - Requires bash, potential security risk
- **Stdin piping**: `cat Makefile | make -f -` - Make doesn't support reading from stdin
- **Modifying original file**: Too dangerous, could corrupt user's Makefile

**Consequences**:
- ✅ No command injection risk
- ✅ Works on any POSIX system
- ✅ Atomic cleanup (defer ensures temp file removal)
- ⚠️ Requires filesystem I/O (but negligible performance impact)
- ⚠️ Must create temp file in same directory as Makefile to preserve relative includes

**Implementation**: See `internal/discovery/filelist.go:18-94` with detailed security comment.

---

### Command Timeouts

**Decision**: All `make` command executions use `context.WithTimeout` (30 seconds) to prevent indefinite hangs.

**Context**: When executing external `make` commands to discover targets or validate Makefiles, infinite loops or recursive includes could cause the tool to hang forever.

**Rationale**:
1. **Robustness**: Prevents tool from hanging on malformed Makefiles
2. **User experience**: Users get an error message rather than indefinite waiting
3. **Resource protection**: Prevents runaway processes on CI/CD systems
4. **Fail-fast**: 30 seconds is generous for most projects but prevents truly pathological cases

**Alternatives Considered**:
- **No timeout**: Hangs forever on problematic Makefiles
- **Shorter timeout** (5-10s): May be too aggressive for large projects with many includes
- **Configurable timeout**: Adds complexity without clear benefit

**Consequences**:
- ✅ Tool never hangs indefinitely
- ✅ Clear error message when timeout occurs
- ✅ Protects against malicious or malformed Makefiles
- ⚠️ Requires context.Context threading through command execution
- ⚠️ 30 seconds might be too short for extremely large projects (but can be increased if needed)

**Implementation**: See `internal/discovery/filelist.go:54-65` and `internal/target/add.go:92-103`.

---

### Atomic File Writes

**Decision**: All file modifications use atomic writes (write to temp file, sync, rename) instead of direct writes.

**Context**: The tool modifies Makefiles and creates help target files. If the process crashes mid-write, the file could be corrupted.

**Rationale**:
1. **Data safety**: Prevents corruption if process crashes or is killed during write
2. **All-or-nothing**: File is either completely written or not modified at all
3. **Standard practice**: Atomic writes are a well-known pattern for critical file operations
4. **Minimal overhead**: Slight performance cost but critical files justify it

**Alternatives Considered**:
- **Direct writes**: Simple but risks corruption on crash/kill
- **Backups before write**: Creates clutter and requires cleanup logic
- **Write-ahead log**: Overkill for this use case

**Consequences**:
- ✅ Files never corrupted mid-write
- ✅ Crash-safe operations
- ✅ Temp file automatically cleaned up on error (via defer)
- ⚠️ Slightly more complex code
- ⚠️ Temporary file must be in same directory (for atomic rename on same filesystem)

**Implementation**: See `internal/target/file.go:9-59` with detailed comments on atomic rename requirements.

---

### Makefile Validation Before Modification

**Decision**: Run `make -n` (dry-run) to validate Makefile syntax before any modifications.

**Context**: Before injecting help targets or include directives, we need to ensure the Makefile is valid.

**Rationale**:
1. **Fail-fast**: Detect syntax errors before making changes
2. **Better errors**: Users get Make's own error messages about what's wrong
3. **Safety**: Avoids modifying broken Makefiles and making diagnosis harder
4. **Catch side effects**: Dry-run (-n) doesn't execute recipes but validates syntax

**Alternatives Considered**:
- **No validation**: Modify file then discover it's broken afterward
- **Parse Makefile ourselves**: Extremely complex, Make's syntax has many edge cases
- **Regex-based validation**: Incomplete and error-prone

**Consequences**:
- ✅ Early detection of syntax errors
- ✅ Clear error messages from Make itself
- ✅ Prevents modifying broken Makefiles
- ⚠️ Requires executing `make` command (with timeout protection)
- ⚠️ Adds slight latency to modification operations

**Implementation**: See `internal/target/add.go:90-104`.

---

## Code Organization

### All Packages in internal/

**Decision**: Place all code in `internal/` directory rather than exporting public packages.

**Context**: Standard Go practice is to export packages that others might use, but this tool is a CLI application, not a library.

**Rationale**:
1. **No API commitment**: Changes don't break external users because there are no external users
2. **Freedom to refactor**: Can change package structure without semantic versioning concerns
3. **Clear intent**: Signals this is an application, not a library
4. **Best practice**: Follows Go project layout guidelines for CLI tools

**Alternatives Considered**:
- **Public packages**: Creates API commitment we don't want
- **Flat structure**: All code in one package becomes unwieldy at scale
- **Mixed approach**: Some internal, some public - confusing and unclear

**Consequences**:
- ✅ Complete freedom to refactor
- ✅ No versioning constraints on internal APIs
- ✅ Clear that this is not a library
- ❌ Code cannot be imported by other projects (this is intentional)

**Implementation**: All packages under `internal/` directory. See project structure in `docs/architecture.md:129-147`.

---

### Cobra as Only External Dependency

**Decision**: Use only Cobra for CLI framework; everything else uses Go standard library.

**Context**: Go ecosystem has many libraries for various tasks (logging, HTTP, etc.). We need to decide on dependency philosophy.

**Rationale**:
1. **Minimalism**: Fewer dependencies = smaller binary, faster builds, less supply chain risk
2. **Stability**: Standard library is stable and well-maintained
3. **Learning**: Cobra is widely known in Go ecosystem, reducing learning curve
4. **Sufficient**: Standard library provides everything we need (file I/O, exec, regex, etc.)
5. **Security**: Each dependency is a potential security risk

**Alternatives Considered**:
- **Zero dependencies**: Implement CLI parsing ourselves. Significant work and error-prone.
- **Multiple frameworks**: Use libraries for HTTP, logging, config, etc. Overkill for this tool.
- **Alternative CLI frameworks** (urfave/cli, flag only): Cobra is more feature-rich and well-known

**Consequences**:
- ✅ Small binary size
- ✅ Fast builds
- ✅ Minimal supply chain risk
- ✅ Standard library knowledge is transferable
- ⚠️ Must implement some things ourselves (e.g., colored output)
- ⚠️ Cobra is one dependency we must maintain

**Implementation**: See `go.mod:1-16` showing only Cobra + test dependencies.

---

### Interface-Based Command Execution

**Decision**: Use `CommandExecutor` interface for all external command execution instead of calling `exec.Command` directly.

**Context**: The tool needs to execute `make` commands to discover targets and validate Makefiles. Direct calls to `exec.Command` are hard to test.

**Rationale**:
1. **Testability**: Enables mocking for unit tests without actually executing `make`
2. **Flexibility**: Can inject alternative implementations (e.g., for debugging)
3. **Standard practice**: Dependency injection via interfaces is idiomatic Go
4. **Timeout support**: Interface includes `ExecuteContext` for timeout handling

**Alternatives Considered**:
- **Direct exec.Command calls**: Simple but untestable
- **Global mock flag**: Fragile and pollutes production code
- **Build tags**: Requires separate test implementations

**Consequences**:
- ✅ Fully testable without executing external commands
- ✅ Easy to mock for tests
- ✅ Supports timeout/cancellation via context
- ⚠️ Slight indirection in code
- ⚠️ Must pass executor through function calls

**Implementation**: See `internal/discovery/executor.go:9-42` for interface definition and default implementation.

---

## User Experience

### All-or-Nothing Categorization

**Decision**: Enforce that either all targets are categorized or none are (unless `--default-category` is provided).

**Context**: When some targets have `@category` and others don't, it's ambiguous whether the uncategorized targets should be shown separately or were simply forgotten.

**Rationale**:
1. **Clarity**: Forces users to be explicit about categorization strategy
2. **Prevents mistakes**: Catches cases where user forgot to categorize some targets
3. **Consistent output**: Help display is either fully categorized or flat, never mixed
4. **Escape hatch**: `--default-category` provides a way to resolve mixed situations

**Alternatives Considered**:
- **Allow mixed**: Show categorized targets in categories, uncategorized at end. Unclear and potentially hides forgotten docs.
- **Auto-categorize uncategorized as "Other"**: Assumes user intent, might hide mistakes
- **Warn but allow**: Users might ignore warnings

**Consequences**:
- ✅ Prevents accidental omissions
- ✅ Forces intentional categorization strategy
- ✅ Consistent help output
- ⚠️ Can be surprising to new users
- ⚠️ Requires explicit `--default-category` to resolve mixed cases
- ✅ Error message provides actionable fix

**Implementation**: See `internal/model/validator.go:7-38` with clear error message suggesting `--default-category`.

---

## Summary

These design decisions prioritize:

1. **Security**: Command injection prevention, timeouts, atomic writes, validation before modification
2. **Simplicity**: Minimal dependencies, single-pass processing, clear separation of concerns
3. **Usability**: Natural CLI design, clear error messages, sensible defaults
4. **Maintainability**: Testable design via interfaces, freedom to refactor via internal packages
5. **Correctness**: Line-order merging, discovery order tracking, all-or-nothing categorization

Each decision involves trade-offs, but the consequences align with the project's goals of being a secure, simple, and reliable tool for Makefile documentation.
