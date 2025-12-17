package parser

// DirectiveType represents the type of a documentation directive.
type DirectiveType int

const (
	// DirectiveFile represents !file directive for file-level documentation.
	DirectiveFile DirectiveType = iota

	// DirectiveCategory represents !category directive for target grouping.
	DirectiveCategory

	// DirectiveVar represents !var directive for environment variable documentation.
	DirectiveVar

	// DirectiveAlias represents !alias directive for target aliases.
	DirectiveAlias

	// DirectiveNotAlias represents !notalias directive to exclude a target from implicit alias detection.
	DirectiveNotAlias

	// DirectiveDoc represents a regular documentation line (not a special directive).
	DirectiveDoc
)

// String returns the string representation of DirectiveType.
func (d DirectiveType) String() string {
	switch d {
	case DirectiveFile:
		return "file"
	case DirectiveCategory:
		return "category"
	case DirectiveVar:
		return "var"
	case DirectiveAlias:
		return "alias"
	case DirectiveNotAlias:
		return "notalias"
	case DirectiveDoc:
		return "doc"
	default:
		return "unknown"
	}
}

// Directive represents a parsed documentation directive from a Makefile.
type Directive struct {
	// Type indicates the directive type (!file, !category, !var, !alias, or doc).
	Type DirectiveType

	// Value contains the directive content after the directive keyword.
	// For !category: the category name
	// For !var: "NAME - description"
	// For !alias: "alias1, alias2, ..."
	// For doc: the documentation text
	Value string

	// SourceFile is the path to the file where this directive appears.
	SourceFile string

	// LineNumber is the 1-based line number where this directive appears.
	LineNumber int
}

// ParsedFile represents the parsing result for a single Makefile.
type ParsedFile struct {
	// Path is the absolute path to the parsed file.
	Path string

	// Directives contains all parsed documentation directives in order.
	Directives []Directive

	// TargetMap maps target names to their line numbers.
	// Used to associate documentation with targets.
	TargetMap map[string]int
}
