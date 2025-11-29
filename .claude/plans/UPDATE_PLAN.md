# Update Plan: Simplify CLI and Enhance Help Target Generation

## Overview

This update refactors make-help to:
1. Replace subcommands (`add-target`, `remove-target`) with flags (`--create-help-target`, `--remove-help-target`)
2. Generate project-local binary installation with proper `.PHONY` handling
3. Generate `help-<target>` targets for detailed per-target help
4. Only show documented targets by default, with `--include-target` and `--include-all-phony` options

## Stage Execution Order

```
Stage 1: Config & Flag Infrastructure
    │
    ├──► Stage 2: Target Filtering (parallel with Stage 3)
    │
    └──► Stage 3: Detailed Help (parallel with Stage 2)
          │
          ▼
      Stage 4: Help File Generator
          │
          ▼
      Stage 5: Create/Remove Help Target Commands
          │
          ▼
      Stage 6: Cleanup & Documentation
```

## Parallel Execution

- **Stages 2 and 3** can be done in parallel (no dependencies between them)
- All other stages must be sequential

## Stage Summary

| Stage | Name | Dependencies | Estimated Files |
|-------|------|--------------|-----------------|
| 1 | Config & Flag Infrastructure | None | 3 |
| 2 | Target Filtering | Stage 1 | 4 |
| 3 | Detailed Help | Stage 1 | 2 |
| 4 | Help File Generator | Stages 2, 3 | 2 |
| 5 | Create/Remove Commands | Stage 4 | 4 |
| 6 | Cleanup & Documentation | Stage 5 | 5 |

## Documentation Updates Required

- `README.md` - Update CLI usage, new flags, examples
- `.claude/CLAUDE.md` - Update build commands if needed
- `docs/design.md` - Update architecture, data flow, CLI structure

## Detailed Stage Plans

- [TODO-01.md](TODO-01.md) - Stage 1: Config & Flag Infrastructure
- [TODO-02.md](TODO-02.md) - Stage 2: Target Filtering
- [TODO-03.md](TODO-03.md) - Stage 3: Detailed Help
- [TODO-04.md](TODO-04.md) - Stage 4: Help File Generator
- [TODO-05.md](TODO-05.md) - Stage 5: Create/Remove Help Target Commands
- [TODO-06.md](TODO-06.md) - Stage 6: Cleanup & Documentation
