# Error Handling

Error classification, types, and handling strategies in make-help.

## Table of contents

- [Error Classification](#error-classification)
- [Error Types](#error-types)
- [Error Scenarios](#error-scenarios)

---

## Overview

### 1 Error Classification

| Error Type | Priority | Behavior |
|------------|----------|----------|
| Makefile not found | CRITICAL | Exit with error message |
| Mixed categorization without --default-category | CRITICAL | Exit with clear error and suggestion |
| Unknown category in --category-order | CRITICAL | Exit with list of available categories |
| Make command execution failure | CRITICAL | Exit with stderr output |
| Invalid directive syntax | WARNING | Log warning, skip directive, continue |
| Malformed !var or !alias | WARNING | Log warning, best-effort parse |
| Duplicate help target | WARNING | Ask user to remove with --remove-help-target first |
| File write failure | CRITICAL | Exit with error message |

### 2 Error types and messages

```go
package errors

type MixedCategorizationError struct {
    Message string
}

func (e *MixedCategorizationError) Error() string {
    return fmt.Sprintf("mixed categorization: %s\nUse --default-category to assign uncategorized targets to a default category", e.Message)
}

type UnknownCategoryError struct {
    CategoryName string
    Available    []string
}

func (e *UnknownCategoryError) Error() string {
    return fmt.Sprintf("unknown category %q in --category-order\nAvailable categories: %s",
        e.CategoryName, strings.Join(e.Available, ", "))
}

type MakefileNotFoundError struct {
    Path string
}

func (e *MakefileNotFoundError) Error() string {
    return fmt.Sprintf("Makefile not found: %s\nUse --makefile-path to specify location", e.Path)
}

type MakeExecutionError struct {
    Command string
    Stderr  string
}

func (e *MakeExecutionError) Error() string {
    return fmt.Sprintf("make command failed: %s\n%s", e.Command, e.Stderr)
}
```

### 3 Error scenarios and handling

**Scenario 1: Mixed Categorization**
```
Problem: Some targets have !category, others don't
Detection: Model validator counts categorized vs uncategorized
Action:
  - If --default-category set: assign uncategorized to default
  - Else: return MixedCategorizationError
```

**Scenario 2: Unknown Category in --category-order**
```
Problem: User specifies category that doesn't exist
Detection: Ordering service validates all names exist
Action: Return UnknownCategoryError with available categories
```

**Scenario 3: File Not Found**
```
Problem: Makefile doesn't exist at specified path
Detection: os.Stat fails in discovery service
Action: Return MakefileNotFoundError
```

**Scenario 4: Invalid Directive Syntax**
```
Problem: Malformed !var or !alias directive
Detection: Parser fails to split on expected delimiter
Action: Log warning, skip directive, continue parsing
Example: "!var NODELIM" -> log "invalid !var directive at line X: missing ' - '"
```

**Scenario 5: Make Command Failure**
```
Problem: make -p or make _list_makefiles fails
Detection: Non-zero exit code from exec.Command
Action: Return MakeExecutionError with stderr
```

**Scenario 6: Duplicate Help Target**
```
Problem: help target already exists when running --create-help-target
Detection: Check for existing help: in Makefile
Action:
  - Return error asking user to run --remove-help-target first
```


Last reviewed: 2025-12-25T16:43Z
