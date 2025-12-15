# make-help Examples

This directory contains working examples demonstrating different features and use cases of `make-help`. Each example is a self-contained project with its own Makefile and generated static help file.

## Quick Start

Each example directory includes a generated `help.mk` file with embedded help text. To see the help output:

```bash
cd examples/<example-name>
make help
```

To regenerate the help file:

```bash
make-help
```

Or display help dynamically without generating a file:

```bash
make-help --show-help --makefile-path examples/<example-name>/Makefile
```

## Examples

### uncategorized-targets

**When to use:** Starting point for simple projects or when you don't need categorization.

Demonstrates basic target documentation without categories. Targets appear in a flat list with their descriptions. Shows:
- Basic `##` comment documentation
- `!var` directives for environment variables
- `!alias` directives for target shortcuts
- Multi-line target descriptions

Best for small projects with fewer than 10 targets where grouping isn't necessary.

### categorized-project

**When to use:** Organizing medium to large projects with logical groupings.

Shows how to use `!category` to organize targets into sections like Build, Test, Development, and Maintenance. Each category appears as a header in the help output. Demonstrates:
- `!category` directives for grouping related targets
- Consistent organization across multiple categories
- Complete project structure (build, test, dev, maintenance workflows)

Best for projects with 10+ targets that benefit from logical grouping by function or workflow.

### filtering-demo

**When to use:** Understanding how to selectively display targets.

Demonstrates target filtering capabilities with both documented and undocumented targets. Shows:
- Default behavior (only documented targets shown)
- `--include-target` flag to expose specific undocumented targets
- `--include-all-phony` flag to show all .PHONY targets
- Convention for internal targets (prefix with `_`)

Best for understanding how to create public vs. internal/utility targets and control what appears in help output.

### full-featured

**When to use:** Comprehensive reference for all make-help features.

Complete showcase of all available directives and features:
- `!file` for project-level documentation
- `!category` for organizing targets
- `!var` for documenting environment variables
- `!alias` for target shortcuts
- Multi-line documentation with detailed descriptions
- Real-world examples (Docker, releases, development, code quality)

Best as a reference implementation when you need to use advanced features or see examples of complex documentation patterns.

## Running Examples

### View Help Output

From any example directory:

```bash
make help
```

### View Detailed Help for a Specific Target

```bash
make help-<target-name>

# Example:
make help-build
```

### Generate Fresh Help File

If you modify a Makefile, regenerate the help file:

```bash
make-help
```

The help file will also auto-regenerate when you run `make help` if the source Makefile is newer than the help file.

## Example Comparison

| Feature | uncategorized | categorized | filtering | full-featured |
|---------|---------------|-------------|-----------|---------------|
| !file documentation | ✓ | ✓ | ✓ | ✓ |
| !category | ✗ | ✓ | ✓ | ✓ |
| !var | ✓ | ✓ | ✓ | ✓ |
| !alias | ✓ | ✓ | ✗ | ✓ |
| Multi-line docs | ✓ | ✗ | ✗ | ✓ |
| Filtering demo | ✗ | ✗ | ✓ | ✗ |
| Complexity | Simple | Medium | Medium | Complex |
| Target count | 5 | 11 | 8 | 13 |

## Using Examples as Templates

To use an example as a starting point for your project:

1. Copy the Makefile structure you need
2. Update the `!file` documentation with your project description
3. Modify targets and categories to match your workflow
4. Generate your help file:
   ```bash
   cd path/to/your/project
   make-help
   ```

## Common Patterns

### Minimal Setup (uncategorized-targets)

Start here if you just need basic help output without categories.

### Standard Project (categorized-project)

Use this pattern for most projects. Common categories include:
- **Core/Build**: Compilation and building
- **Test**: Testing and quality assurance
- **Development**: Local development and debugging
- **Deployment/Release**: Production deployment
- **Maintenance/Cleanup**: Housekeeping tasks

### Advanced Documentation (full-featured)

Use multi-line documentation when targets need detailed explanations, prerequisites, or step-by-step instructions.

## Tips

- **Keep it simple**: Start with basic documentation (uncategorized or categorized) and add complexity only when needed
- **Consistency matters**: Use consistent category names and documentation style across targets
- **Document variables**: Always use `!var` for environment variables users might need to set
- **Use aliases**: Add `!alias` for common shortcuts (e.g., `b` for `build`, `t` for `test`)
- **Test filtering**: Use the filtering-demo patterns to keep internal/utility targets hidden from regular users
- **Commit help files**: The generated help.mk files are self-contained and should be committed to your repository
- **Auto-regeneration**: Help files will automatically regenerate when source Makefiles change (when running `make help`)

Last reviewed: 2025-12-25T16:43Z
