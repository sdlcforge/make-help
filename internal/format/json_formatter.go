package format

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/sdlcforge/make-help/internal/model"
)

// JSONFormatter generates JSON output for programmatic consumption.
// The output is valid JSON with 2-space indentation.
type JSONFormatter struct {
	config *FormatterConfig
}

// NewJSONFormatter creates a new JSONFormatter with the given configuration.
func NewJSONFormatter(config *FormatterConfig) *JSONFormatter {
	config = normalizeConfig(config)

	return &JSONFormatter{
		config: config,
	}
}

// jsonHelpOutput represents the complete help output in JSON format.
type jsonHelpOutput struct {
	Usage          string              `json:"usage"`
	Description    string              `json:"description,omitempty"`
	IncludedFiles  []jsonIncludedFile  `json:"includedFiles,omitempty"`
	Categories     []jsonCategory      `json:"categories,omitempty"`
}

// jsonIncludedFile represents a single included file.
type jsonIncludedFile struct {
	Path        string `json:"path"`
	Description string `json:"description,omitempty"`
}

// jsonCategory represents a category with its targets.
type jsonCategory struct {
	Name    string       `json:"name"`
	Targets []jsonTarget `json:"targets"`
}

// jsonTarget represents a target in the help output.
type jsonTarget struct {
	Name       string         `json:"name"`
	Summary    string         `json:"summary,omitempty"`
	Aliases    []string       `json:"aliases,omitempty"`
	Variables  []jsonVariable `json:"variables,omitempty"`
	SourceFile string         `json:"sourceFile,omitempty"`
	LineNumber int            `json:"lineNumber,omitempty"`
}

// jsonVariable represents a documented variable.
type jsonVariable struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// jsonDetailedTarget represents a detailed target view.
type jsonDetailedTarget struct {
	Name          string         `json:"name"`
	Summary       string         `json:"summary,omitempty"`
	Documentation []string       `json:"documentation,omitempty"`
	Aliases       []string       `json:"aliases,omitempty"`
	Variables     []jsonVariable `json:"variables,omitempty"`
	SourceFile    string         `json:"sourceFile,omitempty"`
	LineNumber    int            `json:"lineNumber,omitempty"`
}

// jsonBasicTarget represents a basic target without documentation.
type jsonBasicTarget struct {
	Name       string `json:"name"`
	SourceFile string `json:"sourceFile,omitempty"`
	LineNumber int    `json:"lineNumber,omitempty"`
}

// RenderHelp generates the complete help output from a HelpModel in JSON format.
func (f *JSONFormatter) RenderHelp(helpModel *model.HelpModel, w io.Writer) error {
	if helpModel == nil {
		return fmt.Errorf("json formatter: help model cannot be nil")
	}

	output := jsonHelpOutput{
		Usage: "make [<target>...] [<ENV_VAR>=<value>...]",
	}

	// Extract entry point description and included files
	if len(helpModel.FileDocs) > 0 {
		// Entry point file documentation
		for _, fileDoc := range helpModel.FileDocs {
			if fileDoc.IsEntryPoint && len(fileDoc.Documentation) > 0 {
				// Join all documentation lines with newlines
				output.Description = strings.Join(fileDoc.Documentation, "\n")
				break
			}
		}

		// Included files
		for _, fileDoc := range helpModel.FileDocs {
			if !fileDoc.IsEntryPoint && len(fileDoc.Documentation) > 0 {
				output.IncludedFiles = append(output.IncludedFiles, jsonIncludedFile{
					Path:        fileDoc.SourceFile,
					Description: strings.Join(fileDoc.Documentation, "\n"),
				})
			}
		}
	}

	// Convert categories and targets
	for _, category := range helpModel.Categories {
		jsonCat := jsonCategory{
			Name:    category.Name,
			Targets: make([]jsonTarget, 0, len(category.Targets)),
		}

		for _, target := range category.Targets {
			jsonTgt := jsonTarget{
				Name:       target.Name,
				Summary:    target.Summary.PlainText(), // Use plain text for JSON consumers
				SourceFile: target.SourceFile,
				LineNumber: target.LineNumber,
			}

			// Add aliases if present
			if len(target.Aliases) > 0 {
				jsonTgt.Aliases = target.Aliases
			}

			// Add variables if present
			if len(target.Variables) > 0 {
				jsonTgt.Variables = make([]jsonVariable, len(target.Variables))
				for i, v := range target.Variables {
					jsonTgt.Variables[i] = jsonVariable{
						Name:        v.Name,
						Description: v.Description,
					}
				}
			}

			jsonCat.Targets = append(jsonCat.Targets, jsonTgt)
		}

		output.Categories = append(output.Categories, jsonCat)
	}

	// Marshal to JSON with 2-space indentation
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// RenderDetailedTarget renders a detailed view of a single target in JSON format.
func (f *JSONFormatter) RenderDetailedTarget(target *model.Target, w io.Writer) error {
	if target == nil {
		return fmt.Errorf("json formatter: target cannot be nil")
	}

	output := jsonDetailedTarget{
		Name:          target.Name,
		Summary:       target.Summary.PlainText(),
		Documentation: target.Documentation,
		SourceFile:    target.SourceFile,
		LineNumber:    target.LineNumber,
	}

	// Add aliases if present
	if len(target.Aliases) > 0 {
		output.Aliases = target.Aliases
	}

	// Add variables if present
	if len(target.Variables) > 0 {
		output.Variables = make([]jsonVariable, len(target.Variables))
		for i, v := range target.Variables {
			output.Variables[i] = jsonVariable{
				Name:        v.Name,
				Description: v.Description,
			}
		}
	}

	// Marshal to JSON with 2-space indentation
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// RenderBasicTarget renders minimal info for a target without documentation in JSON format.
func (f *JSONFormatter) RenderBasicTarget(name string, sourceFile string, lineNumber int, w io.Writer) error {
	output := jsonBasicTarget{
		Name:       name,
		SourceFile: sourceFile,
		LineNumber: lineNumber,
	}

	// Marshal to JSON with 2-space indentation
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// ContentType returns the MIME type for JSON format.
func (f *JSONFormatter) ContentType() string {
	return "application/json"
}

// DefaultExtension returns the default file extension for JSON format.
func (f *JSONFormatter) DefaultExtension() string {
	return ".json"
}
