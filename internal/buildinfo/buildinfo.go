// Package buildinfo exposes metadata stamped into release binaries with -ldflags.
package buildinfo

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)
