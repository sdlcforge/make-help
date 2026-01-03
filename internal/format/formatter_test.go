package format

import (
	"strings"
	"testing"
)

// TestNewFormatter tests the formatter factory function
func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name        string
		formatType  string
		wantType    string // expected type name
		wantErr     bool
		errContains string
	}{
		{
			name:       "make format",
			formatType: "make",
			wantType:   "*format.MakeFormatter",
			wantErr:    false,
		},
		{
			name:       "mk alias",
			formatType: "mk",
			wantType:   "*format.MakeFormatter",
			wantErr:    false,
		},
		{
			name:       "text format",
			formatType: "text",
			wantType:   "*format.TextFormatter",
			wantErr:    false,
		},
		{
			name:       "txt alias",
			formatType: "txt",
			wantType:   "*format.TextFormatter",
			wantErr:    false,
		},
		{
			name:       "html format",
			formatType: "html",
			wantType:   "*format.HTMLFormatter",
			wantErr:    false,
		},
		{
			name:       "markdown format",
			formatType: "markdown",
			wantType:   "*format.MarkdownFormatter",
			wantErr:    false,
		},
		{
			name:       "md alias",
			formatType: "md",
			wantType:   "*format.MarkdownFormatter",
			wantErr:    false,
		},
		{
			name:        "unknown format",
			formatType:  "invalid",
			wantErr:     true,
			errContains: "unknown format type",
		},
	}

	config := &FormatterConfig{UseColor: false}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter, err := NewFormatter(tt.formatType, config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewFormatter() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errContains != "" {
					if !strings.Contains(err.Error(), tt.errContains) {
						t.Errorf("NewFormatter() error = %v, want error containing %q", err, tt.errContains)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("NewFormatter() unexpected error = %v", err)
				return
			}

			if formatter == nil {
				t.Error("NewFormatter() returned nil formatter")
				return
			}

			// Verify formatter type (using type assertion)
			switch tt.wantType {
			case "*format.MakeFormatter":
				if _, ok := formatter.(*MakeFormatter); !ok {
					t.Errorf("NewFormatter() returned %T, want %s", formatter, tt.wantType)
				}
			case "*format.TextFormatter":
				if _, ok := formatter.(*TextFormatter); !ok {
					t.Errorf("NewFormatter() returned %T, want %s", formatter, tt.wantType)
				}
			case "*format.HTMLFormatter":
				if _, ok := formatter.(*HTMLFormatter); !ok {
					t.Errorf("NewFormatter() returned %T, want %s", formatter, tt.wantType)
				}
			case "*format.MarkdownFormatter":
				if _, ok := formatter.(*MarkdownFormatter); !ok {
					t.Errorf("NewFormatter() returned %T, want %s", formatter, tt.wantType)
				}
			}
		})
	}
}

// TestFormatterInterface verifies that all formatters implement the Formatter interface
func TestFormatterInterface(t *testing.T) {
	config := &FormatterConfig{UseColor: false}

	// Test that each formatter implements the interface
	formatters := []Formatter{
		NewMakeFormatter(config),
		NewTextFormatter(config),
		NewHTMLFormatter(config),
		NewMarkdownFormatter(config),
	}

	for _, f := range formatters {
		if f == nil {
			t.Error("formatter is nil")
			continue
		}

		// Verify interface methods are callable
		if f.ContentType() == "" {
			t.Errorf("%T.ContentType() returned empty string", f)
		}
		if f.DefaultExtension() == "" {
			t.Errorf("%T.DefaultExtension() returned empty string", f)
		}
	}
}

// TestFormatterContentTypes verifies content types for each formatter
func TestFormatterContentTypes(t *testing.T) {
	tests := []struct {
		name        string
		formatter   Formatter
		wantContent string
		wantExt     string
	}{
		{
			name:        "MakeFormatter",
			formatter:   NewMakeFormatter(&FormatterConfig{}),
			wantContent: "text/x-makefile",
			wantExt:     ".mk",
		},
		{
			name:        "TextFormatter",
			formatter:   NewTextFormatter(&FormatterConfig{}),
			wantContent: "text/plain",
			wantExt:     ".txt",
		},
		{
			name:        "HTMLFormatter",
			formatter:   NewHTMLFormatter(&FormatterConfig{}),
			wantContent: "text/html",
			wantExt:     ".html",
		},
		{
			name:        "MarkdownFormatter",
			formatter:   NewMarkdownFormatter(&FormatterConfig{}),
			wantContent: "text/markdown",
			wantExt:     ".md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.formatter.ContentType(); got != tt.wantContent {
				t.Errorf("ContentType() = %q, want %q", got, tt.wantContent)
			}
			if got := tt.formatter.DefaultExtension(); got != tt.wantExt {
				t.Errorf("DefaultExtension() = %q, want %q", got, tt.wantExt)
			}
		})
	}
}

// TestFormatterNilConfig verifies that NewMakeFormatter and NewTextFormatter handle nil config
func TestFormatterNilConfig(t *testing.T) {
	t.Run("NewMakeFormatter with nil config", func(t *testing.T) {
		formatter := NewMakeFormatter(nil)
		if formatter == nil {
			t.Error("NewMakeFormatter(nil) should not return nil")
		}
	})

	t.Run("NewTextFormatter with nil config", func(t *testing.T) {
		formatter := NewTextFormatter(nil)
		if formatter == nil {
			t.Error("NewTextFormatter(nil) should not return nil")
		}
	})

	t.Run("NewHTMLFormatter with nil config", func(t *testing.T) {
		formatter := NewHTMLFormatter(nil)
		if formatter == nil {
			t.Error("NewHTMLFormatter(nil) should not return nil")
		}
	})

	t.Run("NewMarkdownFormatter with nil config", func(t *testing.T) {
		formatter := NewMarkdownFormatter(nil)
		if formatter == nil {
			t.Error("NewMarkdownFormatter(nil) should not return nil")
		}
	})
}
