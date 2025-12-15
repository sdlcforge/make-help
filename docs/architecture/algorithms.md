# Key Algorithms

Core algorithms and their implementations in make-help.

## Table of Contents

- [File Discovery via MAKEFILE_LIST](#file-discovery-via-makefile_list)
- [Documentation Parsing](#documentation-parsing)
- [Ordering Logic](#ordering-logic)
- [Summary Extraction](#summary-extraction)
- [Include Pattern Detection](#include-pattern-detection)
- [Numbered Prefix Determination](#numbered-prefix-determination)

---

## Overview

### 1 File Discovery via MAKEFILE_LIST

**Algorithm:**
```
Input: makefilePath (path to main Makefile)
Output: []string (ordered list of Makefile paths)

1. Read main Makefile content into memory

2. Create temporary physical file in same directory as main Makefile:
   - Use os.CreateTemp() to create temp file with pattern ".makefile-discovery-*.mk"
   - Write main Makefile content to temp file
   - Append discovery target:
     .PHONY: _list_makefiles
     _list_makefiles:
         @echo $(MAKEFILE_LIST)
   - Close temp file
   - Ensure cleanup with defer os.Remove()

3. Execute make command with 30-second timeout:
   make -s --no-print-directory -f <temp-file-path> _list_makefiles

   Flags explained:
   - "-s" (silent): Suppress make's own output
   - "--no-print-directory": Prevent directory change messages
   - These prevent output corruption when running within another make

4. Parse stdout:
   - Split on whitespace
   - Each token is a Makefile path
   - First file will be temp file - replace with original Makefile path

5. Resolve to absolute paths:
   - For each path:
     - If relative, resolve from Makefile directory
     - Validate file exists with os.Stat()
     - Return absolute path

6. Return ordered list
```

**Security Note:** This implementation uses temporary physical files instead of bash process
substitution (e.g., `<(...)`) to prevent command injection vulnerabilities. The temp file
is created in the same directory as the Makefile to ensure relative includes work correctly.

**Important:** Included files appear in MAKEFILE_LIST after their parent file completes, not at the include point. This matches Make's processing order.

**Error Handling:**
- File read failure -> "failed to read Makefile: <error>"
- Temp file creation failure -> "failed to create temp file: <error>"
- Make command timeout (30s) -> "make command timed out after 30s"
- Make command failure -> "failed to discover makefiles: <error>"
- Empty output -> "no Makefiles found in MAKEFILE_LIST"
- Invalid paths -> "Makefile not found: <path>"

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
     - Parse directive (detect !file, !category, !var, !alias, or doc)
     - If !category: update currentCategory
     - If !file: add to FileDocs immediately
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

### 5 Include Pattern Detection

**Algorithm:**
```
Input: content []byte (Makefile content)
Output: *IncludePattern (pattern details) or nil

1. Define regex to match include directives for make/* pattern:
   Pattern: (?m)^-?include\s+(?:\$\([^)]+\))?(\./)?make/\*(\.[a-zA-Z0-9]+)?(?:\s|$)

   Matches:
   - "include make/*.mk"
   - "-include make/*.mk"
   - "include ./make/*.mk"
   - "-include $(dir ...)make/*.mk"

   Capture groups:
   - Group 1: Optional ./ prefix
   - Group 2: File extension suffix (e.g., .mk)

2. Execute regex on Makefile content
   matches := includeRegex.FindSubmatch(content)

3. If no match found:
   return nil

4. Extract suffix from capture group 2:
   suffix := ""
   if len(matches) > 2 && len(matches[2]) > 0:
       suffix = string(matches[2])

5. Determine pattern prefix from capture group 1:
   patternPrefix := "make/"
   if len(matches) > 1 && len(matches[1]) > 0:
       patternPrefix = "./make/"

6. Return IncludePattern:
   return &IncludePattern{
       Suffix:        suffix,
       FullPattern:   string(matches[0]),
       PatternPrefix: patternPrefix,
   }
```

**Examples:**
```
Input: "include make/*.mk"
Output: &IncludePattern{
    Suffix:        ".mk",
    FullPattern:   "include make/*.mk",
    PatternPrefix: "make/",
}

Input: "-include ./make/*"
Output: &IncludePattern{
    Suffix:        "",
    FullPattern:   "-include ./make/*",
    PatternPrefix: "./make/",
}

Input: "include some/other/*.mk"
Output: nil
```

**Usage:**
This algorithm is used during help file generation to:
1. Determine if a make/ directory pattern already exists
2. Extract the file extension convention (.mk vs no extension)
3. Decide whether to add a new include directive

### 6 Numbered Prefix Determination

**Algorithm:**
```
Input:
  - makeDir string (absolute path to make/ directory)
  - suffix string (file extension, e.g., ".mk")
  - pattern *IncludePattern (detected include pattern, may be nil)
Output: prefix string (e.g., "00-", "000-", or "")

1. Try to read directory entries:
   entries, err := os.ReadDir(makeDir)
   if err != nil:
       return "" // Directory doesn't exist or can't be read

2. Build regex to match numbered files with same suffix:
   numberedFileRegex := `^(\d+)-.*` + regexp.QuoteMeta(suffix) + `$`

   Examples:
   - For suffix ".mk": matches "01-foo.mk", "10-bar.mk", "100-baz.mk"
   - For suffix "": matches "01-foo", "10-bar"

3. Scan directory entries to find max digit count:
   maxDigits := 0
   for _, entry := range entries:
       if entry.IsDir():
           continue

       matches := numberedFileRegex.FindStringSubmatch(entry.Name())
       if matches != nil:
           digitCount := len(matches[1])
           if digitCount > maxDigits:
               maxDigits = digitCount

4. If no numbered files found:
   return ""

5. Generate zero-padded prefix with matching digit count:
   zeros := ""
   for i := 0; i < maxDigits; i++:
       zeros += "0"
   return zeros + "-"
```

**Examples:**
```
Input:
  makeDir="/path/to/make"
  suffix=".mk"
  Directory contains: ["10-constants.mk", "20-utils.mk"]
Output: "00-"

Input:
  makeDir="/path/to/make"
  suffix=".mk"
  Directory contains: ["100-constants.mk", "200-utils.mk"]
Output: "000-"

Input:
  makeDir="/path/to/make"
  suffix=".mk"
  Directory contains: ["foo.mk", "bar.mk"]
Output: ""

Input:
  makeDir="/path/to/make"
  suffix=".mk"
  Directory contains: ["10-foo.mk", "bar.mk"]
Output: "00-" (only numbered files are considered)
```

**Usage:**
This algorithm ensures the generated help file follows existing numbering conventions:
- If files use "10-", "20-" format, generates "00-help.mk"
- If files use "100-", "200-" format, generates "000-help.mk"
- If no numbered files exist, generates "help.mk" without prefix

**Integration:**
Combined with include pattern detection, the full file location strategy is:
1. Use explicit --help-file-rel-path if provided
2. Otherwise, default to make/ directory:
   - Detect include pattern to determine suffix
   - Detect numbered prefix from existing files
   - Generate filename: `{prefix}help{suffix}`
   - Example: "00-help.mk" or "help.mk"
3. Add include directive if no pattern exists

