// Package version holds build-time version information.
// Values are injected via ldflags at build time.
package version

var (
	Version   = "dev"
	BuildTime = "unknown"
)
