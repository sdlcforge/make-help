// Package model defines the core data structures for help documentation.
//
// The central type is HelpModel, which aggregates file documentation,
// categories, and targets with their associated aliases and variables.
// The Builder type (implemented separately) constructs HelpModel from
// parsed directives using the builder pattern for complex object construction.
//
// # Data Model
//
// The model hierarchy is:
//
//	HelpModel
//	├── FileDocs []string           // @file documentation
//	└── Categories []Category
//	    └── Targets []Target
//	        ├── Aliases []string    // @alias directives
//	        └── Variables []Variable // @var directives
//
// # Discovery Order
//
// Both Category and Target include a DiscoveryOrder field that tracks
// when each element was first encountered during Makefile parsing.
// This supports the --keep-order-categories and --keep-order-targets
// flags which preserve discovery order instead of alphabetical sorting.
package model
