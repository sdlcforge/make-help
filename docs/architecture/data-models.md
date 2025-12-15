# Data Models

Core data structures used throughout the make-help system.

## Table of Contents

- [Makefile Documentation Model](#makefile-documentation-model)
- [Configuration Model](#configuration-model)
- [Discovery Model](#discovery-model)
- [Rendering Model](#rendering-model)

---

## Overview

### 1 Target Service Models

```go
// Config holds configuration for target manipulation operations.
type Config struct {
    MakefilePath        string
    TargetFileRelPath   string   // Relative path for help target file
    KeepOrderCategories bool
    KeepOrderTargets    bool
    CategoryOrder       []string
    DefaultCategory     string
}

// IncludePattern holds information about a detected include directive pattern.
type IncludePattern struct {
    // Suffix is the file extension (e.g., ".mk" or "")
    Suffix string
    // FullPattern is the complete include pattern (e.g., "make/*.mk")
    FullPattern string
    // PatternPrefix is the prefix part before the wildcard (e.g., "make/" or "./make/")
    PatternPrefix string
}
```

### 2 Lint Service Models

```go
// Check represents a lint check with optional auto-fix capability.
type Check struct {
    // Name is a unique identifier for the check (e.g., "summary-punctuation").
    Name string
    // CheckFunc performs the check and returns any warnings found.
    CheckFunc CheckFunc
    // FixFunc generates a fix for a warning. May be nil if not auto-fixable.
    FixFunc FixFunc
}

// FixFunc generates a fix for a warning.
// Returns nil if the warning cannot be auto-fixed.
type FixFunc func(w Warning) *Fix

// Fix represents a single file modification to fix a lint warning.
type Fix struct {
    // File is the absolute path to the file to modify.
    File string
    // Line is the 1-indexed line number to modify.
    Line int
    // Operation specifies the type of modification.
    Operation FixOperation
    // OldContent is the expected current content of the line (for validation).
    OldContent string
    // NewContent is the replacement content (for FixReplace operation).
    NewContent string
}

// FixOperation specifies the type of file modification.
type FixOperation int

const (
    // FixReplace replaces the entire line with new content.
    FixReplace FixOperation = iota
    // FixDelete removes the line entirely.
    FixDelete
)

// Fixer applies fixes to source files.
type Fixer struct {
    // DryRun when true shows what would be fixed without modifying files.
    DryRun bool
}

// FixResult contains the results of applying fixes.
type FixResult struct {
    // TotalFixed is the number of fixes successfully applied.
    TotalFixed int
    // FilesModified maps file paths to the number of fixes applied.
    FilesModified map[string]int
}
```

### 3 Makefile Documentation Model

```go
// HelpModel represents the complete parsed help documentation
type HelpModel struct {
    FileDocs     []string              // !file documentation sections (ordered)
    Categories   []Category            // Ordered list of categories
    HasCategories bool                 // True if any !category directives found
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
    Aliases        []string            // Alternative names (!alias)
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
    Type          DirectiveType        // !file, !category, !var, !alias, or doc
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

### 4 Configuration Model

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
    HelpCategory        string        // Category for generated help targets (default: "Help")

    // Add-target options
    HelpFileRelPath string            // Relative path for generated help target file
    CreateHelpTarget bool             // Whether to generate help target file
    RemoveHelpTarget bool             // Whether to remove help target from Makefile
    Version         string            // Version for go install (e.g., "v1.2.3")
    IncludeTargets  []string          // Undocumented targets to include in help
    IncludeAllPhony bool              // Include all .PHONY targets in help output
    Target          string            // Target name for detailed help view

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

### 5 Discovery Model

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

### 6 Rendering Model

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

