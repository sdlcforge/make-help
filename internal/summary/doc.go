// Package summary extracts topic sentences from documentation text.
//
// This is a Go port of the extract-topic JavaScript library. It processes
// text through several stages:
//  1. Strip markdown headers
//  2. Remove markdown formatting (bold, italic, code, links)
//  3. Remove HTML tags
//  4. Normalize whitespace
//  5. Extract first sentence using regex
//
// # Sentence Extraction
//
// The sentence extraction handles edge cases like ellipsis (...) and
// IP addresses (127.0.0.1) which should not be treated as sentence endings.
//
// The regex pattern used is:
//
//	^((?:[^.!?]|\.\.\.|\.[^\s])+[.?!])(\s|$)
//
// This matches:
//   - Any character except .!?
//   - OR ellipsis (...)
//   - OR period followed by non-whitespace (e.g., IP addresses)
//   - Ending with a sentence terminator (.!?)
//   - Followed by whitespace or end of string
//
// # Performance
//
// All regex patterns are pre-compiled at Extractor construction time for
// performance when processing many targets. Create a single Extractor
// instance and reuse it for all targets.
package summary
