package richtext

import "strings"

// SegmentType represents the type of text segment
type SegmentType int

const (
	SegmentPlain  SegmentType = iota // Plain text
	SegmentBold                       // **text** or __text__
	SegmentItalic                     // *text* or _text_
	SegmentCode                       // `code`
	SegmentLink                       // [text](url)
)

// Segment represents a piece of text with optional formatting
type Segment struct {
	Type    SegmentType
	Content string // The text content (without markdown markers)
	URL     string // For links only (empty for other types)
}

// RichText represents formatted text as a sequence of segments
type RichText []Segment

// PlainText returns the text with all formatting stripped
func (rt RichText) PlainText() string {
	var buf strings.Builder
	for _, seg := range rt {
		buf.WriteString(seg.Content)
	}
	return buf.String()
}

// Markdown returns the text with markdown formatting preserved
func (rt RichText) Markdown() string {
	var buf strings.Builder
	for _, seg := range rt {
		switch seg.Type {
		case SegmentBold:
			buf.WriteString("**")
			buf.WriteString(seg.Content)
			buf.WriteString("**")
		case SegmentItalic:
			buf.WriteString("*")
			buf.WriteString(seg.Content)
			buf.WriteString("*")
		case SegmentCode:
			buf.WriteString("`")
			buf.WriteString(seg.Content)
			buf.WriteString("`")
		case SegmentLink:
			buf.WriteString("[")
			buf.WriteString(seg.Content)
			buf.WriteString("](")
			buf.WriteString(seg.URL)
			buf.WriteString(")")
		default:
			buf.WriteString(seg.Content)
		}
	}
	return buf.String()
}

// String returns the markdown representation (alias for Markdown)
func (rt RichText) String() string {
	return rt.Markdown()
}
