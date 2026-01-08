package target

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/sdlcforge/make-help/internal/format"
	"github.com/sdlcforge/make-help/internal/model"
)

// GeneratorConfig holds configuration for help file generation.
type GeneratorConfig struct {
	// Options for rendering
	KeepOrderCategories bool
	KeepOrderTargets    bool
	CategoryOrder       []string
	DefaultCategory     string
	IncludeTargets      []string
	IncludeAllPhony     bool

	// UseColor controls whether ANSI color codes are embedded in the output
	UseColor bool

	// HelpCategory is the category name for generated help targets (help, update-help).
	// Defaults to "Help" if empty.
	HelpCategory string

	// Makefiles is the list of discovered Makefiles for dependency tracking
	Makefiles []string

	// HelpModel is the built model to render
	HelpModel *model.HelpModel

	// MakefileDir is the directory containing the main Makefile (for relative paths)
	MakefileDir string

	// HelpFilename is the basename of the help file (e.g., "help.mk", "00-help.mk")
	HelpFilename string

	// CommandLine is the full command line used to generate this file (for restoration)
	CommandLine string
}

// GenerateHelpFile creates the complete help Makefile content with static help text.
// The generated file includes:
//   - Static help content embedded in @echo statements
//   - Individual help-<target> targets with detailed information
//   - Auto-regeneration target that rebuilds when source Makefiles change
func GenerateHelpFile(config *GeneratorConfig) (string, error) {
	var buf strings.Builder

	// Create formatter with color configuration
	// We use the LineRenderer interface to decouple from the concrete MakeFormatter type
	var renderer format.LineRenderer = format.NewMakeFormatter(&format.FormatterConfig{
		UseColor:    config.UseColor,
		MakefileDir: config.MakefileDir,
	})

	// Header with new format
	buf.WriteString("# generated-by: make-help\n")
	commandLine := config.CommandLine
	if commandLine == "" {
		commandLine = "make-help" + buildRegenerateFlags(config)
	}
	buf.WriteString(fmt.Sprintf("# command: %s\n", commandLine))
	buf.WriteString(fmt.Sprintf("# date: %s\n", time.Now().UTC().Format("2006-01-02T15:04:05 UTC")))
	buf.WriteString("# ---\n")
	buf.WriteString("# DO NOT EDIT\n")
	buf.WriteString("\n")

	// Variables
	buf.WriteString("MAKE_HELP_DIR := $(dir $(lastword $(MAKEFILE_LIST)))\n")

	// Makefile dependencies
	relativeMakefiles := relativizeMakefilePaths(config.Makefiles, config.MakefileDir)
	if len(relativeMakefiles) > 0 {
		buf.WriteString(fmt.Sprintf("MAKE_HELP_MAKEFILES := %s\n", strings.Join(relativeMakefiles, " ")))
	}
	buf.WriteString("\n")

	// Main help target with static content
	// If source Makefiles use categories, add category directive for consistency
	if config.HelpModel.HasCategories {
		helpCategory := config.HelpCategory
		if helpCategory == "" {
			helpCategory = "Help"
		}
		buf.WriteString(fmt.Sprintf("## !category %s\n", helpCategory))
	}
	buf.WriteString(".PHONY: help\n")
	buf.WriteString("## Displays help for available targets.\n")
	buf.WriteString("help:\n")

	// Add timestamp check to warn if help.mk may be stale
	helpFilename := config.HelpFilename
	if helpFilename == "" {
		helpFilename = "help.mk"
	}
	buf.WriteString("\t@for f in $(MAKE_HELP_MAKEFILES); do \\\n")
	buf.WriteString(fmt.Sprintf("\t  if [ \"$$f\" -nt \"$(MAKE_HELP_DIR)%s\" ]; then \\\n", helpFilename))
	if config.UseColor {
		buf.WriteString(fmt.Sprintf("\t    printf '\\033[0;33mWarning: %%s is newer than %s. Run make update-help to refresh.\\033[0m\\n' \"$$f\"; \\\n", helpFilename))
	} else {
		buf.WriteString(fmt.Sprintf("\t    printf 'Warning: %%s is newer than %s. Run make update-help to refresh.\\n' \"$$f\"; \\\n", helpFilename))
	}
	buf.WriteString("\t  fi; \\\n")
	buf.WriteString("\tdone\n")

	// Render help content
	helpLines, err := renderer.RenderHelpLines(config.HelpModel)
	if err != nil {
		return "", fmt.Errorf("failed to render help content: %w", err)
	}

	for _, line := range helpLines {
		buf.WriteString(fmt.Sprintf("\t@printf '%%b\\n' \"%s\"\n", line))
	}

	// Generate help-<target> targets for each documented target
	for _, category := range config.HelpModel.Categories {
		for _, target := range category.Targets {
			buf.WriteString("\n")
			buf.WriteString(fmt.Sprintf(".PHONY: help-%s\n", target.Name))
			buf.WriteString(fmt.Sprintf("help-%s:\n", target.Name))

			detailedLines := renderer.RenderDetailedTargetLines(&target)
			for _, line := range detailedLines {
				buf.WriteString(fmt.Sprintf("\t@printf '%%b\\n' \"%s\"\n", line))
			}
		}
	}

	// Auto-regeneration target
	buf.WriteString("\n")
	buf.WriteString(generateRegenerationTarget(config))

	return buf.String(), nil
}

// buildRegenerateFlags builds the flag string for the regeneration comment.
// This shows users how to regenerate the help file.
func buildRegenerateFlags(config *GeneratorConfig) string {
	var flags []string

	// Add --no-color if colors are disabled
	if !config.UseColor {
		flags = append(flags, "--no-color")
	}

	// Add ordering flags
	if config.KeepOrderCategories {
		flags = append(flags, "--keep-order-categories")
	}
	if config.KeepOrderTargets {
		flags = append(flags, "--keep-order-targets")
	}

	// Add category order
	if len(config.CategoryOrder) > 0 {
		flags = append(flags, fmt.Sprintf("--category-order %s", strings.Join(config.CategoryOrder, ",")))
	}

	// Add default category
	if config.DefaultCategory != "" {
		flags = append(flags, fmt.Sprintf("--default-category %s", config.DefaultCategory))
	}

	// Add include targets
	if len(config.IncludeTargets) > 0 {
		for _, target := range config.IncludeTargets {
			flags = append(flags, fmt.Sprintf("--include-target %s", target))
		}
	}

	// Add include all phony
	if config.IncludeAllPhony {
		flags = append(flags, "--include-all-phony")
	}

	// Add help category if not default
	if config.HelpCategory != "" && config.HelpCategory != "Help" {
		flags = append(flags, fmt.Sprintf("--help-category %s", config.HelpCategory))
	}

	if len(flags) == 0 {
		return ""
	}
	return " " + strings.Join(flags, " ")
}

// generateRegenerationTarget creates the update-help target.
// This is an explicit target users can run to regenerate help.mk.
func generateRegenerationTarget(config *GeneratorConfig) string {
	var buf strings.Builder

	// Build flags to pass to regeneration command (same flags used for original generation)
	flags := buildRegenerateFlags(config)

	// If source Makefiles use categories, add category directive for consistency
	// (uses same category as the help target)
	if config.HelpModel.HasCategories {
		helpCategory := config.HelpCategory
		if helpCategory == "" {
			helpCategory = "Help"
		}
		buf.WriteString(fmt.Sprintf("## !category %s\n", helpCategory))
	}
	buf.WriteString(".PHONY: update-help\n")
	buf.WriteString("## Regenerates help.mk from source Makefiles.\n")
	buf.WriteString("update-help:\n")
	buf.WriteString(fmt.Sprintf("\t@make-help --makefile-path $(MAKE_HELP_DIR)Makefile%s || \\\n", flags))
	buf.WriteString(fmt.Sprintf("\t npx make-help --makefile-path $(MAKE_HELP_DIR)Makefile%s || \\\n", flags))
	buf.WriteString("\t echo \"make-help not found; install with 'go install github.com/sdlcforge/make-help/cmd/make-help@latest' or 'npm install -g make-help'\"\n")

	return buf.String()
}

// relativizeMakefilePaths converts absolute Makefile paths to relative paths using $(MAKE_HELP_DIR).
// This ensures the generated help.mk works regardless of where it's included from.
func relativizeMakefilePaths(makefiles []string, makefileDir string) []string {
	var relative []string

	for _, mf := range makefiles {
		// Clean both paths for proper comparison
		cleanMF := filepath.Clean(mf)
		cleanDir := filepath.Clean(makefileDir)

		// Try to make relative to makefileDir
		relPath, err := filepath.Rel(cleanDir, cleanMF)
		if err != nil {
			// If we can't make it relative, use the absolute path
			// This shouldn't happen in normal usage
			relative = append(relative, cleanMF)
			continue
		}

		// Convert to $(MAKE_HELP_DIR) relative path
		// Use forward slashes for Makefile compatibility
		relPath = filepath.ToSlash(relPath)
		relative = append(relative, "$(MAKE_HELP_DIR)"+relPath)
	}

	return relative
}

// generateHelpTarget is a DEPRECATED compatibility wrapper for add.go.
// It will be removed in Phase 4/5 when the orchestration layer is updated.
// This function cannot generate proper static help without a HelpModel,
// so it returns a minimal placeholder that directs users to use the new workflow.
func generateHelpTarget(config *Config) string {
	var buf strings.Builder

	buf.WriteString("# Generated by make-help. DO NOT EDIT.\n")
	buf.WriteString("# This is a placeholder - regenerate with updated make-help for static content.\n")
	buf.WriteString("\n")
	buf.WriteString("MAKE_HELP_DIR := $(dir $(lastword $(MAKEFILE_LIST)))\n")
	buf.WriteString("\n")
	buf.WriteString(".PHONY: help\n")
	buf.WriteString("## Displays help for available targets.\n")
	buf.WriteString("help:\n")
	buf.WriteString("\t@echo \"Help generation is being upgraded. Please run 'make-help' to regenerate.\"\n")

	return buf.String()
}
