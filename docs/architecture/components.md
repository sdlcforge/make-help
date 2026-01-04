# Component Architecture

Detailed specifications for each major component in the make-help system.

## Table of contents

- [CLI Parser (Cobra-based)](#cli-parser-cobra-based)
- [Discovery Service](#discovery-service)
- [Parser Service](#parser-service)
- [Model Builder](#model-builder)
- [Ordering Service](#ordering-service)
- [Summary Extractor](#summary-extractor)
- [Formatter Service](#formatter-service)
- [Create-Help-Target Service](#create-help-target-service)
- [Remove-Help-Target Service](#remove-help-target-service)

---

## Overview

### 1 CLI Parser (Cobra-based)

**Package:** `internal/cli`

**Design:** Use spf13/cobra with flag-based commands (no subcommands)

**Pseudocode:**
```
function NewRootCmd():
    config = create default configuration

    rootCmd = create Cobra command with:
        - usage information
        - flag registration (mode, input, output, misc)
        - RunE handler that:
            1. validates flag combinations
            2. resolves color mode
            3. dispatches to appropriate handler based on flags

    register flags organized by category:
        - Mode: --output, --remove-help, --dry-run, --lint, --target
        - Input: --makefile-path, --help-file-rel-path
        - Output: --color, --no-color, --include-target, ordering flags
        - Misc: --verbose

    return configured command
```

[View source: NewRootCmd](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/cli/root.go#L26-L146)

**Responsibilities:**
- Parse command-line arguments and validate flag combinations
- Detect terminal capabilities (isatty) and resolve color mode
- Delegate to appropriate service based on mode flags

**Mode Flags:**
1. **Default (no flags)**: Generate static help file via `runCreateHelpTarget()`
2. **`--output -`**: Display help dynamically via `runHelp()`
3. **`--output - --target <name>`**: Display detailed help for single target via `runDetailedHelp()`
4. **`--remove-help`**: Remove generated help files via `runRemoveHelpTarget()`
5. **`--lint`**: Run lint checks (with optional `--fix` and `--dry-run`)

[View source: runHelp orchestration](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/cli/help.go#L24-L130)

**Target Filtering:**
- **`--include-target`**: Include specific undocumented targets (repeatable, comma-separated)
- **`--include-all-phony`**: Include all .PHONY targets
- By default, only documented targets (with `## ` comments) are shown

**Generated Help File:**
When generating a help file (default mode), creates a Makefile include with:
- Static `@echo` statements containing formatted help text
- `.PHONY: help` target for summary view
- `.PHONY: help-<target>` for each documented target (detailed view)
- Auto-regeneration target that regenerates help.mk when source Makefiles change
- Fallback chain: tries `make-help`, then `npx make-help`, then shows error

**Error Handling:**
- Invalid flag combinations (e.g., `--target` without `--output -`)
- File path validation via `ResolveMakefilePath()` and `ValidateMakefileExists()`
- Conflicting color flags (`--color` + `--no-color`)
- Mode flag restrictions (enforced in PreRunE validation)

### 2 Discovery Service

**Package:** `internal/discovery`

**Design:** Execute make commands and parse output

**Pseudocode:**
```
function DiscoverMakefiles(mainPath):
    // SECURITY: Uses temporary physical file to prevent command injection
    1. read main Makefile content
    2. create temporary file with appended _list_makefiles target
    3. execute make with 30s timeout: make -f tmpFile _list_makefiles
    4. parse space-separated output (MAKEFILE_LIST variable)
    5. resolve to absolute paths
    6. cleanup temporary file
    return list of absolute Makefile paths

function DiscoverTargets(makefilePath):
    1. execute make with 30s timeout: make -f makefilePath -p -r
    2. parse make database output using regex
    3. filter out special targets, pattern rules, built-ins
    4. extract .PHONY status, dependencies, and recipe status
    return DiscoverTargetsResult with targets and metadata
```

[View source: Service.DiscoverMakefiles](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/discovery/service.go#L22-L33)
[View source: Service.DiscoverTargets](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/discovery/service.go#L40-L46)

**Key Algorithms:**
- MAKEFILE_LIST discovery via temporary target injection
- Target extraction via regex from `make -p` database output
- Absolute path resolution for included files
- 30-second timeout on all make command executions

**Error Handling:**
- Make command execution failures with stderr capture
- Makefile not found errors
- Invalid Makefile syntax (caught by make during execution)
- Timeout detection (context.DeadlineExceeded)
- Shell command injection prevention (uses exec.Command, not shell)

### 3 Parser Service

**Package:** `internal/parser`

**Design:** Stateful scanner with directive detection

**Pseudocode:**
```
type Scanner with state:
    currentFile     - file being scanned
    currentCategory - sticky category from !category directive
    pendingDocs     - documentation awaiting target association

function ScanFile(path):
    1. read file content
    2. reset scanner state
    3. for each line:
        if line starts with "##":
            parse directive (!file, !category, !var, !alias, or doc)
            if !file: add immediately to result
            else: queue in pendingDocs
        else if line is target definition:
            record target name and line number
            attach pendingDocs to this target
            clear pendingDocs
        else:
            clear pendingDocs (breaks association)
    return ParsedFile with directives and target map

function parseDirective(line):
    1. strip "## " prefix
    2. detect directive type by prefix:
        !file -> DirectiveFile
        !category -> DirectiveCategory (updates currentCategory state)
        !var -> DirectiveVar
        !alias -> DirectiveAlias
        default -> DirectiveDoc
    3. extract value after directive keyword
    return Directive object
```

[View source: Scanner.ScanFile](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/parser/scanner.go#L28-L94)
[View source: Scanner.parseDirective](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/parser/scanner.go#L99-L142)

**Key Design Decisions:**
- **Stateful scanning to track current category**: `!category` sets the current category that applies to all following targets until another `!category` directive is encountered
- **Sticky category behavior**: Once set, category applies to subsequent targets until changed or reset with `!category _`
- **Pending documentation queue**: Documentation lines are queued and associated with the next target definition
- **Simple regex-free parsing for robustness**: Target parsing uses string operations instead of complex regex
- **Target name extraction handles grouped and variable targets**: Supports `foo:`, `foo&:`, and `$(VAR):` patterns

**Error Handling:**
- File read failures return errors
- Invalid directive syntax is handled gracefully with best-effort parsing
- Malformed !var or !alias directives parse what they can

### 4 Model Builder

**Package:** `internal/model`

**Design:** Aggregate directives into structured model using two-pointer line-order merge algorithm

**Pseudocode:**
```
function Build(parsedFiles):
    1. initialize model, categoryMap, targetMap, targetToCategory
    2. for each file:
        processFile(file, ...)  // uses two-pointer merge algorithm
    3. detect implicit aliases (phony targets with single phony dep, no recipe)
    4. for each target:
        - skip if implicit alias of another target
        - apply filtering (shouldIncludeTarget)
        - add implicit aliases to target.Aliases
        - compute summary from documentation
        - assign to category
    5. validate categorization (no mixed categorized/uncategorized)
    6. apply default category if needed
    return HelpModel

function processFile(file, ...):
    // Two-pointer algorithm to merge directives and targets in line order
    1. sort targets by line number
    2. maintain state: currentCategory, pendingDocs, pendingVars, pendingAliases
    3. advance through directives and targets by line number:
        if directive comes first:
            handle !file, !category, !var, !alias, !doc
        if target comes first:
            create Target with pending directives
            assign to currentCategory
            clear pending state

function shouldIncludeTarget(target):
    return target has documentation
        OR target in IncludeTargets list
        OR (target is .PHONY AND IncludeAllPhony is true)

function detectImplicitAliases(targetMap):
    for each target:
        if .PHONY AND has 1 dependency AND dep is .PHONY AND no recipe:
            mark as implicit alias of dependency
```

[View source: Builder.Build](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/model/builder.go#L70-L149)
[View source: Builder.processFile (two-pointer merge)](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/model/builder.go#L268-L390)
[View source: Builder.shouldIncludeTarget](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/model/builder.go#L156-L175)
[View source: Builder.detectImplicitAliases](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/model/builder.go#L185-L217)

**Key Design Decisions:**
- **Builder pattern** for complex construction with immutable result
- **Two-pointer line-order merge**: Ensures directives associate with correct targets by processing in source line order
- **Discovery order tracking** via counter for `--keep-order` flags
- **Implicit alias detection**: Automatically detects phony targets that alias other targets
- **Target filtering**: Three inclusion criteria (documented, explicitly included, or all phony)
- **Validation of categorization rules**: Prevents mixing categorized and uncategorized targets

**Error Handling:**
- Mixed categorization without default category returns `MixedCategorizationError`
- Duplicate target definitions: first definition wins, subsequent ignored
- Invalid directive format handled with best-effort parsing

### 5 Ordering Service

**Package:** `internal/ordering`

**Design:** Strategy pattern for flexible ordering

**Pseudocode:**
```
function ApplyOrdering(model):
    1. orderCategories(model)
    2. for each category in model:
        orderTargets(category)

function orderCategories(model):
    if explicit --category-order provided:
        1. place specified categories in given order
        2. append remaining categories alphabetically
        3. error if specified category not found
    else if --keep-order-categories:
        sort by discovery order
    else:
        sort alphabetically (default)

function orderTargets(category):
    if --keep-order-targets:
        sort by discovery order
    else:
        sort alphabetically (default)
```

[View source: Service.ApplyOrdering](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/ordering/service.go#L25-L36)
[View source: Service.orderCategories](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/ordering/service.go#L39-L54)

**Key Design Decisions:**
- **Clear separation** of category vs target ordering (independent strategies)
- **Explicit category order with alphabetical fallback** for unspecified categories
- **Error on unknown category** in `--category-order` to catch typos early

**Error Handling:**
- Unknown category in `--category-order` returns error before applying any sorting
- All sorting operations modify model in-place

### 6 Summary Extractor

**Package:** `internal/summary`

**Design:** Port of extract-topic algorithm with pre-compiled regex patterns

**Pseudocode:**
```
type Extractor:
    // Pre-compiled regex patterns for performance
    sentenceRegex, headerRegex, boldRegex, italicRegex,
    boldUnderRegex, italicUnderRegex, codeRegex, linkRegex,
    htmlTagRegex, whitespaceRegex

function Extract(documentation):
    if empty: return ""

    1. join all documentation lines into single string
    2. strip markdown headers (## Title -> Title)
    3. strip markdown formatting (**bold**, *italic*, `code`, [links])
    4. strip HTML tags (<tag>)
    5. normalize whitespace (collapse multiple spaces/newlines)
    6. extract first sentence using regex:
        - handles ellipsis (...)
        - handles IP addresses (127.0.0.1.)
        - matches sentence ending with .!?
    7. if no sentence ending found, return full text
    return first sentence
```

[View source: Extractor struct](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/summary/extractor.go#L10-L21)
[View source: NewExtractor (pre-compiles regex)](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/summary/extractor.go#L24-L40)
[View source: Extract method](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/summary/extractor.go#L44-L64)

**Key Design Decisions:**
- **Faithful port of extract-topic algorithm** from the Node.js ecosystem
- **Pre-compiled regex patterns** for performance (avoids recompilation per target)
- **Regex handles edge cases**: ellipsis (...), IP addresses (127.0.0.1.), abbreviations
- **Fallback to full text** if no sentence boundary found
- **Order-sensitive stripping**: Removes markdown formatting before extracting sentence

**Testing:** Critical to unit test with:
- Standard sentences, ellipsis handling, IP address handling
- Markdown formatting (**bold**, *italic*, `code`, [links])
- HTML tags, multiple sentences (ensure only first extracted)

### 7 Formatter Service

**Package:** `internal/format`

**Design:** String builder-based rendering with conditional ANSI color support

**Pseudocode:**
```
type Renderer:
    colors - ColorScheme (ANSI codes or empty strings based on useColor)

function Render(model):
    // Main help output for stdout or static help.mk generation
    1. write usage line
    2. if file docs exist: write each doc line
    3. write "Targets:" header
    4. for each category:
        if category has name: write colored category header
        for each target:
            write "  - " + colored target name
            if aliases: write colored aliases
            write ": " + colored summary
            if variables: write "Vars: " + variable names
    return formatted string

function RenderDetailedTarget(target):
    // Detailed help for single target (--output - --target <name>)
    1. write "Target: " + colored target name
    2. if aliases: write "Aliases: " + aliases
    3. if variables: write "Variables:" + detailed list
    4. write full documentation (all lines)
    5. write source location
    return formatted string

function RenderForMakefile(model):
    // Generate help as list of strings for @echo embedding
    1. similar to Render() but return []string
    2. each string is escaped for Makefile @echo:
        - $ -> $$
        - " -> \"
        - \x1b -> \033 (ANSI escape literal form)
    return array of escaped lines
```

[View source: Renderer.Render](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/format/renderer.go#L30-L55)
[View source: Renderer.RenderDetailedTarget](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/format/renderer.go#L125-L183)
[View source: Renderer.RenderForMakefile](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/format/renderer.go#L216-L242)
[View source: escapeForMakefileEcho](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/format/renderer.go#L378-L403)

**Key Design Decisions:**
- **String builder for efficient concatenation** (single allocation, minimal copying)
- **Color codes injected conditionally** via ColorScheme (ANSI codes vs empty strings)
- **Three rendering modes**: Render (stdout), RenderDetailedTarget (single target), RenderForMakefile (static @echo)
- **Escape strategy for Makefile generation**: Converts ANSI codes to literal \033 form for @echo compatibility
- **Structured rendering methods** for consistency across all output modes

### 8 Static Help File Generator

**Package:** `internal/target`

**Design:** Generates static help files with embedded help text and auto-regeneration logic. Includes smart file location detection that supports make/ directory patterns, numbered prefixes, and automatic include directive detection.

**Pseudocode:**
```
function GenerateHelpFile(config):
    1. create renderer with color configuration
    2. write header with metadata:
        - generated-by: make-help
        - command: <full command line>
        - date: <UTC timestamp>
    3. write variables (MAKE_HELP_DIR, MAKE_HELP_MAKEFILES)
    4. generate main help target:
        - if has categories: add !category directive
        - .PHONY: help
        - timestamp check (warn if Makefiles newer than help.mk)
        - render help content as @printf '%b\n' statements
    5. generate help-<target> targets for each documented target:
        - .PHONY: help-<target>
        - render detailed help as @printf '%b\n' statements
    6. generate update-help target:
        - tries make-help, npx make-help, then error
    return complete file content

function determineTargetFile(makefilePath, explicitRelPath):
    // Smart file location detection
    if explicitRelPath provided:
        use it (needs include directive)
    else:
        1. scan Makefile for existing include patterns
        2. determine suffix (.mk or custom)
        3. create make/ directory if needed
        4. detect numbered prefix (00-, 10-, etc.)
        5. construct filename: make/<prefix>help<suffix>
        6. needs include only if no pattern found
    return (targetPath, needsInclude)

function addIncludeDirective(makefilePath, targetFile):
    if targetFile in make/ directory:
        if no pattern exists: add "-include make/*.mk"
    else:
        add self-referential include:
            -include $(dir $(lastword $(MAKEFILE_LIST)))<relpath>
    use atomic write (prevents corruption)

function atomicWriteFile(filename, data):
    1. create temp file in same directory
    2. write content to temp file
    3. sync to disk
    4. set permissions
    5. atomic rename (temp -> final)
    cleanup temp file on any error
```

[View source: GenerateHelpFile](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/target/generator.go#L48-L129)
[View source: determineTargetFile logic](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/target/add.go)
[View source: atomicWriteFile](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/target/write.go)

**Key Design Decisions:**
- **Static generation**: Help text is embedded as `@echo` statements, not generated dynamically
- **Auto-regeneration**: Generated file includes target that regenerates when source Makefiles change
- **Fallback chain**: Tries `make-help`, then `npx make-help`, then shows error message
- **Smart file location**: Defaults to make/ directory (./make/help.mk) instead of root directory
- **Include pattern detection**: Scans for existing include directives to match project conventions
- **Numbered prefix support**: Detects numbered files (10-foo.mk) and generates matching prefix (00-help.mk)
- **Suffix detection**: Matches existing file extensions (.mk, no extension, etc.)
- **Pattern vs. specific includes**: Uses -include make/*.mk for make/ directory, self-referential $(dir $(lastword $(MAKEFILE_LIST))) for other locations
- **Include directive injection**: Adds include statement at end of Makefile if needed
- **Atomic writes**: Prevents file corruption on crashes

**Error Handling:**
- File write failures
- Directory creation failures
- Makefile syntax validation before modification

### 9 Remove-Help Service

**Package:** `internal/target`

**Design:** Clean removal of help artifacts

**Pseudocode:**
```
function RemoveTarget():
    1. validate Makefile syntax (make -n with 10s timeout)
    2. removeIncludeDirectives():
        - filter lines matching regex: ^include\s+.*help.*\.mk
        - use atomic write if changes made
    3. removeInlineHelpTarget():
        - detect help target start (.PHONY: help or help:)
        - skip recipe lines (tab/space prefixed)
        - detect end (next target or non-recipe line)
        - use atomic write if changes made
    4. removeHelpTargetFiles():
        - remove help.mk from project root (if exists)
        - remove make/help.mk (if exists)
        - file not found is not an error
```

[View source: RemoveService](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/target/remove.go)

**Key Design Decisions:**
- **Multi-step removal** (include directives, inline targets, help files)
- **Pattern matching** for include directives using regex
- **Recipe detection** for inline target removal (tab/space prefix detection)
- **Atomic writes** for Makefile modifications
- **Silent success** when files don't exist (idempotent operation)

**Error Handling:**
- File not found is not an error (already removed, idempotent)
- Multiple help targets are all removed
- Makefile syntax validated before any modifications

### 10 Lint Service

**Package:** `internal/lint`

**Design:** Validates documentation quality with optional auto-fix capability

**Pseudocode:**
```
type Check:
    Name      - unique identifier (e.g., "summary-punctuation")
    CheckFunc - performs check, returns warnings
    FixFunc   - generates fixes (nil if not auto-fixable)

type Fix:
    File       - absolute path to file to modify
    Line       - 1-indexed line number
    Operation  - FixReplace or FixDelete
    OldContent - expected current content (validation)
    NewContent - replacement content (for FixReplace)

function ApplyFixes(fixes):
    1. group fixes by file
    2. for each file:
        a. sort fixes by line number (descending)
        b. read file content
        c. for each fix (bottom-to-top):
            validate OldContent matches current line
            apply fix (replace or delete)
        d. atomically write modified content
    3. return FixResult with counts
    all fixes to a file succeed or none do (atomic)
```

[View source: Lint Service](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/lint)

**Key Design Decisions:**
- **Check/Fix separation**: Checks can exist without fixes (error-only checks)
- **Atomic file modifications**: All fixes to a file succeed or none do (transactional)
- **Reverse line order**: Fixes applied from bottom to top to preserve line numbers
- **Validation before fixing**: Checks OldContent matches current line to detect file changes
- **Dry-run support**: Preview fixes without modifying files (`--fix --dry-run`)
- **Fix filtering**: Fixed warnings are hidden from output (only unfixed warnings displayed)

**CLI Integration:**
- `--lint`: Run lint checks and display warnings
- `--lint --fix`: Apply auto-fixes for safe issues, display remaining warnings
- `--lint --fix --dry-run`: Preview fixes without modifying files

**Error Handling:**
- Line content mismatch returns error (file changed since check)
- Line number out of range returns error
- File write failures propagate errors
- All-or-nothing per file (atomic commits)

### 11 Version Package

**Package:** `internal/version`

**Design:** Provides build-time version information via ldflags injection

**Implementation:**
```go
// Version is set at build time via ldflags
// Default: "dev"
var Version = "dev"
```

**Build Integration:**
```makefile
VERSION := $(shell node -p "require('./package.json').version")
LDFLAGS := -X github.com/sdlcforge/make-help/internal/version.Version=$(VERSION)

build:
    go build -ldflags "$(LDFLAGS)" -o bin/make-help ./cmd/make-help
```

[View source: version.go](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/version/version.go)

**CLI Integration:**
- `--version`: Display version and exit

**Key Design Decisions:**
- **Single source of truth**: Version comes from package.json (not duplicated in Go code)
- **Ldflags injection**: Version set at build time, no code changes needed for version bumps
- **Default fallback**: Shows "dev" for local development builds (when built without ldflags)
- **Simple implementation**: Single exported variable, minimal code


Last reviewed: 2025-12-25T16:43Z
