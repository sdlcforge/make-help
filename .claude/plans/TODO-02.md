# Stage 2: Target Filtering

## Objective

Implement filtering to show only documented targets by default, with `--include-target` and `--include-all-phony` options to include additional targets.

**Can be done in parallel with Stage 3.**

## Files to Modify

### 1. `internal/discovery/targets.go`

**Modify to track .PHONY status:**

Current function returns `[]string` of target names. Modify to also return phony status:

```go
// DiscoverTargetsResult contains discovered targets and their metadata
type DiscoverTargetsResult struct {
    Targets []string          // All discovered target names in order
    IsPhony map[string]bool   // Map of target name to phony status
}

// DiscoverTargets finds all user-defined targets and their .PHONY status
func (s *Service) DiscoverTargets(makefilePath string) (*DiscoverTargetsResult, error) {
    // Existing logic to run `make -f <makefile> -p -r`
    //
    // Additionally parse lines like:
    //   .PHONY: build test clean
    // To populate IsPhony map
}
```

Parse `.PHONY` declarations from `make -p` output:
- Lines matching `^\.PHONY:` followed by target names
- Multiple `.PHONY` lines may exist

### 2. `internal/model/builder.go`

**Add filtering config:**

```go
type BuilderConfig struct {
    // ... existing fields
    IncludeTargets   []string  // Explicitly include these undocumented targets
    IncludeAllPhony  bool      // Include all .PHONY targets even if undocumented
    PhonyTargets     map[string]bool  // Which targets are .PHONY
}
```

**Modify Build() to filter:**

```go
func (b *Builder) Build() (*HelpModel, error) {
    // ... existing parsing logic

    // After collecting all targets, filter:
    for _, target := range allTargets {
        shouldInclude := false

        // Include if documented
        if len(target.Documentation) > 0 {
            shouldInclude = true
        }

        // Include if in IncludeTargets list
        if contains(b.config.IncludeTargets, target.Name) {
            shouldInclude = true
        }

        // Include if .PHONY and IncludeAllPhony is set
        if b.config.IncludeAllPhony && b.config.PhonyTargets[target.Name] {
            shouldInclude = true
        }

        if shouldInclude {
            // Add to result
        }
    }
}
```

### 3. `internal/cli/help.go`

**Pass filtering config to builder:**

```go
func runHelp(config *Config) error {
    // ... existing discovery logic

    // Discover targets with phony status
    targetsResult, err := discoveryService.DiscoverTargets(makefilePath)

    // Parse include targets
    includeTargets := parseIncludeTargets(config.IncludeTargets)

    // Build model with filtering
    builderConfig := &model.BuilderConfig{
        // ... existing config
        IncludeTargets:  includeTargets,
        IncludeAllPhony: config.IncludeAllPhony,
        PhonyTargets:    targetsResult.IsPhony,
    }

    builder := model.NewBuilder(builderConfig)
    // ...
}
```

### 4. `internal/model/types.go`

**Add IsPhony field to Target (optional, for future use):**

```go
type Target struct {
    // ... existing fields
    IsPhony bool  // Whether target is declared .PHONY
}
```

## Acceptance Criteria

- [ ] By default, only targets with `## ` documentation appear in help output
- [ ] `--include-target foo` includes `foo` in output even without documentation
- [ ] `--include-target foo,bar` works (comma-separated)
- [ ] `--include-target foo --include-target bar` works (repeated)
- [ ] `--include-all-phony` includes all `.PHONY` targets
- [ ] Targets included via `--include-target` show name but no description (since undocumented)
- [ ] `.PHONY` status is correctly parsed from `make -p` output

## New Tests

### `internal/discovery/targets_test.go`
```go
func TestDiscoverTargets_PhonyStatus(t *testing.T) {
    // Test that .PHONY targets are correctly identified
    // Mock make -p output with .PHONY declarations
}
```

### `internal/model/builder_test.go`
```go
func TestBuilder_FilterUndocumented(t *testing.T) {
    // Test that undocumented targets are excluded by default
}

func TestBuilder_IncludeTargets(t *testing.T) {
    // Test --include-target includes specific undocumented targets
}

func TestBuilder_IncludeAllPhony(t *testing.T) {
    // Test --include-all-phony includes all .PHONY targets
}
```

### `test/fixtures/makefiles/undocumented.mk` (new fixture)
```makefile
## Documented target.
build:
	@echo "build"

# No documentation - should be hidden by default
clean:
	@echo "clean"

.PHONY: build clean lint

# Also undocumented
lint:
	@echo "lint"
```

### `test/integration/cli_test.go`
```go
func TestHelpFilteredOutput(t *testing.T) {
    // Test default shows only documented targets
}

func TestHelpIncludeTarget(t *testing.T) {
    // Test --include-target shows specified undocumented targets
}

func TestHelpIncludeAllPhony(t *testing.T) {
    // Test --include-all-phony shows all .PHONY targets
}
```
