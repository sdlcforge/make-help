# Design Document: make-help

## 1. Architecture Overview

### 1.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                          CLI Layer                               │
│  (Cobra-based command parser, flag validation, color detection) │
└───────────────┬─────────────────────────────────────────────────┘
                │
                ├──> help (default command)
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
                │    │  - Template rendering                      │
                │    │  - Color application                       │
                │    │  - Layout formatting                       │
                │    └──────┬─────────────────────────────────────┘
                │           │
                │           ▼
                │       [STDOUT]
                │
                ├──> add-target
                │    ┌────────────────────────────────────────────┐
                │    │  Add-Target Service                        │
                │    │  - Detect include pattern                  │
                │    │  - Generate help target file               │
                │    │  - Inject include directive                │
                │    └────────────────────────────────────────────┘
                │
                └──> remove-target
                     ┌────────────────────────────────────────────┐
                     │  Remove-Target Service                     │
                     │  - Identify help target artifacts          │
                     │  - Remove include directives               │
                     │  - Clean up generated files                │
                     └────────────────────────────────────────────┘
```

### 1.2 Component Responsibilities

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

## 2. Package Structure

```
make-help/
├── cmd/
│   └── genmake/
│       └── main.go                    # Entry point
├── internal/
│   ├── cli/
│   │   ├── root.go                    # Root command setup
│   │   ├── help.go                    # Help command (default)
│   │   ├── add_target.go              # Add-target command
│   │   ├── remove_target.go           # Remove-target command
│   │   └── config.go                  # Shared config struct
│   ├── discovery/
│   │   ├── makefile.go                # Resolve Makefile path
│   │   ├── filelist.go                # MAKEFILE_LIST discovery
│   │   └── targets.go                 # Target discovery (make -p)
│   ├── parser/
│   │   ├── scanner.go                 # Line-by-line scanning
│   │   ├── directive.go               # Directive parsing
│   │   └── association.go             # Associate docs with targets
│   ├── model/
│   │   ├── help.go                    # HelpModel data structures
│   │   ├── builder.go                 # Build model from directives
│   │   └── validator.go               # Validation logic
│   ├── ordering/
│   │   ├── service.go                 # Apply ordering rules
│   │   └── strategy.go                # Ordering strategies
│   ├── summary/
│   │   ├── extractor.go               # Summary extraction logic
│   │   └── markdown.go                # Markdown stripping utilities
│   ├── format/
│   │   ├── renderer.go                # Template-based rendering
│   │   ├── color.go                   # Color utilities
│   │   └── templates.go               # Help output templates
│   ├── target/
│   │   ├── add.go                     # Add-target logic
│   │   ├── remove.go                  # Remove-target logic
│   │   └── generator.go               # Generate help target content
│   └── errors/
│       └── errors.go                  # Custom error types
├── pkg/
│   └── (empty - no public API initially)
├── test/
│   ├── fixtures/
│   │   ├── makefiles/                 # Sample Makefiles
│   │   └── expected/                  # Expected outputs
│   └── integration/
│       └── cli_test.go                # End-to-end tests
├── go.mod
├── go.sum
└── README.md
```

### 2.1 Package Design Rationale

- **`internal/`**: All code is internal (not intended as library) to prevent API commitment
- **`internal/cli/`**: Thin layer, delegates to services; uses Cobra for consistency with Go ecosystem
- **`internal/discovery/`**: Isolates external `make` command execution for testability
- **`internal/parser/`**: Pure functions for parsing; no side effects
- **`internal/model/`**: Central data structures; builder pattern for construction
- **`internal/ordering/`**: Strategy pattern for flexible ordering algorithms
- **`internal/summary/`**: Port of extract-topic; isolated for unit testing
- **`internal/format/`**: Template-based rendering for flexibility and testability
- **`internal/target/`**: File manipulation logic; isolated for safety
- **`internal/errors/`**: Centralized error definitions for consistent handling

### 2.2 Package-Level Godoc Comments

Each package MUST include a doc.go file with package-level documentation:

```go
// internal/cli/doc.go

// Package cli provides the command-line interface for make-help using Cobra.
//
// This package handles argument parsing, flag validation, terminal detection,
// and delegates to the appropriate service packages for actual functionality.
// It is the only package that interacts with os.Args and stdout/stderr.
package cli
```

```go
// internal/discovery/doc.go

// Package discovery handles finding Makefiles and extracting targets.
//
// It uses Make's MAKEFILE_LIST variable to discover all included files and
// the `make -p` database output to enumerate available targets. All external
// command execution uses context with timeout to prevent indefinite hangs.
//
// Security note: This package uses temporary physical files instead of bash
// process substitution to prevent command injection vulnerabilities.
package discovery
```

```go
// internal/parser/doc.go

// Package parser scans Makefile content and extracts documentation directives.
//
// It recognizes the following directive types:
//   - @file: File-level documentation
//   - @category: Category grouping for targets
//   - @var: Environment variable documentation
//   - @alias: Target aliases
//
// The parser maintains state to track the current category and pending
// documentation lines that will be associated with the next target definition.
package parser
```

```go
// internal/model/doc.go

// Package model defines the core data structures for help documentation.
//
// The central type is HelpModel, which aggregates file documentation,
// categories, and targets with their associated aliases and variables.
// The Builder type constructs HelpModel from parsed directives using the
// builder pattern for complex object construction.
package model
```

```go
// internal/ordering/doc.go

// Package ordering applies sorting rules to categories and targets.
//
// It supports three ordering strategies:
//   - Alphabetical (default)
//   - Discovery order (--keep-order-* flags)
//   - Explicit order (--category-order flag)
//
// The package uses the strategy pattern to select the appropriate sorting
// algorithm based on configuration flags.
package ordering
```

```go
// internal/summary/doc.go

// Package summary extracts topic sentences from documentation text.
//
// This is a Go port of the extract-topic JavaScript library. It processes
// text through several stages:
//   1. Strip markdown headers
//   2. Remove markdown formatting (bold, italic, code, links)
//   3. Remove HTML tags
//   4. Normalize whitespace
//   5. Extract first sentence using regex
//
// The sentence extraction handles edge cases like ellipsis (...) and
// IP addresses (127.0.0.1) which should not be treated as sentence endings.
//
// All regex patterns are pre-compiled at Extractor construction time for
// performance when processing many targets.
package summary
```

```go
// internal/format/doc.go

// Package format renders HelpModel as formatted text output.
//
// It supports colorized output using ANSI escape codes, with automatic
// detection of terminal capabilities. Colors can be forced on or off
// via --color and --no-color flags respectively.
//
// The Renderer type uses a template-like approach with string builders
// for efficient concatenation.
package format
```

```go
// internal/target/doc.go

// Package target handles adding and removing help targets from Makefiles.
//
// The add-target command uses a three-tier strategy to determine where
// to place the help target:
//   1. Explicit --target-file path
//   2. make/01-help.mk if include make/*.mk pattern exists
//   3. Append directly to main Makefile
//
// All file modifications use atomic writes (write to temp, then rename)
// to prevent file corruption on process crashes. Makefiles are validated
// with `make -n` before modification to catch syntax errors early.
package target
```

```go
// internal/errors/doc.go

// Package errors defines custom error types for make-help.
//
// Error types include:
//   - MixedCategorizationError: categorized and uncategorized targets mixed
//   - UnknownCategoryError: invalid category in --category-order
//   - MakefileNotFoundError: Makefile not found at specified path
//   - MakeExecutionError: make command failed
//
// All errors provide actionable suggestions in their error messages.
package errors
```

## 3. Core Data Structures

### 3.1 Makefile Documentation Model

```go
// HelpModel represents the complete parsed help documentation
type HelpModel struct {
    FileDocs     []string              // @file documentation sections (ordered)
    Categories   []Category            // Ordered list of categories
    HasCategories bool                 // True if any @category directives found
    DefaultCategory string             // Category for uncategorized targets
}

// Category represents a documentation category
type Category struct {
    Name           string              // Category name
    Targets        []Target            // Targets in this category
    DiscoveryOrder int                 // Order of first appearance
}

// Target represents a documented target
type Target struct {
    Name           string              // Primary target name
    Aliases        []string            // Alternative names (@alias)
    Documentation  []string            // Full documentation lines
    Summary        string              // Extracted summary (computed)
    Variables      []Variable          // Associated environment variables
    DiscoveryOrder int                 // Order of first appearance
    SourceFile     string              // File where target was documented
    LineNumber     int                 // Line number for error reporting
}

// Variable represents a documented environment variable
type Variable struct {
    Name          string               // Variable name
    Description   string               // Full description text
}

// Directive represents a parsed documentation directive
type Directive struct {
    Type          DirectiveType        // @file, @category, @var, @alias, or doc
    Value         string               // Directive value/content
    SourceFile    string               // File where directive appears
    LineNumber    int                  // Line number
}

type DirectiveType int

const (
    DirectiveFile DirectiveType = iota
    DirectiveCategory
    DirectiveVar
    DirectiveAlias
    DirectiveDoc  // Regular documentation line
)
```

### 3.2 Configuration Model

```go
// Config holds all CLI configuration
type Config struct {
    // Global options
    MakefilePath  string              // Path to Makefile (resolved absolute)
    ColorMode     ColorMode           // Auto, Always, Never
    Verbose       bool                // Enable verbose output for debugging

    // Help generation options
    KeepOrderCategories bool          // Preserve category discovery order
    KeepOrderTargets    bool          // Preserve target discovery order
    CategoryOrder       []string      // Explicit category order
    DefaultCategory     string        // Default category name

    // Add-target options
    TargetFile    string              // Explicit path for help target file

    // Derived state
    UseColor      bool                // Computed based on ColorMode + terminal detection
}

type ColorMode int

const (
    ColorAuto ColorMode = iota
    ColorAlways
    ColorNever
)
```

### 3.3 Discovery Model

```go
// MakefileInfo holds discovered Makefile information
type MakefileInfo struct {
    MainFile      string              // Main Makefile path (absolute)
    IncludedFiles []string            // Included files in MAKEFILE_LIST order
    AllTargets    []string            // All available targets from make -p
}

// ParsedFile represents a single parsed Makefile
type ParsedFile struct {
    Path          string              // File path
    Directives    []Directive         // Parsed directives
    TargetMap     map[string]int      // Target name -> first line number
}
```

### 3.4 Rendering Model

```go
// RenderContext holds data for template rendering
type RenderContext struct {
    Model         *HelpModel
    Config        *Config
    ColorScheme   *ColorScheme        // Color codes based on UseColor
}

// ColorScheme defines color codes for different elements
type ColorScheme struct {
    CategoryName  string              // ANSI code for category headers
    TargetName    string              // ANSI code for target names
    Alias         string              // ANSI code for aliases
    Variable      string              // ANSI code for variable names
    Documentation string              // ANSI code for doc text
    Reset         string              // ANSI reset code
}

// NewColorScheme creates a scheme based on color mode
func NewColorScheme(useColor bool) *ColorScheme {
    if !useColor {
        return &ColorScheme{} // All empty strings
    }
    return &ColorScheme{
        CategoryName:  "\033[1;36m",  // Bold Cyan
        TargetName:    "\033[1;32m",  // Bold Green
        Alias:         "\033[0;33m",  // Yellow
        Variable:      "\033[0;35m",  // Magenta
        Documentation: "\033[0;37m",  // White
        Reset:         "\033[0m",     // Reset
    }
}
```

## 4. Major Components

### 4.1 CLI Parser (Cobra-based)

**Package:** `internal/cli`

**Design:** Use spf13/cobra for standard Go CLI patterns

```go
// Root command setup
func NewRootCmd() *cobra.Command {
    var config Config

    rootCmd := &cobra.Command{
        Use:   "make-help",
        Short: "Dynamic help generation for Makefiles",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runHelp(&config)
        },
    }

    // Global flags
    rootCmd.PersistentFlags().StringVar(&config.MakefilePath,
        "makefile-path", "", "Path to Makefile")
    rootCmd.PersistentFlags().BoolVar(&noColor,
        "no-color", false, "Disable colored output")
    rootCmd.PersistentFlags().BoolVar(&forceColor,
        "color", false, "Force colored output")
    rootCmd.PersistentFlags().BoolVar(&config.Verbose,
        "verbose", false, "Enable verbose output for debugging file discovery and parsing")

    // Help command flags (on root)
    rootCmd.Flags().BoolVar(&config.KeepOrderCategories,
        "keep-order-categories", false, "Preserve category discovery order")
    // ... other flags

    // Subcommands
    rootCmd.AddCommand(newAddTargetCmd(&config))
    rootCmd.AddCommand(newRemoveTargetCmd(&config))

    return rootCmd
}
```

**Responsibilities:**
- Parse command-line arguments
- Validate flag combinations
- Detect terminal capabilities (isatty)
- Resolve color mode
- Delegate to appropriate service

**Error Handling:**
- Invalid flag combinations (e.g., `--keep-order-all` + `--category-order`)
- File path validation
- Conflicting color flags

### 4.2 Discovery Service

**Package:** `internal/discovery`

**Design:** Execute make commands and parse output

```go
type Service struct {
    executor CommandExecutor  // Interface for testability
    verbose  bool             // Enable verbose output
}

type CommandExecutor interface {
    Execute(cmd string, args ...string) (stdout, stderr string, err error)
    ExecuteContext(ctx context.Context, cmd string, args ...string) (stdout, stderr string, err error)
}

// DiscoverMakefiles finds all Makefiles using MAKEFILE_LIST
// SECURITY: Uses temporary physical file instead of bash process substitution
// to prevent command injection vulnerabilities.
func (s *Service) DiscoverMakefiles(mainPath string) ([]string, error) {
    // Read main Makefile content
    mainContent, err := os.ReadFile(mainPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read Makefile: %w", err)
    }

    // Create temporary file with appended _list_makefiles target
    tmpFile, err := os.CreateTemp("", "makefile-discovery-*.mk")
    if err != nil {
        return nil, fmt.Errorf("failed to create temp file: %w", err)
    }
    defer os.Remove(tmpFile.Name())

    // Write main content + discovery target
    discoveryTarget := "\n\n.PHONY: _list_makefiles\n_list_makefiles:\n\t@echo $(MAKEFILE_LIST)\n"
    if _, err := tmpFile.Write(mainContent); err != nil {
        tmpFile.Close()
        return nil, fmt.Errorf("failed to write temp file: %w", err)
    }
    if _, err := tmpFile.WriteString(discoveryTarget); err != nil {
        tmpFile.Close()
        return nil, fmt.Errorf("failed to write temp file: %w", err)
    }
    tmpFile.Close()

    // Execute make with timeout to prevent indefinite hangs
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    stdout, stderr, err := s.executor.ExecuteContext(ctx, "make", "-f", tmpFile.Name(), "_list_makefiles")
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return nil, fmt.Errorf("make command timed out after 30s")
        }
        return nil, fmt.Errorf("failed to discover makefiles: %w\nstderr: %s", err, stderr)
    }

    // Parse space-separated file list
    files := strings.Fields(stdout)

    // Resolve to absolute paths
    resolved, err := resolveAbsolutePaths(files, filepath.Dir(mainPath))
    if err != nil {
        return nil, err
    }

    if s.verbose {
        fmt.Printf("Discovered %d Makefiles:\n", len(resolved))
        for i, f := range resolved {
            fmt.Printf("  %d. %s\n", i+1, f)
        }
    }

    return resolved, nil
}

// DiscoverTargets extracts all targets from make -p output
func (s *Service) DiscoverTargets(makefilePath string) ([]string, error) {
    // Execute make with timeout to prevent indefinite hangs
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    stdout, stderr, err := s.executor.ExecuteContext(ctx, "make", "-f", makefilePath, "-p", "-r")
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return nil, fmt.Errorf("make command timed out after 30s")
        }
        return nil, fmt.Errorf("failed to discover targets: %w\nstderr: %s", err, stderr)
    }

    targets := parseTargetsFromDatabase(stdout)

    if s.verbose {
        fmt.Printf("Discovered %d targets from make database\n", len(targets))
    }

    return targets, nil
}

// parseTargetsFromDatabase extracts target names from make -p output
func parseTargetsFromDatabase(output string) []string {
    var targets []string
    targetRegex := regexp.MustCompile(`^([a-zA-Z0-9_.-]+):`)

    for _, line := range strings.Split(output, "\n") {
        // Skip comments and whitespace-prefixed lines
        if strings.HasPrefix(line, "#") || strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
            continue
        }

        if matches := targetRegex.FindStringSubmatch(line); matches != nil {
            targets = append(targets, matches[1])
        }
    }

    return targets
}
```

**Key Algorithms:**
- MAKEFILE_LIST discovery via temporary target injection
- Target extraction via regex from `make -p` database output
- Absolute path resolution for included files

**Error Handling:**
- Make command execution failures
- Makefile not found
- Invalid Makefile syntax (caught by make)
- Shell command injection prevention (use exec.Command, not shell)

### 4.3 Parser Service

**Package:** `internal/parser`

**Design:** Stateful scanner with directive detection

```go
type Scanner struct {
    currentFile     string
    currentCategory string
    pendingDocs     []Directive  // Documentation lines awaiting target
}

// ScanFile parses a single Makefile and extracts directives
func (s *Scanner) ScanFile(path string) (*ParsedFile, error) {
    content, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read %s: %w", path, err)
    }

    s.currentFile = path
    s.currentCategory = ""
    s.pendingDocs = nil

    result := &ParsedFile{
        Path:       path,
        Directives: []Directive{},
        TargetMap:  make(map[string]int),
    }

    lines := strings.Split(string(content), "\n")

    for lineNum, line := range lines {
        // Check for documentation line
        if strings.HasPrefix(line, "## ") {
            directive := s.parseDirective(line, lineNum+1)
            result.Directives = append(result.Directives, directive)
            continue
        }

        // Check for target definition
        if target := s.parseTarget(line); target != "" {
            result.TargetMap[target] = lineNum + 1

            // Associate pending docs with target
            if len(s.pendingDocs) > 0 {
                result.Directives = append(result.Directives, s.pendingDocs...)
                s.pendingDocs = nil
            }
        } else {
            // Non-doc, non-target line clears pending docs
            s.pendingDocs = nil
        }
    }

    return result, nil
}

// parseDirective detects and parses a documentation directive
func (s *Scanner) parseDirective(line string, lineNum int) Directive {
    content := strings.TrimPrefix(line, "## ")

    directive := Directive{
        SourceFile: s.currentFile,
        LineNumber: lineNum,
    }

    switch {
    case strings.HasPrefix(content, "@file"):
        directive.Type = DirectiveFile
        directive.Value = strings.TrimSpace(strings.TrimPrefix(content, "@file"))

    case strings.HasPrefix(content, "@category "):
        directive.Type = DirectiveCategory
        directive.Value = strings.TrimSpace(strings.TrimPrefix(content, "@category "))
        s.currentCategory = directive.Value

    case strings.HasPrefix(content, "@var "):
        directive.Type = DirectiveVar
        directive.Value = strings.TrimSpace(strings.TrimPrefix(content, "@var "))

    case strings.HasPrefix(content, "@alias "):
        directive.Type = DirectiveAlias
        directive.Value = strings.TrimSpace(strings.TrimPrefix(content, "@alias "))

    default:
        directive.Type = DirectiveDoc
        directive.Value = content
    }

    // Queue for association with next target
    if directive.Type != DirectiveFile {
        s.pendingDocs = append(s.pendingDocs, directive)
    }

    return directive
}

// parseTarget extracts target name from a target definition line
func (s *Scanner) parseTarget(line string) string {
    // Match: <target>: or <target>&:
    // Handle grouped targets: foo bar baz:
    // Handle variable targets: $(VAR):

    colonIdx := strings.Index(line, ":")
    if colonIdx == -1 {
        return ""
    }

    // Check for &: (grouped target)
    beforeColon := line[:colonIdx]
    if strings.HasSuffix(beforeColon, "&") {
        beforeColon = strings.TrimSuffix(beforeColon, "&")
    }

    // Extract first word/token as target name
    targetPart := strings.TrimSpace(beforeColon)

    // Handle whitespace-prefixed (recipe lines)
    if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
        return ""
    }

    // Extract first token
    fields := strings.Fields(targetPart)
    if len(fields) > 0 {
        return fields[0]
    }

    return ""
}
```

**Key Design Decisions:**
- Stateful scanning to track current category
- Pending documentation queue (associated with next target)
- Simple regex-free parsing for robustness
- Target name extraction handles grouped and variable targets

**Error Handling:**
- Invalid directive syntax (log warning, skip)
- Malformed @var or @alias (log warning, skip)

### 4.4 Model Builder

**Package:** `internal/model`

**Design:** Aggregate directives into structured model

```go
type Builder struct {
    config *cli.Config
}

// Build constructs HelpModel from parsed files
func (b *Builder) Build(parsedFiles []*parser.ParsedFile) (*HelpModel, error) {
    model := &HelpModel{
        FileDocs:   []string{},
        Categories: []Category{},
    }

    categoryMap := make(map[string]*Category)  // Name -> Category
    targetMap := make(map[string]*Target)       // Name -> Target

    discoveryOrder := 0

    for _, file := range parsedFiles {
        var currentCategory *Category
        var currentTarget *Target

        for _, directive := range file.Directives {
            switch directive.Type {
            case parser.DirectiveFile:
                if directive.Value != "" {
                    model.FileDocs = append(model.FileDocs, directive.Value)
                }

            case parser.DirectiveCategory:
                model.HasCategories = true

                // Find or create category
                cat, exists := categoryMap[directive.Value]
                if !exists {
                    cat = &Category{
                        Name:           directive.Value,
                        Targets:        []Target{},
                        DiscoveryOrder: discoveryOrder,
                    }
                    discoveryOrder++
                    categoryMap[directive.Value] = cat
                }
                currentCategory = cat

            case parser.DirectiveDoc:
                if currentTarget != nil {
                    currentTarget.Documentation = append(
                        currentTarget.Documentation,
                        directive.Value,
                    )
                }

            case parser.DirectiveVar:
                if currentTarget != nil {
                    varData := b.parseVarDirective(directive.Value)
                    currentTarget.Variables = append(currentTarget.Variables, varData)
                }

            case parser.DirectiveAlias:
                if currentTarget != nil {
                    aliases := b.parseAliasDirective(directive.Value)
                    currentTarget.Aliases = append(currentTarget.Aliases, aliases...)
                }
            }
        }

        // Associate targets with categories
        for targetName, lineNum := range file.TargetMap {
            target, exists := targetMap[targetName]
            if !exists {
                target = &Target{
                    Name:           targetName,
                    DiscoveryOrder: discoveryOrder,
                    SourceFile:     file.Path,
                    LineNumber:     lineNum,
                }
                discoveryOrder++
                targetMap[targetName] = target
            }

            if currentCategory != nil {
                currentCategory.Targets = append(currentCategory.Targets, *target)
            }
        }
    }

    // Validate mixed categorization
    if err := b.validateCategorization(model, targetMap); err != nil {
        return nil, err
    }

    // Apply default category if needed
    if model.HasCategories && b.config.DefaultCategory != "" {
        b.applyDefaultCategory(model, targetMap, categoryMap)
    }

    // Convert map to slice
    for _, cat := range categoryMap {
        model.Categories = append(model.Categories, *cat)
    }

    return model, nil
}

// validateCategorization ensures no mixing of categorized and uncategorized targets
func (b *Builder) validateCategorization(model *HelpModel, targetMap map[string]*Target) error {
    if !model.HasCategories {
        return nil  // No categories, all targets uncategorized - OK
    }

    categorizedTargets := 0
    uncategorizedTargets := 0

    for _, target := range targetMap {
        hasCategory := false
        for _, cat := range model.Categories {
            for _, catTarget := range cat.Targets {
                if catTarget.Name == target.Name {
                    hasCategory = true
                    break
                }
            }
        }

        if hasCategory {
            categorizedTargets++
        } else {
            uncategorizedTargets++
        }
    }

    if categorizedTargets > 0 && uncategorizedTargets > 0 {
        if b.config.DefaultCategory == "" {
            return errors.NewMixedCategorizationError(
                "found both categorized and uncategorized targets; use --default-category to resolve",
            )
        }
    }

    return nil
}

// parseVarDirective parses @var directive: <NAME> - <description>
func (b *Builder) parseVarDirective(value string) Variable {
    parts := strings.SplitN(value, " - ", 2)
    if len(parts) != 2 {
        return Variable{Name: value, Description: ""}
    }
    return Variable{
        Name:        strings.TrimSpace(parts[0]),
        Description: strings.TrimSpace(parts[1]),
    }
}

// parseAliasDirective parses @alias directive: <name>[, <name>...]
func (b *Builder) parseAliasDirective(value string) []string {
    parts := strings.Split(value, ",")
    aliases := make([]string, 0, len(parts))
    for _, part := range parts {
        if alias := strings.TrimSpace(part); alias != "" {
            aliases = append(aliases, alias)
        }
    }
    return aliases
}
```

**Key Design Decisions:**
- Builder pattern for complex construction
- Discovery order tracking via counter
- Validation of categorization rules
- Default category application

**Error Handling:**
- Mixed categorization without default category (CRITICAL)
- Duplicate category definitions (merge targets)
- Invalid directive format (log warning, best-effort parse)

### 4.5 Ordering Service

**Package:** `internal/ordering`

**Design:** Strategy pattern for flexible ordering

```go
type Service struct {
    config *cli.Config
}

// ApplyOrdering sorts categories and targets based on config
func (s *Service) ApplyOrdering(model *HelpModel) error {
    // Apply category ordering
    if err := s.orderCategories(model); err != nil {
        return err
    }

    // Apply target ordering within each category
    for i := range model.Categories {
        s.orderTargets(&model.Categories[i])
    }

    return nil
}

// orderCategories applies category ordering rules
func (s *Service) orderCategories(model *HelpModel) error {
    if len(s.config.CategoryOrder) > 0 {
        return s.applyExplicitCategoryOrder(model)
    }

    if s.config.KeepOrderCategories {
        s.sortByDiscoveryOrder(model.Categories)
    } else {
        s.sortAlphabetically(model.Categories)
    }

    return nil
}

// applyExplicitCategoryOrder uses --category-order flag
func (s *Service) applyExplicitCategoryOrder(model *HelpModel) error {
    categoryMap := make(map[string]*Category)
    for i := range model.Categories {
        categoryMap[model.Categories[i].Name] = &model.Categories[i]
    }

    ordered := make([]Category, 0, len(model.Categories))
    remaining := make(map[string]*Category)
    for k, v := range categoryMap {
        remaining[k] = v
    }

    // Add categories in specified order
    for _, name := range s.config.CategoryOrder {
        cat, exists := categoryMap[name]
        if !exists {
            return fmt.Errorf("category %q specified in --category-order not found", name)
        }
        ordered = append(ordered, *cat)
        delete(remaining, name)
    }

    // Append remaining categories alphabetically
    remainingSlice := make([]Category, 0, len(remaining))
    for _, cat := range remaining {
        remainingSlice = append(remainingSlice, *cat)
    }
    s.sortAlphabetically(remainingSlice)
    ordered = append(ordered, remainingSlice...)

    model.Categories = ordered
    return nil
}

// orderTargets applies target ordering rules
func (s *Service) orderTargets(category *Category) {
    if s.config.KeepOrderTargets {
        s.sortTargetsByDiscoveryOrder(category.Targets)
    } else {
        s.sortTargetsAlphabetically(category.Targets)
    }
}

// sortAlphabetically sorts categories by name
func (s *Service) sortAlphabetically(categories []Category) {
    sort.Slice(categories, func(i, j int) bool {
        return categories[i].Name < categories[j].Name
    })
}

// sortByDiscoveryOrder sorts categories by first appearance
func (s *Service) sortByDiscoveryOrder(categories []Category) {
    sort.Slice(categories, func(i, j int) bool {
        return categories[i].DiscoveryOrder < categories[j].DiscoveryOrder
    })
}

// sortTargetsAlphabetically sorts targets by name
func (s *Service) sortTargetsAlphabetically(targets []Target) {
    sort.Slice(targets, func(i, j int) bool {
        return targets[i].Name < targets[j].Name
    })
}

// sortTargetsByDiscoveryOrder sorts targets by first appearance
func (s *Service) sortTargetsByDiscoveryOrder(targets []Target) {
    sort.Slice(targets, func(i, j int) bool {
        return targets[i].DiscoveryOrder < targets[j].DiscoveryOrder
    })
}
```

**Key Design Decisions:**
- Clear separation of category vs target ordering
- Explicit category order with alphabetical fallback
- Error on unknown category in explicit order

**Error Handling:**
- Unknown category in `--category-order` (CRITICAL)
- Validate all specified categories exist before applying

### 4.6 Summary Extractor

**Package:** `internal/summary`

**Design:** Port of extract-topic algorithm

```go
// Extractor pre-compiles all regex patterns at construction time for performance.
// This avoids repeated regex compilation when processing many targets.
type Extractor struct {
    sentenceRegex    *regexp.Regexp
    headerRegex      *regexp.Regexp
    boldRegex        *regexp.Regexp
    italicRegex      *regexp.Regexp
    boldUnderRegex   *regexp.Regexp
    italicUnderRegex *regexp.Regexp
    codeRegex        *regexp.Regexp
    linkRegex        *regexp.Regexp
    htmlTagRegex     *regexp.Regexp
    whitespaceRegex  *regexp.Regexp
}

func NewExtractor() *Extractor {
    return &Extractor{
        // Regex from extract-topic: first sentence ending in .!?
        // Handles: ellipsis (...), IPs (127.0.0.1.), abbreviations
        sentenceRegex:    regexp.MustCompile(`^((?:[^.!?]|\.\.\.|\.[^\s])+[.?!])(\s|$)`),
        headerRegex:      regexp.MustCompile(`(?m)^#+\s+`),
        boldRegex:        regexp.MustCompile(`\*\*([^*]+)\*\*`),
        italicRegex:      regexp.MustCompile(`\*([^*]+)\*`),
        boldUnderRegex:   regexp.MustCompile(`__([^_]+)__`),
        italicUnderRegex: regexp.MustCompile(`_([^_]+)_`),
        codeRegex:        regexp.MustCompile("`([^`]+)`"),
        linkRegex:        regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`),
        htmlTagRegex:     regexp.MustCompile(`<[^>]+>`),
        whitespaceRegex:  regexp.MustCompile(`\s+`),
    }
}

// Extract generates summary from full documentation
func (e *Extractor) Extract(documentation []string) string {
    if len(documentation) == 0 {
        return ""
    }

    // Join all documentation lines
    fullText := strings.Join(documentation, " ")

    // Strip markdown headers
    fullText = e.stripMarkdownHeaders(fullText)

    // Strip markdown formatting
    fullText = e.stripMarkdownFormatting(fullText)

    // Strip HTML tags
    fullText = e.stripHTMLTags(fullText)

    // Normalize whitespace
    fullText = e.normalizeWhitespace(fullText)

    // Extract first sentence
    return e.extractFirstSentence(fullText)
}

// stripMarkdownHeaders removes # headers (uses pre-compiled regex)
func (e *Extractor) stripMarkdownHeaders(text string) string {
    return e.headerRegex.ReplaceAllString(text, "")
}

// stripMarkdownFormatting removes **bold**, *italic*, `code`, [links]
// All regexes are pre-compiled for performance
func (e *Extractor) stripMarkdownFormatting(text string) string {
    // Remove bold/italic (order matters: ** before *, __ before _)
    text = e.boldRegex.ReplaceAllString(text, "$1")
    text = e.italicRegex.ReplaceAllString(text, "$1")
    text = e.boldUnderRegex.ReplaceAllString(text, "$1")
    text = e.italicUnderRegex.ReplaceAllString(text, "$1")

    // Remove inline code
    text = e.codeRegex.ReplaceAllString(text, "$1")

    // Remove links [text](url) -> text
    text = e.linkRegex.ReplaceAllString(text, "$1")

    return text
}

// stripHTMLTags removes HTML tags (uses pre-compiled regex)
func (e *Extractor) stripHTMLTags(text string) string {
    return e.htmlTagRegex.ReplaceAllString(text, "")
}

// normalizeWhitespace collapses multiple spaces and newlines (uses pre-compiled regex)
func (e *Extractor) normalizeWhitespace(text string) string {
    // Replace newlines with spaces
    text = strings.ReplaceAll(text, "\n", " ")

    // Collapse multiple spaces
    text = e.whitespaceRegex.ReplaceAllString(text, " ")

    return strings.TrimSpace(text)
}

// extractFirstSentence uses regex to extract first sentence
func (e *Extractor) extractFirstSentence(text string) string {
    matches := e.sentenceRegex.FindStringSubmatch(text)
    if len(matches) > 1 {
        return strings.TrimSpace(matches[1])
    }

    // No sentence ending found, return full text
    return text
}
```

**Key Design Decisions:**
- Faithful port of extract-topic algorithm
- Regex handles edge cases (ellipsis, IPs)
- Fallback to full text if no sentence boundary found

**Testing:** Critical to unit test with:
- Standard sentences
- Ellipsis handling
- IP address handling
- Markdown formatting
- HTML tags
- Multiple sentences (ensure only first extracted)

### 4.7 Formatter Service

**Package:** `internal/format`

**Design:** Template-based rendering with color support

```go
type Renderer struct {
    config *cli.Config
    colors *ColorScheme
}

func NewRenderer(config *cli.Config) *Renderer {
    return &Renderer{
        config: config,
        colors: NewColorScheme(config.UseColor),
    }
}

// Render generates help output from model
func (r *Renderer) Render(model *HelpModel) (string, error) {
    var buf strings.Builder

    // Render usage line
    buf.WriteString("Usage: make [<target>...] [<ENV_VAR>=<value>...]\n\n")

    // Render file documentation
    if len(model.FileDocs) > 0 {
        for _, doc := range model.FileDocs {
            buf.WriteString(doc)
            buf.WriteString("\n")
        }
        buf.WriteString("\n")
    }

    // Render targets header
    buf.WriteString("Targets:\n")

    // Render categories
    for _, category := range model.Categories {
        r.renderCategory(&buf, &category)
    }

    return buf.String(), nil
}

// renderCategory renders a single category section
func (r *Renderer) renderCategory(buf *strings.Builder, category *Category) {
    // Category header (if named)
    if category.Name != "" {
        buf.WriteString("\n")
        buf.WriteString(r.colors.CategoryName)
        buf.WriteString(category.Name)
        buf.WriteString(":")
        buf.WriteString(r.colors.Reset)
        buf.WriteString("\n")
    }

    // Render targets
    for _, target := range category.Targets {
        r.renderTarget(buf, &target)
    }
}

// renderTarget renders a single target entry
func (r *Renderer) renderTarget(buf *strings.Builder, target *Target) {
    // Indent
    buf.WriteString("  - ")

    // Target name
    buf.WriteString(r.colors.TargetName)
    buf.WriteString(target.Name)
    buf.WriteString(r.colors.Reset)

    // Aliases
    if len(target.Aliases) > 0 {
        buf.WriteString(" ")
        buf.WriteString(r.colors.Alias)
        buf.WriteString(strings.Join(target.Aliases, ", "))
        buf.WriteString(r.colors.Reset)
    }

    // Summary
    if target.Summary != "" {
        buf.WriteString(": ")
        buf.WriteString(r.colors.Documentation)
        buf.WriteString(target.Summary)
        buf.WriteString(r.colors.Reset)
    }

    buf.WriteString("\n")

    // Variables
    if len(target.Variables) > 0 {
        buf.WriteString("    Vars: ")
        varNames := make([]string, len(target.Variables))
        for i, v := range target.Variables {
            varNames[i] = v.Name
        }
        buf.WriteString(r.colors.Variable)
        buf.WriteString(strings.Join(varNames, ", "))
        buf.WriteString(r.colors.Reset)
        buf.WriteString("\n")
    }
}

// RenderDetailedTarget renders help-<target> detailed view
func (r *Renderer) RenderDetailedTarget(target *Target) string {
    var buf strings.Builder

    // Target name
    buf.WriteString(r.colors.TargetName)
    buf.WriteString(target.Name)
    buf.WriteString(r.colors.Reset)
    buf.WriteString("\n\n")

    // Full documentation
    for _, line := range target.Documentation {
        buf.WriteString(line)
        buf.WriteString("\n")
    }

    // Aliases section
    if len(target.Aliases) > 0 {
        buf.WriteString("\nAliases: ")
        buf.WriteString(r.colors.Alias)
        buf.WriteString(strings.Join(target.Aliases, ", "))
        buf.WriteString(r.colors.Reset)
        buf.WriteString("\n")
    }

    // Variables section
    if len(target.Variables) > 0 {
        buf.WriteString("\nVariables:\n")
        for _, v := range target.Variables {
            buf.WriteString("  ")
            buf.WriteString(r.colors.Variable)
            buf.WriteString(v.Name)
            buf.WriteString(r.colors.Reset)
            buf.WriteString(": ")
            buf.WriteString(v.Description)
            buf.WriteString("\n")
        }
    }

    return buf.String()
}
```

**Key Design Decisions:**
- String builder for efficient concatenation
- Color codes injected conditionally
- Separate methods for main help vs detailed target help
- Template-like rendering for consistency

### 4.8 Add-Target Service

**Package:** `internal/target`

**Design:** File generation and injection

```go
type AddService struct {
    config   *cli.Config
    executor CommandExecutor // For Makefile validation
    verbose  bool            // Enable verbose output
}

// AddTarget generates and injects help target into Makefile
func (s *AddService) AddTarget() error {
    makefilePath := s.config.MakefilePath

    // Validate Makefile syntax before modifying
    if err := s.validateMakefile(makefilePath); err != nil {
        return fmt.Errorf("Makefile validation failed: %w", err)
    }

    // Determine target file location
    targetFile, needsInclude, err := s.determineTargetFile(makefilePath)
    if err != nil {
        return err
    }

    // Generate help target content
    content := s.generateHelpTarget()

    // Write target file using atomic write (write to temp, then rename)
    if err := atomicWriteFile(targetFile, []byte(content), 0644); err != nil {
        return fmt.Errorf("failed to write target file %s: %w", targetFile, err)
    }

    if s.verbose {
        fmt.Printf("Created help target file: %s\n", targetFile)
    }

    // Add include directive if needed
    if needsInclude {
        if err := s.addIncludeDirective(makefilePath, targetFile); err != nil {
            return err
        }
        if s.verbose {
            fmt.Printf("Added include directive to: %s\n", makefilePath)
        }
    }

    return nil
}

// validateMakefile runs `make -n` to check for syntax errors
func (s *AddService) validateMakefile(makefilePath string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Run make -n (dry-run) to check syntax without executing recipes
    _, stderr, err := s.executor.ExecuteContext(ctx, "make", "-n", "-f", makefilePath)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return fmt.Errorf("validation timed out")
        }
        return fmt.Errorf("syntax error in Makefile:\n%s", stderr)
    }
    return nil
}

// atomicWriteFile writes data to a file atomically by writing to a temp file
// first, then renaming. This prevents file corruption if the process crashes.
func atomicWriteFile(filename string, data []byte, perm os.FileMode) error {
    // Create temp file in same directory (required for atomic rename)
    dir := filepath.Dir(filename)
    tmpFile, err := os.CreateTemp(dir, ".tmp-*")
    if err != nil {
        return fmt.Errorf("failed to create temp file: %w", err)
    }
    tmpName := tmpFile.Name()

    // Clean up temp file on any error
    success := false
    defer func() {
        if !success {
            os.Remove(tmpName)
        }
    }()

    // Write content
    if _, err := tmpFile.Write(data); err != nil {
        tmpFile.Close()
        return fmt.Errorf("failed to write temp file: %w", err)
    }

    // Sync to disk
    if err := tmpFile.Sync(); err != nil {
        tmpFile.Close()
        return fmt.Errorf("failed to sync temp file: %w", err)
    }

    if err := tmpFile.Close(); err != nil {
        return fmt.Errorf("failed to close temp file: %w", err)
    }

    // Set permissions
    if err := os.Chmod(tmpName, perm); err != nil {
        return fmt.Errorf("failed to set permissions: %w", err)
    }

    // Atomic rename
    if err := os.Rename(tmpName, filename); err != nil {
        return fmt.Errorf("failed to rename temp file: %w", err)
    }

    success = true
    return nil
}

// determineTargetFile decides where to create help target
func (s *AddService) determineTargetFile(makefilePath string) (targetFile string, needsInclude bool, err error) {
    // 1. Explicit --target-file
    if s.config.TargetFile != "" {
        return s.config.TargetFile, true, nil
    }

    // 2. Check for include make/*.mk pattern
    content, err := os.ReadFile(makefilePath)
    if err != nil {
        return "", false, fmt.Errorf("failed to read Makefile: %w", err)
    }

    includeRegex := regexp.MustCompile(`(?m)^include\s+make/\*\.mk`)
    if includeRegex.Match(content) {
        // Create make/01-help.mk
        makeDir := filepath.Join(filepath.Dir(makefilePath), "make")
        if err := os.MkdirAll(makeDir, 0755); err != nil {
            return "", false, fmt.Errorf("failed to create make/ directory: %w", err)
        }
        return filepath.Join(makeDir, "01-help.mk"), false, nil
    }

    // 3. Append directly to Makefile
    return makefilePath, false, nil
}

// generateHelpTarget creates help target content
func (s *AddService) generateHelpTarget() string {
    var buf strings.Builder

    buf.WriteString(".PHONY: help\n")
    buf.WriteString("help:\n")
    buf.WriteString("\t@make-help")

    // Add flags from config
    if s.config.KeepOrderCategories {
        buf.WriteString(" --keep-order-categories")
    }
    if s.config.KeepOrderTargets {
        buf.WriteString(" --keep-order-targets")
    }
    if len(s.config.CategoryOrder) > 0 {
        buf.WriteString(" --category-order ")
        buf.WriteString(strings.Join(s.config.CategoryOrder, ","))
    }
    if s.config.DefaultCategory != "" {
        buf.WriteString(" --default-category ")
        buf.WriteString(s.config.DefaultCategory)
    }

    buf.WriteString("\n")

    return buf.String()
}

// addIncludeDirective injects include statement into Makefile using atomic write
func (s *AddService) addIncludeDirective(makefilePath, targetFile string) error {
    content, err := os.ReadFile(makefilePath)
    if err != nil {
        return err
    }

    // Make path relative to Makefile
    relPath, err := filepath.Rel(filepath.Dir(makefilePath), targetFile)
    if err != nil {
        return err
    }

    includeDirective := fmt.Sprintf("\ninclude %s\n", relPath)

    // Append to end of file
    newContent := append(content, []byte(includeDirective)...)

    // Use atomic write to prevent corruption
    return atomicWriteFile(makefilePath, newContent, 0644)
}
```

**Key Design Decisions:**
- Three-tier target file resolution strategy
- Flag pass-through from add-target to generated help command
- Include directive injection at end of Makefile
- Directory creation for make/ pattern

**Error Handling:**
- File write failures
- Directory creation failures
- Duplicate help target detection (check before adding)

### 4.9 Remove-Target Service

**Package:** `internal/target`

**Design:** Clean removal of help artifacts

```go
type RemoveService struct {
    config   *cli.Config
    executor CommandExecutor // For Makefile validation
    verbose  bool            // Enable verbose output
}

// RemoveTarget removes help target artifacts
func (s *RemoveService) RemoveTarget() error {
    makefilePath := s.config.MakefilePath

    // Validate Makefile syntax before modifying
    if err := s.validateMakefile(makefilePath); err != nil {
        return fmt.Errorf("Makefile validation failed: %w", err)
    }

    // Find and remove include directives
    if err := s.removeIncludeDirectives(makefilePath); err != nil {
        return err
    }

    // Find and remove inline help target
    if err := s.removeInlineHelpTarget(makefilePath); err != nil {
        return err
    }

    // Remove help target files (make/01-help.mk or custom)
    if err := s.removeHelpTargetFiles(makefilePath); err != nil {
        return err
    }

    return nil
}

// validateMakefile runs `make -n` to check for syntax errors
func (s *RemoveService) validateMakefile(makefilePath string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    _, stderr, err := s.executor.ExecuteContext(ctx, "make", "-n", "-f", makefilePath)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            return fmt.Errorf("validation timed out")
        }
        return fmt.Errorf("syntax error in Makefile:\n%s", stderr)
    }
    return nil
}

// removeIncludeDirectives removes include lines for help targets using atomic write
func (s *RemoveService) removeIncludeDirectives(makefilePath string) error {
    content, err := os.ReadFile(makefilePath)
    if err != nil {
        return err
    }

    lines := strings.Split(string(content), "\n")
    filtered := []string{}

    includeRegex := regexp.MustCompile(`^include\s+.*help.*\.mk`)
    removed := false

    for _, line := range lines {
        if !includeRegex.MatchString(line) {
            filtered = append(filtered, line)
        } else {
            removed = true
            if s.verbose {
                fmt.Printf("Removed include directive: %s\n", line)
            }
        }
    }

    if !removed {
        return nil // No changes needed
    }

    newContent := strings.Join(filtered, "\n")
    return atomicWriteFile(makefilePath, []byte(newContent), 0644)
}

// removeInlineHelpTarget removes help target from Makefile using atomic write
func (s *RemoveService) removeInlineHelpTarget(makefilePath string) error {
    content, err := os.ReadFile(makefilePath)
    if err != nil {
        return err
    }

    lines := strings.Split(string(content), "\n")
    filtered := []string{}

    inHelpTarget := false
    removed := false

    for _, line := range lines {
        // Detect start of help target
        if strings.HasPrefix(line, "help:") || strings.HasPrefix(line, ".PHONY: help") {
            inHelpTarget = true
            removed = true
            if s.verbose {
                fmt.Printf("Removing help target starting at: %s\n", line)
            }
            continue
        }

        // Detect end of help target (next target or non-recipe line)
        if inHelpTarget {
            if strings.HasPrefix(line, "\t") || strings.HasPrefix(line, "  ") {
                continue  // Skip recipe lines
            }
            inHelpTarget = false
        }

        filtered = append(filtered, line)
    }

    if !removed {
        return nil // No changes needed
    }

    newContent := strings.Join(filtered, "\n")
    return atomicWriteFile(makefilePath, []byte(newContent), 0644)
}

// removeHelpTargetFiles deletes help target files
func (s *RemoveService) removeHelpTargetFiles(makefilePath string) error {
    makeDir := filepath.Join(filepath.Dir(makefilePath), "make")
    helpFile := filepath.Join(makeDir, "01-help.mk")

    if _, err := os.Stat(helpFile); err == nil {
        if err := os.Remove(helpFile); err != nil {
            return fmt.Errorf("failed to remove %s: %w", helpFile, err)
        }
    }

    return nil
}
```

**Key Design Decisions:**
- Multi-step removal (directives, inline target, files)
- Pattern matching for include directives
- Recipe detection for inline target removal

**Error Handling:**
- File not found (not an error, already removed)
- Multiple help targets (remove all)

## 5. Program Flow

### 5.1 Help Generation Flow

```
1. CLI Parsing
   ├─> Parse flags and validate
   ├─> Resolve Makefile path (cwd/Makefile or --makefile-path)
   ├─> Detect color mode (terminal detection + flags)
   └─> Build Config object

2. Discovery Phase
   ├─> Discover Makefiles (MAKEFILE_LIST)
   │   ├─> Generate temporary Makefile with _list_makefiles target
   │   ├─> Execute: make -f <temp> _list_makefiles
   │   └─> Parse space-separated output -> []string
   └─> Discover Targets (make -p)
       ├─> Execute: make -f <makefile> -p -r
       └─> Parse database output -> []string

3. Parsing Phase
   ├─> For each Makefile in discovery order:
   │   ├─> Scan line-by-line
   │   ├─> Detect ## documentation lines
   │   ├─> Parse directives (@file, @category, @var, @alias)
   │   ├─> Detect target definitions (lines with :)
   │   └─> Associate pending docs with targets
   └─> Result: []*ParsedFile

4. Model Building Phase
   ├─> Aggregate directives from all files
   ├─> Group targets by category
   ├─> Validate categorization (no mixing unless --default-category)
   ├─> Associate aliases and variables with targets
   └─> Result: *HelpModel

5. Ordering Phase
   ├─> Apply category ordering
   │   ├─> If --category-order: explicit order + alphabetical remainder
   │   ├─> Else if --keep-order-categories: discovery order
   │   └─> Else: alphabetical
   └─> Apply target ordering (within each category)
       ├─> If --keep-order-targets: discovery order
       └─> Else: alphabetical

6. Summary Extraction Phase
   ├─> For each target:
   │   ├─> Join documentation lines
   │   ├─> Strip markdown headers
   │   ├─> Strip markdown formatting
   │   ├─> Strip HTML tags
   │   ├─> Normalize whitespace
   │   └─> Extract first sentence (regex)
   └─> Update target.Summary

7. Formatting Phase
   ├─> Initialize ColorScheme based on config.UseColor
   ├─> Render usage line
   ├─> Render file docs
   ├─> Render "Targets:" header
   ├─> For each category:
   │   ├─> Render category name (if not default)
   │   └─> For each target:
   │       ├─> Render target name + aliases
   │       ├─> Render summary
   │       └─> Render variables (if any)
   └─> Result: formatted string

8. Output
   └─> Write to STDOUT
```

### 5.2 Add-Target Flow

```
1. CLI Parsing
   ├─> Parse flags (including help generation flags)
   ├─> Validate --target-file if specified
   └─> Build Config object

2. Determine Target File Location
   ├─> If --target-file specified:
   │   └─> Use specified path, mark needsInclude=true
   ├─> Else if Makefile contains "include make/*.mk":
   │   ├─> Create make/ directory if needed
   │   └─> Set targetFile=make/01-help.mk, needsInclude=false
   └─> Else:
       └─> Set targetFile=<Makefile>, needsInclude=false (append)

3. Generate Help Target Content
   ├─> Build .PHONY: help line
   └─> Build help: target with make-help + flags

4. Write Target File
   ├─> If appending to Makefile:
   │   ├─> Read existing content
   │   ├─> Append help target
   │   └─> Write back
   └─> Else:
       └─> Write new file with help target

5. Add Include Directive (if needed)
   ├─> Compute relative path from Makefile to target file
   ├─> Generate include directive
   └─> Append to Makefile

6. Success
   └─> Print confirmation message
```

### 5.3 Remove-Target Flow

```
1. CLI Parsing
   ├─> Parse flags
   └─> Resolve Makefile path

2. Remove Include Directives
   ├─> Read Makefile
   ├─> Filter out lines matching: ^include\s+.*help.*\.mk
   └─> Write back

3. Remove Inline Help Target
   ├─> Read Makefile
   ├─> Detect help: target and .PHONY: help
   ├─> Skip target and its recipe lines (tab/space-prefixed)
   └─> Write back

4. Remove Help Target Files
   ├─> Check for make/01-help.mk
   ├─> Delete if exists
   └─> Check for other help-related .mk files in make/

5. Success
   └─> Print confirmation message
```

## 6. Key Algorithms

### 6.1 File Discovery via MAKEFILE_LIST

**Algorithm:**
```
Input: makefilePath (path to main Makefile)
Output: []string (ordered list of Makefile paths)

1. Create temporary Makefile content:
   - Cat main Makefile content
   - Append blank line
   - Append target:
     .PHONY: _list_makefiles
     _list_makefiles:
         @echo $(MAKEFILE_LIST)

2. Execute shell command:
   make -f <(cat Makefile && echo && echo -e '<target>') _list_makefiles

3. Parse stdout:
   - Split on whitespace
   - Each token is a Makefile path

4. Resolve to absolute paths:
   - For each path:
     - If relative, resolve from Makefile directory
     - Return absolute path

5. Return ordered list
```

**Important:** Included files appear in MAKEFILE_LIST after their parent file completes, not at the include point. This matches Make's processing order.

**Error Handling:**
- Shell command failure -> wrap error with context
- Empty output -> error "no Makefiles found"
- Invalid paths -> error "Makefile not found: <path>"

### 6.2 Documentation Parsing and Directive Handling

**Algorithm:**
```
Input: fileContent (string), targetMap (map[string]int)
Output: []Directive

State:
- currentCategory: string (current category name)
- pendingDocs: []Directive (docs awaiting target association)

For each line:
  1. If line starts with "## ":
     - Parse directive (detect @file, @category, @var, @alias, or doc)
     - If @category: update currentCategory
     - If @file: add to FileDocs immediately
     - Else: add to pendingDocs

  2. Else if line is target definition (contains : or &:):
     - Extract target name
     - Associate pendingDocs with target
     - Clear pendingDocs

  3. Else:
     - Clear pendingDocs (non-doc line breaks association)

Return all directives
```

**Target Name Extraction:**
```
1. Find first : in line
2. Extract everything before :
3. If ends with &, remove it (grouped target)
4. Split on whitespace, take first token
5. Return token as target name
```

**Edge Cases:**
- Grouped targets (foo bar baz:) -> extract "foo"
- Variable targets ($(VAR):) -> extract "$(VAR)"
- Comments -> skip
- Indented lines -> skip (recipe lines)

### 6.3 Category/Target Ordering Logic

**Algorithm:**
```
Input: HelpModel, Config
Output: HelpModel (with ordered categories and targets)

Category Ordering:
  If --category-order specified:
    1. Create map of category name -> Category
    2. Build ordered list:
       - For each name in --category-order:
         - Add corresponding Category
         - Remove from map
    3. Sort remaining categories alphabetically
    4. Append to ordered list
    5. Validate: error if any --category-order name not found

  Else if --keep-order-categories:
    Sort categories by DiscoveryOrder field

  Else:
    Sort categories alphabetically by Name

Target Ordering (within each category):
  If --keep-order-targets:
    Sort targets by DiscoveryOrder field

  Else:
    Sort targets alphabetically by Name
```

**Discovery Order Tracking:**
- Global counter incremented for each category/target first appearance
- Stored in DiscoveryOrder field
- Split categories use DiscoveryOrder of first appearance

### 6.4 Summary Extraction Algorithm (Port of extract-topic)

**Algorithm:**
```
Input: documentation []string (full target docs)
Output: summary string (first sentence)

1. Join documentation lines with space
   text = strings.Join(documentation, " ")

2. Strip markdown headers
   Remove all lines matching ^#+\s+

3. Strip markdown formatting
   - Remove **bold** -> bold
   - Remove *italic* -> italic
   - Remove `code` -> code
   - Remove [text](url) -> text

4. Strip HTML tags
   Remove all <tag> patterns

5. Normalize whitespace
   - Replace \n with space
   - Collapse multiple spaces -> single space
   - Trim leading/trailing whitespace

6. Extract first sentence
   Regex: ^((?:[^.!?]|\.\.\.|\.[^\s])+[.?!])(\s|$)

   Explanation:
   - (?:[^.!?]|\.\.\.|\.[^\s])+ : Match anything except .!? OR ... OR .<non-space>
   - [.?!] : Match sentence terminator
   - (\s|$) : Must be followed by whitespace or end-of-string

   Edge Cases:
   - "..." (ellipsis) -> not sentence boundary
   - "127.0.0.1." (IP) -> not sentence boundary (. followed by digit)
   - "This is it." -> sentence boundary (. followed by space/EOL)

7. Return matched sentence or full text if no match
```

**Test Cases:**
```
Input: "Build the project. Run tests."
Output: "Build the project."

Input: "Supports IPv4 addresses like 127.0.0.1. Cool!"
Output: "Supports IPv4 addresses like 127.0.0.1."

Input: "Wait for it... then proceed. Done."
Output: "Wait for it... then proceed."

Input: "**Bold text** and *italic* formatting"
Output: "Bold text and italic formatting"

Input: "No sentence terminator here"
Output: "No sentence terminator here"
```

## 7. Error Handling

### 7.1 Error Classification

| Error Type | Priority | Behavior |
|------------|----------|----------|
| Makefile not found | CRITICAL | Exit with error message |
| Mixed categorization without --default-category | CRITICAL | Exit with clear error and suggestion |
| Unknown category in --category-order | CRITICAL | Exit with list of available categories |
| Make command execution failure | CRITICAL | Exit with stderr output |
| Invalid directive syntax | WARNING | Log warning, skip directive, continue |
| Malformed @var or @alias | WARNING | Log warning, best-effort parse |
| Duplicate help target | WARNING | Ask user to remove manually or use --force |
| File write failure | CRITICAL | Exit with error message |

### 7.2 Error Types and Messages

```go
package errors

type MixedCategorizationError struct {
    Message string
}

func (e *MixedCategorizationError) Error() string {
    return fmt.Sprintf("mixed categorization: %s\nUse --default-category to assign uncategorized targets to a default category", e.Message)
}

type UnknownCategoryError struct {
    CategoryName string
    Available    []string
}

func (e *UnknownCategoryError) Error() string {
    return fmt.Sprintf("unknown category %q in --category-order\nAvailable categories: %s",
        e.CategoryName, strings.Join(e.Available, ", "))
}

type MakefileNotFoundError struct {
    Path string
}

func (e *MakefileNotFoundError) Error() string {
    return fmt.Sprintf("Makefile not found: %s\nUse --makefile-path to specify location", e.Path)
}

type MakeExecutionError struct {
    Command string
    Stderr  string
}

func (e *MakeExecutionError) Error() string {
    return fmt.Sprintf("make command failed: %s\n%s", e.Command, e.Stderr)
}
```

### 7.3 Error Scenarios and Handling

**Scenario 1: Mixed Categorization**
```
Problem: Some targets have @category, others don't
Detection: Model validator counts categorized vs uncategorized
Action:
  - If --default-category set: assign uncategorized to default
  - Else: return MixedCategorizationError
```

**Scenario 2: Unknown Category in --category-order**
```
Problem: User specifies category that doesn't exist
Detection: Ordering service validates all names exist
Action: Return UnknownCategoryError with available categories
```

**Scenario 3: File Not Found**
```
Problem: Makefile doesn't exist at specified path
Detection: os.Stat fails in discovery service
Action: Return MakefileNotFoundError
```

**Scenario 4: Invalid Directive Syntax**
```
Problem: Malformed @var or @alias directive
Detection: Parser fails to split on expected delimiter
Action: Log warning, skip directive, continue parsing
Example: "@var NODELIM" -> log "invalid @var directive at line X: missing ' - '"
```

**Scenario 5: Make Command Failure**
```
Problem: make -p or make _list_makefiles fails
Detection: Non-zero exit code from exec.Command
Action: Return MakeExecutionError with stderr
```

**Scenario 6: Duplicate Help Target**
```
Problem: help target already exists when running add-target
Detection: Check for existing help: in Makefile
Action:
  - Return error asking user to remove manually
  - Or add --force flag to overwrite
```

## 8. Testing Strategy

### 8.1 Unit Testing Approach

**Parser Tests** (`internal/parser/scanner_test.go`)
```go
func TestScanFile(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected *ParsedFile
    }{
        {
            name: "file directive",
            input: `## @file
## This is file documentation
## Second line`,
            expected: &ParsedFile{
                Directives: []Directive{
                    {Type: DirectiveFile, Value: ""},
                    {Type: DirectiveDoc, Value: "This is file documentation"},
                    {Type: DirectiveDoc, Value: "Second line"},
                },
            },
        },
        {
            name: "category and target",
            input: `## @category Build
## Build the project
build:
	go build`,
            expected: &ParsedFile{
                Directives: []Directive{
                    {Type: DirectiveCategory, Value: "Build"},
                    {Type: DirectiveDoc, Value: "Build the project"},
                },
                TargetMap: map[string]int{"build": 3},
            },
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            scanner := NewScanner()
            result, err := scanner.ScanFile("test.mk")
            // Assertions...
        })
    }
}
```

**Summary Extractor Tests** (`internal/summary/extractor_test.go`)
```go
func TestExtract(t *testing.T) {
    tests := []struct {
        name     string
        docs     []string
        expected string
    }{
        {
            name:     "simple sentence",
            docs:     []string{"This is a test.", "More text."},
            expected: "This is a test.",
        },
        {
            name:     "ellipsis handling",
            docs:     []string{"Wait for it... then proceed.", "Done."},
            expected: "Wait for it... then proceed.",
        },
        {
            name:     "IP address handling",
            docs:     []string{"Connect to 127.0.0.1. Then test.", "More."},
            expected: "Connect to 127.0.0.1.",
        },
        {
            name:     "markdown stripping",
            docs:     []string{"**Bold** and *italic* text."},
            expected: "Bold and italic text.",
        },
        {
            name:     "no sentence terminator",
            docs:     []string{"No terminator"},
            expected: "No terminator",
        },
    }

    extractor := NewExtractor()
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := extractor.Extract(tt.docs)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

**Model Builder Tests** (`internal/model/builder_test.go`)
```go
func TestBuild(t *testing.T) {
    tests := []struct {
        name        string
        parsedFiles []*parser.ParsedFile
        config      *cli.Config
        expected    *HelpModel
        expectError bool
    }{
        {
            name: "mixed categorization without default",
            parsedFiles: []*parser.ParsedFile{
                {
                    Directives: []Directive{
                        {Type: DirectiveCategory, Value: "Build"},
                        {Type: DirectiveDoc, Value: "Build target"},
                    },
                    TargetMap: map[string]int{"build": 2},
                },
                {
                    Directives: []Directive{
                        {Type: DirectiveDoc, Value: "Test target"},
                    },
                    TargetMap: map[string]int{"test": 1},
                },
            },
            config:      &cli.Config{},
            expectError: true,
        },
        // More test cases...
    }
}
```

**Ordering Service Tests** (`internal/ordering/service_test.go`)
```go
func TestApplyOrdering(t *testing.T) {
    tests := []struct {
        name     string
        model    *HelpModel
        config   *cli.Config
        expected []string  // Expected category order
    }{
        {
            name: "alphabetical category order",
            model: &HelpModel{
                Categories: []Category{
                    {Name: "Zebra", DiscoveryOrder: 1},
                    {Name: "Alpha", DiscoveryOrder: 2},
                },
            },
            config:   &cli.Config{},
            expected: []string{"Alpha", "Zebra"},
        },
        {
            name: "discovery order preserved",
            model: &HelpModel{
                Categories: []Category{
                    {Name: "Zebra", DiscoveryOrder: 1},
                    {Name: "Alpha", DiscoveryOrder: 2},
                },
            },
            config:   &cli.Config{KeepOrderCategories: true},
            expected: []string{"Zebra", "Alpha"},
        },
        {
            name: "explicit category order",
            model: &HelpModel{
                Categories: []Category{
                    {Name: "Build", DiscoveryOrder: 1},
                    {Name: "Test", DiscoveryOrder: 2},
                    {Name: "Deploy", DiscoveryOrder: 3},
                },
            },
            config:   &cli.Config{CategoryOrder: []string{"Deploy", "Build"}},
            expected: []string{"Deploy", "Build", "Test"},  // Test appended alphabetically
        },
    }
}
```

### 8.2 Integration Testing Approach

**Fixture-Based Tests** (`test/integration/cli_test.go`)

```go
func TestHelpGeneration(t *testing.T) {
    tests := []struct {
        name         string
        fixture      string  // Path to test Makefile
        args         []string
        expectedFile string  // Path to expected output
    }{
        {
            name:         "basic help",
            fixture:      "fixtures/makefiles/basic.mk",
            args:         []string{},
            expectedFile: "fixtures/expected/basic_help.txt",
        },
        {
            name:         "categorized targets",
            fixture:      "fixtures/makefiles/categorized.mk",
            args:         []string{},
            expectedFile: "fixtures/expected/categorized_help.txt",
        },
        {
            name:         "explicit category order",
            fixture:      "fixtures/makefiles/categorized.mk",
            args:         []string{"--category-order", "Deploy,Build"},
            expectedFile: "fixtures/expected/categorized_ordered_help.txt",
        },
        {
            name:         "included files",
            fixture:      "fixtures/makefiles/with_includes.mk",
            args:         []string{},
            expectedFile: "fixtures/expected/with_includes_help.txt",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Execute CLI command
            cmd := exec.Command("make-help", append([]string{"--makefile-path", tt.fixture}, tt.args...)...)
            output, err := cmd.Output()
            require.NoError(t, err)

            // Read expected output
            expected, err := os.ReadFile(tt.expectedFile)
            require.NoError(t, err)

            // Compare (strip colors for comparison)
            assert.Equal(t, stripANSI(string(expected)), stripANSI(string(output)))
        })
    }
}

func TestAddTarget(t *testing.T) {
    tests := []struct {
        name         string
        fixture      string
        args         []string
        expectedMake string  // Path to expected Makefile after add
        expectedFile string  // Path to expected help target file (if separate)
    }{
        {
            name:         "append to makefile",
            fixture:      "fixtures/makefiles/empty.mk",
            args:         []string{},
            expectedMake: "fixtures/expected/empty_with_help.mk",
        },
        {
            name:         "create make/01-help.mk",
            fixture:      "fixtures/makefiles/with_make_include.mk",
            args:         []string{},
            expectedMake: "fixtures/expected/with_make_include_updated.mk",
            expectedFile: "fixtures/expected/01-help.mk",
        },
        {
            name:         "explicit target file",
            fixture:      "fixtures/makefiles/basic.mk",
            args:         []string{"--target-file", "custom-help.mk"},
            expectedMake: "fixtures/expected/basic_with_include.mk",
            expectedFile: "fixtures/expected/custom-help.mk",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Copy fixture to temp location
            tmpDir := t.TempDir()
            tmpMakefile := filepath.Join(tmpDir, "Makefile")
            copyFile(tt.fixture, tmpMakefile)

            // Execute add-target
            cmd := exec.Command("make-help", append([]string{"add-target", "--makefile-path", tmpMakefile}, tt.args...)...)
            err := cmd.Run()
            require.NoError(t, err)

            // Compare Makefile
            assertFileEquals(t, tt.expectedMake, tmpMakefile)

            // Compare help target file if separate
            if tt.expectedFile != "" {
                targetFile := filepath.Join(tmpDir, filepath.Base(tt.expectedFile))
                assertFileEquals(t, tt.expectedFile, targetFile)
            }
        })
    }
}
```

**Test Fixtures Structure:**
```
test/fixtures/
├── makefiles/
│   ├── basic.mk                   # Simple Makefile with targets
│   ├── categorized.mk             # Targets with @category
│   ├── with_includes.mk           # Makefile with includes
│   ├── mixed_categorization.mk    # Error case: mixed
│   └── empty.mk                   # Empty Makefile
└── expected/
    ├── basic_help.txt             # Expected help output
    ├── categorized_help.txt
    ├── with_includes_help.txt
    ├── empty_with_help.mk         # Expected Makefile after add
    └── 01-help.mk                 # Expected help target file
```

### 8.3 Mock Strategy

**CommandExecutor Mock** (for testing discovery without executing make)
```go
type MockCommandExecutor struct {
    outputs map[string]string  // Command -> stdout
    errors  map[string]error   // Command -> error
}

func (m *MockCommandExecutor) Execute(cmd string, args ...string) (string, string, error) {
    key := fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
    return m.outputs[key], "", m.errors[key]
}

// Usage in tests:
func TestDiscoverMakefiles(t *testing.T) {
    mock := &MockCommandExecutor{
        outputs: map[string]string{
            "make ... _list_makefiles": "Makefile include/common.mk",
        },
    }

    service := &discovery.Service{executor: mock}
    files, err := service.DiscoverMakefiles("Makefile")

    assert.NoError(t, err)
    assert.Equal(t, []string{"Makefile", "include/common.mk"}, files)
}
```

### 8.4 Test Coverage Goals

| Package | Coverage Target | Focus Areas |
|---------|----------------|-------------|
| `internal/parser` | 95% | All directive types, edge cases |
| `internal/summary` | 100% | All regex edge cases (ellipsis, IPs) |
| `internal/model` | 90% | Categorization validation, alias/var handling |
| `internal/ordering` | 95% | All ordering strategies |
| `internal/format` | 85% | Template rendering, color schemes |
| `internal/discovery` | 80% | Mock-based tests (not real make execution) |
| `internal/target` | 85% | File operations, pattern detection |
| Overall | 90% | Focus on critical paths |

## Architecture Review Summary

This design follows Go idioms and emphasizes:

### Strengths

**Simplicity:**
- Clear separation of concerns (discovery, parsing, building, ordering, rendering)
- Minimal dependencies (Cobra only for CLI)
- No unnecessary abstractions

**Standard Approaches:**
- Cobra for CLI (de facto standard in Go ecosystem)
- Builder pattern for complex model construction
- Strategy pattern for ordering flexibility
- Template-based rendering

**Usability:**
- Clear error messages with actionable suggestions
- Intuitive data structures
- Well-documented package responsibilities
- `--verbose` flag for debugging file discovery and parsing

**Security:**
- Temporary physical files instead of bash process substitution (prevents command injection)
- No hardcoded credentials
- File path validation
- Input sanitization in parser

**Maintainability:**
- Testable design (interfaces for external commands)
- Clear package boundaries
- Comprehensive error handling
- Fixture-based integration tests
- Package-level godoc comments in doc.go files

**Performance:**
- Minimal allocations (string builder for rendering)
- Single-pass parsing
- No unnecessary file reads
- Pre-compiled regex patterns in summary extractor

**Robustness:**
- Comprehensive error handling
- Graceful degradation (best-effort directive parsing)
- Validation at model building stage
- Clear error types
- Timeout on all make command executions (30s default)
- Atomic file writes for Makefile modifications
- Makefile syntax validation before modification

### Key Design Decisions

1. **Security - No Shell Injection:** MAKEFILE_LIST discovery uses temporary physical files instead of bash process substitution `<(...)` to eliminate command injection risk.

2. **Robustness - Command Timeouts:** All `make` command executions use `context.WithTimeout` (30 seconds) to prevent indefinite hangs on malformed Makefiles.

3. **Robustness - Atomic File Writes:** All file modifications (add-target, remove-target) write to a temporary file first, then rename. This prevents file corruption if the process crashes mid-write.

4. **Robustness - Makefile Validation:** Before modifying any Makefile, run `make -n` (dry-run) to validate syntax and catch errors early.

5. **Performance - Pre-compiled Regex:** All regex patterns in the summary extractor are compiled once at construction time, avoiding repeated compilation when processing many targets.

6. **Maintainability - Package Documentation:** Each package includes a doc.go file with comprehensive godoc comments explaining purpose, design decisions, and security considerations.

7. **Usability - Verbose Mode:** The `--verbose` flag enables detailed output about file discovery, target parsing, and file modifications for debugging.

### Recommendations for Implementation

1. **Start with core data structures** (`internal/model/help.go`)
2. **Implement parser** (most critical component)
3. **Port summary extractor** (isolated, testable)
4. **Build model builder** (depends on parser)
5. **Add ordering and formatting** (depends on model)
6. **Implement discovery** (can be mocked initially)
7. **Build CLI layer** (thin, delegates to services)
8. **Add target manipulation** (add/remove)

This design is ready for implementation with clear component boundaries, comprehensive error handling, and a robust testing strategy.
