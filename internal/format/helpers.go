package format

import (
	"fmt"

	"github.com/sdlcforge/make-help/internal/model"
)

// extractEntryPointDocs returns the documentation from the entry point file.
// Returns nil if no entry point documentation exists.
func extractEntryPointDocs(fileDocs []model.FileDoc) []string {
	for _, fileDoc := range fileDocs {
		if fileDoc.IsEntryPoint && len(fileDoc.Documentation) > 0 {
			return fileDoc.Documentation
		}
	}
	return nil
}

// extractIncludedFiles returns all non-entry-point files with documentation.
func extractIncludedFiles(fileDocs []model.FileDoc) []model.FileDoc {
	var includedFiles []model.FileDoc
	for _, fileDoc := range fileDocs {
		if !fileDoc.IsEntryPoint && len(fileDoc.Documentation) > 0 {
			includedFiles = append(includedFiles, fileDoc)
		}
	}
	return includedFiles
}

// initColorScheme creates a ColorScheme from config, using provided scheme or creating default.
func initColorScheme(config *FormatterConfig) *ColorScheme {
	colors := config.ColorScheme
	if colors == nil {
		colors = NewColorScheme(config.UseColor)
	}
	return colors
}

// errNilHelpModel returns an error for a nil help model.
func errNilHelpModel(formatterName string) error {
	return fmt.Errorf("%s formatter: help model cannot be nil", formatterName)
}

// errNilTarget returns an error for a nil target.
func errNilTarget(formatterName string) error {
	return fmt.Errorf("%s formatter: target cannot be nil", formatterName)
}
