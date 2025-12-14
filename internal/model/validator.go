package model

import (
	"fmt"
	"strings"

	"github.com/sdlcforge/make-help/internal/errors"
)

// ValidateCategorization ensures that the categorization rules are followed:
// - If any categories exist, all targets must be categorized (unless a default category will be used)
// - Mixed categorization (some with categories, some without) is an error without a default category
func ValidateCategorization(model *HelpModel, defaultCategory string) error {
	if !model.HasCategories {
		// No categories defined, all targets are uncategorized - OK
		return nil
	}

	// Count categorized and uncategorized targets, and collect uncategorized names
	categorizedCount := 0
	var uncategorizedTargets []string

	for _, cat := range model.Categories {
		if cat.Name == "" {
			for _, t := range cat.Targets {
				uncategorizedTargets = append(uncategorizedTargets, t.Name)
			}
		} else {
			categorizedCount += len(cat.Targets)
		}
	}

	// Check for mixed categorization
	if categorizedCount > 0 && len(uncategorizedTargets) > 0 {
		if defaultCategory == "" {
			msg := fmt.Sprintf(
				"found both categorized and uncategorized targets\nUncategorized targets: %s",
				strings.Join(uncategorizedTargets, ", "),
			)
			return errors.NewMixedCategorizationError(msg)
		}
	}

	return nil
}

// ApplyDefaultCategory moves all targets from the empty category
// to the specified default category.
func ApplyDefaultCategory(model *HelpModel, defaultCategory string) {
	if defaultCategory == "" {
		return
	}

	// Find the empty category and the default category
	var emptyCategory *Category
	var defaultCat *Category
	emptyIdx := -1

	for i := range model.Categories {
		if model.Categories[i].Name == "" {
			emptyCategory = &model.Categories[i]
			emptyIdx = i
		}
		if model.Categories[i].Name == defaultCategory {
			defaultCat = &model.Categories[i]
		}
	}

	// Nothing to do if no uncategorized targets
	if emptyCategory == nil || len(emptyCategory.Targets) == 0 {
		return
	}

	// Create default category if it doesn't exist
	if defaultCat == nil {
		newCat := Category{
			Name:           defaultCategory,
			Targets:        emptyCategory.Targets,
			DiscoveryOrder: emptyCategory.DiscoveryOrder,
		}
		model.Categories = append(model.Categories, newCat)
	} else {
		// Move targets to existing default category
		defaultCat.Targets = append(defaultCat.Targets, emptyCategory.Targets...)
	}

	// Remove the empty category
	if emptyIdx >= 0 {
		model.Categories = append(model.Categories[:emptyIdx], model.Categories[emptyIdx+1:]...)
	}
}

// CountTargetsByCategory returns the number of targets in each category.
// Used for debugging and validation.
func CountTargetsByCategory(model *HelpModel) map[string]int {
	counts := make(map[string]int)
	for _, cat := range model.Categories {
		counts[cat.Name] = len(cat.Targets)
	}
	return counts
}

// GetCategoryNames returns a list of all category names in the model.
// Used for error messages and validation.
func GetCategoryNames(model *HelpModel) []string {
	names := make([]string, 0, len(model.Categories))
	for _, cat := range model.Categories {
		if cat.Name != "" {
			names = append(names, cat.Name)
		}
	}
	return names
}

// HasCategory checks if a category with the given name exists.
func HasCategory(model *HelpModel, name string) bool {
	for _, cat := range model.Categories {
		if cat.Name == name {
			return true
		}
	}
	return false
}

// GetTarget finds a target by name across all categories.
// Returns nil if the target is not found.
func GetTarget(model *HelpModel, name string) *Target {
	for _, cat := range model.Categories {
		for i := range cat.Targets {
			if cat.Targets[i].Name == name {
				return &cat.Targets[i]
			}
		}
	}
	return nil
}

// GetTargetCount returns the total number of targets in the model.
func GetTargetCount(model *HelpModel) int {
	count := 0
	for _, cat := range model.Categories {
		count += len(cat.Targets)
	}
	return count
}
