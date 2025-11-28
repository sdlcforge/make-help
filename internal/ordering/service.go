package ordering

import (
	"github.com/sdlcforge/make-help/internal/cli"
	"github.com/sdlcforge/make-help/internal/model"
)

// Service handles category and target ordering based on configuration.
type Service struct {
	config *cli.Config
}

// NewService creates a new ordering service with the given configuration.
func NewService(config *cli.Config) *Service {
	return &Service{
		config: config,
	}
}

// ApplyOrdering applies the configured ordering strategy to categories and targets.
// It modifies the HelpModel in place.
func (s *Service) ApplyOrdering(helpModel *model.HelpModel) error {
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
	if len(s.config.CategoryOrder) > 0 {
		return applyExplicitCategoryOrder(helpModel, s.config.CategoryOrder)
	}

	// If keep-order-categories is set, sort by discovery order
	if s.config.KeepOrderCategories {
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
	if s.config.KeepOrderTargets {
		sortTargetsByDiscoveryOrder(category.Targets)
		return
	}

	// Default: sort alphabetically
	sortTargetsAlphabetically(category.Targets)
}
