package ordering

import (
	"testing"

	"github.com/sdlcforge/make-help/internal/errors"
	"github.com/sdlcforge/make-help/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create test categories
func createTestCategories() []model.Category {
	return []model.Category{
		{
			Name:           "Development",
			DiscoveryOrder: 2,
			Targets: []model.Target{
				{Name: "test", DiscoveryOrder: 3},
				{Name: "build", DiscoveryOrder: 1},
				{Name: "lint", DiscoveryOrder: 2},
			},
		},
		{
			Name:           "Deployment",
			DiscoveryOrder: 1,
			Targets: []model.Target{
				{Name: "deploy", DiscoveryOrder: 2},
				{Name: "package", DiscoveryOrder: 1},
			},
		},
		{
			Name:           "CI",
			DiscoveryOrder: 3,
			Targets: []model.Target{
				{Name: "ci-test", DiscoveryOrder: 1},
				{Name: "ci-build", DiscoveryOrder: 2},
			},
		},
	}
}

// Helper function to create test model
func createTestModel() *model.HelpModel {
	return &model.HelpModel{
		Categories:    createTestCategories(),
		HasCategories: true,
	}
}

func TestNewService(t *testing.T) {
	service := NewService(false, false, false, []string{})

	assert.NotNil(t, service)
	assert.NotNil(t, service)
}

func TestApplyOrdering_DefaultAlphabeticalCategories(t *testing.T) {
	service := NewService(false, false, false, []string{})
	helpModel := createTestModel()

	err := service.ApplyOrdering(helpModel)
	require.NoError(t, err)

	// Categories should be in alphabetical order
	assert.Equal(t, "CI", helpModel.Categories[0].Name)
	assert.Equal(t, "Deployment", helpModel.Categories[1].Name)
	assert.Equal(t, "Development", helpModel.Categories[2].Name)
}

func TestApplyOrdering_DefaultAlphabeticalTargets(t *testing.T) {
	service := NewService(false, false, false, []string{})
	helpModel := createTestModel()

	err := service.ApplyOrdering(helpModel)
	require.NoError(t, err)

	// Find Development category and check target order
	for _, cat := range helpModel.Categories {
		if cat.Name == "Development" {
			assert.Equal(t, "build", cat.Targets[0].Name)
			assert.Equal(t, "lint", cat.Targets[1].Name)
			assert.Equal(t, "test", cat.Targets[2].Name)
		}
	}
}

func TestApplyOrdering_KeepOrderCategories(t *testing.T) {
	service := NewService(true, false, false, []string{})
	helpModel := createTestModel()

	err := service.ApplyOrdering(helpModel)
	require.NoError(t, err)

	// Categories should be in discovery order
	assert.Equal(t, 1, helpModel.Categories[0].DiscoveryOrder)
	assert.Equal(t, 2, helpModel.Categories[1].DiscoveryOrder)
	assert.Equal(t, 3, helpModel.Categories[2].DiscoveryOrder)
}

func TestApplyOrdering_KeepOrderTargets(t *testing.T) {
	service := NewService(false, true, false, []string{})
	helpModel := createTestModel()

	err := service.ApplyOrdering(helpModel)
	require.NoError(t, err)

	// Find Development category and check target order (by discovery order)
	for _, cat := range helpModel.Categories {
		if cat.Name == "Development" {
			assert.Equal(t, 1, cat.Targets[0].DiscoveryOrder)
			assert.Equal(t, 2, cat.Targets[1].DiscoveryOrder)
			assert.Equal(t, 3, cat.Targets[2].DiscoveryOrder)
		}
	}
}

func TestApplyOrdering_KeepOrderBoth(t *testing.T) {
	service := NewService(true, true, false, []string{})
	helpModel := createTestModel()

	err := service.ApplyOrdering(helpModel)
	require.NoError(t, err)

	// Categories in discovery order
	assert.Equal(t, "Deployment", helpModel.Categories[0].Name)
	assert.Equal(t, "Development", helpModel.Categories[1].Name)
	assert.Equal(t, "CI", helpModel.Categories[2].Name)

	// Targets within each category in discovery order
	for _, cat := range helpModel.Categories {
		if cat.Name == "Deployment" {
			assert.Equal(t, "package", cat.Targets[0].Name)
			assert.Equal(t, "deploy", cat.Targets[1].Name)
		}
	}
}

func TestApplyOrdering_ExplicitCategoryOrder(t *testing.T) {
	service := NewService(false, false, false, []string{"Development", "CI"})
	helpModel := createTestModel()

	err := service.ApplyOrdering(helpModel)
	require.NoError(t, err)

	// Categories should be in specified order, with remaining appended alphabetically
	assert.Equal(t, "Development", helpModel.Categories[0].Name)
	assert.Equal(t, "CI", helpModel.Categories[1].Name)
	assert.Equal(t, "Deployment", helpModel.Categories[2].Name) // Remaining, alphabetically
}

func TestApplyOrdering_ExplicitCategoryOrder_AllSpecified(t *testing.T) {
	service := NewService(false, false, false, []string{"CI", "Development", "Deployment"})
	helpModel := createTestModel()

	err := service.ApplyOrdering(helpModel)
	require.NoError(t, err)

	assert.Equal(t, "CI", helpModel.Categories[0].Name)
	assert.Equal(t, "Development", helpModel.Categories[1].Name)
	assert.Equal(t, "Deployment", helpModel.Categories[2].Name)
}

func TestApplyOrdering_ExplicitCategoryOrder_UnknownCategory(t *testing.T) {
	service := NewService(false, false, false, []string{"Development", "NonExistent", "CI"})
	helpModel := createTestModel()

	err := service.ApplyOrdering(helpModel)
	require.Error(t, err)

	var unknownCatErr *errors.UnknownCategoryError
	require.ErrorAs(t, err, &unknownCatErr)
	assert.Equal(t, "NonExistent", unknownCatErr.CategoryName)
	assert.Contains(t, unknownCatErr.Available, "CI")
	assert.Contains(t, unknownCatErr.Available, "Development")
	assert.Contains(t, unknownCatErr.Available, "Deployment")
}

func TestApplyOrdering_ExplicitCategoryOrder_WithKeepOrderTargets(t *testing.T) {
	service := NewService(false, true, false, []string{"Deployment"})
	helpModel := createTestModel()

	err := service.ApplyOrdering(helpModel)
	require.NoError(t, err)

	// Categories in explicit order
	assert.Equal(t, "Deployment", helpModel.Categories[0].Name)

	// Targets in discovery order
	assert.Equal(t, "package", helpModel.Categories[0].Targets[0].Name)
	assert.Equal(t, "deploy", helpModel.Categories[0].Targets[1].Name)
}

func TestApplyOrdering_EmptyModel(t *testing.T) {
	service := NewService(false, false, false, []string{})
	helpModel := &model.HelpModel{
		Categories:    []model.Category{},
		HasCategories: false,
	}

	err := service.ApplyOrdering(helpModel)
	require.NoError(t, err)

	assert.Len(t, helpModel.Categories, 0)
}

func TestApplyOrdering_SingleCategory(t *testing.T) {
	service := NewService(false, false, false, []string{})
	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name:           "Build",
				DiscoveryOrder: 1,
				Targets: []model.Target{
					{Name: "test", DiscoveryOrder: 2},
					{Name: "build", DiscoveryOrder: 1},
				},
			},
		},
		HasCategories: true,
	}

	err := service.ApplyOrdering(helpModel)
	require.NoError(t, err)

	assert.Len(t, helpModel.Categories, 1)
	assert.Equal(t, "Build", helpModel.Categories[0].Name)
	assert.Equal(t, "build", helpModel.Categories[0].Targets[0].Name)
	assert.Equal(t, "test", helpModel.Categories[0].Targets[1].Name)
}

func TestSortCategoriesAlphabetically(t *testing.T) {
	categories := []model.Category{
		{Name: "Zebra"},
		{Name: "apple"},
		{Name: "Banana"},
	}

	sortCategoriesAlphabetically(categories)

	assert.Equal(t, "apple", categories[0].Name)
	assert.Equal(t, "Banana", categories[1].Name)
	assert.Equal(t, "Zebra", categories[2].Name)
}

func TestSortCategoriesByDiscoveryOrder(t *testing.T) {
	categories := []model.Category{
		{Name: "Third", DiscoveryOrder: 3},
		{Name: "First", DiscoveryOrder: 1},
		{Name: "Second", DiscoveryOrder: 2},
	}

	sortCategoriesByDiscoveryOrder(categories)

	assert.Equal(t, "First", categories[0].Name)
	assert.Equal(t, "Second", categories[1].Name)
	assert.Equal(t, "Third", categories[2].Name)
}

func TestSortTargetsAlphabetically(t *testing.T) {
	targets := []model.Target{
		{Name: "zebra"},
		{Name: "Apple"},
		{Name: "banana"},
	}

	sortTargetsAlphabetically(targets)

	assert.Equal(t, "Apple", targets[0].Name)
	assert.Equal(t, "banana", targets[1].Name)
	assert.Equal(t, "zebra", targets[2].Name)
}

func TestSortTargetsByDiscoveryOrder(t *testing.T) {
	targets := []model.Target{
		{Name: "third", DiscoveryOrder: 3},
		{Name: "first", DiscoveryOrder: 1},
		{Name: "second", DiscoveryOrder: 2},
	}

	sortTargetsByDiscoveryOrder(targets)

	assert.Equal(t, "first", targets[0].Name)
	assert.Equal(t, "second", targets[1].Name)
	assert.Equal(t, "third", targets[2].Name)
}

func TestApplyExplicitCategoryOrder_PartialOrder(t *testing.T) {
	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{Name: "E"},
			{Name: "D"},
			{Name: "C"},
			{Name: "B"},
			{Name: "A"},
		},
	}

	order := []string{"C", "A"}
	err := applyExplicitCategoryOrder(helpModel, order)
	require.NoError(t, err)

	// First two should be in explicit order
	assert.Equal(t, "C", helpModel.Categories[0].Name)
	assert.Equal(t, "A", helpModel.Categories[1].Name)

	// Remaining should be alphabetical
	assert.Equal(t, "B", helpModel.Categories[2].Name)
	assert.Equal(t, "D", helpModel.Categories[3].Name)
	assert.Equal(t, "E", helpModel.Categories[4].Name)
}

func TestApplyExplicitCategoryOrder_DuplicatesInOrder(t *testing.T) {
	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{Name: "A"},
			{Name: "B"},
			{Name: "C"},
		},
	}

	// Test that duplicates in order don't cause issues
	order := []string{"B", "B", "A"}
	err := applyExplicitCategoryOrder(helpModel, order)
	require.NoError(t, err)

	// B should appear first (only once)
	assert.Equal(t, "B", helpModel.Categories[0].Name)
	assert.Equal(t, "A", helpModel.Categories[1].Name)
	assert.Equal(t, "C", helpModel.Categories[2].Name)
}

func TestService_String(t *testing.T) {
	service := NewService(true, false, false, []string{"Build", "Deploy"})

	result := service.String()
	assert.Contains(t, result, "keepOrderCategories=true")
	assert.Contains(t, result, "keepOrderTargets=false")
	assert.Contains(t, result, "categoryOrder=[Build Deploy]")
}

func TestApplyOrdering_CaseInsensitiveSorting(t *testing.T) {
	service := NewService(false, false, false, []string{})
	helpModel := &model.HelpModel{
		Categories: []model.Category{
			{
				Name: "Build",
				Targets: []model.Target{
					{Name: "Test"},
					{Name: "build"},
					{Name: "LINT"},
					{Name: "compile"},
				},
			},
		},
		HasCategories: true,
	}

	err := service.ApplyOrdering(helpModel)
	require.NoError(t, err)

	// Targets should be case-insensitively sorted
	targets := helpModel.Categories[0].Targets
	assert.Equal(t, "build", targets[0].Name)
	assert.Equal(t, "compile", targets[1].Name)
	assert.Equal(t, "LINT", targets[2].Name)
	assert.Equal(t, "Test", targets[3].Name)
}

func TestApplyOrdering_PreservesOtherFields(t *testing.T) {
	service := NewService(false, false, false, []string{})
	helpModel := &model.HelpModel{
		FileDocs: []model.FileDoc{
			{
				SourceFile:     "Makefile",
				Documentation:  []string{"File documentation"},
				DiscoveryOrder: 0,
				IsEntryPoint:   true,
			},
		},
		Categories: []model.Category{
			{
				Name:           "Build",
				DiscoveryOrder: 1,
				Targets: []model.Target{
					{
						Name:           "test",
						Aliases:        []string{"t"},
						Documentation:  []string{"Run tests"},
						Summary:        []string{"Run tests"},
						Variables:      []model.Variable{{Name: "VERBOSE"}},
						DiscoveryOrder: 1,
						SourceFile:     "/path/to/Makefile",
						LineNumber:     10,
					},
				},
			},
		},
		HasCategories:   true,
		DefaultCategory: "Other",
	}

	err := service.ApplyOrdering(helpModel)
	require.NoError(t, err)

	// Check that other fields are preserved
	assert.Len(t, helpModel.FileDocs, 1)
	assert.Equal(t, "Makefile", helpModel.FileDocs[0].SourceFile)
	assert.Equal(t, []string{"File documentation"}, helpModel.FileDocs[0].Documentation)
	assert.True(t, helpModel.HasCategories)
	assert.Equal(t, "Other", helpModel.DefaultCategory)

	target := helpModel.Categories[0].Targets[0]
	assert.Equal(t, "test", target.Name)
	assert.Equal(t, []string{"t"}, target.Aliases)
	assert.Equal(t, []string{"Run tests"}, target.Documentation)
	assert.Equal(t, "Run tests", target.Summary[0])
	assert.Len(t, target.Variables, 1)
	assert.Equal(t, "VERBOSE", target.Variables[0].Name)
	assert.Equal(t, "/path/to/Makefile", target.SourceFile)
	assert.Equal(t, 10, target.LineNumber)
}

func TestApplyOrdering_KeepOrderFiles(t *testing.T) {
	service := NewService(false, false, true, []string{})
	helpModel := &model.HelpModel{
		FileDocs: []model.FileDoc{
			{
				SourceFile:     "Makefile",
				Documentation:  []string{"Entry point"},
				DiscoveryOrder: 0,
				IsEntryPoint:   true,
			},
			{
				SourceFile:     "make/zulu.mk",
				Documentation:  []string{"Third file"},
				DiscoveryOrder: 2,
				IsEntryPoint:   false,
			},
			{
				SourceFile:     "make/alpha.mk",
				Documentation:  []string{"Second file"},
				DiscoveryOrder: 1,
				IsEntryPoint:   false,
			},
		},
	}

	err := service.ApplyOrdering(helpModel)
	require.NoError(t, err)

	// Files should be in discovery order
	require.Len(t, helpModel.FileDocs, 3)
	assert.Equal(t, 0, helpModel.FileDocs[0].DiscoveryOrder)
	assert.Equal(t, "Makefile", helpModel.FileDocs[0].SourceFile)
	assert.Equal(t, 1, helpModel.FileDocs[1].DiscoveryOrder)
	assert.Equal(t, "make/alpha.mk", helpModel.FileDocs[1].SourceFile)
	assert.Equal(t, 2, helpModel.FileDocs[2].DiscoveryOrder)
	assert.Equal(t, "make/zulu.mk", helpModel.FileDocs[2].SourceFile)
}

func TestSortFilesAlphabetically_EntryPointFirst(t *testing.T) {
	files := []model.FileDoc{
		{
			SourceFile:     "make/zulu.mk",
			DiscoveryOrder: 2,
			IsEntryPoint:   false,
		},
		{
			SourceFile:     "Makefile",
			DiscoveryOrder: 0,
			IsEntryPoint:   true,
		},
		{
			SourceFile:     "make/alpha.mk",
			DiscoveryOrder: 1,
			IsEntryPoint:   false,
		},
	}

	sortFilesAlphabetically(files)

	// Entry point should be first, followed by alphabetically sorted files
	assert.Equal(t, "Makefile", files[0].SourceFile)
	assert.True(t, files[0].IsEntryPoint)
	assert.Equal(t, "make/alpha.mk", files[1].SourceFile)
	assert.Equal(t, "make/zulu.mk", files[2].SourceFile)
}
