package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCommandLineFromHelpFile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		cmdLine     string
		expectError bool
		errorMsg    string
		checkConfig func(*testing.T, *Config)
	}{
		{
			name:        "empty command line",
			cmdLine:     "",
			expectError: false,
			checkConfig: func(t *testing.T, config *Config) {
				// Should have default values
				assert.Equal(t, ColorAuto, config.ColorMode)
			},
		},
		{
			name:        "command line with just make-help",
			cmdLine:     "make-help",
			expectError: false,
			checkConfig: func(t *testing.T, config *Config) {
				assert.Equal(t, ColorAuto, config.ColorMode)
			},
		},
		{
			name:        "command line with color flag",
			cmdLine:     "make-help --color",
			expectError: false,
			checkConfig: func(t *testing.T, config *Config) {
				assert.Equal(t, ColorAlways, config.ColorMode)
			},
		},
		{
			name:        "command line with no-color flag",
			cmdLine:     "make-help --no-color",
			expectError: false,
			checkConfig: func(t *testing.T, config *Config) {
				assert.Equal(t, ColorNever, config.ColorMode)
			},
		},
		{
			name:        "command line with keep-order flags",
			cmdLine:     "make-help --keep-order-categories --keep-order-targets",
			expectError: false,
			checkConfig: func(t *testing.T, config *Config) {
				assert.True(t, config.KeepOrderCategories)
				assert.True(t, config.KeepOrderTargets)
			},
		},
		{
			name:        "command line with keep-order-all",
			cmdLine:     "make-help --keep-order-all",
			expectError: false,
			checkConfig: func(t *testing.T, config *Config) {
				assert.True(t, config.KeepOrderCategories)
				assert.True(t, config.KeepOrderTargets)
				assert.True(t, config.KeepOrderFiles)
			},
		},
		{
			name:        "command line with default-category",
			cmdLine:     "make-help --default-category General",
			expectError: false,
			checkConfig: func(t *testing.T, config *Config) {
				assert.Equal(t, "General", config.DefaultCategory)
			},
		},
		{
			name:        "command line with category-order",
			cmdLine:     "make-help --category-order Build,Test,Deploy",
			expectError: false,
			checkConfig: func(t *testing.T, config *Config) {
				assert.Equal(t, []string{"Build", "Test", "Deploy"}, config.CategoryOrder)
			},
		},
		{
			name:        "command line with format",
			cmdLine:     "make-help --format text",
			expectError: false,
			checkConfig: func(t *testing.T, config *Config) {
				assert.Equal(t, "text", config.Format)
			},
		},
		{
			name:        "command line with verbose",
			cmdLine:     "make-help --verbose",
			expectError: false,
			checkConfig: func(t *testing.T, config *Config) {
				assert.True(t, config.Verbose)
			},
		},
		{
			name:        "disallowed flag: remove-help",
			cmdLine:     "make-help --remove-help",
			expectError: true,
			errorMsg:    "mode flag --remove-help is not allowed",
		},
		{
			name:        "disallowed flag: dry-run",
			cmdLine:     "make-help --dry-run",
			expectError: true,
			errorMsg:    "mode flag --dry-run is not allowed",
		},
		{
			name:        "disallowed flag: lint",
			cmdLine:     "make-help --lint",
			expectError: true,
			errorMsg:    "mode flag --lint is not allowed",
		},
		{
			name:        "disallowed flag: fix",
			cmdLine:     "make-help --fix",
			expectError: true,
			errorMsg:    "mode flag --fix is not allowed",
		},
		{
			name:        "disallowed flag: target",
			cmdLine:     "make-help --target build",
			expectError: true,
			errorMsg:    "mode flag --target is not allowed",
		},
		{
			name:        "disallowed flag with equals sign",
			cmdLine:     "make-help --target=build",
			expectError: true,
			errorMsg:    "mode flag --target is not allowed",
		},
		{
			name:        "multiple allowed flags",
			cmdLine:     "make-help --no-color --verbose --default-category General",
			expectError: false,
			checkConfig: func(t *testing.T, config *Config) {
				assert.Equal(t, ColorNever, config.ColorMode)
				assert.True(t, config.Verbose)
				assert.Equal(t, "General", config.DefaultCategory)
			},
		},
		{
			name:        "command line without make-help prefix",
			cmdLine:     "--color --verbose",
			expectError: false,
			checkConfig: func(t *testing.T, config *Config) {
				assert.Equal(t, ColorAlways, config.ColorMode)
				assert.True(t, config.Verbose)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			config := NewConfig()
			err := ParseCommandLineFromHelpFile(tt.cmdLine, config)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				if tt.checkConfig != nil {
					tt.checkConfig(t, config)
				}
			}
		})
	}
}
