# Component Architecture

Detailed specifications for each major component in the make-help system.

## Table of Contents

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

```go
// Root command setup
func NewRootCmd() *cobra.Command {
    config := NewConfig()

    rootCmd := &cobra.Command{
        Use:   "make-help",
        Short: "Static help generation for Makefiles",
        Long: `make-help generates static help files from Makefile documentation.

Default behavior generates help.mk with embedded help text. Use flags for other operations:
  --show-help           Display help dynamically (don't generate file)
  --show-help --target <name>  Show detailed help for a target
  --remove-help         Remove generated help files

Documentation directives (in ## comments):
  !file         File-level documentation
  !category     Group targets into categories
  !var          Document environment variables
  !alias        Define target aliases`,
        RunE: func(cmd *cobra.Command, args []string) error {
            // Normalize IncludeTargets from comma-separated + repeatable flags
            config.IncludeTargets = parseIncludeTargets(config.IncludeTargets)

            // Resolve color mode
            config.UseColor = ResolveColorMode(config)

            // Dispatch to appropriate handler
            if config.RemoveHelp {
                return runRemoveHelp(config)
            } else if config.ShowHelp {
                if config.Target != "" {
                    return runDetailedHelp(config)
                } else {
                    return runShowHelp(config)
                }
            } else {
                return runGenerateHelpFile(config)
            }
        },
    }

    // Global flags
    rootCmd.PersistentFlags().StringVar(&config.MakefilePath,
        "makefile-path", "", "Path to Makefile")
    rootCmd.PersistentFlags().BoolVar(&noColor,
        "no-color", false, "Disable colored output")
    rootCmd.PersistentFlags().BoolVar(&forceColor,
        "color", false, "Force colored output")
    rootCmd.PersistentFlags().BoolVarP(&config.Verbose,
        "verbose", "v", false, "Enable verbose output for debugging")

    // Mode flags
    rootCmd.Flags().BoolVar(&config.ShowHelp,
        "show-help", false, "Display help dynamically instead of generating file")
    rootCmd.Flags().BoolVar(&config.RemoveHelp,
        "remove-help", false, "Remove generated help files")
    rootCmd.Flags().StringVar(&config.Target,
        "target", "", "Show detailed help for a specific target (requires --show-help)")
    rootCmd.Flags().BoolVar(&config.DryRun,
        "dry-run", false, "Preview what files would be created/modified without making changes")

    // Target filtering flags
    rootCmd.Flags().StringSliceVar(&config.IncludeTargets,
        "include-target", []string{}, "Include undocumented target (repeatable, comma-separated)")
    rootCmd.Flags().BoolVar(&config.IncludeAllPhony,
        "include-all-phony", false, "Include all .PHONY targets in help output")

    // Ordering flags
    rootCmd.Flags().BoolVar(&config.KeepOrderCategories,
        "keep-order-categories", false, "Preserve category discovery order")
    rootCmd.Flags().BoolVar(&config.KeepOrderTargets,
        "keep-order-targets", false, "Preserve target discovery order")
    rootCmd.Flags().StringSliceVar(&config.CategoryOrder,
        "category-order", []string{}, "Explicit category order (comma-separated)")
    rootCmd.Flags().StringVar(&config.DefaultCategory,
        "default-category", "", "Default category for uncategorized targets")
    rootCmd.Flags().StringVar(&config.HelpCategory,
        "help-category", "Help", "Category name for generated help targets (help, update-help)")

    // Help file generation flags
    rootCmd.Flags().StringVar(&config.HelpFileRelPath,
        "help-file-rel-path", "", "Explicit relative path for generated help file")

    return rootCmd
}
```

**Responsibilities:**
- Parse command-line arguments
- Validate flag combinations
- Detect terminal capabilities (isatty)
- Resolve color mode
- Delegate to appropriate service based on mode flags

**Mode Flags:**
1. **Default (no flags)**: Generate static help file via `runGenerateHelpFile()`
2. **`--show-help`**: Display help dynamically via `runShowHelp()`
3. **`--show-help --target <name>`**: Display detailed help for single target via `runDetailedHelp()`
4. **`--remove-help`**: Remove generated help files via `runRemoveHelp()`

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
- Invalid flag combinations (e.g., `--target` without `--show-help`)
- File path validation
- Conflicting color flags (`--color` + `--no-color`)
- Mode flag restrictions

### 2 Discovery Service

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

### 3 Parser Service

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
    case strings.HasPrefix(content, "!file"):
        directive.Type = DirectiveFile
        directive.Value = strings.TrimSpace(strings.TrimPrefix(content, "!file"))

    case strings.HasPrefix(content, "!category "):
        directive.Type = DirectiveCategory
        directive.Value = strings.TrimSpace(strings.TrimPrefix(content, "!category "))
        // SWITCH BEHAVIOR: Category is "sticky" - applies to all subsequent targets
        // until changed. Use "!category _" to reset to uncategorized (nil).
        s.currentCategory = directive.Value

    case strings.HasPrefix(content, "!var "):
        directive.Type = DirectiveVar
        directive.Value = strings.TrimSpace(strings.TrimPrefix(content, "!var "))

    case strings.HasPrefix(content, "!alias "):
        directive.Type = DirectiveAlias
        directive.Value = strings.TrimSpace(strings.TrimPrefix(content, "!alias "))

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
- **Stateful scanning to track current category**: `!category` sets the current category that applies to all following targents until another `!category` directive is encountered
- **Pending documentation queue**: Documentation lines are queued and associated with the next target definition
- **Simple regex-free parsing for robustness**: Target parsing uses string operations instead of complex regex
- **Target name extraction handles grouped and variable targets**: Supports `foo:`, `foo&:`, and `$(VAR):` patterns

**Error Handling:**
- Invalid directive syntax (log warning, skip)
- Malformed !var or !alias (log warning, skip)

### 4 Model Builder

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

// parseVarDirective parses !var directive: <NAME> - <description>
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

// parseAliasDirective parses !alias directive: <name>[, <name>...]
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

### 5 Ordering Service

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

### 6 Summary Extractor

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

### 7 Formatter Service

**Package:** `internal/format`

**Design:** String builder-based rendering with color support

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
- Structured rendering methods for consistency

### 8 Static Help File Generator

**Package:** `internal/target`

**Design:** Generates static help files with embedded help text and auto-regeneration logic. Includes smart file location detection that supports make/ directory patterns, numbered prefixes, and automatic include directive detection.

```go
type AddService struct {
    config   *Config
    executor discovery.CommandExecutor
    verbose  bool
}

// AddTarget generates and injects a help target into the Makefile.
// It follows a three-tier strategy for target file placement:
//  1. Use explicit --help-file-rel-path if specified (needs include directive)
//  2. Create make/NN-help.mk if include make/*.mk pattern found (no include needed)
//  3. Otherwise create help.mk in same directory as Makefile (needs include directive)
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
    content := generateHelpTarget(s.config)

    // Write target file using atomic write (write to temp, then rename)
    if err := AtomicWriteFile(targetFile, []byte(content), 0644); err != nil {
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

// previewDryRun shows what files would be created without making changes
func (s *AddService) previewDryRun(targetFile, content string, needsInclude bool, makefilePath string) {
    fmt.Printf("DRY RUN: Preview of files that would be created/modified\n\n")

    // Show target file that would be created
    fmt.Printf("Would create file: %s\n", targetFile)
    fmt.Printf("Content preview:\n%s\n", strings.Repeat("-", 70))
    fmt.Printf("%s\n", content)
    fmt.Printf("%s\n\n", strings.Repeat("-", 70))

    // Show include directive if needed
    if needsInclude {
        relPath, _ := filepath.Rel(filepath.Dir(makefilePath), targetFile)
        fmt.Printf("Would append to file: %s\n", makefilePath)
        fmt.Printf("Content to append:\n%s\n", strings.Repeat("-", 70))
        fmt.Printf("\ninclude %s\n", relPath)
        fmt.Printf("%s\n", strings.Repeat("-", 70))
    }
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

// IncludePattern holds information about a detected include directive pattern.
type IncludePattern struct {
    // Suffix is the file extension (e.g., ".mk" or "")
    Suffix string
    // FullPattern is the complete include pattern (e.g., "make/*.mk")
    FullPattern string
    // PatternPrefix is the prefix part before the wildcard (e.g., "make/" or "./make/")
    PatternPrefix string
}

// determineTargetFileImpl decides where to create the help target.
// Strategy:
//  1. If explicit --help-file-rel-path is provided, use that (needs include directive)
//  2. Default to make/help.mk (or make/NN-help.mk if numbered files exist)
//  3. Scan Makefile for existing include patterns to determine suffix
//  4. If no include pattern exists, one will be added
func determineTargetFileImpl(makefilePath, explicitRelPath string, createDirs bool) (string, bool, error) {
    makefileDir := filepath.Dir(makefilePath)

    // 1. Explicit --help-file-rel-path (always relative)
    if explicitRelPath != "" {
        absPath := filepath.Join(makefileDir, explicitRelPath)
        if createDirs {
            parentDir := filepath.Dir(absPath)
            if err := os.MkdirAll(parentDir, 0755); err != nil {
                return "", false, fmt.Errorf("failed to create directory %s: %w", parentDir, err)
            }
        }
        return absPath, true, nil
    }

    // 2. Read Makefile to check for include patterns
    content, err := os.ReadFile(makefilePath)
    if err != nil {
        return "", false, fmt.Errorf("failed to read Makefile: %w", err)
    }

    // 3. Find include pattern for make/* files
    pattern := findMakeIncludePattern(content)

    // 4. Determine the suffix to use for our file
    suffix := ".mk" // default
    if pattern != nil {
        suffix = pattern.Suffix
    }

    // 5. Create make/ directory if needed
    makeDir := filepath.Join(makefileDir, "make")
    if createDirs {
        if err := os.MkdirAll(makeDir, 0755); err != nil {
            return "", false, fmt.Errorf("failed to create make/ directory: %w", err)
        }
    }

    // 6. Check for numbered files in make/ directory
    prefix := determineNumberPrefix(makeDir, suffix, pattern)

    // 7. Construct filename
    filename := prefix + "help" + suffix
    targetPath := filepath.Join(makeDir, filename)

    // Need include directive if no existing pattern was found
    needsInclude := pattern == nil

    return targetPath, needsInclude, nil
}

// findMakeIncludePattern scans Makefile content for include directives matching make/*
// Returns nil if no matching pattern found.
// Matches patterns like: include make/*.mk, -include ./make/*.mk, etc.
func findMakeIncludePattern(content []byte) *IncludePattern {
    includeRegex := regexp.MustCompile(`(?m)^-?include\s+(?:\$\([^)]+\))?(\./)?make/\*(\.[a-zA-Z0-9]+)?(?:\s|$)`)
    matches := includeRegex.FindSubmatch(content)
    if matches == nil {
        return nil
    }

    suffix := ""
    if len(matches) > 2 && len(matches[2]) > 0 {
        suffix = string(matches[2])
    }

    patternPrefix := "make/"
    if len(matches) > 1 && len(matches[1]) > 0 {
        patternPrefix = "./make/"
    }

    return &IncludePattern{
        Suffix:        suffix,
        FullPattern:   string(matches[0]),
        PatternPrefix: patternPrefix,
    }
}

// determineNumberPrefix checks if files in the make directory use numeric prefixes.
// If numbered files exist (e.g., "10-foo.mk"), returns a prefix with matching digit count
// using zeros (e.g., "00-"). Otherwise returns empty string.
func determineNumberPrefix(makeDir, suffix string, pattern *IncludePattern) string {
    entries, err := os.ReadDir(makeDir)
    if err != nil {
        return ""
    }

    numberedFileRegex := regexp.MustCompile(`^(\d+)-.*` + regexp.QuoteMeta(suffix) + `$`)

    maxDigits := 0
    for _, entry := range entries {
        if entry.IsDir() {
            continue
        }
        matches := numberedFileRegex.FindStringSubmatch(entry.Name())
        if matches != nil {
            digitCount := len(matches[1])
            if digitCount > maxDigits {
                maxDigits = digitCount
            }
        }
    }

    if maxDigits == 0 {
        return ""
    }

    // Generate prefix with zeros matching the digit count
    zeros := strings.Repeat("0", maxDigits)
    return zeros + "-"
}

// generateStaticHelpFile creates static help file with embedded help text
func (g *Generator) generateStaticHelpFile() string {
    var buf strings.Builder

    // Header comment
    buf.WriteString("# Generated by make-help. DO NOT EDIT.\n")
    buf.WriteString("# Regenerate with: make-help\n\n")

    // Generate help target with @echo statements
    buf.WriteString(".PHONY: help\n")
    buf.WriteString("help:\n")

    // Render help model as @echo statements
    for _, line := range g.renderHelpAsEcho() {
        buf.WriteString("\t@echo ")
        buf.WriteString(escapeForEcho(line))
        buf.WriteString("\n")
    }

    buf.WriteString("\n")

    // Generate help-<target> targets for each documented target
    for _, category := range g.model.Categories {
        for _, target := range category.Targets {
            buf.WriteString(fmt.Sprintf(".PHONY: help-%s\n", target.Name))
            buf.WriteString(fmt.Sprintf("help-%s:\n", target.Name))
            for _, line := range g.renderTargetHelpAsEcho(target) {
                buf.WriteString("\t@echo ")
                buf.WriteString(escapeForEcho(line))
                buf.WriteString("\n")
            }
            buf.WriteString("\n")
        }
    }

    // Auto-regeneration target
    buf.WriteString("# Auto-regenerate help when Makefiles change\n")
    buf.WriteString("help.mk: Makefile\n")
    buf.WriteString("\t@command -v make-help >/dev/null 2>&1 || \\\n")
    buf.WriteString("\tcommand -v npx >/dev/null 2>&1 && npx -y @sdlcforge/make-help || \\\n")
    buf.WriteString("\t{ echo \"Error: make-help not found. Install via: npm install -g @sdlcforge/make-help\"; exit 1; }\n")

    return buf.String()
}

// AddIncludeDirective injects an include statement into the Makefile using atomic write.
// When targetFile is in the make/ directory and no existing include pattern is found,
// adds a pattern include (-include make/*.mk). Otherwise, uses the self-referential pattern
// $(dir $(lastword $(MAKEFILE_LIST))) to ensure the include works regardless of the working
// directory when make is invoked.
func AddIncludeDirective(makefilePath, targetFile string) error {
    content, err := os.ReadFile(makefilePath)
    if err != nil {
        return err
    }

    makefileDir := filepath.Dir(makefilePath)
    relPath, err := filepath.Rel(makefileDir, targetFile)
    if err != nil {
        relPath = filepath.Base(targetFile)
    }

    // Check if target file is in make/ directory
    isInMakeDir := strings.HasPrefix(relPath, "make"+string(filepath.Separator))

    if isInMakeDir {
        // Target is in make/ directory - check for existing pattern
        pattern := findMakeIncludePattern(content)
        if pattern != nil {
            return nil // Pattern already exists
        }

        // Check if pattern include already exists
        patternIncludeRegex := regexp.MustCompile(`(?m)^-?include\s+(?:\./)?make/\*\.mk\s*$`)
        if patternIncludeRegex.Match(content) {
            return nil
        }

        // No pattern found, add -include make/*.mk
        includeDirective := "\n-include make/*.mk\n"
        newContent := append(content, []byte(includeDirective)...)
        return AtomicWriteFile(makefilePath, newContent, 0644)
    }

    // Target is not in make/ directory - add specific file include
    escapedRelPath := regexp.QuoteMeta(relPath)
    includePattern := fmt.Sprintf(`(?m)^-?include\s+(\$\(dir \$\(lastword \$\(MAKEFILE_LIST\)\)\))?%s\s*$`, escapedRelPath)
    existingIncludeRegex := regexp.MustCompile(includePattern)
    if existingIncludeRegex.Match(content) {
        return nil // Include directive already exists
    }

    // Use self-referential include pattern that works from any directory
    includeDirective := fmt.Sprintf("\n-include $(dir $(lastword $(MAKEFILE_LIST)))%s\n", relPath)
    newContent := append(content, []byte(includeDirective)...)
    return AtomicWriteFile(makefilePath, newContent, 0644)
}
```

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

// removeHelpTargetFiles deletes help files (help.mk or make/help.mk)
func (s *RemoveService) removeHelpTargetFiles(makefilePath string) error {
    baseDir := filepath.Dir(makefilePath)

    // Check for help.mk in project root
    helpFile := filepath.Join(baseDir, "help.mk")
    if _, err := os.Stat(helpFile); err == nil {
        if err := os.Remove(helpFile); err != nil {
            return fmt.Errorf("failed to remove %s: %w", helpFile, err)
        }
        if s.verbose {
            fmt.Printf("Removed: %s\n", helpFile)
        }
    }

    // Check for make/help.mk
    makeDir := filepath.Join(baseDir, "make")
    makeHelpFile := filepath.Join(makeDir, "help.mk")
    if _, err := os.Stat(makeHelpFile); err == nil {
        if err := os.Remove(makeHelpFile); err != nil {
            return fmt.Errorf("failed to remove %s: %w", makeHelpFile, err)
        }
        if s.verbose {
            fmt.Printf("Removed: %s\n", makeHelpFile)
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

### 10 Lint Service

**Package:** `internal/lint`

**Design:** Validates documentation quality with optional auto-fix capability

```go
// Check represents a lint check with optional auto-fix capability.
type Check struct {
    Name      string    // Unique identifier (e.g., "summary-punctuation")
    CheckFunc CheckFunc // Performs the check and returns warnings
    FixFunc   FixFunc   // Generates fixes (nil if not auto-fixable)
}

// Fix represents a single file modification to fix a lint warning.
type Fix struct {
    File       string       // Absolute path to the file to modify
    Line       int          // 1-indexed line number to modify
    Operation  FixOperation // Type of modification (FixReplace or FixDelete)
    OldContent string       // Expected current content (for validation)
    NewContent string       // Replacement content (for FixReplace)
}

type FixOperation int

const (
    FixReplace FixOperation = iota // Replace the entire line
    FixDelete                        // Remove the line entirely
)

// Fixer applies fixes to source files.
type Fixer struct {
    DryRun bool // Show what would be fixed without modifying files
}

// ApplyFixes groups fixes by file and applies them atomically.
// Fixes are applied in reverse line order to avoid offset invalidation.
func (f *Fixer) ApplyFixes(fixes []Fix) (*FixResult, error) {
    // Group fixes by file
    fileFixes := make(map[string][]Fix)
    for _, fix := range fixes {
        fileFixes[fix.File] = append(fileFixes[fix.File], fix)
    }

    result := &FixResult{
        FilesModified: make(map[string]int),
    }

    // Apply fixes file by file
    for file, fixes := range fileFixes {
        count, err := f.applyFileFixes(file, fixes)
        if err != nil {
            return result, fmt.Errorf("failed to fix %s: %w", file, err)
        }
        result.FilesModified[file] = count
        result.TotalFixed += count
    }

    return result, nil
}
```

**Key Design Decisions:**
- **Check/Fix separation**: Checks can exist without fixes (error-only checks)
- **Atomic file modifications**: All fixes to a file succeed or none do
- **Reverse line order**: Fixes applied from bottom to top to preserve line numbers
- **Validation before fixing**: Checks OldContent matches current line to detect file changes
- **Dry-run support**: Preview fixes without modifying files (--fix --dry-run)
- **Fix filtering**: Fixed warnings are removed from display output

**CLI Integration:**
- `--lint`: Run lint checks and display warnings
- `--lint --fix`: Apply auto-fixes for safe issues, display remaining warnings
- `--lint --fix --dry-run`: Preview fixes without modifying files

**Error Handling:**
- Line content mismatch (file changed since check)
- Line number out of range
- File write failures

### 11 Version Package

**Package:** `internal/version`

**Design:** Provides build-time version information via ldflags injection

```go
// Version is set at build time via ldflags:
//   go build -ldflags "-X github.com/sdlcforge/make-help/internal/version.Version=1.0.0"
// If not set, defaults to "dev".
var Version = "dev"
```

**Build Integration:**
Version is injected from package.json during build:

```makefile
VERSION := $(shell node -p "require('./package.json').version")
LDFLAGS := -X github.com/sdlcforge/make-help/internal/version.Version=$(VERSION)

build:
    go build -ldflags "$(LDFLAGS)" -o bin/make-help ./cmd/make-help
```

**CLI Integration:**
- `--version`: Display version and exit

**Key Design Decisions:**
- **Single source of truth**: Version comes from package.json
- **Ldflags injection**: No need to update Go code when version changes
- **Default fallback**: Shows "dev" for local development builds
- **Simple implementation**: Just a single exported variable

