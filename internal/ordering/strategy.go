package ordering

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sdlcforge/make-help/internal/errors"
	"github.com/sdlcforge/make-help/internal/model"
)

// sortCategoriesAlphabetically sorts categories by name in ascending order.
// Case-insensitive comparison is used for natural sorting.
func sortCategoriesAlphabetically(categories []model.Category) {
	sort.Slice(categories, func(i, j int) bool {
		return strings.ToLower(categories[i].Name) < strings.ToLower(categories[j].Name)
	})
}

// sortCategoriesByDiscoveryOrder sorts categories by their discovery order.
// This preserves the order in which categories were first encountered during parsing.
func sortCategoriesByDiscoveryOrder(categories []model.Category) {
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].DiscoveryOrder < categories[j].DiscoveryOrder
	})
}

// applyExplicitCategoryOrder applies an explicit category order.
// Categories in the order list are placed first (in the specified order),
// and remaining categories are appended alphabetically.
// Returns an error if any category in the order list doesn't exist.
func applyExplicitCategoryOrder(helpModel *model.HelpModel, order []string) error {
	// Build a map of category names to categories for quick lookup
	categoryMap := make(map[string]*model.Category)
	for i := range helpModel.Categories {
		categoryMap[helpModel.Categories[i].Name] = &helpModel.Categories[i]
	}

	// Validate that all categories in the order list exist
	for _, name := range order {
		if _, exists := categoryMap[name]; !exists {
			// Collect all available category names for the error message
			availableCategories := make([]string, 0, len(categoryMap))
			for catName := range categoryMap {
				availableCategories = append(availableCategories, catName)
			}
			sort.Strings(availableCategories)
			return &errors.UnknownCategoryError{
				CategoryName: name,
				Available:    availableCategories,
			}
		}
	}

	// Build the ordered category list
	ordered := make([]model.Category, 0, len(helpModel.Categories))
	usedCategories := make(map[string]bool)

	// First, add categories in the specified order
	for _, name := range order {
		if cat, exists := categoryMap[name]; exists && !usedCategories[name] {
			ordered = append(ordered, *cat)
			usedCategories[name] = true
		}
	}

	// Then, add remaining categories alphabetically
	remaining := make([]model.Category, 0)
	for i := range helpModel.Categories {
		if !usedCategories[helpModel.Categories[i].Name] {
			remaining = append(remaining, helpModel.Categories[i])
		}
	}
	sortCategoriesAlphabetically(remaining)
	ordered = append(ordered, remaining...)

	// Replace the categories in the model
	helpModel.Categories = ordered
	return nil
}

// sortTargetsAlphabetically sorts targets by name in ascending order.
// Case-insensitive comparison is used for natural sorting.
func sortTargetsAlphabetically(targets []model.Target) {
	sort.Slice(targets, func(i, j int) bool {
		return strings.ToLower(targets[i].Name) < strings.ToLower(targets[j].Name)
	})
}

// sortTargetsByDiscoveryOrder sorts targets by their discovery order.
// This preserves the order in which targets were first encountered during parsing.
func sortTargetsByDiscoveryOrder(targets []model.Target) {
	sort.Slice(targets, func(i, j int) bool {
		return targets[i].DiscoveryOrder < targets[j].DiscoveryOrder
	})
}

// String representation for debugging
func (s *Service) String() string {
	return fmt.Sprintf("OrderingService{keepOrderCategories=%v, keepOrderTargets=%v, categoryOrder=%v}",
		s.config.KeepOrderCategories, s.config.KeepOrderTargets, s.config.CategoryOrder)
}
