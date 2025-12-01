# Testing Strategy

Comprehensive testing approach for make-help.

## Table of Contents

- [Unit Testing Approach](#unit-testing-approach)
- [Integration Testing Approach](#integration-testing-approach)
- [Mock Strategy](#mock-strategy)
- [Test Coverage Goals](#test-coverage-goals)

---

## Overview

### 1 Unit Testing Approach

**Parser Tests** (`internal/parser/scanner_test.go`)
```go
func TestScanFile(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected *ParsedFile
    }{
        {
            name: "file directive",
            input: `## @file
## This is file documentation
## Second line`,
            expected: &ParsedFile{
                Directives: []Directive{
                    {Type: DirectiveFile, Value: ""},
                    {Type: DirectiveDoc, Value: "This is file documentation"},
                    {Type: DirectiveDoc, Value: "Second line"},
                },
            },
        },
        {
            name: "category and target",
            input: `## @category Build
## Build the project
build:
	go build`,
            expected: &ParsedFile{
                Directives: []Directive{
                    {Type: DirectiveCategory, Value: "Build"},
                    {Type: DirectiveDoc, Value: "Build the project"},
                },
                TargetMap: map[string]int{"build": 3},
            },
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            scanner := NewScanner()
            result, err := scanner.ScanFile("test.mk")
            // Assertions...
        })
    }
}
```

**Summary Extractor Tests** (`internal/summary/extractor_test.go`)
```go
func TestExtract(t *testing.T) {
    tests := []struct {
        name     string
        docs     []string
        expected string
    }{
        {
            name:     "simple sentence",
            docs:     []string{"This is a test.", "More text."},
            expected: "This is a test.",
        },
        {
            name:     "ellipsis handling",
            docs:     []string{"Wait for it... then proceed.", "Done."},
            expected: "Wait for it... then proceed.",
        },
        {
            name:     "IP address handling",
            docs:     []string{"Connect to 127.0.0.1. Then test.", "More."},
            expected: "Connect to 127.0.0.1.",
        },
        {
            name:     "markdown stripping",
            docs:     []string{"**Bold** and *italic* text."},
            expected: "Bold and italic text.",
        },
        {
            name:     "no sentence terminator",
            docs:     []string{"No terminator"},
            expected: "No terminator",
        },
    }

    extractor := NewExtractor()
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := extractor.Extract(tt.docs)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

**Model Builder Tests** (`internal/model/builder_test.go`)
```go
func TestBuild(t *testing.T) {
    tests := []struct {
        name        string
        parsedFiles []*parser.ParsedFile
        config      *cli.Config
        expected    *HelpModel
        expectError bool
    }{
        {
            name: "mixed categorization without default",
            parsedFiles: []*parser.ParsedFile{
                {
                    Directives: []Directive{
                        {Type: DirectiveCategory, Value: "Build"},
                        {Type: DirectiveDoc, Value: "Build target"},
                    },
                    TargetMap: map[string]int{"build": 2},
                },
                {
                    Directives: []Directive{
                        {Type: DirectiveDoc, Value: "Test target"},
                    },
                    TargetMap: map[string]int{"test": 1},
                },
            },
            config:      &cli.Config{},
            expectError: true,
        },
        // More test cases...
    }
}
```

**Ordering Service Tests** (`internal/ordering/service_test.go`)
```go
func TestApplyOrdering(t *testing.T) {
    tests := []struct {
        name     string
        model    *HelpModel
        config   *cli.Config
        expected []string  // Expected category order
    }{
        {
            name: "alphabetical category order",
            model: &HelpModel{
                Categories: []Category{
                    {Name: "Zebra", DiscoveryOrder: 1},
                    {Name: "Alpha", DiscoveryOrder: 2},
                },
            },
            config:   &cli.Config{},
            expected: []string{"Alpha", "Zebra"},
        },
        {
            name: "discovery order preserved",
            model: &HelpModel{
                Categories: []Category{
                    {Name: "Zebra", DiscoveryOrder: 1},
                    {Name: "Alpha", DiscoveryOrder: 2},
                },
            },
            config:   &cli.Config{KeepOrderCategories: true},
            expected: []string{"Zebra", "Alpha"},
        },
        {
            name: "explicit category order",
            model: &HelpModel{
                Categories: []Category{
                    {Name: "Build", DiscoveryOrder: 1},
                    {Name: "Test", DiscoveryOrder: 2},
                    {Name: "Deploy", DiscoveryOrder: 3},
                },
            },
            config:   &cli.Config{CategoryOrder: []string{"Deploy", "Build"}},
            expected: []string{"Deploy", "Build", "Test"},  // Test appended alphabetically
        },
    }
}
```

### 2 Integration Testing Approach

**Fixture-Based Tests** (`test/integration/cli_test.go`)

```go
func TestHelpGeneration(t *testing.T) {
    tests := []struct {
        name         string
        fixture      string  // Path to test Makefile
        args         []string
        expectedFile string  // Path to expected output
    }{
        {
            name:         "basic help",
            fixture:      "fixtures/makefiles/basic.mk",
            args:         []string{},
            expectedFile: "fixtures/expected/basic_help.txt",
        },
        {
            name:         "categorized targets",
            fixture:      "fixtures/makefiles/categorized.mk",
            args:         []string{},
            expectedFile: "fixtures/expected/categorized_help.txt",
        },
        {
            name:         "explicit category order",
            fixture:      "fixtures/makefiles/categorized.mk",
            args:         []string{"--category-order", "Deploy,Build"},
            expectedFile: "fixtures/expected/categorized_ordered_help.txt",
        },
        {
            name:         "included files",
            fixture:      "fixtures/makefiles/with_includes.mk",
            args:         []string{},
            expectedFile: "fixtures/expected/with_includes_help.txt",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Execute CLI command
            cmd := exec.Command("make-help", append([]string{"--makefile-path", tt.fixture}, tt.args...)...)
            output, err := cmd.Output()
            require.NoError(t, err)

            // Read expected output
            expected, err := os.ReadFile(tt.expectedFile)
            require.NoError(t, err)

            // Compare (strip colors for comparison)
            assert.Equal(t, stripANSI(string(expected)), stripANSI(string(output)))
        })
    }
}

func TestAddTarget(t *testing.T) {
    tests := []struct {
        name         string
        fixture      string
        args         []string
        expectedMake string  // Path to expected Makefile after add
        expectedFile string  // Path to expected help target file (if separate)
    }{
        {
            name:         "append to makefile",
            fixture:      "fixtures/makefiles/empty.mk",
            args:         []string{},
            expectedMake: "fixtures/expected/empty_with_help.mk",
        },
        {
            name:         "create make/01-help.mk",
            fixture:      "fixtures/makefiles/with_make_include.mk",
            args:         []string{},
            expectedMake: "fixtures/expected/with_make_include_updated.mk",
            expectedFile: "fixtures/expected/01-help.mk",
        },
        {
            name:         "explicit target file",
            fixture:      "fixtures/makefiles/basic.mk",
            args:         []string{"--target-file", "custom-help.mk"},
            expectedMake: "fixtures/expected/basic_with_include.mk",
            expectedFile: "fixtures/expected/custom-help.mk",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Copy fixture to temp location
            tmpDir := t.TempDir()
            tmpMakefile := filepath.Join(tmpDir, "Makefile")
            copyFile(tt.fixture, tmpMakefile)

            // Execute --create-help-target
            cmd := exec.Command("make-help", append([]string{"--create-help-target", "--makefile-path", tmpMakefile}, tt.args...)...)
            err := cmd.Run()
            require.NoError(t, err)

            // Compare Makefile
            assertFileEquals(t, tt.expectedMake, tmpMakefile)

            // Compare help target file if separate
            if tt.expectedFile != "" {
                targetFile := filepath.Join(tmpDir, filepath.Base(tt.expectedFile))
                assertFileEquals(t, tt.expectedFile, targetFile)
            }
        })
    }
}
```

**Test Fixtures Structure:**
```
test/fixtures/
├── makefiles/
│   ├── basic.mk                   # Simple Makefile with targets
│   ├── categorized.mk             # Targets with @category
│   ├── with_includes.mk           # Makefile with includes
│   ├── mixed_categorization.mk    # Error case: mixed
│   └── empty.mk                   # Empty Makefile
└── expected/
    ├── basic_help.txt             # Expected help output
    ├── categorized_help.txt
    ├── with_includes_help.txt
    ├── empty_with_help.mk         # Expected Makefile after add
    └── 01-help.mk                 # Expected help target file
```

### 3 Mock Strategy

**CommandExecutor Mock** (for testing discovery without executing make)
```go
type MockCommandExecutor struct {
    outputs map[string]string  // Command -> stdout
    errors  map[string]error   // Command -> error
}

func (m *MockCommandExecutor) Execute(cmd string, args ...string) (string, string, error) {
    key := fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
    return m.outputs[key], "", m.errors[key]
}

// Usage in tests:
func TestDiscoverMakefiles(t *testing.T) {
    mock := &MockCommandExecutor{
        outputs: map[string]string{
            "make ... _list_makefiles": "Makefile include/common.mk",
        },
    }

    service := &discovery.Service{executor: mock}
    files, err := service.DiscoverMakefiles("Makefile")

    assert.NoError(t, err)
    assert.Equal(t, []string{"Makefile", "include/common.mk"}, files)
}
```

### 4 Test Coverage Goals

| Package | Coverage Target | Focus Areas |
|---------|----------------|-------------|
| `internal/parser` | 95% | All directive types, edge cases |
| `internal/summary` | 100% | All regex edge cases (ellipsis, IPs) |
| `internal/model` | 90% | Categorization validation, alias/var handling |
| `internal/ordering` | 95% | All ordering strategies |
| `internal/format` | 85% | Template rendering, color schemes |
| `internal/discovery` | 80% | Mock-based tests (not real make execution) |
| `internal/target` | 85% | File operations, pattern detection |
| Overall | 90% | Focus on critical paths |

