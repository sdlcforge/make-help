# Stage 5: Create/Remove Help Target Commands

## Objective

Implement `--create-help-target` and `--remove-help-target` flag handlers, including conflict detection for `help-<target>` names.

## Files to Create

### 1. `internal/cli/create_help_target.go` (new file)

```go
package cli

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/sdlcforge/make-help/internal/discovery"
    "github.com/sdlcforge/make-help/internal/model"
    "github.com/sdlcforge/make-help/internal/parser"
    "github.com/sdlcforge/make-help/internal/target"
)

// runCreateHelpTarget generates and writes the help target file.
func runCreateHelpTarget(config *Config) error {
    // 1. Resolve Makefile path
    makefilePath, err := resolveMakefilePath(config.MakefilePath)
    if err != nil {
        return err
    }

    // 2. Validate Makefile syntax
    executor := discovery.NewDefaultExecutor()
    if err := target.ValidateMakefile(executor, makefilePath); err != nil {
        return fmt.Errorf("Makefile validation failed: %w", err)
    }

    // 3. Discover files and targets
    discoveryService := discovery.NewService(executor)
    files, err := discoveryService.DiscoverFiles(makefilePath)
    if err != nil {
        return fmt.Errorf("failed to discover Makefile includes: %w", err)
    }

    targetsResult, err := discoveryService.DiscoverTargets(makefilePath)
    if err != nil {
        return fmt.Errorf("failed to discover targets: %w", err)
    }

    // 4. Parse and build model to get documented targets
    parsedFiles, err := parser.ParseFiles(files)
    if err != nil {
        return err
    }

    builderConfig := &model.BuilderConfig{
        // ... config
    }
    builder := model.NewBuilder(builderConfig)
    for _, pf := range parsedFiles {
        builder.AddFile(pf)
    }
    helpModel, err := builder.Build()
    if err != nil {
        return err
    }

    // 5. Collect documented target names
    var documentedTargets []string
    for _, category := range helpModel.Categories {
        for _, t := range category.Targets {
            documentedTargets = append(documentedTargets, t.Name)
        }
    }

    // 6. Check for help-<target> conflicts
    existingTargets := make(map[string]bool)
    for _, t := range targetsResult.Targets {
        existingTargets[t] = true
    }

    for _, t := range documentedTargets {
        helpTargetName := "help-" + t
        if existingTargets[helpTargetName] {
            return fmt.Errorf("cannot generate %s target - target already exists in Makefile\n"+
                "Consider renaming your existing %s target", helpTargetName, helpTargetName)
        }
    }

    // Also check for 'help' conflict
    if existingTargets["help"] {
        return fmt.Errorf("cannot generate help target - target already exists in Makefile")
    }

    // 7. Determine target file location
    targetFile, needsInclude, err := target.DetermineTargetFile(makefilePath, config.HelpFilePath)
    if err != nil {
        return err
    }

    // 8. Generate help file content
    genConfig := &target.GeneratorConfig{
        KeepOrderCategories: config.KeepOrderCategories,
        KeepOrderTargets:    config.KeepOrderTargets,
        CategoryOrder:       config.CategoryOrder,
        DefaultCategory:     config.DefaultCategory,
        IncludeTargets:      parseIncludeTargets(config.IncludeTargets),
        IncludeAllPhony:     config.IncludeAllPhony,
        Version:             config.Version,
    }
    content := target.GenerateHelpFile(genConfig, documentedTargets)

    // 9. Write file atomically
    if err := target.AtomicWriteFile(targetFile, []byte(content), 0644); err != nil {
        return fmt.Errorf("failed to write help target file %s: %w", targetFile, err)
    }

    if config.Verbose {
        fmt.Printf("Created help target file: %s\n", targetFile)
    }

    // 10. Add include directive if needed
    if needsInclude {
        if err := target.AddIncludeDirective(makefilePath, targetFile); err != nil {
            return err
        }
        if config.Verbose {
            fmt.Printf("Added include directive to: %s\n", makefilePath)
        }
    }

    fmt.Printf("Successfully created help target: %s\n", targetFile)
    return nil
}
```

### 2. `internal/cli/remove_help_target.go` (new file)

```go
package cli

import (
    "fmt"

    "github.com/sdlcforge/make-help/internal/discovery"
    "github.com/sdlcforge/make-help/internal/target"
)

// runRemoveHelpTarget removes help targets from the Makefile.
func runRemoveHelpTarget(config *Config) error {
    // 1. Resolve Makefile path
    makefilePath, err := resolveMakefilePath(config.MakefilePath)
    if err != nil {
        return err
    }

    // 2. Create remove service and execute
    executor := discovery.NewDefaultExecutor()
    removeService := target.NewRemoveService(makefilePath, executor, config.Verbose)

    if err := removeService.RemoveTarget(); err != nil {
        return fmt.Errorf("failed to remove help target: %w", err)
    }

    fmt.Println("Successfully removed help target")
    return nil
}
```

## Files to Modify

### 3. `internal/target/add.go` - Refactor to utilities

Convert methods to standalone functions that can be called from `create_help_target.go`:

```go
// ValidateMakefile runs `make -n` to check for syntax errors.
func ValidateMakefile(executor discovery.CommandExecutor, makefilePath string) error {
    // Move logic from AddService.validateMakefile()
}

// DetermineTargetFile decides where to create the help target.
// Returns: (targetFile path, needsInclude directive, error)
func DetermineTargetFile(makefilePath, explicitPath string) (string, bool, error) {
    // Move logic from AddService.determineTargetFile()
}

// AddIncludeDirective injects an include statement into the Makefile.
func AddIncludeDirective(makefilePath, targetFile string) error {
    // Move logic from AddService.addIncludeDirective()
}
```

### 4. `internal/target/file.go` - Export atomic write

```go
// AtomicWriteFile writes data atomically using temp file + rename.
// Export this function (was atomicWriteFile)
func AtomicWriteFile(filename string, data []byte, perm os.FileMode) error {
    // Existing implementation
}
```

### 5. `internal/cli/root.go` - Wire up commands

Update RunE to call actual implementations:

```go
if config.RemoveHelpTarget {
    return runRemoveHelpTarget(config)
} else if config.CreateHelpTarget {
    return runCreateHelpTarget(config)
} else if config.Target != "" {
    return runDetailedHelp(config)
} else {
    return runHelp(config)
}
```

## Acceptance Criteria

- [ ] `make-help --create-help-target` generates `make/01-help.mk` (or appropriate location)
- [ ] Generated file contains all `help-<target>` targets for documented targets
- [ ] Conflict detection errors if `help` or `help-<target>` already exists
- [ ] Error message suggests renaming conflicting target
- [ ] Include directive added to main Makefile if needed
- [ ] `--help-file-path` overrides default location
- [ ] `--version` is passed through to generator
- [ ] All `--keep-order-*`, `--category-order`, `--default-category` flags are embedded
- [ ] `--remove-help-target` removes help file and include directives
- [ ] Verbose mode shows progress messages
- [ ] Atomic file writes prevent corruption

## New Tests

### `internal/cli/create_help_target_test.go`
```go
func TestRunCreateHelpTarget_Basic(t *testing.T) {
    // Test basic help target creation
}

func TestRunCreateHelpTarget_ConflictDetection(t *testing.T) {
    // Test error when help-<target> already exists
}

func TestRunCreateHelpTarget_HelpConflict(t *testing.T) {
    // Test error when help target already exists
}

func TestRunCreateHelpTarget_WithOptions(t *testing.T) {
    // Test all options are passed through
}

func TestRunCreateHelpTarget_ExplicitPath(t *testing.T) {
    // Test --help-file-path override
}
```

### `internal/cli/remove_help_target_test.go`
```go
func TestRunRemoveHelpTarget(t *testing.T) {
    // Test removal of help target file and includes
}
```

### `test/fixtures/makefiles/with_help_conflict.mk` (new fixture)
```makefile
## Existing help-build target.
help-build:
	@echo "This conflicts"

## Build target.
build:
	@echo "build"
```

### `test/integration/cli_test.go`
```go
func TestCreateHelpTarget(t *testing.T) {
    // Integration test for --create-help-target
}

func TestCreateHelpTarget_Conflict(t *testing.T) {
    // Integration test for conflict detection
}

func TestRemoveHelpTarget(t *testing.T) {
    // Integration test for --remove-help-target
}
```
