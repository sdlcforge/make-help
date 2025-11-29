# Stage 3: Detailed Help

## Objective

Implement `--target <name>` flag to show detailed help for a single target, including full documentation (not summary) and full variable descriptions.

**Can be done in parallel with Stage 2.**

## Files to Modify

### 1. `internal/cli/help.go`

**Add `runDetailedHelp()` function:**

```go
// runDetailedHelp displays detailed information for a single target.
// Shows full documentation, all variables with descriptions, aliases, and source location.
func runDetailedHelp(config *Config) error {
    // 1. Resolve Makefile path
    makefilePath, err := resolveMakefilePath(config.MakefilePath)
    if err != nil {
        return err
    }

    // 2. Discover all targets
    executor := discovery.NewDefaultExecutor()
    discoveryService := discovery.NewService(executor)
    targets, err := discoveryService.DiscoverTargets(makefilePath)
    if err != nil {
        return err
    }

    // 3. Check if target exists at all
    targetExists := false
    for _, t := range targets.Targets {
        if t == config.Target {
            targetExists = true
            break
        }
    }
    if !targetExists {
        return fmt.Errorf("target '%s' not found", config.Target)
    }

    // 4. Parse and build model (need full model to get documentation)
    files, err := discoveryService.DiscoverFiles(makefilePath)
    // ... parsing and model building

    // 5. Find target in model
    var foundTarget *model.Target
    for _, category := range helpModel.Categories {
        for i := range category.Targets {
            if category.Targets[i].Name == config.Target {
                foundTarget = &category.Targets[i]
                break
            }
        }
    }

    // 6. Render output
    renderer := format.NewRenderer(config.UseColor)

    if foundTarget != nil {
        // Target has documentation - use detailed renderer
        output := renderer.RenderDetailedTarget(foundTarget)
        fmt.Print(output)
    } else {
        // Target exists but has no documentation - show basic info
        output := renderer.RenderBasicTarget(config.Target, /* sourceFile, lineNumber if available */)
        fmt.Print(output)
    }

    return nil
}
```

**Update RunE dispatch in root.go (from Stage 1 stub):**

```go
} else if config.Target != "" {
    return runDetailedHelp(config)
}
```

### 2. `internal/format/renderer.go`

**Add `RenderBasicTarget()` for undocumented targets:**

```go
// RenderBasicTarget renders minimal info for a target without documentation.
// Shows target name and source location if available.
func (r *Renderer) RenderBasicTarget(name string, sourceFile string, lineNumber int) string {
    var buf strings.Builder

    buf.WriteString(r.colors.TargetName)
    buf.WriteString("Target: ")
    buf.WriteString(name)
    buf.WriteString(r.colors.Reset)
    buf.WriteString("\n")

    buf.WriteString("\n")
    buf.WriteString(r.colors.Documentation)
    buf.WriteString("No documentation available.\n")
    buf.WriteString(r.colors.Reset)

    if sourceFile != "" {
        buf.WriteString(fmt.Sprintf("\nSource: %s:%d\n", sourceFile, lineNumber))
    }

    return buf.String()
}
```

**Verify `RenderDetailedTarget()` shows full docs:**

The existing implementation already shows full documentation. Verify it includes:
- Full documentation text (not summary)
- All variables with full descriptions
- Aliases
- Source location

## Acceptance Criteria

- [ ] `make-help --target build` shows detailed help for `build` target
- [ ] Detailed view shows full documentation (all lines, not just first sentence)
- [ ] Detailed view shows all `@var` with full descriptions
- [ ] Detailed view shows aliases
- [ ] Detailed view shows source file and line number
- [ ] If target doesn't exist, error: "target 'foo' not found"
- [ ] If target exists but has no documentation, shows basic info with "No documentation available"
- [ ] Color output works correctly in detailed view

## New Tests

### `internal/cli/help_test.go`
```go
func TestRunDetailedHelp_ExistingTarget(t *testing.T) {
    // Test --target with a documented target shows full docs
}

func TestRunDetailedHelp_UndocumentedTarget(t *testing.T) {
    // Test --target with undocumented target shows basic info
}

func TestRunDetailedHelp_NonexistentTarget(t *testing.T) {
    // Test --target with nonexistent target returns error
}
```

### `internal/format/renderer_test.go`
```go
func TestRenderBasicTarget(t *testing.T) {
    // Test rendering of undocumented target
}

func TestRenderDetailedTarget_FullDocs(t *testing.T) {
    // Verify full documentation is shown, not summary
}
```

### `test/integration/cli_test.go`
```go
func TestDetailedHelp(t *testing.T) {
    // Integration test for --target flag
    // Verify output format matches expected
}
```

### `test/fixtures/expected/detailed_build.txt` (new fixture)
Example expected output for `--target build`:
```
Target: build

Aliases: b, compile

Variables:
  - BUILD_FLAGS: Flags passed to go build
  - OUTPUT_DIR: Directory for build output

Documentation:
  Builds the application binary.

  This target compiles the Go source code and produces
  an executable in the output directory.

Source: Makefile:42
```
