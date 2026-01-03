# Data Models

Core data structures used throughout the make-help system.

## Table of contents

- [Makefile Documentation Model](#makefile-documentation-model)
- [Configuration Model](#configuration-model)
- [Discovery Model](#discovery-model)
- [Rendering Model](#rendering-model)

---

## Overview

### 1. Target Service Models

#### Config
Configuration for target manipulation operations (add/remove help targets).

**Key fields:**
- `MakefilePath` - Path to the Makefile being modified
- `TargetFileRelPath` - Relative path for generated help file (e.g., "make/help.mk")
- `KeepOrderCategories`, `KeepOrderTargets` - Preserve discovery order
- `CategoryOrder` - Explicit category ordering
- `DefaultCategory` - Category name for uncategorized targets

[View source](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/target/add.go#L16-L23)

#### AddService
Service for adding help targets to Makefiles. Uses Config to determine placement and includes the logic for validating Makefiles, determining target file location, generating content, and adding include directives.

[View source](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/target/add.go#L26-L30)

#### RemoveService
Service for removing help target artifacts from Makefiles. Cleans up include directives and deletes generated help files.

[View source](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/target/remove.go#L16-L20)

#### IncludePattern
Information about a detected include directive pattern in the Makefile.

**Key fields:**
- `Suffix` - File extension (e.g., ".mk" or "")
- `FullPattern` - Complete include pattern (e.g., "make/*.mk")
- `PatternPrefix` - Prefix before wildcard (e.g., "make/" or "./make/")

[View source](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/target/add.go#L128-L135)

#### GeneratorConfig
Configuration for static help file generation, including rendering options and the help model.

**Key fields:**
- `HelpModel` - The built model to render
- `UseColor` - Whether to embed ANSI color codes
- `Makefiles` - List of discovered Makefiles for dependency tracking
- `CommandLine` - Full command line used to generate the file (for restoration)

[View source](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/target/generator.go#L14-L41)

### 2. Lint Service Models

#### Check
Represents a lint check with optional auto-fix capability.

**Key fields:**
- `Name` - Unique identifier (e.g., "summary-punctuation")
- `CheckFunc` - Function that performs the check
- `FixFunc` - Function that generates a fix (may be nil if not auto-fixable)

[View source](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/lint/types.go#L20-L28)

#### Fix
Represents a single file modification to fix a lint warning.

**Key fields:**
- `File` - Absolute path to the file to modify
- `Line` - 1-indexed line number to modify
- `Operation` - Type of modification (FixReplace or FixDelete)
- `OldContent` - Expected current content (for validation)
- `NewContent` - Replacement content

[View source](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/lint/types.go#L36-L51)

#### FixOperation
Enum specifying the type of file modification: `FixReplace` (replace entire line) or `FixDelete` (remove line).

[View source](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/lint/types.go#L54-L60)

#### Fixer
Applies fixes to source files, with support for dry-run mode.

[View source](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/lint/types.go#L63-L66)

#### FixResult
Contains the results of applying fixes, including total fixes applied and files modified.

[View source](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/lint/types.go#L69-L74)

### 3. Makefile Documentation Model

#### HelpModel
The complete parsed help documentation from all Makefiles. This is the core data structure built by the model builder.

**Key fields:**
- `FileDocs` - !file documentation sections in discovery order
- `Categories` - All documented categories with their targets
- `HasCategories` - True if any !category directives were found
- `DefaultCategory` - Category name for uncategorized targets

[View source](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/model/types.go#L8-L22)

#### Category
A documentation category containing related targets.

**Key fields:**
- `Name` - Category name from !category directive (empty string = uncategorized)
- `Targets` - All targets in this category
- `DiscoveryOrder` - When this category was first encountered (for --keep-order-categories)

[View source](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/model/types.go#L24-L36)

#### Target
A documented Makefile target with all associated metadata.

**Key fields:**
- `Name` - Primary target name
- `Aliases` - Alternative names from !alias directives
- `Documentation` - Full documentation lines (without ## prefix)
- `Summary` - Extracted first sentence (computed from Documentation)
- `Variables` - Associated environment variables from !var directives
- `DiscoveryOrder` - When target was first encountered (for --keep-order-targets)
- `SourceFile`, `LineNumber` - Location information
- `IsPhony` - Whether target is declared as .PHONY

[View source](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/model/types.go#L38-L67)

#### Variable
A documented environment variable associated with a target.

**Key fields:**
- `Name` - Variable name (e.g., "DEBUG", "PORT")
- `Description` - Full description text from !var directive

[View source](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/model/types.go#L69-L76)

#### Directive
A parsed documentation directive from a Makefile. Used during parsing before being assembled into the HelpModel.

**Key fields:**
- `Type` - Directive type (see DirectiveType below)
- `Value` - Directive content after the directive keyword
- `SourceFile`, `LineNumber` - Location information

[View source](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/parser/types.go#L41-L58)

#### DirectiveType
Enum representing the type of documentation directive: `DirectiveFile`, `DirectiveCategory`, `DirectiveVar`, `DirectiveAlias`, or `DirectiveDoc` (regular documentation line).

[View source](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/parser/types.go#L3-L21)

#### ParsedFile
The parsing result for a single Makefile, containing directives and target information.

**Key fields:**
- `Path` - Absolute path to the parsed file
- `Directives` - All parsed documentation directives in order
- `TargetMap` - Maps target names to their line numbers

[View source](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/parser/types.go#L60-L71)

### 4. Configuration Model

#### Config (CLI)
All CLI configuration options, including global settings, help generation options, and derived state.

**Key fields:**
- **Global:** `MakefilePath`, `ColorMode`, `Verbose`
- **Help generation:** `KeepOrderCategories`, `KeepOrderTargets`, `CategoryOrder`, `DefaultCategory`, `HelpCategory`
- **Mode control:** `ShowHelp`, `RemoveHelpTarget`, `Lint`, `Fix`, `DryRun`
- **Include options:** `IncludeTargets`, `IncludeAllPhony`
- **Target detail:** `Target` (for detailed help view)
- **Derived:** `UseColor` (computed based on ColorMode and terminal detection)

[View source](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/cli/config.go#L31-L102)

#### ColorMode
Enum for color output mode: `ColorAuto` (terminal detection), `ColorAlways`, or `ColorNever`.

[View source](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/cli/config.go#L3-L15)

### 5. Discovery Model

#### Service
Provides Makefile and target discovery functionality using the `make` command.

**Key methods:**
- `DiscoverMakefiles()` - Find all Makefiles using MAKEFILE_LIST
- `DiscoverTargets()` - Extract targets from make -p database output

[View source](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/discovery/service.go#L8-L12)

#### DiscoverTargetsResult
Contains discovered targets and their metadata extracted from make -p output.

**Key fields:**
- `Targets` - All discovered target names in order
- `IsPhony` - Maps target names to their .PHONY status
- `Dependencies` - Maps target names to their prerequisite targets
- `HasRecipe` - Maps target names to whether they have a recipe

[View source](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/discovery/targets.go#L12-L24)

### 6. Rendering Model

#### ColorScheme
Defines ANSI color codes for different elements in the rendered output.

**Key fields:**
- `CategoryName`, `TargetName`, `Alias`, `Variable`, `Documentation` - ANSI codes for each element type
- `Reset` - ANSI reset code

**Note:** All fields are empty strings when color is disabled.

[View source](https://github.com/sdlcforge/make-help/blob/86a8eea0cb298def52ddd7dcbe70107532e5ef69/internal/format/colors.go#L3-L14)


Last reviewed: 2025-12-25T16:43Z
