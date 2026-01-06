package format

import (
	"bytes"
	"testing"

	"github.com/sdlcforge/make-help/internal/model"
)

// createBenchmarkModel creates a realistic help model for benchmarking.
// It contains ~50 targets across ~5 categories with variables, aliases, and file docs.
func createBenchmarkModel() *model.HelpModel {
	return &model.HelpModel{
		FileDocs: []model.FileDoc{
			{
				SourceFile:     "Makefile",
				Documentation:  []string{"Main Makefile for the project.", "Contains common development tasks."},
				IsEntryPoint:   true,
				DiscoveryOrder: 0,
			},
			{
				SourceFile:     "make/build.mk",
				Documentation:  []string{"Build and compilation tasks."},
				IsEntryPoint:   false,
				DiscoveryOrder: 1,
			},
			{
				SourceFile:     "make/test.mk",
				Documentation:  []string{"Testing and quality assurance tasks."},
				IsEntryPoint:   false,
				DiscoveryOrder: 2,
			},
		},
		HasCategories: true,
		Categories: []model.Category{
			{
				Name: "Build",
				Targets: []model.Target{
					{Name: "build", Summary: []string{"Build the project."}, Aliases: []string{"b"}},
					{Name: "compile", Summary: []string{"Compile source files."}, Variables: []model.Variable{{Name: "GOOS"}, {Name: "GOARCH"}}},
					{Name: "clean", Summary: []string{"Clean build artifacts."}},
					{Name: "install", Summary: []string{"Install the binary."}, Variables: []model.Variable{{Name: "PREFIX", Description: "Installation prefix"}}},
					{Name: "uninstall", Summary: []string{"Uninstall the binary."}},
					{Name: "rebuild", Summary: []string{"Clean and rebuild."}, Aliases: []string{"rb"}},
					{Name: "build-linux", Summary: []string{"Build for Linux."}, Variables: []model.Variable{{Name: "GOARCH"}}},
					{Name: "build-darwin", Summary: []string{"Build for macOS."}, Variables: []model.Variable{{Name: "GOARCH"}}},
					{Name: "build-windows", Summary: []string{"Build for Windows."}, Variables: []model.Variable{{Name: "GOARCH"}}},
					{Name: "build-all", Summary: []string{"Build for all platforms."}, Aliases: []string{"ba"}},
				},
			},
			{
				Name: "Test",
				Targets: []model.Target{
					{Name: "test", Summary: []string{"Run all tests."}, Aliases: []string{"t"}},
					{Name: "test-unit", Summary: []string{"Run unit tests."}, Variables: []model.Variable{{Name: "VERBOSE", Description: "Enable verbose output"}}},
					{Name: "test-integration", Summary: []string{"Run integration tests."}},
					{Name: "test-e2e", Summary: []string{"Run end-to-end tests."}},
					{Name: "test-coverage", Summary: []string{"Generate test coverage report."}, Aliases: []string{"cov"}},
					{Name: "test-benchmark", Summary: []string{"Run benchmarks."}, Aliases: []string{"bench"}},
					{Name: "test-race", Summary: []string{"Run tests with race detector."}},
					{Name: "test-watch", Summary: []string{"Run tests in watch mode."}},
					{Name: "test-verbose", Summary: []string{"Run tests with verbose output."}, Aliases: []string{"tv"}},
					{Name: "test-short", Summary: []string{"Run tests in short mode."}},
				},
			},
			{
				Name: "Development",
				Targets: []model.Target{
					{Name: "dev", Summary: []string{"Start development server."}, Variables: []model.Variable{{Name: "PORT", Description: "Server port"}}},
					{Name: "watch", Summary: []string{"Watch files and rebuild."}, Aliases: []string{"w"}},
					{Name: "fmt", Summary: []string{"Format source code."}},
					{Name: "lint", Summary: []string{"Run linters."}, Variables: []model.Variable{{Name: "FIX", Description: "Auto-fix issues"}}},
					{Name: "vet", Summary: []string{"Run go vet."}},
					{Name: "check", Summary: []string{"Run all checks."}, Aliases: []string{"c"}},
					{Name: "generate", Summary: []string{"Run code generators."}, Aliases: []string{"gen"}},
					{Name: "mod-tidy", Summary: []string{"Tidy Go modules."}},
					{Name: "mod-download", Summary: []string{"Download Go modules."}},
					{Name: "mod-verify", Summary: []string{"Verify Go modules."}},
				},
			},
			{
				Name: "Docker",
				Targets: []model.Target{
					{Name: "docker-build", Summary: []string{"Build Docker image."}, Variables: []model.Variable{{Name: "TAG", Description: "Image tag"}}},
					{Name: "docker-run", Summary: []string{"Run Docker container."}, Variables: []model.Variable{{Name: "PORT"}}},
					{Name: "docker-push", Summary: []string{"Push Docker image."}},
					{Name: "docker-clean", Summary: []string{"Clean Docker resources."}},
					{Name: "docker-compose-up", Summary: []string{"Start Docker Compose services."}, Aliases: []string{"up"}},
					{Name: "docker-compose-down", Summary: []string{"Stop Docker Compose services."}, Aliases: []string{"down"}},
					{Name: "docker-logs", Summary: []string{"View Docker logs."}},
					{Name: "docker-shell", Summary: []string{"Open shell in container."}},
					{Name: "docker-exec", Summary: []string{"Execute command in container."}, Variables: []model.Variable{{Name: "CMD"}}},
					{Name: "docker-prune", Summary: []string{"Prune unused Docker resources."}},
				},
			},
			{
				Name: "Release",
				Targets: []model.Target{
					{Name: "release", Summary: []string{"Create a new release."}, Variables: []model.Variable{{Name: "VERSION", Description: "Release version"}}},
					{Name: "release-patch", Summary: []string{"Create patch release."}},
					{Name: "release-minor", Summary: []string{"Create minor release."}},
					{Name: "release-major", Summary: []string{"Create major release."}},
					{Name: "changelog", Summary: []string{"Generate changelog."}},
					{Name: "tag", Summary: []string{"Create Git tag."}, Variables: []model.Variable{{Name: "VERSION"}}},
					{Name: "publish", Summary: []string{"Publish release."}, Aliases: []string{"pub"}},
					{Name: "snapshot", Summary: []string{"Create snapshot build."}},
					{Name: "release-notes", Summary: []string{"Generate release notes."}},
					{Name: "release-dry-run", Summary: []string{"Dry run release process."}},
				},
			},
		},
	}
}

// BenchmarkMakeFormatter_RenderHelp benchmarks MakeFormatter with realistic data
func BenchmarkMakeFormatter_RenderHelp(b *testing.B) {
	formatter := NewMakeFormatter(&FormatterConfig{UseColor: false})
	model := createBenchmarkModel()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if err := formatter.RenderHelp(model, &buf); err != nil {
			b.Fatalf("RenderHelp() error = %v", err)
		}
	}
}

// BenchmarkMakeFormatter_RenderHelp_WithColor benchmarks MakeFormatter with colors
func BenchmarkMakeFormatter_RenderHelp_WithColor(b *testing.B) {
	formatter := NewMakeFormatter(&FormatterConfig{UseColor: true})
	model := createBenchmarkModel()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if err := formatter.RenderHelp(model, &buf); err != nil {
			b.Fatalf("RenderHelp() error = %v", err)
		}
	}
}

// BenchmarkTextFormatter_RenderHelp benchmarks TextFormatter with realistic data
func BenchmarkTextFormatter_RenderHelp(b *testing.B) {
	formatter := NewTextFormatter(&FormatterConfig{UseColor: false})
	model := createBenchmarkModel()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if err := formatter.RenderHelp(model, &buf); err != nil {
			b.Fatalf("RenderHelp() error = %v", err)
		}
	}
}

// BenchmarkTextFormatter_RenderHelp_WithColor benchmarks TextFormatter with colors
func BenchmarkTextFormatter_RenderHelp_WithColor(b *testing.B) {
	formatter := NewTextFormatter(&FormatterConfig{UseColor: true})
	model := createBenchmarkModel()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if err := formatter.RenderHelp(model, &buf); err != nil {
			b.Fatalf("RenderHelp() error = %v", err)
		}
	}
}

// BenchmarkHTMLFormatter_RenderHelp benchmarks HTMLFormatter with realistic data
func BenchmarkHTMLFormatter_RenderHelp(b *testing.B) {
	formatter := NewHTMLFormatter(&FormatterConfig{UseColor: false})
	model := createBenchmarkModel()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if err := formatter.RenderHelp(model, &buf); err != nil {
			b.Fatalf("RenderHelp() error = %v", err)
		}
	}
}

// BenchmarkHTMLFormatter_RenderHelp_WithCSS benchmarks HTMLFormatter with embedded CSS
func BenchmarkHTMLFormatter_RenderHelp_WithCSS(b *testing.B) {
	formatter := NewHTMLFormatter(&FormatterConfig{UseColor: true})
	model := createBenchmarkModel()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if err := formatter.RenderHelp(model, &buf); err != nil {
			b.Fatalf("RenderHelp() error = %v", err)
		}
	}
}

// BenchmarkMarkdownFormatter_RenderHelp benchmarks MarkdownFormatter with realistic data
func BenchmarkMarkdownFormatter_RenderHelp(b *testing.B) {
	formatter := NewMarkdownFormatter(&FormatterConfig{UseColor: false})
	model := createBenchmarkModel()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if err := formatter.RenderHelp(model, &buf); err != nil {
			b.Fatalf("RenderHelp() error = %v", err)
		}
	}
}

// BenchmarkJSONFormatter_RenderHelp benchmarks JSONFormatter with realistic data
func BenchmarkJSONFormatter_RenderHelp(b *testing.B) {
	formatter := NewJSONFormatter(&FormatterConfig{UseColor: false})
	model := createBenchmarkModel()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if err := formatter.RenderHelp(model, &buf); err != nil {
			b.Fatalf("RenderHelp() error = %v", err)
		}
	}
}

// BenchmarkMakeFormatter_RenderDetailedTarget benchmarks detailed target rendering
func BenchmarkMakeFormatter_RenderDetailedTarget(b *testing.B) {
	formatter := NewMakeFormatter(&FormatterConfig{UseColor: false})
	target := &model.Target{
		Name:    "build",
		Aliases: []string{"b", "compile"},
		Summary: []string{"Build the project."},
		Documentation: []string{
			"Build the project.",
			"",
			"This compiles all source files and generates the binary.",
			"You can customize the build with environment variables.",
		},
		Variables: []model.Variable{
			{Name: "GOOS", Description: "Target operating system"},
			{Name: "GOARCH", Description: "Target architecture"},
			{Name: "CGO_ENABLED", Description: "Enable CGO"},
		},
		SourceFile: "Makefile",
		LineNumber: 42,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if err := formatter.RenderDetailedTarget(target, &buf); err != nil {
			b.Fatalf("RenderDetailedTarget() error = %v", err)
		}
	}
}

// BenchmarkTextFormatter_RenderDetailedTarget benchmarks detailed target rendering
func BenchmarkTextFormatter_RenderDetailedTarget(b *testing.B) {
	formatter := NewTextFormatter(&FormatterConfig{UseColor: false})
	target := &model.Target{
		Name:    "build",
		Aliases: []string{"b", "compile"},
		Summary: []string{"Build the project."},
		Documentation: []string{
			"Build the project.",
			"",
			"This compiles all source files and generates the binary.",
			"You can customize the build with environment variables.",
		},
		Variables: []model.Variable{
			{Name: "GOOS", Description: "Target operating system"},
			{Name: "GOARCH", Description: "Target architecture"},
			{Name: "CGO_ENABLED", Description: "Enable CGO"},
		},
		SourceFile: "Makefile",
		LineNumber: 42,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if err := formatter.RenderDetailedTarget(target, &buf); err != nil {
			b.Fatalf("RenderDetailedTarget() error = %v", err)
		}
	}
}

// BenchmarkHTMLFormatter_RenderDetailedTarget benchmarks detailed target rendering
func BenchmarkHTMLFormatter_RenderDetailedTarget(b *testing.B) {
	formatter := NewHTMLFormatter(&FormatterConfig{UseColor: false})
	target := &model.Target{
		Name:    "build",
		Aliases: []string{"b", "compile"},
		Summary: []string{"Build the project."},
		Documentation: []string{
			"Build the project.",
			"",
			"This compiles all source files and generates the binary.",
			"You can customize the build with environment variables.",
		},
		Variables: []model.Variable{
			{Name: "GOOS", Description: "Target operating system"},
			{Name: "GOARCH", Description: "Target architecture"},
			{Name: "CGO_ENABLED", Description: "Enable CGO"},
		},
		SourceFile: "Makefile",
		LineNumber: 42,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if err := formatter.RenderDetailedTarget(target, &buf); err != nil {
			b.Fatalf("RenderDetailedTarget() error = %v", err)
		}
	}
}

// BenchmarkMarkdownFormatter_RenderDetailedTarget benchmarks detailed target rendering
func BenchmarkMarkdownFormatter_RenderDetailedTarget(b *testing.B) {
	formatter := NewMarkdownFormatter(&FormatterConfig{UseColor: false})
	target := &model.Target{
		Name:    "build",
		Aliases: []string{"b", "compile"},
		Summary: []string{"Build the project."},
		Documentation: []string{
			"Build the project.",
			"",
			"This compiles all source files and generates the binary.",
			"You can customize the build with environment variables.",
		},
		Variables: []model.Variable{
			{Name: "GOOS", Description: "Target operating system"},
			{Name: "GOARCH", Description: "Target architecture"},
			{Name: "CGO_ENABLED", Description: "Enable CGO"},
		},
		SourceFile: "Makefile",
		LineNumber: 42,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if err := formatter.RenderDetailedTarget(target, &buf); err != nil {
			b.Fatalf("RenderDetailedTarget() error = %v", err)
		}
	}
}

// BenchmarkJSONFormatter_RenderDetailedTarget benchmarks detailed target rendering
func BenchmarkJSONFormatter_RenderDetailedTarget(b *testing.B) {
	formatter := NewJSONFormatter(&FormatterConfig{UseColor: false})
	target := &model.Target{
		Name:    "build",
		Aliases: []string{"b", "compile"},
		Summary: []string{"Build the project."},
		Documentation: []string{
			"Build the project.",
			"",
			"This compiles all source files and generates the binary.",
			"You can customize the build with environment variables.",
		},
		Variables: []model.Variable{
			{Name: "GOOS", Description: "Target operating system"},
			{Name: "GOARCH", Description: "Target architecture"},
			{Name: "CGO_ENABLED", Description: "Enable CGO"},
		},
		SourceFile: "Makefile",
		LineNumber: 42,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		if err := formatter.RenderDetailedTarget(target, &buf); err != nil {
			b.Fatalf("RenderDetailedTarget() error = %v", err)
		}
	}
}

// BenchmarkAllFormatters_RenderHelp compares all formatters side-by-side
func BenchmarkAllFormatters_RenderHelp(b *testing.B) {
	model := createBenchmarkModel()
	config := &FormatterConfig{UseColor: false}

	benchmarks := []struct {
		name      string
		formatter Formatter
	}{
		{"Make", NewMakeFormatter(config)},
		{"Text", NewTextFormatter(config)},
		{"HTML", NewHTMLFormatter(config)},
		{"Markdown", NewMarkdownFormatter(config)},
		{"JSON", NewJSONFormatter(config)},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var buf bytes.Buffer
				if err := bm.formatter.RenderHelp(model, &buf); err != nil {
					b.Fatalf("RenderHelp() error = %v", err)
				}
			}
		})
	}
}
