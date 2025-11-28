package model

import (
	"sort"
	"strings"

	"github.com/sdlcforge/make-help/internal/cli"
	"github.com/sdlcforge/make-help/internal/parser"
	"github.com/sdlcforge/make-help/internal/summary"
)

// Builder constructs a HelpModel from parsed Makefile directives.
// It aggregates file documentation, groups targets by category,
// and associates aliases and variables with targets.
type Builder struct {
	config    *cli.Config
	extractor *summary.Extractor
}

// NewBuilder creates a new Builder with the given configuration.
func NewBuilder(config *cli.Config) *Builder {
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

	// Assign targets to categories
	for targetName, target := range targetMap {
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
	if err := ValidateCategorization(model, b.config); err != nil {
		return nil, err
	}

	// Apply default category if needed
	if model.HasCategories && b.config.DefaultCategory != "" {
		ApplyDefaultCategory(model, b.config.DefaultCategory)
	}

	return model, nil
}

// processFile handles directives and targets from a single parsed file.
// The scanner outputs directives that belong to targets - directives immediately
// before a target (without intervening non-doc lines) are associated with that target.
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
		// Determine what comes next: directive or target
		var nextDirectiveLine int = int(^uint(0) >> 1) // Max int
		var nextTargetLine int = int(^uint(0) >> 1)

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

// parseVarDirective parses @var directive: NAME - description
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

// parseAliasDirective parses @alias directive: alias1, alias2, ...
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
