package model

import (
	"sort"
	"strings"

	"github.com/sdlcforge/make-help/internal/parser"
	"github.com/sdlcforge/make-help/internal/summary"
)

// BuilderConfig holds configuration for the Builder.
type BuilderConfig struct {
	// DefaultCategory is used for uncategorized targets when categories are mixed.
	DefaultCategory string

	// IncludeTargets lists undocumented targets to include in help.
	IncludeTargets []string

	// IncludeAllPhony includes all .PHONY targets in help output.
	IncludeAllPhony bool

	// PhonyTargets maps target names to their .PHONY status.
	PhonyTargets map[string]bool
}

// Builder constructs a HelpModel from parsed Makefile directives.
// It aggregates file documentation, groups targets by category,
// and associates aliases and variables with targets.
type Builder struct {
	config    *BuilderConfig
	extractor *summary.Extractor
}

// NewBuilder creates a new Builder with the given configuration.
func NewBuilder(config *BuilderConfig) *Builder {
	if config == nil {
		config = &BuilderConfig{}
	}
	if config.PhonyTargets == nil {
		config.PhonyTargets = make(map[string]bool)
	}
	return &Builder{
		config:    config,
		extractor: summary.NewExtractor(),
	}
}

// Build constructs a HelpModel from parsed files.
// It processes directives in order, groups targets by category,
// and validates categorization rules.
func (b *Builder) Build(parsedFiles []*parser.ParsedFile) (*HelpModel, error) {
	model := &HelpModel{
		FileDocs:   []string{},
		Categories: []Category{},
	}

	categoryMap := make(map[string]*Category)
	targetMap := make(map[string]*Target)
	targetToCategory := make(map[string]string) // target name -> category name

	categoryOrder := 0
	targetOrder := 0

	for _, file := range parsedFiles {
		b.processFile(file, model, categoryMap, targetMap, targetToCategory, &categoryOrder, &targetOrder)
	}

	// Assign targets to categories with filtering
	for targetName, target := range targetMap {
		// Apply filtering logic
		shouldInclude := b.shouldIncludeTarget(target)
		if !shouldInclude {
			continue
		}

		// Set phony status
		target.IsPhony = b.config.PhonyTargets[targetName]

		categoryName := targetToCategory[targetName]

		// Compute summary from documentation
		target.Summary = b.extractor.Extract(target.Documentation)

		// Get or create category
		cat, exists := categoryMap[categoryName]
		if !exists {
			cat = &Category{
				Name:           categoryName,
				Targets:        []Target{},
				DiscoveryOrder: categoryOrder,
			}
			categoryOrder++
			categoryMap[categoryName] = cat
		}

		cat.Targets = append(cat.Targets, *target)
	}

	// Convert category map to slice
	for _, cat := range categoryMap {
		model.Categories = append(model.Categories, *cat)
	}

	// Validate categorization
	if err := ValidateCategorization(model, b.config.DefaultCategory); err != nil {
		return nil, err
	}

	// Apply default category if needed
	if model.HasCategories && b.config.DefaultCategory != "" {
		ApplyDefaultCategory(model, b.config.DefaultCategory)
	}

	return model, nil
}

// shouldIncludeTarget determines if a target should be included in the help output.
// A target is included if:
// 1. It has documentation (len(Documentation) > 0), OR
// 2. It's in the IncludeTargets list, OR
// 3. It's .PHONY and IncludeAllPhony is true
func (b *Builder) shouldIncludeTarget(target *Target) bool {
	// Include if documented
	if len(target.Documentation) > 0 {
		return true
	}

	// Include if explicitly listed
	for _, includedName := range b.config.IncludeTargets {
		if target.Name == includedName {
			return true
		}
	}

	// Include if phony and IncludeAllPhony is set
	if b.config.IncludeAllPhony && b.config.PhonyTargets[target.Name] {
		return true
	}

	return false
}

// processFile handles directives and targets from a single parsed file.
//
// # Algorithm: Two-Pointer Line-Order Merge
//
// This function uses a two-pointer algorithm to merge directives and targets
// in line-number order. This ensures that directives (documentation, categories,
// variables, aliases) are correctly associated with the target that follows them.
//
// The algorithm maintains two indices:
//   - directiveIdx: points to the next unprocessed directive
//   - targetIdx: points to the next unprocessed target
//
// On each iteration, we compare line numbers and process whichever comes first.
// Directives accumulate in "pending" state until a target is encountered, at which
// point all pending directives are associated with that target.
//
// # Example
//
// Given this Makefile content:
//
//	Line 1:  ## !category Build      <- directive (category)
//	Line 2:  ## !var CC Compiler     <- directive (var)
//	Line 3:  ## Build the project    <- directive (doc)
//	Line 4:  build:                  <- target
//	Line 5:  ## Run tests            <- directive (doc)
//	Line 6:  test:                   <- target
//
// Processing order:
//  1. Line 1: Set currentCategory = "Build"
//  2. Line 2: Add CC to pendingVars
//  3. Line 3: Add "Build the project" to pendingDocs
//  4. Line 4: Create target "build" with pendingDocs=["Build the project"],
//     pendingVars=[CC], category="Build". Clear pending state.
//  5. Line 5: Add "Run tests" to pendingDocs
//  6. Line 6: Create target "test" with pendingDocs=["Run tests"],
//     category="Build" (inherited). Clear pending state.
//
// # Special Cases
//
//   - !file directives: Added to model.FileDocs (not associated with targets)
//   - !category directives: Update currentCategory for subsequent targets
//   - Duplicate targets: If a target was already processed from another file,
//     skip it and clear pending state (first definition wins)
//
// # Why This Approach
//
// The parser extracts directives and targets separately for simplicity. This
// function reunites them based on line order, which matches how developers
// write Makefiles: documentation immediately precedes the target it documents.
func (b *Builder) processFile(
	file *parser.ParsedFile,
	model *HelpModel,
	categoryMap map[string]*Category,
	targetMap map[string]*Target,
	targetToCategory map[string]string,
	categoryOrder *int,
	targetOrder *int,
) {
	// Build a sorted list of target line numbers for association
	type targetLine struct {
		name string
		line int
	}
	var targetLines []targetLine
	for name, line := range file.TargetMap {
		targetLines = append(targetLines, targetLine{name: name, line: line})
	}
	sort.Slice(targetLines, func(i, j int) bool {
		return targetLines[i].line < targetLines[j].line
	})

	// Track current state
	var currentCategory string

	// Accumulate directives for the next target
	var pendingDocs []string
	var pendingVars []Variable
	var pendingAliases []string

	// Process directives in file order
	directiveIdx := 0
	targetIdx := 0

	for directiveIdx < len(file.Directives) || targetIdx < len(targetLines) {
		// Determine what comes next by comparing line numbers.
		// Initialize to max int so exhausted lists sort to the end.
		const maxInt = int(^uint(0) >> 1)
		var nextDirectiveLine int = maxInt
		var nextTargetLine int = maxInt

		if directiveIdx < len(file.Directives) {
			nextDirectiveLine = file.Directives[directiveIdx].LineNumber
		}
		if targetIdx < len(targetLines) {
			nextTargetLine = targetLines[targetIdx].line
		}

		if nextDirectiveLine < nextTargetLine {
			// Process directive
			directive := file.Directives[directiveIdx]
			directiveIdx++

			switch directive.Type {
			case parser.DirectiveFile:
				if directive.Value != "" {
					model.FileDocs = append(model.FileDocs, directive.Value)
				}

			case parser.DirectiveCategory:
				model.HasCategories = true
				currentCategory = directive.Value

				// Handle !category _ as reset to uncategorized
				if currentCategory == "_" {
					currentCategory = ""
					// Don't create a category entry for "_"
					break
				}

				// Create category if it doesn't exist
				if _, exists := categoryMap[currentCategory]; !exists {
					categoryMap[currentCategory] = &Category{
						Name:           currentCategory,
						Targets:        []Target{},
						DiscoveryOrder: *categoryOrder,
					}
					*categoryOrder++
				}

			case parser.DirectiveDoc:
				pendingDocs = append(pendingDocs, directive.Value)

			case parser.DirectiveVar:
				pendingVars = append(pendingVars, b.parseVarDirective(directive.Value))

			case parser.DirectiveAlias:
				pendingAliases = append(pendingAliases, b.parseAliasDirective(directive.Value)...)
			}
		} else {
			// Process target - associate pending directives with it
			tl := targetLines[targetIdx]
			targetIdx++

			// Skip if target already processed from another file
			if _, exists := targetMap[tl.name]; exists {
				pendingDocs = nil
				pendingVars = nil
				pendingAliases = nil
				continue
			}

			// Create target
			target := &Target{
				Name:           tl.name,
				Aliases:        pendingAliases,
				Documentation:  pendingDocs,
				Variables:      pendingVars,
				DiscoveryOrder: *targetOrder,
				SourceFile:     file.Path,
				LineNumber:     tl.line,
			}
			*targetOrder++

			targetMap[tl.name] = target
			targetToCategory[tl.name] = currentCategory

			// Clear pending state
			pendingDocs = nil
			pendingVars = nil
			pendingAliases = nil
		}
	}
}

// parseVarDirective parses !var directive: NAME - description
// or just NAME if no description is provided.
func (b *Builder) parseVarDirective(value string) Variable {
	parts := strings.SplitN(value, " - ", 2)
	if len(parts) == 2 {
		return Variable{
			Name:        strings.TrimSpace(parts[0]),
			Description: strings.TrimSpace(parts[1]),
		}
	}
	return Variable{
		Name:        strings.TrimSpace(value),
		Description: "",
	}
}

// parseAliasDirective parses !alias directive: alias1, alias2, ...
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
