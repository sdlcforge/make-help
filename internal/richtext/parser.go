package richtext

import (
	"regexp"
	"sort"
)

const (
	// MaxInputLength is the maximum allowed input text length (10KB)
	MaxInputLength = 10 * 1024
	// MaxSegmentLength is the maximum length for a single segment (2000 chars)
	MaxSegmentLength = 2000
)

// ansiEscapeRegex matches ANSI escape codes for stripping
var ansiEscapeRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// Parser parses markdown inline formatting into RichText segments
type Parser struct {
	linkRegex        *regexp.Regexp
	codeRegex        *regexp.Regexp
	boldRegex        *regexp.Regexp
	boldUnderRegex   *regexp.Regexp
	italicRegex      *regexp.Regexp
	italicUnderRegex *regexp.Regexp
}

// NewParser creates a new Parser with pre-compiled regex patterns
func NewParser() *Parser {
	return &Parser{
		// Links: [text](url) - use non-greedy +? to prevent excessive backtracking
		linkRegex: regexp.MustCompile(`\[([^\]]+?)\]\(([^)]+?)\)`),
		// Code: `code` - non-greedy match
		codeRegex: regexp.MustCompile("`([^`]+)`"),
		// Bold: **text** - match content that doesn't contain **
		boldRegex: regexp.MustCompile(`\*\*(.+?)\*\*`),
		// Bold underscore: __text__ - match content that doesn't contain __
		boldUnderRegex: regexp.MustCompile(`__(.+?)__`),
		// Italic: *text* - single asterisks only, will be filtered by overlap check
		italicRegex: regexp.MustCompile(`\*([^*]+)\*`),
		// Italic underscore: _text_ - single underscores only
		italicUnderRegex: regexp.MustCompile(`_([^_]+)_`),
	}
}

// match represents a formatting match with its position
type match struct {
	start   int
	end     int
	segment Segment
}

// Parse converts a markdown string into RichText segments
// Processing order: links → code → bold → italic (highest to lowest precedence)
func (p *Parser) Parse(text string) RichText {
	// Strip ANSI escape codes to prevent ANSI injection
	text = ansiEscapeRegex.ReplaceAllString(text, "")

	// Enforce input length limit
	if len(text) > MaxInputLength {
		// On parse error, return plain text segment
		return RichText{
			{Type: SegmentPlain, Content: text[:MaxInputLength]},
		}
	}

	if text == "" {
		return RichText{}
	}

	// Find all matches
	matches := p.findAllMatches(text)

	// If no matches, return plain text
	if len(matches) == 0 {
		return RichText{{Type: SegmentPlain, Content: text}}
	}

	// Build segments
	return p.buildSegments(text, matches)
}

// findAllMatches finds all formatting matches in the text
func (p *Parser) findAllMatches(text string) []match {
	var matches []match

	// 1. Find links (highest precedence)
	for _, loc := range p.linkRegex.FindAllStringSubmatchIndex(text, -1) {
		content := text[loc[2]:loc[3]]
		if len(content) > MaxSegmentLength {
			continue // Skip oversized segments
		}
		matches = append(matches, match{
			start: loc[0],
			end:   loc[1],
			segment: Segment{
				Type:    SegmentLink,
				Content: content,
				URL:     text[loc[4]:loc[5]],
			},
		})
	}

	// 2. Find code (second precedence)
	for _, loc := range p.codeRegex.FindAllStringSubmatchIndex(text, -1) {
		content := text[loc[2]:loc[3]]
		if len(content) > MaxSegmentLength {
			continue // Skip oversized segments
		}
		if !p.overlaps(matches, loc[0], loc[1]) {
			matches = append(matches, match{
				start: loc[0],
				end:   loc[1],
				segment: Segment{
					Type:    SegmentCode,
					Content: content,
				},
			})
		}
	}

	// 3. Find bold (third precedence) - both ** and __
	for _, loc := range p.boldRegex.FindAllStringSubmatchIndex(text, -1) {
		content := text[loc[2]:loc[3]]
		if len(content) > MaxSegmentLength {
			continue // Skip oversized segments
		}
		if !p.overlaps(matches, loc[0], loc[1]) {
			matches = append(matches, match{
				start: loc[0],
				end:   loc[1],
				segment: Segment{
					Type:    SegmentBold,
					Content: content,
				},
			})
		}
	}

	for _, loc := range p.boldUnderRegex.FindAllStringSubmatchIndex(text, -1) {
		content := text[loc[2]:loc[3]]
		if len(content) > MaxSegmentLength {
			continue // Skip oversized segments
		}
		if !p.overlaps(matches, loc[0], loc[1]) {
			matches = append(matches, match{
				start: loc[0],
				end:   loc[1],
				segment: Segment{
					Type:    SegmentBold,
					Content: content,
				},
			})
		}
	}

	// 4. Find italic (lowest precedence) - both * and _
	// Manually scan for italic to avoid regex consuming wrong positions
	matches = p.findItalicMatches(text, matches, '*')
	matches = p.findItalicMatches(text, matches, '_')

	return matches
}

// findItalicMatches manually scans for italic patterns using the given delimiter
func (p *Parser) findItalicMatches(text string, existingMatches []match, delim byte) []match {
	matches := existingMatches
	pos := 0

	for pos < len(text) {
		// Find opening delimiter
		start := -1
		for i := pos; i < len(text); i++ {
			if text[i] == delim {
				nextIsDelim := i+1 < len(text) && text[i+1] == delim

				// Skip if next char is same delimiter (this is start of **)
				if nextIsDelim {
					continue
				}

				// Skip if this position is inside an existing match
				if p.isInsideMatch(matches, i) {
					continue
				}

				// This is a valid opening delimiter
				start = i
				break
			}
		}

		if start == -1 {
			break // No more opening delimiters
		}

		// Find closing delimiter
		end := -1
		for i := start + 2; i < len(text); i++ { // +2 to ensure at least one char content
			if text[i] == delim {
				// Check it's not the first of a ** pair
				nextIsDelim := i+1 < len(text) && text[i+1] == delim
				if !nextIsDelim {
					end = i + 1 // end is exclusive
					break
				}
			}
		}

		if end == -1 {
			pos = start + 1
			continue // No closing delimiter found
		}

		// Check if this match overlaps with existing matches
		if !p.overlaps(matches, start, end) {
			content := text[start+1 : end-1]
			// Skip oversized segments
			if len(content) <= MaxSegmentLength {
				matches = append(matches, match{
					start: start,
					end:   end,
					segment: Segment{
						Type:    SegmentItalic,
						Content: content,
					},
				})
			}
		}

		pos = end
	}

	return matches
}

// isInsideMatch checks if a position is inside any existing match
func (p *Parser) isInsideMatch(matches []match, pos int) bool {
	for _, m := range matches {
		if pos >= m.start && pos < m.end {
			return true
		}
	}
	return false
}

// overlaps checks if a new match overlaps with existing matches
func (p *Parser) overlaps(matches []match, start, end int) bool {
	for _, m := range matches {
		if (start >= m.start && start < m.end) || (end > m.start && end <= m.end) || (start <= m.start && end >= m.end) {
			return true
		}
	}
	return false
}

// buildSegments builds the final RichText from matches and plain text
func (p *Parser) buildSegments(text string, matches []match) RichText {
	// Sort matches by start position
	p.sortMatches(matches)

	var segments RichText
	pos := 0

	for _, m := range matches {
		// Add plain text before this match
		if m.start > pos {
			plainText := text[pos:m.start]
			if plainText != "" {
				segments = append(segments, Segment{
					Type:    SegmentPlain,
					Content: plainText,
				})
			}
		}

		// Add the formatted segment
		segments = append(segments, m.segment)
		pos = m.end
	}

	// Add remaining plain text
	if pos < len(text) {
		plainText := text[pos:]
		if plainText != "" {
			segments = append(segments, Segment{
				Type:    SegmentPlain,
				Content: plainText,
			})
		}
	}

	return segments
}

// sortMatches sorts matches by start position using stdlib sort
func (p *Parser) sortMatches(matches []match) {
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].start < matches[j].start
	})
}
