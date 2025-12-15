# Research Report: Makefile Help Systems Comparison

## Executive Summary

This report compares `make-help` with other tools and patterns specifically designed to document Makefiles and generate help output. All tools in this comparison work **with** Makefiles rather than replacing Make.

---

## 1. AWK-Based Self-Documenting Patterns (Inline)

The most common approach uses inline AWK/grep/sed to parse `##` comments.

### Simple Pattern (Most Popular)
```makefile
help:
    @grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
        awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the project
    go build ./...
```

### Grouped Pattern (with Categories)
```makefile
help:
    @awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

##@ Build
build: ## Build the project
    go build ./...

##@ Test
test: ## Run tests
    go test ./...
```

| Strengths | Weaknesses |
|-----------|------------|
| Zero dependencies | Hard to read/maintain AWK one-liners |
| Self-contained in Makefile | No variable documentation |
| Works everywhere | No aliases |
| Instant setup | No per-target detailed help |
| Categories via `##@` | Alphabetical sort or none |
| | No static file generation |

**Workflow Fit:** Quick setup for simple projects. Copy-paste a pattern and go.

**Sources:**
- [Marmelab: Self-Documented Makefile](https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html)
- [Well-Documented Makefiles](https://www.thapaliya.com/en/writings/well-documented-makefiles/)
- [klmr's Self-documenting Makefile Gist](https://gist.github.com/klmr/575726c7e05d8780505a)
- [prwhite's Gist](https://gist.github.com/prwhite/8168133)

---

## 2. mmake (Modern Make) by TJ Holowaychuk

**GitHub:** [tj/mmake](https://github.com/tj/mmake) (1.7k stars)
**Language:** Go
**Install:** `brew install tj/mmake/mmake` or `go install`

A wrapper around `make` that adds features. Requires aliasing: `alias make=mmake`

### How It Works
```makefile
# Build the project for production
build:
    go build -ldflags ...

# - Internal target (hidden from help)
_internal:
    ...
```

Run `mmake help` to see formatted output. Comments starting with `-` are hidden.

### Unique Features
- **Remote includes**: Include Makefiles from GitHub URLs
- **Glob patterns**: Filter help with `mmake help 'build*'`
- **Verbose mode**: `mmake help -v`

| Strengths | Weaknesses |
|-----------|------------|
| Simple `# comment` syntax | Requires wrapping all make calls |
| Remote include support | Last release 2019 (1.4.0) |
| Glob-based help filtering | No categories |
| Pass-through to make | No variable docs |
| | No per-target detailed help |
| | Help from includes can be buggy |

**Workflow Fit:** Teams willing to alias make→mmake who want remote includes.

---

## 3. makefile-help (npm package) by ianstormtaylor

**GitHub:** [ianstormtaylor/makefile-help](https://github.com/ianstormtaylor/makefile-help)
**npm:** `makefile-help`
**Language:** Makefile

### How It Works
Install via npm, then include in your Makefile:
```makefile
ifneq ($(wildcard ./node_modules),)
  include ./node_modules/makefile-help/Makefile
endif

# Remove derived files.
clean:
    rm -rf build

# Enable debug mode. (default: false)
DEBUG ?= false
```

Comments above targets and `?=` variables are automatically extracted.

| Strengths | Weaknesses |
|-----------|------------|
| Simple `# comment` syntax | Node.js/npm dependency |
| Auto-documents variables | No categories |
| No AWK knowledge needed | No aliases |
| | No detailed per-target help |
| | Only for npm projects |

**Workflow Fit:** Node.js projects wanting simple help with minimal setup.

---

## 4. makefile-parser by kba

**GitHub:** [kba/makefile-parser](https://github.com/kba/makefile-parser)
**npm:** `@kba/makefile-parser`
**Language:** JavaScript

A full Makefile parser that can generate help output or AST.

### How It Works
```bash
# CLI usage
makefile-parser --make-help Makefile

# Or embed in Makefile
# BEGIN-EVAL makefile-parser --make-help Makefile
# END-EVAL
```

Uses single-line comments above targets/variables.

| Strengths | Weaknesses |
|-----------|------------|
| Full AST parsing | JavaScript/npm dependency |
| Programmatic API | Complex for simple use |
| Variables documented | No categories |
| Integrates with shinclude | Less popular |

**Workflow Fit:** Projects needing programmatic Makefile analysis.

---

## 5. makehelp by ryanvolpe

**GitHub:** [ryanvolpe/makehelp](https://github.com/ryanvolpe/makehelp) (2 stars)
**Language:** Shell

### How It Works
Uses `#:` prefix for documentation comments:
```makefile
#: Build the application.
#: This compiles all sources.
build:
    go build ./...
```

Supports text formatting: `*bold*`, `_underline_`, `~inverse~`

### Unique Features
- **Static mode**: `--static` embeds help directly in Makefile
- **Text formatting** in docs
- **Header vs target docs**: Unassociated `#:` blocks become header text

| Strengths | Weaknesses |
|-----------|------------|
| Static output option | Very low adoption |
| Text formatting | Shell script dependency |
| Header documentation | Different syntax (`#:`) |
| | No categories |
| | No variable docs |

**Workflow Fit:** Niche use case, similar concept to make-help's static generation.

---

## 6. TangoMan Makefile Generator

**GitHub:** [TangoMan75/makefile-generator](https://github.com/TangoMan75/makefile-generator)
**Language:** Shell

A template-based generator that creates Makefiles with embedded help.

### How It Works
- Define templates in `makefiles/`, `vars/`, `header/` directories
- Configure via `config.yaml`
- Run generator to produce Makefile with AWK-based help

### Syntax
```makefile
### Category Name
## Description of target
target:
    command
```

| Strengths | Weaknesses |
|-----------|------------|
| Generates full Makefile | Not for existing Makefiles |
| Categories via `###` | Template-based workflow |
| README generation | Requires GAWK, SED |
| License file generation | Overkill for simple needs |

**Workflow Fit:** Starting new projects with standardized Makefile structure.

---

## 7. make-help (This Project)

**Language:** Go
**Install:** `npm install -g @sdlcforge/make-help` or `go install`

### How It Works
```makefile
## !file
## My Project Build System

## !category Build
## !var CC C compiler to use
## !alias b
## Build the project.
## Compiles and links the binary.
build:
    $(CC) -o app main.c
```

Run `make-help` to generate `./make/help.mk` with static help.

### Unique Features
- **Static generation**: Help embedded as `@echo` statements
- **Auto-regeneration**: Help file rebuilds when Makefiles change
- **Rich syntax**: `!file`, `!category`, `!var`, `!alias` directives
- **Detailed help**: `make help-<target>` for full documentation
- **Summary extraction**: Intelligent first-sentence extraction
- **Smart file placement**: Auto-creates `./make/`, handles numbered prefixes
- **Include directive insertion**: Auto-adds `-include make/*.mk`
- **Fallback chain**: `make-help` → `npx make-help` → error

| Strengths | Weaknesses |
|-----------|------------|
| Static output (no runtime dep) | Requires installation |
| Rich documentation syntax | Learning curve for directives |
| Categories, aliases, variables | Adds file to project |
| Per-target detailed help | Custom syntax vs `##` |
| Auto-regeneration | |
| Works with existing Makefiles | |
| Lint/fix capabilities | |

**Workflow Fit:** Teams with existing Makefiles wanting professional documentation.

---

## Feature Comparison Matrix

| Feature | AWK Pattern | mmake | makefile-help | makefile-parser | makehelp | TangoMan | make-help |
|---------|-------------|-------|---------------|-----------------|----------|----------|-----------|
| **Zero dependencies** | Yes | No (Go) | No (npm) | No (npm) | No (sh) | No (awk/sed) | No (Go/npm) |
| **Categories** | `##@` | No | No | No | No | `###` | `!category` |
| **Variable docs** | No | No | `?=` vars | Yes | No | No | `!var` |
| **Aliases** | No | No | No | No | No | No | `!alias` |
| **Per-target detail** | No | `-v` flag | No | No | No | No | `help-<target>` |
| **Static output** | No | No | No | No | `--static` | Generated | Default |
| **Auto-regeneration** | N/A | N/A | N/A | N/A | No | N/A | Yes |
| **Summary extraction** | No | No | No | No | No | No | Yes |
| **Remote includes** | No | Yes | No | No | No | No | No |
| **File-level docs** | No | No | No | No | Header | No | `!file` |
| **Lint/fix** | No | No | No | No | No | No | Yes |
| **Comment syntax** | `##` | `#` | `#` | `#` | `#:` | `##` | `##` |

---

## Workflow Integration Comparison

### Development Experience

| Tool | Setup Effort | Maintenance | Editor Support |
|------|--------------|-------------|----------------|
| AWK Pattern | Copy-paste | Edit AWK | None |
| mmake | Install + alias | None | None |
| makefile-help | npm install | None | None |
| make-help | Install + run | Auto-regen | VSCode (planned) |

### CI/CD Integration

| Tool | CI Dependencies | Reproducibility |
|------|-----------------|-----------------|
| AWK Pattern | None | Excellent |
| mmake | Go binary or brew | Good |
| makefile-help | npm | Good |
| make-help | Go/npm OR none (static) | Excellent |

---

## Strengths & Weaknesses Summary

### AWK Patterns
- **Best for:** Simple projects, minimal setup, no dependencies
- **Avoid when:** Need categories + variables + aliases together

### mmake
- **Best for:** Teams wanting remote include sharing
- **Avoid when:** Can't alias make, need active maintenance

### makefile-help
- **Best for:** Node.js projects, simple variable docs
- **Avoid when:** Not using npm, need categories

### make-help
- **Best for:** Comprehensive docs, existing Makefiles, static output
- **Avoid when:** Want zero new tools, prefer `##` inline syntax

---

## Recommendations by Use Case

| Use Case | Recommended Tool |
|----------|------------------|
| "I want basic help with zero dependencies" | AWK Pattern |
| "I'm a Node.js project" | makefile-help |
| "I need to share Makefile snippets across repos" | mmake |
| "I have existing Makefiles and want comprehensive docs" | make-help |
| "I need categories + variables + aliases + detailed help" | make-help |
| "I want static output with no runtime dependency" | make-help |

---

## Competitive Position of make-help

### Unique to make-help
1. **Only tool with all four**: categories + variables + aliases + detailed help
2. **Static generation by default**: No runtime dependency for `make help`
3. **Auto-regeneration**: Help stays in sync automatically
4. **Summary extraction**: Intelligent first-sentence parsing
5. **Lint/fix**: Documentation quality checking
6. **Smart defaults**: Directory creation, numbered prefixes, include insertion

### Potential Gaps to Consider
Based on this research, features other tools have that make-help doesn't:
- **Remote includes** (mmake): Could be interesting for shared team standards
- **Simpler `#` comment syntax** (mmake, makefile-help): Lower barrier to entry

---

## Sources

### Tools
- [tj/mmake](https://github.com/tj/mmake) - Modern Make wrapper
- [ianstormtaylor/makefile-help](https://github.com/ianstormtaylor/makefile-help) - npm package
- [kba/makefile-parser](https://github.com/kba/makefile-parser) - Full parser
- [ryanvolpe/makehelp](https://github.com/ryanvolpe/makehelp) - Shell-based with static mode
- [TangoMan75/makefile-generator](https://github.com/TangoMan75/makefile-generator) - Template generator

### Patterns & Tutorials
- [Marmelab: Self-Documented Makefile](https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html)
- [Well-Documented Makefiles](https://www.thapaliya.com/en/writings/well-documented-makefiles/)
- [klmr's Self-documenting Makefile Gist](https://gist.github.com/klmr/575726c7e05d8780505a)
- [prwhite's Gist](https://gist.github.com/prwhite/8168133)
- [dwmkerr: Makefile Help Command](https://dwmkerr.com/makefile-help-command/)
- [FreeCodeCamp: Self-Documenting Makefile](https://www.freecodecamp.org/news/self-documenting-makefile/)
- [Jiby's Toolbox: Documenting Makefiles](https://jiby.tech/post/make-help-documenting-makefile/)

Last reviewed: 2025-12-25T16:43Z
