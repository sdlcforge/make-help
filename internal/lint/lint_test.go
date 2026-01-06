package lint

import (
	"strings"
	"testing"

	"github.com/sdlcforge/make-help/internal/model"
)

func TestCheckUndocumentedPhony_NoWarnings(t *testing.T) {
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{},
		PhonyTargets: map[string]bool{
			"build": true,
			"test":  true,
		},
		DocumentedTargets: map[string]bool{
			"build": true,
			"test":  true,
		},
		Aliases: map[string]bool{},
	}

	warnings := CheckUndocumentedPhony(ctx)
	if len(warnings) != 0 {
		t.Errorf("Expected no warnings, got %d", len(warnings))
	}
}

func TestCheckUndocumentedPhony_WithUndocumented(t *testing.T) {
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{},
		PhonyTargets: map[string]bool{
			"build": true,
			"test":  true,
			"setup": true,
			"check": true,
		},
		DocumentedTargets: map[string]bool{
			"build": true,
			"test":  true,
		},
		Aliases: map[string]bool{},
	}

	warnings := CheckUndocumentedPhony(ctx)
	if len(warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(warnings))
	}

	// Check that the right targets are flagged
	flaggedTargets := make(map[string]bool)
	for _, w := range warnings {
		if w.Severity != SeverityWarning {
			t.Errorf("Expected severity 'warning', got '%s'", w.Severity)
		}
		// Extract target name from message
		if len(w.Message) > 0 {
			// Message format: "undocumented phony target 'target'"
			// Simple check that setup and check are mentioned
			if w.Message == "undocumented phony target 'setup'" {
				flaggedTargets["setup"] = true
			}
			if w.Message == "undocumented phony target 'check'" {
				flaggedTargets["check"] = true
			}
		}
	}

	if !flaggedTargets["setup"] || !flaggedTargets["check"] {
		t.Error("Expected 'setup' and 'check' to be flagged as undocumented")
	}
}

func TestCheckUndocumentedPhony_ImplicitAliases(t *testing.T) {
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{},
		PhonyTargets: map[string]bool{
			"build": true,
			"b":     true, // implicit alias
		},
		DocumentedTargets: map[string]bool{
			"build": true,
		},
		Aliases: map[string]bool{
			"b": true, // b is an alias, should not trigger warning
		},
	}

	warnings := CheckUndocumentedPhony(ctx)
	if len(warnings) != 0 {
		t.Errorf("Expected no warnings (b is an alias), got %d", len(warnings))
	}
}

func TestCheckUndocumentedPhony_GeneratedHelpTargets(t *testing.T) {
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{},
		PhonyTargets: map[string]bool{
			"build":      true,
			"help":       true, // generated
			"update-help": true, // generated
			"help-build": true, // generated for documented target
			"help-random": true, // NOT generated (user-created)
		},
		DocumentedTargets: map[string]bool{
			"build": true,
		},
		Aliases: map[string]bool{},
		GeneratedHelpTargets: map[string]bool{
			"help":       true,
			"update-help": true,
			"help-build": true,
			// help-random is NOT in this list
		},
	}

	warnings := CheckUndocumentedPhony(ctx)
	// Should only warn about help-random (user-created, not generated)
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning (for help-random), got %d", len(warnings))
	}

	if len(warnings) > 0 && warnings[0].Message != "undocumented phony target 'help-random'" {
		t.Errorf("Expected warning for 'help-random', got: %s", warnings[0].Message)
	}
}

func TestCheckSummaryPunctuation_AllValid(t *testing.T) {
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{
			Categories: []model.Category{
				{
					Name: "Build",
					Targets: []model.Target{
						{
							Name:       "build",
							Summary:    []string{"Build the project."},
							SourceFile: "Makefile",
							LineNumber: 10,
						},
						{
							Name:       "test",
							Summary:    []string{"Run tests!"},
							SourceFile: "Makefile",
							LineNumber: 15,
						},
						{
							Name:       "check",
							Summary:    []string{"Check everything?"},
							SourceFile: "Makefile",
							LineNumber: 20,
						},
					},
				},
			},
		},
	}

	warnings := CheckSummaryPunctuation(ctx)
	if len(warnings) != 0 {
		t.Errorf("Expected no warnings, got %d", len(warnings))
	}
}

func TestCheckSummaryPunctuation_MissingPunctuation(t *testing.T) {
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{
			Categories: []model.Category{
				{
					Name: "Build",
					Targets: []model.Target{
						{
							Name:          "build",
							Documentation: []string{"Build the project"},
							SourceFile:    "Makefile",
							LineNumber:    10,
						},
						{
							Name:          "test",
							Documentation: []string{"Run tests"},
							SourceFile:    "Makefile",
							LineNumber:    15,
						},
					},
				},
			},
		},
	}

	warnings := CheckSummaryPunctuation(ctx)
	if len(warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(warnings))
	}

	for _, w := range warnings {
		if w.Severity != SeverityWarning {
			t.Errorf("Expected severity 'warning', got '%s'", w.Severity)
		}
		if w.File != "Makefile" {
			t.Errorf("Expected File 'Makefile', got '%s'", w.File)
		}
	}
}

func TestCheckSummaryPunctuation_AllPunctuationTypes(t *testing.T) {
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{
			Categories: []model.Category{
				{
					Name: "Build",
					Targets: []model.Target{
						{
							Name:          "period",
							Documentation: []string{"Ends with period."},
							SourceFile:    "Makefile",
							LineNumber:    10,
						},
						{
							Name:          "exclamation",
							Documentation: []string{"Ends with exclamation!"},
							SourceFile:    "Makefile",
							LineNumber:    15,
						},
						{
							Name:          "question",
							Documentation: []string{"Ends with question?"},
							SourceFile:    "Makefile",
							LineNumber:    20,
						},
					},
				},
			},
		},
	}

	warnings := CheckSummaryPunctuation(ctx)
	if len(warnings) != 0 {
		t.Errorf("Expected no warnings for all punctuation types, got %d", len(warnings))
	}
}

func TestLint_MultipleChecks(t *testing.T) {
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{
			Categories: []model.Category{
				{
					Name: "Build",
					Targets: []model.Target{
						{
							Name:          "build",
							Documentation: []string{"Build the project"},
							SourceFile:    "Makefile",
							LineNumber:    10,
						},
					},
				},
			},
		},
		PhonyTargets: map[string]bool{
			"build": true,
			"setup": true,
		},
		DocumentedTargets: map[string]bool{
			"build": true,
		},
		Aliases: map[string]bool{},
	}

	checks := []Check{
		{Name: "undocumented-phony", CheckFunc: CheckUndocumentedPhony},
		{Name: "summary-punctuation", CheckFunc: CheckSummaryPunctuation},
	}

	result := Lint(ctx, checks)

	if !result.HasWarnings {
		t.Error("Expected warnings to be found")
	}

	// Should have 2 warnings: 1 undocumented phony, 1 missing punctuation
	if len(result.Warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(result.Warnings))
	}
}

func TestFormatWarning_WithLine(t *testing.T) {
	w := Warning{
		File:     "Makefile",
		Line:     42,
		Severity: SeverityWarning,
		Message:  "test warning",
		Context:  "some context",
	}

	formatted := FormatWarning(w)
	expected := "Makefile:42: warning: test warning\n  | some context"
	if formatted != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, formatted)
	}
}

func TestFormatWarning_WithoutLine(t *testing.T) {
	w := Warning{
		File:     "Makefile",
		Line:     0,
		Severity: SeverityWarning,
		Message:  "test warning",
	}

	formatted := FormatWarning(w)
	expected := "Makefile: warning: test warning"
	if formatted != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, formatted)
	}
}

// Tests for CheckOrphanAliases

func TestCheckOrphanAliases_NoWarnings(t *testing.T) {
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{
			Categories: []model.Category{
				{
					Name: "Build",
					Targets: []model.Target{
						{
							Name:       "build",
							Aliases:    []string{"b", "compile"},
							SourceFile: "Makefile",
							LineNumber: 10,
						},
					},
				},
			},
		},
		DocumentedTargets: map[string]bool{
			"build": true,
		},
		PhonyTargets: map[string]bool{
			"build":   true,
			"b":       true,
			"compile": true,
		},
		HasRecipe: map[string]bool{},
		Aliases: map[string]bool{
			"b":       true,
			"compile": true,
		},
	}

	warnings := CheckOrphanAliases(ctx)
	if len(warnings) != 0 {
		t.Errorf("Expected no warnings (all aliases exist), got %d", len(warnings))
	}
}

func TestCheckOrphanAliases_WithOrphans(t *testing.T) {
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{
			Categories: []model.Category{
				{
					Name: "Build",
					Targets: []model.Target{
						{
							Name:       "build",
							Aliases:    []string{"b", "nonexistent"},
							SourceFile: "Makefile",
							LineNumber: 10,
						},
						{
							Name:       "test",
							Aliases:    []string{"t", "missing-target"},
							SourceFile: "Makefile",
							LineNumber: 20,
						},
					},
				},
			},
		},
		DocumentedTargets: map[string]bool{
			"build": true,
			"test":  true,
		},
		PhonyTargets: map[string]bool{
			"build": true,
			"test":  true,
			"b":     true,
			"t":     true,
		},
		HasRecipe: map[string]bool{},
		Aliases: map[string]bool{
			"b": true,
			"t": true,
		},
	}

	warnings := CheckOrphanAliases(ctx)
	if len(warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(warnings))
	}

	// Check that both orphaned aliases are flagged
	flaggedAliases := make(map[string]bool)
	for _, w := range warnings {
		if w.Severity != SeverityWarning {
			t.Errorf("Expected severity 'warning', got '%s'", w.Severity)
		}
		if w.File != "Makefile" {
			t.Errorf("Expected File 'Makefile', got '%s'", w.File)
		}
		// Extract alias name from message to verify correct aliases are flagged
		// Message format: "alias 'X' points to non-existent target (referenced by 'Y')"
		if strings.Contains(w.Message, "'nonexistent'") {
			flaggedAliases["nonexistent"] = true
		}
		if strings.Contains(w.Message, "'missing-target'") {
			flaggedAliases["missing-target"] = true
		}
	}

	if !flaggedAliases["nonexistent"] || !flaggedAliases["missing-target"] {
		t.Error("Expected 'nonexistent' and 'missing-target' to be flagged as orphaned aliases")
	}
}

// Tests for CheckLongSummaries

func TestCheckLongSummaries_NoWarnings(t *testing.T) {
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{
			Categories: []model.Category{
				{
					Name: "Build",
					Targets: []model.Target{
						{
							Name:       "build",
							Summary:    []string{"Build the project."},
							SourceFile: "Makefile",
							LineNumber: 10,
						},
						{
							Name:       "test",
							Summary:    []string{"Run all tests."},
							SourceFile: "Makefile",
							LineNumber: 15,
						},
						{
							Name:       "exactly-80",
							Summary:    []string{"1234567890123456789012345678901234567890123456789012345678901234567890123456789."},
							SourceFile: "Makefile",
							LineNumber: 20,
						},
					},
				},
			},
		},
	}

	warnings := CheckLongSummaries(ctx)
	if len(warnings) != 0 {
		t.Errorf("Expected no warnings, got %d", len(warnings))
	}
}

func TestCheckLongSummaries_WithLong(t *testing.T) {
	longSummary := []string{"This is a very long summary that exceeds the eighty character limit and should trigger a warning."}
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{
			Categories: []model.Category{
				{
					Name: "Build",
					Targets: []model.Target{
						{
							Name:       "build",
							Summary:    []string{"Build the project."},
							SourceFile: "Makefile",
							LineNumber: 10,
						},
						{
							Name:       "complex-task",
							Summary:    longSummary,
							SourceFile: "Makefile",
							LineNumber: 15,
						},
					},
				},
			},
		},
	}

	warnings := CheckLongSummaries(ctx)
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warnings))
	}

	if len(warnings) > 0 {
		w := warnings[0]
		if w.Severity != SeverityWarning {
			t.Errorf("Expected severity 'warning', got '%s'", w.Severity)
		}
		if w.File != "Makefile" {
			t.Errorf("Expected File 'Makefile', got '%s'", w.File)
		}
		if w.Line != 15 {
			t.Errorf("Expected Line 15, got %d", w.Line)
		}
		// Check that the message includes the actual length
		expectedMsg := "summary for 'complex-task' is too long (97 characters, max 80)"
		if w.Message != expectedMsg {
			t.Errorf("Expected message '%s', got '%s'", expectedMsg, w.Message)
		}
		if w.Context != longSummary[0] {
			t.Errorf("Expected context to be the long summary")
		}
	}
}

func TestCheckLongSummaries_BoundaryCase(t *testing.T) {
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{
			Categories: []model.Category{
				{
					Name: "Build",
					Targets: []model.Target{
						{
							Name:       "exactly-80",
							Summary:    []string{"1234567890123456789012345678901234567890123456789012345678901234567890123456789."},
							SourceFile: "Makefile",
							LineNumber: 10,
						},
						{
							Name:       "exactly-81",
							Summary:    []string{"12345678901234567890123456789012345678901234567890123456789012345678901234567890."},
							SourceFile: "Makefile",
							LineNumber: 15,
						},
					},
				},
			},
		},
	}

	warnings := CheckLongSummaries(ctx)
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning (for 81-char summary), got %d", len(warnings))
	}

	if len(warnings) > 0 {
		w := warnings[0]
		if w.Line != 15 {
			t.Errorf("Expected warning for line 15 (81-char summary), got line %d", w.Line)
		}
		expectedMsg := "summary for 'exactly-81' is too long (81 characters, max 80)"
		if w.Message != expectedMsg {
			t.Errorf("Expected message '%s', got '%s'", expectedMsg, w.Message)
		}
	}
}

// Tests for CheckEmptyDocumentation

func TestCheckEmptyDocumentation_NoWarnings(t *testing.T) {
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{
			Categories: []model.Category{
				{
					Name: "Build",
					Targets: []model.Target{
						{
							Name:          "build",
							Documentation: []string{"Build the project."},
							SourceFile:    "Makefile",
							LineNumber:    10,
						},
						{
							Name:          "test",
							Documentation: []string{"First line.", "", "Second line."},
							SourceFile:    "Makefile",
							LineNumber:    15,
						},
					},
				},
			},
		},
	}

	warnings := CheckEmptyDocumentation(ctx)
	if len(warnings) != 0 {
		t.Errorf("Expected no warnings, got %d", len(warnings))
	}
}

func TestCheckEmptyDocumentation_LeadingEmpty(t *testing.T) {
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{
			Categories: []model.Category{
				{
					Name: "Build",
					Targets: []model.Target{
						{
							Name:          "build",
							Documentation: []string{"", "Build the project."},
							SourceFile:    "Makefile",
							LineNumber:    10,
						},
					},
				},
			},
		},
	}

	warnings := CheckEmptyDocumentation(ctx)
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warnings))
	}

	if len(warnings) > 0 {
		if warnings[0].Severity != SeverityWarning {
			t.Errorf("Expected severity 'warning', got '%s'", warnings[0].Severity)
		}
		if warnings[0].Message != "target 'build' has empty documentation line at the beginning" {
			t.Errorf("Unexpected message: %s", warnings[0].Message)
		}
	}
}

func TestCheckEmptyDocumentation_TrailingEmpty(t *testing.T) {
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{
			Categories: []model.Category{
				{
					Name: "Build",
					Targets: []model.Target{
						{
							Name:          "build",
							Documentation: []string{"Build the project.", ""},
							SourceFile:    "Makefile",
							LineNumber:    10,
						},
					},
				},
			},
		},
	}

	warnings := CheckEmptyDocumentation(ctx)
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warnings))
	}

	if len(warnings) > 0 {
		if warnings[0].Severity != SeverityWarning {
			t.Errorf("Expected severity 'warning', got '%s'", warnings[0].Severity)
		}
		if warnings[0].Message != "target 'build' has empty documentation line at the end" {
			t.Errorf("Unexpected message: %s", warnings[0].Message)
		}
	}
}

func TestCheckEmptyDocumentation_InternalEmpty(t *testing.T) {
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{
			Categories: []model.Category{
				{
					Name: "Build",
					Targets: []model.Target{
						{
							Name:          "build",
							Documentation: []string{"First paragraph.", "", "Second paragraph."},
							SourceFile:    "Makefile",
							LineNumber:    10,
						},
					},
				},
			},
		},
	}

	warnings := CheckEmptyDocumentation(ctx)
	if len(warnings) != 0 {
		t.Errorf("Expected no warnings (internal empty lines are OK), got %d", len(warnings))
		for _, w := range warnings {
			t.Errorf("Unexpected warning: %s", w.Message)
		}
	}
}

// Tests for CheckMissingVarDescriptions

func TestCheckMissingVarDescriptions_NoWarnings(t *testing.T) {
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{
			Categories: []model.Category{
				{
					Name: "Build",
					Targets: []model.Target{
						{
							Name:       "build",
							Summary:    []string{"Build the project."},
							SourceFile: "Makefile",
							LineNumber: 10,
							Variables: []model.Variable{
								{
									Name:        "CC",
									Description: "C compiler to use (default: gcc)",
								},
								{
									Name:        "CFLAGS",
									Description: "C compiler flags",
								},
							},
						},
					},
				},
			},
		},
	}

	warnings := CheckMissingVarDescriptions(ctx)
	if len(warnings) != 0 {
		t.Errorf("Expected no warnings, got %d", len(warnings))
	}
}

func TestCheckMissingVarDescriptions_WithMissing(t *testing.T) {
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{
			Categories: []model.Category{
				{
					Name: "Build",
					Targets: []model.Target{
						{
							Name:       "build",
							Summary:    []string{"Build the project."},
							SourceFile: "Makefile",
							LineNumber: 10,
							Variables: []model.Variable{
								{
									Name:        "CC",
									Description: "C compiler to use (default: gcc)",
								},
								{
									Name:        "PORT",
									Description: "", // Missing description
								},
							},
						},
						{
							Name:       "serve",
							Summary:    []string{"Run the server."},
							SourceFile: "Makefile",
							LineNumber: 20,
							Variables: []model.Variable{
								{
									Name:        "HOST",
									Description: "   ", // Whitespace-only description
								},
							},
						},
					},
				},
			},
		},
	}

	warnings := CheckMissingVarDescriptions(ctx)
	if len(warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(warnings))
	}

	// Check that the right variables are flagged
	flaggedVars := make(map[string]bool)
	for _, w := range warnings {
		if w.Severity != SeverityWarning {
			t.Errorf("Expected severity 'warning', got '%s'", w.Severity)
		}
		if w.Message == "variable 'PORT' in target 'build' is missing a description" {
			flaggedVars["PORT"] = true
		}
		if w.Message == "variable 'HOST' in target 'serve' is missing a description" {
			flaggedVars["HOST"] = true
		}
	}

	if !flaggedVars["PORT"] || !flaggedVars["HOST"] {
		t.Error("Expected 'PORT' and 'HOST' to be flagged as missing descriptions")
	}
}

// Tests for CheckInconsistentNaming

func TestCheckInconsistentNaming_NoWarnings(t *testing.T) {
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{
			Categories: []model.Category{
				{
					Name: "Build",
					Targets: []model.Target{
						{
							Name:       "build",
							SourceFile: "Makefile",
							LineNumber: 10,
						},
						{
							Name:       "test",
							SourceFile: "Makefile",
							LineNumber: 15,
						},
						{
							Name:       "build-all",
							SourceFile: "Makefile",
							LineNumber: 20,
						},
						{
							Name:       "run-tests",
							SourceFile: "Makefile",
							LineNumber: 25,
						},
						{
							Name:       "docker-build-image",
							SourceFile: "Makefile",
							LineNumber: 30,
						},
					},
				},
			},
		},
	}

	warnings := CheckInconsistentNaming(ctx)
	if len(warnings) != 0 {
		t.Errorf("Expected no warnings for valid kebab-case names, got %d", len(warnings))
	}
}

func TestCheckInconsistentNaming_WithInvalid(t *testing.T) {
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{
			Categories: []model.Category{
				{
					Name: "Build",
					Targets: []model.Target{
						{
							Name:       "buildAll",
							SourceFile: "Makefile",
							LineNumber: 10,
						},
						{
							Name:       "Build",
							SourceFile: "Makefile",
							LineNumber: 15,
						},
						{
							Name:       "build_all",
							SourceFile: "Makefile",
							LineNumber: 20,
						},
						{
							Name:       "DEPLOY",
							SourceFile: "Makefile",
							LineNumber: 25,
						},
					},
				},
			},
		},
	}

	warnings := CheckInconsistentNaming(ctx)
	if len(warnings) != 4 {
		t.Errorf("Expected 4 warnings, got %d", len(warnings))
	}

	// Verify all invalid targets are flagged
	flaggedTargets := make(map[string]bool)
	for _, w := range warnings {
		if w.Severity != SeverityWarning {
			t.Errorf("Expected severity 'warning', got '%s'", w.Severity)
		}
		// Extract target name from context
		flaggedTargets[w.Context] = true
	}

	expectedInvalid := []string{"buildAll", "Build", "build_all", "DEPLOY"}
	for _, target := range expectedInvalid {
		if !flaggedTargets[target] {
			t.Errorf("Expected '%s' to be flagged as invalid naming", target)
		}
	}
}

// Tests for CheckCircularDependencies

func TestCheckCircularDependencies_NoWarnings(t *testing.T) {
	ctx := &CheckContext{
		HelpModel:    &model.HelpModel{},
		MakefilePath: "Makefile",
		PhonyTargets: map[string]bool{
			"build": true,
			"test":  true,
			"b":     true,
		},
		Dependencies: map[string][]string{
			"b":     {"build"},
			"build": {},
			"test":  {},
		},
		HasRecipe: map[string]bool{
			"build": true,
			"test":  true,
			"b":     false,
		},
	}

	warnings := CheckCircularDependencies(ctx)
	if len(warnings) != 0 {
		t.Errorf("Expected no warnings, got %d", len(warnings))
	}
}

func TestCheckCircularDependencies_SimpleCycle(t *testing.T) {
	ctx := &CheckContext{
		HelpModel:    &model.HelpModel{},
		MakefilePath: "Makefile",
		PhonyTargets: map[string]bool{
			"a": true,
			"b": true,
		},
		Dependencies: map[string][]string{
			"a": {"b"},
			"b": {"a"},
		},
		HasRecipe: map[string]bool{
			"a": false,
			"b": false,
		},
	}

	warnings := CheckCircularDependencies(ctx)
	if len(warnings) != 1 {
		t.Fatalf("Expected 1 warning, got %d", len(warnings))
	}

	w := warnings[0]
	if w.Severity != SeverityWarning {
		t.Errorf("Expected severity 'warning', got '%s'", w.Severity)
	}
	if w.File != "Makefile" {
		t.Errorf("Expected File 'Makefile', got '%s'", w.File)
	}

	// Check that the warning mentions both targets in a cycle
	if !strings.Contains(w.Message, "a") || !strings.Contains(w.Message, "b") {
		t.Errorf("Expected warning to mention both 'a' and 'b', got: %s", w.Message)
	}
	if !strings.Contains(w.Message, "circular dependency chain") {
		t.Errorf("Expected warning to mention 'circular dependency chain', got: %s", w.Message)
	}
}

func TestCheckCircularDependencies_LongerCycle(t *testing.T) {
	ctx := &CheckContext{
		HelpModel:    &model.HelpModel{},
		MakefilePath: "Makefile",
		PhonyTargets: map[string]bool{
			"a": true,
			"b": true,
			"c": true,
		},
		Dependencies: map[string][]string{
			"a": {"b"},
			"b": {"c"},
			"c": {"a"},
		},
		HasRecipe: map[string]bool{
			"a": false,
			"b": false,
			"c": false,
		},
	}

	warnings := CheckCircularDependencies(ctx)
	if len(warnings) != 1 {
		t.Fatalf("Expected 1 warning, got %d", len(warnings))
	}

	w := warnings[0]
	if w.Severity != SeverityWarning {
		t.Errorf("Expected severity 'warning', got '%s'", w.Severity)
	}

	// Check that the warning mentions all three targets in a cycle
	if !strings.Contains(w.Message, "a") || !strings.Contains(w.Message, "b") || !strings.Contains(w.Message, "c") {
		t.Errorf("Expected warning to mention 'a', 'b', and 'c', got: %s", w.Message)
	}
	if !strings.Contains(w.Message, "circular dependency chain") {
		t.Errorf("Expected warning to mention 'circular dependency chain', got: %s", w.Message)
	}
}

// Tests for CheckRedundantDirectives

func TestCheckRedundantDirectives_NoWarnings(t *testing.T) {
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{
			Categories: []model.Category{
				{
					Name: "Build",
					Targets: []model.Target{
						{
							Name:       "build",
							Aliases:    []string{"b"},
							SourceFile: "Makefile",
							LineNumber: 10,
						},
					},
				},
			},
		},
		NotAliasTargets: map[string]bool{}, // No !notalias directives
		TargetLocations: map[string]TargetLocation{
			"build": {File: "Makefile", Line: 10},
		},
		DocumentedTargets: map[string]bool{
			"build": true,
		},
		PhonyTargets: map[string]bool{
			"build": true,
		},
		HasRecipe: map[string]bool{
			"build": true,
		},
		Dependencies: map[string][]string{
			"build": {},
		},
	}

	warnings := CheckRedundantDirectives(ctx)
	if len(warnings) != 0 {
		t.Errorf("Expected no warnings, got %d", len(warnings))
	}
}

func TestCheckRedundantDirectives_NotAliasOnDocumented(t *testing.T) {
	// !notalias on a documented target is redundant
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{
			Categories: []model.Category{
				{
					Name: "Build",
					Targets: []model.Target{
						{
							Name:       "build",
							SourceFile: "Makefile",
							LineNumber: 10,
						},
					},
				},
			},
		},
		NotAliasTargets: map[string]bool{
			"build": true, // Has !notalias
		},
		TargetLocations: map[string]TargetLocation{
			"build": {File: "Makefile", Line: 10},
		},
		DocumentedTargets: map[string]bool{
			"build": true, // Is documented
		},
		PhonyTargets: map[string]bool{
			"build": true,
		},
		HasRecipe: map[string]bool{
			"build": true,
		},
		Dependencies: map[string][]string{
			"build": {"test"},
		},
	}

	warnings := CheckRedundantDirectives(ctx)
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warnings))
	}

	if len(warnings) > 0 {
		w := warnings[0]
		if w.CheckName != "redundant-notalias" {
			t.Errorf("Expected check name 'redundant-notalias', got '%s'", w.CheckName)
		}
		if !strings.Contains(w.Message, "documented targets are never implicit aliases") {
			t.Errorf("Expected message about documented targets, got: %s", w.Message)
		}
		if !w.Fixable {
			t.Errorf("Expected warning to be fixable")
		}
	}
}

func TestCheckRedundantDirectives_NotAliasOnHasRecipe(t *testing.T) {
	// !notalias on a target with a recipe is redundant
	ctx := &CheckContext{
		HelpModel:    &model.HelpModel{},
		MakefilePath: "Makefile",
		NotAliasTargets: map[string]bool{
			"build": true, // Has !notalias
		},
		TargetLocations: map[string]TargetLocation{
			"build": {File: "Makefile", Line: 10},
		},
		DocumentedTargets: map[string]bool{}, // Not documented
		PhonyTargets: map[string]bool{
			"build": true,
		},
		HasRecipe: map[string]bool{
			"build": true, // Has recipe
		},
		Dependencies: map[string][]string{
			"build": {"test"},
		},
	}

	warnings := CheckRedundantDirectives(ctx)
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warnings))
	}

	if len(warnings) > 0 {
		w := warnings[0]
		if w.CheckName != "redundant-notalias" {
			t.Errorf("Expected check name 'redundant-notalias', got '%s'", w.CheckName)
		}
		if !strings.Contains(w.Message, "targets with recipes are never implicit aliases") {
			t.Errorf("Expected message about recipes, got: %s", w.Message)
		}
	}
}

func TestCheckRedundantDirectives_NotAliasOnNonPhony(t *testing.T) {
	// !notalias on a non-.PHONY target is redundant
	ctx := &CheckContext{
		HelpModel:    &model.HelpModel{},
		MakefilePath: "Makefile",
		NotAliasTargets: map[string]bool{
			"build": true, // Has !notalias
		},
		TargetLocations: map[string]TargetLocation{
			"build": {File: "Makefile", Line: 10},
		},
		DocumentedTargets: map[string]bool{}, // Not documented
		PhonyTargets:      map[string]bool{}, // Not .PHONY
		HasRecipe: map[string]bool{
			"build": false, // No recipe
		},
		Dependencies: map[string][]string{
			"build": {"test"},
		},
	}

	warnings := CheckRedundantDirectives(ctx)
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warnings))
	}

	if len(warnings) > 0 {
		w := warnings[0]
		if w.CheckName != "redundant-notalias" {
			t.Errorf("Expected check name 'redundant-notalias', got '%s'", w.CheckName)
		}
		if !strings.Contains(w.Message, "non-phony targets are never implicit aliases") {
			t.Errorf("Expected message about non-phony, got: %s", w.Message)
		}
	}
}

func TestCheckRedundantDirectives_NotAliasMultipleDeps(t *testing.T) {
	// !notalias on a target with multiple dependencies is redundant
	ctx := &CheckContext{
		HelpModel:    &model.HelpModel{},
		MakefilePath: "Makefile",
		NotAliasTargets: map[string]bool{
			"all": true, // Has !notalias
		},
		TargetLocations: map[string]TargetLocation{
			"all": {File: "Makefile", Line: 10},
		},
		DocumentedTargets: map[string]bool{}, // Not documented
		PhonyTargets: map[string]bool{
			"all":   true,
			"build": true,
			"test":  true,
		},
		HasRecipe: map[string]bool{
			"all": false, // No recipe
		},
		Dependencies: map[string][]string{
			"all": {"build", "test"}, // Multiple dependencies
		},
	}

	warnings := CheckRedundantDirectives(ctx)
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warnings))
	}

	if len(warnings) > 0 {
		w := warnings[0]
		if w.CheckName != "redundant-notalias" {
			t.Errorf("Expected check name 'redundant-notalias', got '%s'", w.CheckName)
		}
		if !strings.Contains(w.Message, "only targets with exactly one dependency") {
			t.Errorf("Expected message about single dependency, got: %s", w.Message)
		}
	}
}

func TestCheckRedundantDirectives_NotAliasNonPhonyDep(t *testing.T) {
	// !notalias on a target whose dependency is not .PHONY is redundant
	ctx := &CheckContext{
		HelpModel:    &model.HelpModel{},
		MakefilePath: "Makefile",
		NotAliasTargets: map[string]bool{
			"b": true, // Has !notalias
		},
		TargetLocations: map[string]TargetLocation{
			"b": {File: "Makefile", Line: 10},
		},
		DocumentedTargets: map[string]bool{}, // Not documented
		PhonyTargets: map[string]bool{
			"b": true, // b is .PHONY
			// "build" is NOT .PHONY
		},
		HasRecipe: map[string]bool{
			"b": false, // No recipe
		},
		Dependencies: map[string][]string{
			"b": {"build"}, // Single dependency but not .PHONY
		},
	}

	warnings := CheckRedundantDirectives(ctx)
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warnings))
	}

	if len(warnings) > 0 {
		w := warnings[0]
		if w.CheckName != "redundant-notalias" {
			t.Errorf("Expected check name 'redundant-notalias', got '%s'", w.CheckName)
		}
		if !strings.Contains(w.Message, "dependency 'build' is not phony") {
			t.Errorf("Expected message about non-phony dependency, got: %s", w.Message)
		}
	}
}

func TestCheckRedundantDirectives_SelfAlias(t *testing.T) {
	// A target that has itself as an alias
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{
			Categories: []model.Category{
				{
					Name: "Build",
					Targets: []model.Target{
						{
							Name:       "build",
							Aliases:    []string{"build", "b"}, // "build" is a self-reference
							SourceFile: "Makefile",
							LineNumber: 10,
						},
					},
				},
			},
		},
		NotAliasTargets: map[string]bool{},
		TargetLocations: map[string]TargetLocation{
			"build": {File: "Makefile", Line: 10},
		},
		DocumentedTargets: map[string]bool{
			"build": true,
		},
		PhonyTargets: map[string]bool{
			"build": true,
		},
		HasRecipe: map[string]bool{
			"build": true,
		},
		Dependencies: map[string][]string{
			"build": {},
		},
	}

	warnings := CheckRedundantDirectives(ctx)
	if len(warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(warnings))
	}

	if len(warnings) > 0 {
		w := warnings[0]
		if w.CheckName != "redundant-alias" {
			t.Errorf("Expected check name 'redundant-alias', got '%s'", w.CheckName)
		}
		if !strings.Contains(w.Message, "has itself as an alias") {
			t.Errorf("Expected message about self-alias, got: %s", w.Message)
		}
		if !w.Fixable {
			t.Errorf("Expected warning to be fixable")
		}
	}
}

func TestCheckRedundantDirectives_ValidNotAlias(t *testing.T) {
	// !notalias that is actually useful - target would be implicit alias without it
	ctx := &CheckContext{
		HelpModel:    &model.HelpModel{},
		MakefilePath: "Makefile",
		NotAliasTargets: map[string]bool{
			"t": true, // Has !notalias
		},
		TargetLocations: map[string]TargetLocation{
			"t": {File: "Makefile", Line: 10},
		},
		DocumentedTargets: map[string]bool{}, // Not documented
		PhonyTargets: map[string]bool{
			"t":    true, // Is .PHONY
			"test": true, // Dependency is also .PHONY
		},
		HasRecipe: map[string]bool{
			"t": false, // No recipe
		},
		Dependencies: map[string][]string{
			"t": {"test"}, // Single .PHONY dependency
		},
	}

	warnings := CheckRedundantDirectives(ctx)
	// No warnings - the !notalias is useful (without it, t would be an implicit alias)
	if len(warnings) != 0 {
		t.Errorf("Expected no warnings (valid !notalias), got %d", len(warnings))
		for _, w := range warnings {
			t.Errorf("Unexpected warning: %s", w.Message)
		}
	}
}
