// Package version provides build-time version information.
package version

// Version is set at build time via ldflags:
//
//	go build -ldflags "-X github.com/sdlcforge/make-help/internal/version.Version=1.0.0"
//
// If not set, defaults to "dev".
var Version = "dev"
