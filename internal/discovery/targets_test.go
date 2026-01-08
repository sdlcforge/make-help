package discovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTargetsFromDatabase_PhonyStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		input            string
		expectedTargets  []string
		expectedPhony    map[string]bool
	}{
		{
			name: "single .PHONY line",
			input: `# Make database
.PHONY: build test clean
build:
	go build
test:
	go test
clean:
	rm -rf bin
`,
			expectedTargets: []string{"build", "test", "clean"},
			expectedPhony: map[string]bool{
				"build": true,
				"test":  true,
				"clean": true,
			},
		},
		{
			name: "multiple .PHONY lines",
			input: `# Make database
.PHONY: build test
.PHONY: clean install
build:
	go build
test:
	go test
clean:
	rm -rf bin
install:
	cp bin/app /usr/local/bin
`,
			expectedTargets: []string{"build", "test", "clean", "install"},
			expectedPhony: map[string]bool{
				"build":   true,
				"test":    true,
				"clean":   true,
				"install": true,
			},
		},
		{
			name: "mixed phony and non-phony targets",
			input: `# Make database
.PHONY: clean test
build:
	go build
test:
	go test
clean:
	rm -rf bin
output.txt:
	echo "hello" > output.txt
`,
			expectedTargets: []string{"build", "test", "clean", "output.txt"},
			expectedPhony: map[string]bool{
				"clean": true,
				"test":  true,
			},
		},
		{
			name: "no .PHONY targets",
			input: `# Make database
build:
	go build
test:
	go test
`,
			expectedTargets: []string{"build", "test"},
			expectedPhony:   map[string]bool{},
		},
		{
			name: ".PHONY with targets not in database",
			input: `# Make database
.PHONY: build test missing-target
build:
	go build
test:
	go test
`,
			expectedTargets: []string{"build", "test"},
			expectedPhony: map[string]bool{
				"build":          true,
				"test":           true,
				"missing-target": true, // Still tracked even if target not defined
			},
		},
		{
			name: ".PHONY with single target",
			input: `# Make database
.PHONY: clean
clean:
	rm -rf bin
build:
	go build
`,
			expectedTargets: []string{"clean", "build"},
			expectedPhony: map[string]bool{
				"clean": true,
			},
		},
		{
			name: "empty .PHONY line",
			input: `# Make database
.PHONY:
build:
	go build
`,
			expectedTargets: []string{"build"},
			expectedPhony:   map[string]bool{},
		},
		{
			name: ".PHONY with extra whitespace",
			input: `# Make database
.PHONY:    build    test    clean
build:
	go build
test:
	go test
clean:
	rm -rf bin
`,
			expectedTargets: []string{"build", "test", "clean"},
			expectedPhony: map[string]bool{
				"build": true,
				"test":  true,
				"clean": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := parseTargetsFromDatabase(tt.input)

			assert.Equal(t, tt.expectedTargets, result.Targets, "targets mismatch")
			assert.Equal(t, tt.expectedPhony, result.IsPhony, "phony status mismatch")
		})
	}
}

func TestParseTargetsFromDatabase_ReturnStruct(t *testing.T) {
	t.Parallel()
	input := `# Make database
.PHONY: clean
build:
	go build
clean:
	rm -rf bin
`

	result := parseTargetsFromDatabase(input)

	// Verify it returns a struct
	assert.NotNil(t, result)
	assert.NotNil(t, result.Targets)
	assert.NotNil(t, result.IsPhony)

	// Verify content
	assert.Contains(t, result.Targets, "build")
	assert.Contains(t, result.Targets, "clean")
	assert.True(t, result.IsPhony["clean"])
	assert.False(t, result.IsPhony["build"])
}
