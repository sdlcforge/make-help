# Research Report: Makefile Help Systems Comparison

## TLDR: When and Why Use make-help?

**Choose make-help if you need:**
- Comprehensive documentation: categories + variables + aliases + detailed per-target help
- Static help output with no runtime dependencies
- Auto-regeneration when Makefiles change
- Documentation linting and fixing

**Choose AWK patterns if you need:**
- Zero dependencies, basic target help only, immediate setup

**Quick decision:**
- Need categories + variables + aliases? → **make-help**
- Want remote Makefile sharing? → **mmake**
- Node.js project? → **makefile-help**
- Just basic help? → **AWK pattern**

---

## Executive summary

This report compares `make-help` with other tools and patterns specifically designed to document Makefiles and generate help output. All tools in this comparison work **with** Makefiles rather than replacing Make.

---

## 1. AWK-Based Self-Documenting Patterns (Inline)

The most common approach uses inline AWK/grep/sed to parse `##` comments.

### Simple pattern (most popular)
```makefile
help:
    @grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
        awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the project
    go build ./...
```

### Grouped pattern (with categories)
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

### How it works
```makefile
# Build the project for production
build:
    go build -ldflags ...

# - Internal target (hidden from help)
_internal:
    ...
```

Run `mmake help` to see formatted output. Comments starting with `-` are hidden.

### Unique features
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

### How it works
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

## 4. Other Tools (Niche)

Several other tools exist with lower adoption or specialized use cases:

- **[makefile-parser](https://github.com/kba/makefile-parser)** (npm): Full AST parser with programmatic API. Best for projects needing Makefile analysis, not just help generation.
- **[makehelp](https://github.com/ryanvolpe/makehelp)** (Shell): Uses `#:` syntax, has static output option. Very low adoption (2 stars).
- **[TangoMan Makefile Generator](https://github.com/TangoMan75/makefile-generator)** (Shell): Template-based generator for new Makefiles, not for documenting existing ones.

---

## 5. make-help (This Project)

**Language:** Go
**Install:** `npm install -g @sdlcforge/make-help` or `go install`

### How it works
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

### Unique features
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

## Feature comparison matrix

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

## Recommendations by use case

| Use Case | Recommended Tool |
|----------|------------------|
| "I want basic help with zero dependencies" | AWK Pattern |
| "I'm a Node.js project" | makefile-help |
| "I need to share Makefile snippets across repos" | mmake |
| "I have existing Makefiles and want comprehensive docs" | make-help |
| "I need categories + variables + aliases + detailed help" | make-help |
| "I want static output with no runtime dependency" | make-help |

---

## Competitive position of make-help

make-help is the only tool combining categories, variables, aliases, detailed help, static generation, and auto-regeneration. Unlike AWK patterns, it provides structured documentation; unlike mmake/makefile-help, it works without runtime wrappers.

**Trade-offs:** Requires installation (vs AWK's zero-dependency approach) and learning directive syntax (`!category`, `!var`).

---

## Sources

### Tools
- [tj/mmake](https://github.com/tj/mmake) - Modern Make wrapper
- [ianstormtaylor/makefile-help](https://github.com/ianstormtaylor/makefile-help) - npm package
- [kba/makefile-parser](https://github.com/kba/makefile-parser) - Full parser
- [ryanvolpe/makehelp](https://github.com/ryanvolpe/makehelp) - Shell-based with static mode
- [TangoMan75/makefile-generator](https://github.com/TangoMan75/makefile-generator) - Template generator

### Patterns & tutorials
- [Marmelab: Self-Documented Makefile](https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html)
- [Well-Documented Makefiles](https://www.thapaliya.com/en/writings/well-documented-makefiles/)
- [klmr's Self-documenting Makefile Gist](https://gist.github.com/klmr/575726c7e05d8780505a)
- [prwhite's Gist](https://gist.github.com/prwhite/8168133)
- [dwmkerr: Makefile Help Command](https://dwmkerr.com/makefile-help-command/)
- [FreeCodeCamp: Self-Documenting Makefile](https://www.freecodecamp.org/news/self-documenting-makefile/)
- [Jiby's Toolbox: Documenting Makefiles](https://jiby.tech/post/make-help-documenting-makefile/)

Last reviewed: 2025-12-17
