package summary

import (
	"regexp"
	"strings"

	"github.com/sdlcforge/make-help/internal/richtext"
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
	parser           *richtext.Parser
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
		parser:           richtext.NewParser(),
	}
}

// Extract generates summary from full documentation by processing through
// several stages: strip markdown headers, detect first sentence boundary
// (using stripped text), then parse the first sentence preserving formatting.
//
// The key insight: we strip formatting to find sentence boundaries correctly
// (so "**Build.** Do it." finds "Build." as the first sentence), but then
// parse the original text to preserve formatting in the output.
//
// HTML is stripped, markdown formatting is preserved via RichText.
func (e *Extractor) Extract(documentation []string) richtext.RichText {
	if len(documentation) == 0 {
		return nil
	}

	// Join all documentation lines
	fullText := strings.Join(documentation, " ")

	// Strip markdown headers
	fullText = e.stripMarkdownHeaders(fullText)

	// For sentence boundary detection, we need to strip formatting
	// Otherwise "**Build.** Next sentence." would not detect "Build." correctly
	strippedText := e.stripMarkdownFormatting(fullText)
	strippedText = e.stripHTMLTags(strippedText)
	strippedText = e.normalizeWhitespace(strippedText)

	// Find the first sentence boundary using stripped text
	firstSentence := e.extractFirstSentence(strippedText)

	// Now find the same sentence in the original text to preserve formatting
	// We need to extract the corresponding portion from fullText
	originalFirstSentence := e.extractMatchingPortion(fullText, firstSentence)

	// Parse the original first sentence to preserve formatting as RichText
	return e.parser.Parse(originalFirstSentence)
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

// extractMatchingPortion finds the portion of originalText that corresponds
// to the strippedSentence. This allows us to preserve formatting in the output.
//
// Strategy: Find the sentence boundary in the original text by looking for
// the same punctuation mark at approximately the same position.
// We normalize whitespace in both texts for comparison.
func (e *Extractor) extractMatchingPortion(originalText, strippedSentence string) string {
	if strippedSentence == "" {
		return ""
	}

	// Normalize whitespace in original text
	normalizedOriginal := e.normalizeWhitespace(originalText)

	// The stripped sentence should be a prefix of the normalized original
	// (after we strip HTML tags from original)
	normalizedOriginal = e.stripHTMLTags(normalizedOriginal)

	// Find the sentence boundary in the normalized original
	// by looking for the same terminating punctuation
	lastChar := strippedSentence[len(strippedSentence)-1]
	if lastChar != '.' && lastChar != '!' && lastChar != '?' {
		// No sentence terminator, return the whole text
		return normalizedOriginal
	}

	// Use the same regex to find the sentence boundary in the normalized original
	matches := e.sentenceRegex.FindStringSubmatch(normalizedOriginal)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Fallback: return the original text
	return normalizedOriginal
}
