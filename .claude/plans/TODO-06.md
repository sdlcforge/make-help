# Stage 6: Cleanup & Documentation

## Objective

Remove deprecated subcommand files and update all documentation to reflect the new CLI structure.

## Files to Delete

### 1. `internal/cli/add_target.go`
No longer needed - functionality moved to `create_help_target.go`.

### 2. `internal/cli/remove_target.go`
No longer needed - functionality moved to `remove_help_target.go`.

## Files to Modify

### 1. `README.md`

Update CLI documentation:

**Before:**
```
make-help                    # Display help
make-help add-target         # Add help target to Makefile
make-help remove-target      # Remove help target
```

**After:**
```markdown
## Installation

### Project-Local (Recommended)

Run once to set up help targets in your project:

    make-help --create-help-target

This creates `make/01-help.mk` with:
- Automatic binary installation via `go install`
- `make help` - displays help summary
- `make help-<target>` - detailed help for each documented target

### Global

    go install github.com/sdlcforge/make-help/cmd/make-help@latest

## Usage

### Display Help (default)

    make-help                              # Show help for ./Makefile
    make-help --makefile-path path/to/Makefile

### Detailed Target Help

    make-help --target build               # Full docs for 'build' target

### Target Filtering

By default, only documented targets appear in help output.

    make-help --include-target clean       # Include undocumented 'clean'
    make-help --include-target foo,bar     # Include multiple (comma-separated)
    make-help --include-all-phony          # Include all .PHONY targets

### Generate Help Targets

    make-help --create-help-target         # Create make/01-help.mk
    make-help --create-help-target --version v1.2.3   # Pin version
    make-help --create-help-target --default-category Misc

### Remove Help Targets

    make-help --remove-help-target         # Remove generated help files

## Flags Reference

| Flag | Description |
|------|-------------|
| `--makefile-path` | Path to Makefile (default: ./Makefile) |
| `--target` | Show detailed help for specific target |
| `--include-target` | Include undocumented target (repeatable) |
| `--include-all-phony` | Include all .PHONY targets |
| `--create-help-target` | Generate help target file |
| `--remove-help-target` | Remove help target from Makefile |
| `--help-file-path` | Override path for generated file |
| `--version` | Pin version in generated go install |
| `--default-category` | Default category for uncategorized targets |
| `--keep-order-categories` | Preserve category discovery order |
| `--keep-order-targets` | Preserve target discovery order |
| `--category-order` | Explicit category order (comma-separated) |
| `--no-color` | Disable colored output |
| `--color` | Force colored output |
| `--verbose` | Enable verbose output |
```

### 2. `.claude/CLAUDE.md`

Update build commands section if needed. Add notes about new flags:

```markdown
## CLI Flags

The CLI uses flags instead of subcommands:

- `--create-help-target` - replaces `add-target` subcommand
- `--remove-help-target` - replaces `remove-target` subcommand
- `--target <name>` - show detailed help for single target
- `--include-target` - include undocumented targets
- `--include-all-phony` - include all .PHONY targets
```

### 3. `docs/design.md`

Update CLI section:

```markdown
## CLI Layer

The CLI is implemented using Cobra with a single root command and flags:

### Mode Flags (mutually exclusive)

1. **Default (no special flags)**: Display help output
2. **`--target <name>`**: Display detailed help for single target
3. **`--create-help-target`**: Generate help target file
4. **`--remove-help-target`**: Remove help targets

### Target Filtering

- `--include-target`: Include specific undocumented targets
- `--include-all-phony`: Include all .PHONY targets
- By default, only documented targets (with `## ` comments) are shown

### Generated Help File

When `--create-help-target` is used, generates a Makefile include with:

- `GOBIN ?= .bin` - configurable binary directory
- Binary installation target using `go install`
- `.PHONY: help` target for summary
- `.PHONY: help-<target>` for each documented target
```

Update data flow diagram if present.

### 4. `internal/cli/root.go`

Update command description:

```go
Long: `make-help generates formatted help output from Makefile documentation.

Default behavior displays help. Use flags for other operations:
  --target <name>       Show detailed help for a target
  --create-help-target  Generate help target file
  --remove-help-target  Remove help targets

Documentation directives (in ## comments):
  @file         File-level documentation
  @category     Group targets into categories
  @var          Document environment variables
  @alias        Define target aliases`,
```

## Acceptance Criteria

- [ ] `internal/cli/add_target.go` is deleted
- [ ] `internal/cli/remove_target.go` is deleted
- [ ] Code still compiles after deletions
- [ ] `README.md` documents all new flags
- [ ] `README.md` shows project-local installation workflow
- [ ] `README.md` includes examples for all major use cases
- [ ] `.claude/CLAUDE.md` reflects new CLI structure
- [ ] `docs/design.md` updated with new architecture
- [ ] `make-help --help` shows accurate, well-organized help text
- [ ] All tests pass

## New Tests

### Final Integration Tests

```go
func TestCLI_HelpOutput(t *testing.T) {
    // Verify --help shows all flags correctly
}

func TestFullWorkflow(t *testing.T) {
    // End-to-end test:
    // 1. Create temp Makefile with targets
    // 2. Run make-help --create-help-target
    // 3. Verify generated file
    // 4. Run make help (using generated file)
    // 5. Run make help-<target>
    // 6. Run make-help --remove-help-target
    // 7. Verify cleanup
}
```

## Verification Checklist

Run before considering stage complete:

```bash
# Build
go build ./cmd/make-help

# All tests pass
go test ./...

# Help output looks correct
./make-help --help

# Quick smoke test
./make-help --makefile-path test/fixtures/makefiles/basic.mk
./make-help --target build --makefile-path test/fixtures/makefiles/basic.mk
```
