package model

import (
	"sort"
	"strings"

	"github.com/sdlcforge/make-help/internal/parser"
	"github.com/sdlcforge/make-help/internal/summary"
)

// maxInt is the maximum value of int on the current platform.
// On 64-bit systems: 9223372036854775807 (2^63 - 1)
// On 32-bit systems: 2147483647 (2^31 - 1)
const maxInt = int(^uint(0) >> 1)

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

	// Dependencies maps target names to their prerequisite targets.
	// Used for detecting implicit aliases (phony target with single phony dep, no recipe).
	Dependencies map[string][]string

	// HasRecipe maps target names to whether they have a recipe.
	// Used for detecting implicit aliases.
	HasRecipe map[string]bool
}

// Builder constructs a HelpModel from parsed Makefile directives.
// It aggregates file documentation, groups targets by category,
// and associates aliases and variables with targets.
type Builder struct {
	config      *BuilderConfig
	extractor   *summary.Extractor
	notAliasSet map[string]bool // Targets marked with !notalias directive
}

// NewBuilder creates a new Builder with the given configuration.
func NewBuilder(config *BuilderConfig) *Builder {
	if config == nil {
		config = &BuilderConfig{}
	}
	if config.PhonyTargets == nil {
		config.PhonyTargets = make(map[string]bool)
	}
	if config.Dependencies == nil {
		config.Dependencies = make(map[string][]string)
	}
	if config.HasRecipe == nil {
		config.HasRecipe = make(map[string]bool)
	}
	return &Builder{
		config:      config,
		extractor:   summary.NewExtractor(),
		notAliasSet: make(map[string]bool),
	}
}

// NotAliasTargets returns the set of targets marked with !notalias directive.
func (b *Builder) NotAliasTargets() map[string]bool {
	return b.notAliasSet
}

// Build constructs a HelpModel from parsed files.
// It processes directives in order, groups targets by category,
// and validates categorization rules.
func (b *Builder) Build(parsedFiles []*parser.ParsedFile) (*HelpModel, error) {
	model := &HelpModel{
		FileDocs:   []FileDoc{},
		Categories: []Category{},
	}

	categoryMap := make(map[string]*Category)
	targetMap := make(map[string]*Target)
	targetToCategory := make(map[string]string) // target name -> category name
	fileDocMap := make(map[string]*FileDoc)     // source file path -> FileDoc

	categoryOrder := 0
	targetOrder := 0
	fileOrder := 0

	for _, file := range parsedFiles {
		b.processFile(file, model, categoryMap, targetMap, targetToCategory, fileDocMap, &categoryOrder, &targetOrder, &fileOrder)
	}

	// Convert fileDocMap to slice
	for _, fileDoc := range fileDocMap {
		model.FileDocs = append(model.FileDocs, *fileDoc)
	}
	// Sort by discovery order for deterministic output
	sort.Slice(model.FileDocs, func(i, j int) bool {
		return model.FileDocs[i].DiscoveryOrder < model.FileDocs[j].DiscoveryOrder
	})

	// Detect implicit aliases: phony targets with single phony dependency and no recipe
	implicitAliases := b.detectImplicitAliases(targetMap)

	// Assign targets to categories with filtering
	for targetName, target := range targetMap {
		// Skip if this target is an implicit alias of another target
		if _, isAlias := implicitAliases[targetName]; isAlias {
			continue
		}

		// Apply filtering logic
		shouldInclude := b.shouldIncludeTarget(target)
		if !shouldInclude {
			continue
		}

		// Add implicit aliases to this target
		for aliasName, depName := range implicitAliases {
			if depName == targetName {
				target.Aliases = append(target.Aliases, aliasName)
			}
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

// detectImplicitAliases finds targets that are implicit aliases of other targets.
// A target is an implicit alias if:
//   - It has no documentation (documented targets are semantically distinct)
//   - It is not marked with !notalias directive
//   - It is .PHONY
//   - It has exactly one dependency
//   - That dependency is also .PHONY
//   - It has no recipe (no commands)
//
// Returns a map from alias target name to the target it aliases.
func (b *Builder) detectImplicitAliases(targetMap map[string]*Target) map[string]string {
	aliases := make(map[string]string)

	for targetName, target := range targetMap {
		// Skip if target has documentation (documented targets are not implicit aliases)
		if len(target.Documentation) > 0 {
			continue
		}

		// Skip if marked with !notalias
		if b.notAliasSet[targetName] {
			continue
		}

		// Check conditions for implicit alias:
		// 1. Target is .PHONY
		if !b.config.PhonyTargets[targetName] {
			continue
		}

		// 2. Target has exactly one dependency
		deps := b.config.Dependencies[targetName]
		if len(deps) != 1 {
			continue
		}

		// 3. The dependency is also .PHONY
		depName := deps[0]
		if !b.config.PhonyTargets[depName] {
			continue
		}

		// 4. Target has no recipe
		if b.config.HasRecipe[targetName] {
			continue
		}

		// This target is an implicit alias of its dependency
		aliases[targetName] = depName
	}

	return aliases
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
	fileDocMap map[string]*FileDoc,
	categoryOrder *int,
	targetOrder *int,
	fileOrder *int,
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
	var pendingNotAlias bool

	// Process directives in file order
	directiveIdx := 0
	targetIdx := 0

	for directiveIdx < len(file.Directives) || targetIdx < len(targetLines) {
		// Determine what comes next by comparing line numbers.
		// Initialize to maxInt so exhausted lists sort to the end.
		var nextDirectiveLine = maxInt
		var nextTargetLine = maxInt

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
					// Get or create FileDoc for this file
					fileDoc, exists := fileDocMap[file.Path]
					if !exists {
						fileDoc = &FileDoc{
							SourceFile:     file.Path,
							Documentation:  []string{},
							DiscoveryOrder: *fileOrder,
							IsEntryPoint:   *fileOrder == 0, // First file is entry point
						}
						*fileOrder++
						fileDocMap[file.Path] = fileDoc
					}

					// Concatenate multiple !file blocks with blank line separation
					if len(fileDoc.Documentation) > 0 {
						fileDoc.Documentation = append(fileDoc.Documentation, "") // Blank line
					}
					fileDoc.Documentation = append(fileDoc.Documentation, directive.Value)
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

			case parser.DirectiveNotAlias:
				pendingNotAlias = true
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

			// Track targets marked with !notalias
			if pendingNotAlias {
				b.notAliasSet[tl.name] = true
			}

			// Clear pending state
			pendingDocs = nil
			pendingVars = nil
			pendingAliases = nil
			pendingNotAlias = false
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
