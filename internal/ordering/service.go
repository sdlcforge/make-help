package ordering

import (
	"github.com/sdlcforge/make-help/internal/model"
)

// Service handles category, target, and file ordering based on configuration.
type Service struct {
	keepOrderCategories bool
	keepOrderTargets    bool
	keepOrderFiles      bool
	categoryOrder       []string
}

// NewService creates a new ordering service with the given ordering preferences.
func NewService(keepOrderCategories, keepOrderTargets, keepOrderFiles bool, categoryOrder []string) *Service {
	return &Service{
		keepOrderCategories: keepOrderCategories,
		keepOrderTargets:    keepOrderTargets,
		keepOrderFiles:      keepOrderFiles,
		categoryOrder:       categoryOrder,
	}
}

// ApplyOrdering applies the configured ordering strategy to files, categories, and targets.
// It modifies the HelpModel in place.
func (s *Service) ApplyOrdering(helpModel *model.HelpModel) error {
	// Order files
	s.orderFiles(helpModel)

	// Order categories
	if err := s.orderCategories(helpModel); err != nil {
		return err
	}

	// Order targets within each category
	for i := range helpModel.Categories {
		s.orderTargets(&helpModel.Categories[i])
	}

	return nil
}

// orderCategories applies the configured category ordering strategy.
func (s *Service) orderCategories(helpModel *model.HelpModel) error {
	// If explicit category order is specified, use it
	if len(s.categoryOrder) > 0 {
		return applyExplicitCategoryOrder(helpModel, s.categoryOrder)
	}

	// If keep-order-categories is set, sort by discovery order
	if s.keepOrderCategories {
		sortCategoriesByDiscoveryOrder(helpModel.Categories)
		return nil
	}

	// Default: sort alphabetically
	sortCategoriesAlphabetically(helpModel.Categories)
	return nil
}

// orderTargets applies the configured target ordering strategy to a category.
func (s *Service) orderTargets(category *model.Category) {
	// If keep-order-targets is set, sort by discovery order
	if s.keepOrderTargets {
		sortTargetsByDiscoveryOrder(category.Targets)
		return
	}

	// Default: sort alphabetically
	sortTargetsAlphabetically(category.Targets)
}

// orderFiles applies the configured file ordering strategy.
// The entry point file is always kept first, then other files are sorted
// alphabetically or by discovery order based on the configuration.
func (s *Service) orderFiles(helpModel *model.HelpModel) {
	if len(helpModel.FileDocs) == 0 {
		return
	}

	// If keep-order-files is set, sort by discovery order
	if s.keepOrderFiles {
		sortFilesByDiscoveryOrder(helpModel.FileDocs)
		return
	}

	// Default: sort alphabetically (but keep entry point first)
	sortFilesAlphabetically(helpModel.FileDocs)
}
