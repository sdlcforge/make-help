# Architect Brief: make-help

## Overview

Design a Go CLI tool called `make-help` that provides dynamic help generation for Makefiles. The tool parses specially formatted documentation comments in Makefiles and generates user-friendly help output.

## Deliverable

Create `docs/design.md` describing the high-level design, major components, data structures, and program flow.

## Requirements Summary

### CLI Commands

1. **`make-help`** (default): Generate help output from Makefile documentation
2. **`make-help add-target`**: Add a `help` target to the Makefile
3. **`make-help remove-target`**: Remove `help` target artifacts

### Global Options

- `--makefile-path <path>`: Override default Makefile location (default: `cwd()/Makefile`)
- `--no-color`: Disable colored output
- `--color`: Force colored output (even when piping)
- `--verbose`: Enable verbose output for debugging file discovery and parsing

### Help Generation Options

- `--keep-order-all`: Preserve discovery order for both categories and targets
- `--keep-order-categories`: Preserve discovery order for categories only
- `--keep-order-targets`: Preserve discovery order for targets only
- `--category-order <list>`: Comma-separated list specifying category display order
- `--default-category <name>`: Default category for uncategorized targets

### Add-Target Options

- `--target-file <path>`: Specify explicit path for generated help target file
- All help generation options (passed through to the generated help target command)

### Documentation Syntax

Documentation lines start with `## ` (matching `/^## /`).

**Directives:**

| Directive | Description |
|-----------|-------------|
| `## @file` | File-level documentation; contiguous `## ` lines following become top-level docs |
| `## @category <name>` | Creates/continues a category section for subsequent targets |
| `## @var <NAME> - <description>` | Documents an environment variable used by a target |
| `## @alias <name>[, <name>...]` | Defines aliases for a target (multiple directives allowed) |

**Rules:**

- Multiple `@file` sections are concatenated
- `@category` directives can be split; targets are grouped by category name
- Uncategorized and categorized targets cannot be mixed (error unless `--default-category` specified)
- Grouped targets (`foo bar baz: ...`) and variable targets (`$(VAR): ...`) treat everything before `:` or `&:` as the target name

### Output Format

**`make help`:**
```
Usage: make [<target>...] [<ENV_VAR>=<value>...]

<@file documentation>

Targets:

[Category Name:]
  - <target>[ <alias1>, ...]: <summary>
    [Vars: <VAR1>, <VAR2>...]
```

**`make help-<target>`:**
Detailed view including:
- Full documentation (not just summary)
- List of aliases
- Variable descriptions with full text

### Summary Extraction

The summary is extracted from target documentation following the logic of the `extract-topic` library. Key behaviors:

1. Process/remove markdown headers
2. Strip comment characters (not applicable here since we strip `## ` prefix)
3. Remove markdown formatting and HTML tags
4. Normalize whitespace (collapse newlines, multiple spaces)
5. Extract first sentence using regex: `/^((?:[^.!?]|\.\.\.|\.[^\s])+[.?!])(\s|$)/`
6. Handle special cases like IP addresses (`127.0.0.1.`) and ellipsis (`...`)

### Makefile Discovery

Use Make itself to discover included files:

```bash
make -f <(cat Makefile && echo && echo && echo -e '.PHONY: _list_makefiles\n_list_makefiles:\n\t@echo $(MAKEFILE_LIST)') _list_makefiles
```

**Important:** Included files are processed after their parent file completes (not at the include point). This matches `MAKEFILE_LIST` order.

### Target Discovery

Extract targets from Make's database:

```bash
make -p -r
```

Then filter in Go to extract target names (lines with `:` that aren't comments or whitespace-prefixed).

### Add-Target Behavior

1. If `--target-file` specified: Create file at path, add `include` directive to main Makefile
2. Else if `include make/*.mk` pattern found: Create `make/01-help.mk`
3. Else: Append help target directly to main Makefile

Generated help target format:
```makefile
.PHONY: help
help:
	make-help [options...]
```

### Ordering Rules

**Default:** Categories and targets sorted alphabetically.

**With flags:**
- `--keep-order-categories`: Categories appear in first-discovery order
- `--keep-order-targets`: Targets appear in first-discovery order within categories
- `--keep-order-all`: Both of the above
- `--category-order`: Explicit category order; unlisted categories appended alphabetically

Split categories are ordered by their first declaration.

### Color Output

- Auto-detect terminal color support
- `--no-color`: Disable colors
- `--color`: Force colors
- Piped output implies `--no-color` unless `--color` specified

## Design Document Requirements

The `docs/design.md` should cover:

1. **Architecture Overview**: High-level component diagram and responsibilities

2. **Package Structure**: Recommended Go package layout with package-level godoc comments (doc.go files)

3. **Core Data Structures**:
   - Makefile documentation model (files, categories, targets, variables, aliases)
   - Configuration/options model (including `--verbose` flag)
   - Output rendering model

4. **Major Components**:
   - CLI parser (consider cobra or similar)
   - Makefile scanner/parser
   - Documentation extractor
   - Topic/summary extractor (Go port of extract-topic with pre-compiled regex patterns)
   - Output formatter (with color support)
   - Target file generator (for add-target)

5. **Program Flow**:
   - Help generation flow
   - Add-target flow
   - Remove-target flow

6. **Key Algorithms**:
   - File discovery via MAKEFILE_LIST (using temporary files, not bash process substitution)
   - Documentation parsing and directive handling
   - Category/target ordering logic
   - Summary extraction algorithm

7. **Error Handling**:
   - Mixed categorized/uncategorized targets
   - Unknown categories in `--category-order`
   - File not found scenarios
   - Invalid directive syntax

8. **Security & Robustness**:
   - Use temporary physical files instead of bash process substitution to prevent command injection
   - Add timeout (30s) to all make command executions to prevent indefinite hangs
   - Use atomic file writes (write to temp, then rename) for Makefile modifications
   - Validate Makefile syntax with `make -n` before modification

9. **Testing Strategy**:
   - Unit test approach for each component
   - Integration test approach for CLI commands
   - Test fixtures for Makefile parsing

## Reference

The `extract-topic` library (JavaScript) provides the summary extraction logic to port:

- **Sentence regex:** `/^((?:[^.!?]|\.\.\.|\.[^\s])+[.?!])(\s|$)/`
- **Handles:** IP addresses, ellipsis, standard punctuation
- **Strips:** Markdown formatting, HTML tags, normalizes whitespace
- **Returns:** First sentence as summary
