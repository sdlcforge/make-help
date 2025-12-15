package model

// UncategorizedCategoryName is the category name used for targets without
// an explicit !category directive. Empty string has special meaning as the
// default/uncategorized group.
const UncategorizedCategoryName = ""

// HelpModel represents the complete parsed help documentation from all Makefiles.
type HelpModel struct {
	// FileDocs contains !file documentation sections in discovery order.
	FileDocs []string

	// Categories contains all documented categories with their targets.
	Categories []Category

	// HasCategories is true if any !category directives were found.
	HasCategories bool

	// DefaultCategory is the category name for uncategorized targets
	// (set via --default-category flag).
	DefaultCategory string
}

// Category represents a documentation category containing related targets.
type Category struct {
	// Name is the category name from !category directive.
	// Empty string represents the default/uncategorized group.
	Name string

	// Targets contains all targets in this category.
	Targets []Target

	// DiscoveryOrder tracks when this category was first encountered
	// (used for --keep-order-categories).
	DiscoveryOrder int
}

// Target represents a documented Makefile target.
type Target struct {
	// Name is the primary target name.
	Name string

	// Aliases contains alternative names from !alias directives.
	Aliases []string

	// Documentation contains the full documentation lines (without ## prefix).
	Documentation []string

	// Summary is the extracted first sentence (computed from Documentation).
	Summary string

	// Variables contains associated environment variables from !var directives.
	Variables []Variable

	// DiscoveryOrder tracks when this target was first encountered
	// (used for --keep-order-targets).
	DiscoveryOrder int

	// SourceFile is the path to the file where this target was documented.
	SourceFile string

	// LineNumber is the line number where the target definition appears.
	LineNumber int

	// IsPhony indicates whether this target is declared as .PHONY.
	IsPhony bool
}

// Variable represents a documented environment variable associated with a target.
type Variable struct {
	// Name is the variable name (e.g., "DEBUG", "PORT").
	Name string

	// Description is the full description text from !var directive.
	Description string
}
