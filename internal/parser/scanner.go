package parser

import (
	"fmt"
	"os"
	"strings"
)

// Scanner scans Makefile content and extracts documentation directives.
// It maintains state to track the current category and pending documentation
// that will be associated with the next target.
type Scanner struct {
	currentFile     string      // Current file being scanned
	currentCategory string      // Current category from !category directive
	pendingDocs     []Directive // Documentation lines awaiting target association
}

// NewScanner creates a new Scanner instance.
func NewScanner() *Scanner {
	return &Scanner{
		pendingDocs: []Directive{},
	}
}

// ScanFile parses a single Makefile and extracts directives.
// It reads the file, scans line-by-line, and returns a ParsedFile
// containing all directives and target information.
func (s *Scanner) ScanFile(path string) (*ParsedFile, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	return s.ScanContent(string(content), path)
}

// ScanContent parses Makefile content and extracts directives.
// This method is useful for testing with in-memory content.
// The path parameter is used for error reporting and tracking source files.
func (s *Scanner) ScanContent(content string, path string) (*ParsedFile, error) {
	// Reset scanner state
	s.currentFile = path
	s.currentCategory = ""
	s.pendingDocs = []Directive{}

	result := &ParsedFile{
		Path:       path,
		Directives: []Directive{},
		TargetMap:  make(map[string]int),
	}

	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		lineNumber := lineNum + 1 // 1-based line numbers

		// Check for documentation line
		if IsDocumentationLine(line) {
			directive := s.parseDirective(line, lineNumber)

			// !file directives are added immediately and not queued
			if directive.Type == DirectiveFile {
				result.Directives = append(result.Directives, directive)
			} else {
				// Queue for association with next target
				s.pendingDocs = append(s.pendingDocs, directive)
			}
			continue
		}

		// Check for target definition
		if IsTargetLine(line) {
			targetName := ExtractTargetName(line)
			if targetName != "" {
				result.TargetMap[targetName] = lineNumber

				// Associate pending docs with this target
				if len(s.pendingDocs) > 0 {
					result.Directives = append(result.Directives, s.pendingDocs...)
					s.pendingDocs = []Directive{}
				}
				continue
			}
		}

		// Non-doc, non-target line clears pending docs
		// (breaks the association between docs and the next target)
		if len(s.pendingDocs) > 0 {
			s.pendingDocs = []Directive{}
		}
	}

	return result, nil
}

// parseDirective detects and parses a documentation directive.
// It identifies the directive type (!file, !category, !var, !alias, or regular doc)
// and extracts the directive value.
func (s *Scanner) parseDirective(line string, lineNum int) Directive {
	// Remove the "## " prefix, or handle bare "##" for empty doc lines
	var content string
	if line == "##" {
		content = ""
	} else {
		content = strings.TrimPrefix(line, "## ")
	}

	directive := Directive{
		SourceFile: s.currentFile,
		LineNumber: lineNum,
	}

	// Detect directive type and extract value
	switch {
	case strings.HasPrefix(content, "!file"):
		directive.Type = DirectiveFile
		// Extract value after "!file" (could be empty or have description)
		value := strings.TrimPrefix(content, "!file")
		directive.Value = strings.TrimSpace(value)

	case strings.HasPrefix(content, "!category "):
		directive.Type = DirectiveCategory
		directive.Value = strings.TrimSpace(strings.TrimPrefix(content, "!category "))
		// Update current category state
		s.currentCategory = directive.Value

	case strings.HasPrefix(content, "!var "):
		directive.Type = DirectiveVar
		directive.Value = strings.TrimSpace(strings.TrimPrefix(content, "!var "))

	case strings.HasPrefix(content, "!alias "):
		directive.Type = DirectiveAlias
		directive.Value = strings.TrimSpace(strings.TrimPrefix(content, "!alias "))

	default:
		// Regular documentation line
		directive.Type = DirectiveDoc
		directive.Value = content
	}

	return directive
}
