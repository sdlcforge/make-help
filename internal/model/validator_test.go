package model

import (
	"testing"

	"github.com/sdlcforge/make-help/internal/cli"
	"github.com/sdlcforge/make-help/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCategorization_NoCategories(t *testing.T) {
	model := &HelpModel{
		HasCategories: false,
		Categories: []Category{
			{Name: "", Targets: []Target{{Name: "build"}, {Name: "test"}}},
		},
	}
	config := cli.NewConfig()

	err := ValidateCategorization(model, config)

	assert.NoError(t, err)
}

func TestValidateCategorization_AllCategorized(t *testing.T) {
	model := &HelpModel{
		HasCategories: true,
		Categories: []Category{
			{Name: "Build", Targets: []Target{{Name: "build"}}},
			{Name: "Test", Targets: []Target{{Name: "test"}}},
		},
	}
	config := cli.NewConfig()

	err := ValidateCategorization(model, config)

	assert.NoError(t, err)
}

func TestValidateCategorization_AllUncategorized(t *testing.T) {
	model := &HelpModel{
		HasCategories: true,
		Categories: []Category{
			{Name: "", Targets: []Target{{Name: "build"}, {Name: "test"}}},
		},
	}
	config := cli.NewConfig()

	err := ValidateCategorization(model, config)

	assert.NoError(t, err)
}

func TestValidateCategorization_MixedWithoutDefault(t *testing.T) {
	model := &HelpModel{
		HasCategories: true,
		Categories: []Category{
			{Name: "Build", Targets: []Target{{Name: "build"}}},
			{Name: "", Targets: []Target{{Name: "test"}}},
		},
	}
	config := cli.NewConfig()

	err := ValidateCategorization(model, config)

	require.Error(t, err)
	assert.IsType(t, &errors.MixedCategorizationError{}, err)
}

func TestValidateCategorization_MixedWithDefault(t *testing.T) {
	model := &HelpModel{
		HasCategories: true,
		Categories: []Category{
			{Name: "Build", Targets: []Target{{Name: "build"}}},
			{Name: "", Targets: []Target{{Name: "test"}}},
		},
	}
	config := cli.NewConfig()
	config.DefaultCategory = "Other"

	err := ValidateCategorization(model, config)

	assert.NoError(t, err)
}

func TestApplyDefaultCategory_NoUncategorized(t *testing.T) {
	model := &HelpModel{
		HasCategories: true,
		Categories: []Category{
			{Name: "Build", Targets: []Target{{Name: "build"}}},
		},
	}

	ApplyDefaultCategory(model, "Other")

	// No change - no uncategorized targets
	assert.Len(t, model.Categories, 1)
	assert.Equal(t, "Build", model.Categories[0].Name)
}

func TestApplyDefaultCategory_EmptyDefaultCategory(t *testing.T) {
	model := &HelpModel{
		HasCategories: true,
		Categories: []Category{
			{Name: "", Targets: []Target{{Name: "test"}}},
		},
	}

	ApplyDefaultCategory(model, "")

	// No change - empty default category
	assert.Len(t, model.Categories, 1)
	assert.Equal(t, "", model.Categories[0].Name)
}

func TestApplyDefaultCategory_CreateNewCategory(t *testing.T) {
	model := &HelpModel{
		HasCategories: true,
		Categories: []Category{
			{Name: "Build", Targets: []Target{{Name: "build"}}},
			{Name: "", Targets: []Target{{Name: "test"}}, DiscoveryOrder: 1},
		},
	}

	ApplyDefaultCategory(model, "Other")

	// Should create "Other" category and remove empty category
	assert.Len(t, model.Categories, 2)

	var other *Category
	for i := range model.Categories {
		if model.Categories[i].Name == "Other" {
			other = &model.Categories[i]
		}
	}
	require.NotNil(t, other)
	assert.Len(t, other.Targets, 1)
	assert.Equal(t, "test", other.Targets[0].Name)
}

func TestApplyDefaultCategory_MergeIntoExisting(t *testing.T) {
	model := &HelpModel{
		HasCategories: true,
		Categories: []Category{
			{Name: "Build", Targets: []Target{{Name: "build"}}},
			{Name: "Other", Targets: []Target{{Name: "lint"}}},
			{Name: "", Targets: []Target{{Name: "test"}}},
		},
	}

	ApplyDefaultCategory(model, "Other")

	// Should merge into existing "Other" category
	assert.Len(t, model.Categories, 2)

	var other *Category
	for i := range model.Categories {
		if model.Categories[i].Name == "Other" {
			other = &model.Categories[i]
		}
	}
	require.NotNil(t, other)
	assert.Len(t, other.Targets, 2)
}

func TestCountTargetsByCategory(t *testing.T) {
	model := &HelpModel{
		Categories: []Category{
			{Name: "Build", Targets: []Target{{Name: "build"}, {Name: "compile"}}},
			{Name: "Test", Targets: []Target{{Name: "test"}}},
			{Name: "", Targets: []Target{{Name: "clean"}}},
		},
	}

	counts := CountTargetsByCategory(model)

	assert.Equal(t, 2, counts["Build"])
	assert.Equal(t, 1, counts["Test"])
	assert.Equal(t, 1, counts[""])
}

func TestGetCategoryNames(t *testing.T) {
	model := &HelpModel{
		Categories: []Category{
			{Name: "Build"},
			{Name: "Test"},
			{Name: ""}, // Empty should not be included
		},
	}

	names := GetCategoryNames(model)

	assert.Len(t, names, 2)
	assert.Contains(t, names, "Build")
	assert.Contains(t, names, "Test")
}

func TestHasCategory(t *testing.T) {
	model := &HelpModel{
		Categories: []Category{
			{Name: "Build"},
			{Name: "Test"},
		},
	}

	assert.True(t, HasCategory(model, "Build"))
	assert.True(t, HasCategory(model, "Test"))
	assert.False(t, HasCategory(model, "Deploy"))
}

func TestGetTarget(t *testing.T) {
	model := &HelpModel{
		Categories: []Category{
			{Name: "Build", Targets: []Target{{Name: "build", Summary: "Build summary"}}},
			{Name: "Test", Targets: []Target{{Name: "test", Summary: "Test summary"}}},
		},
	}

	build := GetTarget(model, "build")
	require.NotNil(t, build)
	assert.Equal(t, "Build summary", build.Summary)

	test := GetTarget(model, "test")
	require.NotNil(t, test)
	assert.Equal(t, "Test summary", test.Summary)

	missing := GetTarget(model, "missing")
	assert.Nil(t, missing)
}

func TestGetTargetCount(t *testing.T) {
	model := &HelpModel{
		Categories: []Category{
			{Name: "Build", Targets: []Target{{Name: "build"}, {Name: "compile"}}},
			{Name: "Test", Targets: []Target{{Name: "test"}}},
		},
	}

	assert.Equal(t, 3, GetTargetCount(model))
}

func TestGetTargetCount_EmptyModel(t *testing.T) {
	model := &HelpModel{
		Categories: []Category{},
	}

	assert.Equal(t, 0, GetTargetCount(model))
}
