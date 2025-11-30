# Key Algorithms

Core algorithms and their implementations in make-help.

## Table of Contents

- [File Discovery via MAKEFILE_LIST](#file-discovery-via-makefile_list)
- [Documentation Parsing](#documentation-parsing)
- [Ordering Logic](#ordering-logic)
- [Summary Extraction](#summary-extraction)

---

## Overview

### 1 File Discovery via MAKEFILE_LIST

**Algorithm:**
```
Input: makefilePath (path to main Makefile)
Output: []string (ordered list of Makefile paths)

1. Create temporary Makefile content:
   - Cat main Makefile content
   - Append blank line
   - Append target:
     .PHONY: _list_makefiles
     _list_makefiles:
         @echo $(MAKEFILE_LIST)

2. Execute shell command:
   make -f <(cat Makefile && echo && echo -e '<target>') _list_makefiles

3. Parse stdout:
   - Split on whitespace
   - Each token is a Makefile path

4. Resolve to absolute paths:
   - For each path:
     - If relative, resolve from Makefile directory
     - Return absolute path

5. Return ordered list
```

**Important:** Included files appear in MAKEFILE_LIST after their parent file completes, not at the include point. This matches Make's processing order.

**Error Handling:**
- Shell command failure -> wrap error with context
- Empty output -> error "no Makefiles found"
- Invalid paths -> error "Makefile not found: <path>"

### 2 Documentation Parsing and Directive Handling

**Algorithm:**
```
Input: fileContent (string), targetMap (map[string]int)
Output: []Directive

State:
- currentCategory: string (current category name)
- pendingDocs: []Directive (docs awaiting target association)

For each line:
  1. If line starts with "## ":
     - Parse directive (detect @file, @category, @var, @alias, or doc)
     - If @category: update currentCategory
     - If @file: add to FileDocs immediately
     - Else: add to pendingDocs

  2. Else if line is target definition (contains : or &:):
     - Extract target name
     - Associate pendingDocs with target
     - Clear pendingDocs

  3. Else:
     - Clear pendingDocs (non-doc line breaks association)

Return all directives
```

**Target Name Extraction:**
```
1. Find first : in line
2. Extract everything before :
3. If ends with &, remove it (grouped target)
4. Split on whitespace, take first token
5. Return token as target name
```

**Edge Cases:**
- Grouped targets (foo bar baz:) -> extract "foo"
- Variable targets ($(VAR):) -> extract "$(VAR)"
- Comments -> skip
- Indented lines -> skip (recipe lines)

### 3 Category/Target Ordering Logic

**Algorithm:**
```
Input: HelpModel, Config
Output: HelpModel (with ordered categories and targets)

Category Ordering:
  If --category-order specified:
    1. Create map of category name -> Category
    2. Build ordered list:
       - For each name in --category-order:
         - Add corresponding Category
         - Remove from map
    3. Sort remaining categories alphabetically
    4. Append to ordered list
    5. Validate: error if any --category-order name not found

  Else if --keep-order-categories:
    Sort categories by DiscoveryOrder field

  Else:
    Sort categories alphabetically by Name

Target Ordering (within each category):
  If --keep-order-targets:
    Sort targets by DiscoveryOrder field

  Else:
    Sort targets alphabetically by Name
```

**Discovery Order Tracking:**
- Global counter incremented for each category/target first appearance
- Stored in DiscoveryOrder field
- Split categories use DiscoveryOrder of first appearance

### 4 Summary Extraction Algorithm (Port of extract-topic)

**Algorithm:**
```
Input: documentation []string (full target docs)
Output: summary string (first sentence)

1. Join documentation lines with space
   text = strings.Join(documentation, " ")

2. Strip markdown headers
   Remove all lines matching ^#+\s+

3. Strip markdown formatting
   - Remove **bold** -> bold
   - Remove *italic* -> italic
   - Remove `code` -> code
   - Remove [text](url) -> text

4. Strip HTML tags
   Remove all <tag> patterns

5. Normalize whitespace
   - Replace \n with space
   - Collapse multiple spaces -> single space
   - Trim leading/trailing whitespace

6. Extract first sentence
   Regex: ^((?:[^.!?]|\.\.\.|\.[^\s])+[.?!])(\s|$)

   Explanation:
   - (?:[^.!?]|\.\.\.|\.[^\s])+ : Match anything except .!? OR ... OR .<non-space>
   - [.?!] : Match sentence terminator
   - (\s|$) : Must be followed by whitespace or end-of-string

   Edge Cases:
   - "..." (ellipsis) -> not sentence boundary
   - "127.0.0.1." (IP) -> not sentence boundary (. followed by digit)
   - "This is it." -> sentence boundary (. followed by space/EOL)

7. Return matched sentence or full text if no match
```

**Test Cases:**
```
Input: "Build the project. Run tests."
Output: "Build the project."

Input: "Supports IPv4 addresses like 127.0.0.1. Cool!"
Output: "Supports IPv4 addresses like 127.0.0.1."

Input: "Wait for it... then proceed. Done."
Output: "Wait for it... then proceed."

Input: "**Bold text** and *italic* formatting"
Output: "Bold text and italic formatting"

Input: "No sentence terminator here"
Output: "No sentence terminator here"
```

