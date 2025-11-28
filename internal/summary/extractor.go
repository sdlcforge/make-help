package summary

import (
	"regexp"
	"strings"
)

// Extractor pre-compiles all regex patterns at construction time for performance.
// This avoids repeated regex compilation when processing many targets.
type Extractor struct {
	sentenceRegex    *regexp.Regexp
	headerRegex      *regexp.Regexp
	boldRegex        *regexp.Regexp
	italicRegex      *regexp.Regexp
	boldUnderRegex   *regexp.Regexp
	italicUnderRegex *regexp.Regexp
	codeRegex        *regexp.Regexp
	linkRegex        *regexp.Regexp
	htmlTagRegex     *regexp.Regexp
	whitespaceRegex  *regexp.Regexp
}

// NewExtractor creates an Extractor with all regex patterns pre-compiled.
func NewExtractor() *Extractor {
	return &Extractor{
		// Regex from extract-topic: first sentence ending in .!?
		// Handles: ellipsis (...), IPs (127.0.0.1.), abbreviations
		sentenceRegex:    regexp.MustCompile(`^((?:[^.!?]|\.\.\.|\.[^\s])+[.?!])(\s|$)`),
		headerRegex:      regexp.MustCompile(`(?m)^#+\s+`),
		boldRegex:        regexp.MustCompile(`\*\*([^*]+)\*\*`),
		italicRegex:      regexp.MustCompile(`\*([^*]+)\*`),
		boldUnderRegex:   regexp.MustCompile(`__([^_]+)__`),
		italicUnderRegex: regexp.MustCompile(`_([^_]+)_`),
		codeRegex:        regexp.MustCompile("`([^`]+)`"),
		linkRegex:        regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`),
		htmlTagRegex:     regexp.MustCompile(`<[^>]+>`),
		whitespaceRegex:  regexp.MustCompile(`\s+`),
	}
}

// Extract generates summary from full documentation by processing through
// several stages: strip markdown headers, strip formatting, strip HTML,
// normalize whitespace, and extract first sentence.
func (e *Extractor) Extract(documentation []string) string {
	if len(documentation) == 0 {
		return ""
	}

	// Join all documentation lines
	fullText := strings.Join(documentation, " ")

	// Strip markdown headers
	fullText = e.stripMarkdownHeaders(fullText)

	// Strip markdown formatting
	fullText = e.stripMarkdownFormatting(fullText)

	// Strip HTML tags
	fullText = e.stripHTMLTags(fullText)

	// Normalize whitespace
	fullText = e.normalizeWhitespace(fullText)

	// Extract first sentence
	return e.extractFirstSentence(fullText)
}

// stripMarkdownHeaders removes # headers (uses pre-compiled regex)
func (e *Extractor) stripMarkdownHeaders(text string) string {
	return e.headerRegex.ReplaceAllString(text, "")
}

// stripMarkdownFormatting removes **bold**, *italic*, `code`, [links]
// All regexes are pre-compiled for performance.
// Order matters: ** before *, __ before _
func (e *Extractor) stripMarkdownFormatting(text string) string {
	// Remove bold/italic (order matters: ** before *, __ before _)
	text = e.boldRegex.ReplaceAllString(text, "$1")
	text = e.italicRegex.ReplaceAllString(text, "$1")
	text = e.boldUnderRegex.ReplaceAllString(text, "$1")
	text = e.italicUnderRegex.ReplaceAllString(text, "$1")

	// Remove inline code
	text = e.codeRegex.ReplaceAllString(text, "$1")

	// Remove links [text](url) -> text
	text = e.linkRegex.ReplaceAllString(text, "$1")

	return text
}

// stripHTMLTags removes HTML tags (uses pre-compiled regex)
func (e *Extractor) stripHTMLTags(text string) string {
	return e.htmlTagRegex.ReplaceAllString(text, "")
}

// normalizeWhitespace collapses multiple spaces and newlines (uses pre-compiled regex)
func (e *Extractor) normalizeWhitespace(text string) string {
	// Replace newlines with spaces
	text = strings.ReplaceAll(text, "\n", " ")

	// Collapse multiple spaces
	text = e.whitespaceRegex.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

// extractFirstSentence uses regex to extract first sentence.
// Handles edge cases:
//   - Ellipsis (...) is NOT a sentence boundary
//   - IP addresses (127.0.0.1.) are NOT sentence boundaries
//   - Standard punctuation (.!?) followed by space or EOL IS a sentence boundary
//
// If no sentence terminator is found, returns the full text.
func (e *Extractor) extractFirstSentence(text string) string {
	matches := e.sentenceRegex.FindStringSubmatch(text)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// No sentence ending found, return full text
	return text
}
