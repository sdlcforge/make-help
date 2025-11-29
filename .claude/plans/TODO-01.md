# Stage 1: Config & Flag Infrastructure

## Objective

Set up the foundational config fields and CLI flags needed by subsequent stages. Remove subcommand registration but keep subcommand files until Stage 6.

## Files to Modify

### 1. `internal/cli/config.go`

Add new fields to Config struct:
```go
CreateHelpTarget  bool
RemoveHelpTarget  bool
Version           string     // Version for go install (e.g., "v1.2.3"), empty = @latest
IncludeTargets    []string   // From --include-target flag
IncludeAllPhony   bool
Target            string     // For --target <name> detailed view
HelpFilePath      string     // Rename from TargetFile
```

Rename `TargetFile` to `HelpFilePath` for clarity.

### 2. `internal/cli/root.go`

**Remove:**
- `rootCmd.AddCommand(newAddTargetCmd(config))`
- `rootCmd.AddCommand(newRemoveTargetCmd(config))`

**Add flags:**
```go
rootCmd.Flags().BoolVar(&config.CreateHelpTarget,
    "create-help-target", false, "Generate help target file with local binary installation")
rootCmd.Flags().BoolVar(&config.RemoveHelpTarget,
    "remove-help-target", false, "Remove help target from Makefile")
rootCmd.Flags().StringVar(&config.Version,
    "version", "", "Version to pin in generated go install (e.g., v1.2.3)")
rootCmd.Flags().StringSliceVar(&config.IncludeTargets,
    "include-target", []string{}, "Include undocumented target in help (repeatable, comma-separated)")
rootCmd.Flags().BoolVar(&config.IncludeAllPhony,
    "include-all-phony", false, "Include all .PHONY targets in help output")
rootCmd.Flags().StringVar(&config.Target,
    "target", "", "Show detailed help for a specific target")
rootCmd.Flags().StringVar(&config.HelpFilePath,
    "help-file-path", "", "Explicit path for generated help target file")
```

**Add validation in PreRunE:**
```go
// Mutual exclusivity
if config.CreateHelpTarget && config.RemoveHelpTarget {
    return fmt.Errorf("cannot use both --create-help-target and --remove-help-target")
}

// --remove-help-target only allows --verbose and --makefile-path
if config.RemoveHelpTarget {
    // Check that no other flags are set (except verbose, makefile-path)
    if config.Target != "" || len(config.IncludeTargets) > 0 || config.IncludeAllPhony || ... {
        return fmt.Errorf("--remove-help-target only accepts --verbose and --makefile-path flags")
    }
}
```

**Stub RunE dispatch (actual implementations in later stages):**
```go
if config.RemoveHelpTarget {
    return fmt.Errorf("--remove-help-target not yet implemented")
} else if config.CreateHelpTarget {
    return fmt.Errorf("--create-help-target not yet implemented")
} else if config.Target != "" {
    return fmt.Errorf("--target not yet implemented")
} else {
    return runHelp(config)
}
```

### 3. `internal/cli/flags.go` (new file)

Create helper for parsing comma-separated + repeatable flags:
```go
package cli

import "strings"

// parseIncludeTargets normalizes the --include-target flag values.
// Handles both comma-separated ("foo,bar") and repeated flags.
// Input: ["foo,bar", "baz"] -> Output: ["foo", "bar", "baz"]
func parseIncludeTargets(input []string) []string {
    var result []string
    for _, item := range input {
        parts := strings.Split(item, ",")
        for _, p := range parts {
            if trimmed := strings.TrimSpace(p); trimmed != "" {
                result = append(result, trimmed)
            }
        }
    }
    return result
}
```

## Acceptance Criteria

- [ ] All new flags are registered and appear in `--help` output
- [ ] `--create-help-target` and `--remove-help-target` are mutually exclusive (error if both set)
- [ ] `--remove-help-target` errors if combined with flags other than `--verbose`, `--makefile-path`
- [ ] Old subcommands (`add-target`, `remove-target`) no longer appear in help
- [ ] `make-help` (default behavior) still works as before
- [ ] `--include-target foo,bar` and `--include-target foo --include-target bar` both parse correctly

## New Tests

### `internal/cli/flags_test.go`
```go
func TestParseIncludeTargets(t *testing.T) {
    tests := []struct {
        name     string
        input    []string
        expected []string
    }{
        {"single value", []string{"foo"}, []string{"foo"}},
        {"comma separated", []string{"foo,bar"}, []string{"foo", "bar"}},
        {"repeated flags", []string{"foo", "bar"}, []string{"foo", "bar"}},
        {"mixed", []string{"foo,bar", "baz"}, []string{"foo", "bar", "baz"}},
        {"with spaces", []string{"foo, bar"}, []string{"foo", "bar"}},
        {"empty", []string{}, []string{}},
    }
    // ... test implementation
}
```

### `internal/cli/root_test.go` (add tests)
```go
func TestMutualExclusivityFlags(t *testing.T) {
    // Test --create-help-target and --remove-help-target error together
}

func TestRemoveHelpTargetFlagRestrictions(t *testing.T) {
    // Test --remove-help-target rejects other flags
}
```
