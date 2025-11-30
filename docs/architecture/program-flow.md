# Processing Pipeline

Step-by-step flows for each operation mode in make-help.

## Table of Contents

- [Help Generation Flow](#help-generation-flow)
- [Detailed Target View Flow](#detailed-target-view-flow)
- [Add-Target Flow](#add-target-flow)
- [Remove-Target Flow](#remove-target-flow)

---

## Overview

### 1 Help Generation Flow

```
1. CLI Parsing
   ├─> Parse flags and validate
   ├─> Resolve Makefile path (cwd/Makefile or --makefile-path)
   ├─> Detect color mode (terminal detection + flags)
   └─> Build Config object

2. Discovery Phase
   ├─> Discover Makefiles (MAKEFILE_LIST)
   │   ├─> Generate temporary Makefile with _list_makefiles target
   │   ├─> Execute: make -f <temp> _list_makefiles
   │   └─> Parse space-separated output -> []string
   └─> Discover Targets (make -p)
       ├─> Execute: make -f <makefile> -p -r
       └─> Parse database output -> []string

3. Parsing Phase
   ├─> For each Makefile in discovery order:
   │   ├─> Scan line-by-line
   │   ├─> Detect ## documentation lines
   │   ├─> Parse directives (@file, @category, @var, @alias)
   │   ├─> Detect target definitions (lines with :)
   │   └─> Associate pending docs with targets
   └─> Result: []*ParsedFile

4. Model Building Phase
   ├─> Aggregate directives from all files
   ├─> Group targets by category
   ├─> Validate categorization (no mixing unless --default-category)
   ├─> Associate aliases and variables with targets
   └─> Result: *HelpModel

5. Ordering Phase
   ├─> Apply category ordering
   │   ├─> If --category-order: explicit order + alphabetical remainder
   │   ├─> Else if --keep-order-categories: discovery order
   │   └─> Else: alphabetical
   └─> Apply target ordering (within each category)
       ├─> If --keep-order-targets: discovery order
       └─> Else: alphabetical

6. Summary Extraction Phase
   ├─> For each target:
   │   ├─> Join documentation lines
   │   ├─> Strip markdown headers
   │   ├─> Strip markdown formatting
   │   ├─> Strip HTML tags
   │   ├─> Normalize whitespace
   │   └─> Extract first sentence (regex)
   └─> Update target.Summary

7. Formatting Phase
   ├─> Initialize ColorScheme based on config.UseColor
   ├─> Render usage line
   ├─> Render file docs
   ├─> Render "Targets:" header
   ├─> For each category:
   │   ├─> Render category name (if not default)
   │   └─> For each target:
   │       ├─> Render target name + aliases
   │       ├─> Render summary
   │       └─> Render variables (if any)
   └─> Result: formatted string

8. Output
   └─> Write to STDOUT
```

### 2 Add-Target Flow

```
1. CLI Parsing
   ├─> Parse flags (including help generation flags)
   ├─> Validate --target-file if specified
   └─> Build Config object

2. Determine Target File Location
   ├─> If --target-file specified:
   │   └─> Use specified path, mark needsInclude=true
   ├─> Else if Makefile contains "include make/*.mk":
   │   ├─> Create make/ directory if needed
   │   └─> Set targetFile=make/01-help.mk, needsInclude=false
   └─> Else:
       └─> Set targetFile=<Makefile>, needsInclude=false (append)

3. Generate Help Target Content
   ├─> Build .PHONY: help line
   └─> Build help: target with make-help + flags

4. Write Target File
   ├─> If appending to Makefile:
   │   ├─> Read existing content
   │   ├─> Append help target
   │   └─> Write back
   └─> Else:
       └─> Write new file with help target

5. Add Include Directive (if needed)
   ├─> Compute relative path from Makefile to target file
   ├─> Generate include directive
   └─> Append to Makefile

6. Success
   └─> Print confirmation message
```

### 3 Remove-Target Flow

```
1. CLI Parsing
   ├─> Parse flags
   └─> Resolve Makefile path

2. Remove Include Directives
   ├─> Read Makefile
   ├─> Filter out lines matching: ^include\s+.*help.*\.mk
   └─> Write back

3. Remove Inline Help Target
   ├─> Read Makefile
   ├─> Detect help: target and .PHONY: help
   ├─> Skip target and its recipe lines (tab/space-prefixed)
   └─> Write back

4. Remove Help Target Files
   ├─> Check for make/01-help.mk
   ├─> Delete if exists
   └─> Check for other help-related .mk files in make/

5. Success
   └─> Print confirmation message
```

