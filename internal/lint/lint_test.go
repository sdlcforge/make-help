package lint

import (
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

func TestCheckSummaryPunctuation_AllValid(t *testing.T) {
	ctx := &CheckContext{
		HelpModel: &model.HelpModel{
			Categories: []model.Category{
				{
					Name: "Build",
					Targets: []model.Target{
						{
							Name:       "build",
							Summary:    "Build the project.",
							SourceFile: "Makefile",
							LineNumber: 10,
						},
						{
							Name:       "test",
							Summary:    "Run tests!",
							SourceFile: "Makefile",
							LineNumber: 15,
						},
						{
							Name:       "check",
							Summary:    "Check everything?",
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
							Name:       "build",
							Summary:    "Build the project",
							SourceFile: "Makefile",
							LineNumber: 10,
						},
						{
							Name:       "test",
							Summary:    "Run tests",
							SourceFile: "Makefile",
							LineNumber: 15,
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
							Name:       "period",
							Summary:    "Ends with period.",
							SourceFile: "Makefile",
							LineNumber: 10,
						},
						{
							Name:       "exclamation",
							Summary:    "Ends with exclamation!",
							SourceFile: "Makefile",
							LineNumber: 15,
						},
						{
							Name:       "question",
							Summary:    "Ends with question?",
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
							Name:       "build",
							Summary:    "Build the project",
							SourceFile: "Makefile",
							LineNumber: 10,
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

	checks := []CheckFunc{
		CheckUndocumentedPhony,
		CheckSummaryPunctuation,
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
