package platform

import (
	"fmt"
	"runtime"
)

var (
	Version = "0.1.0"
	Commit  = "dev"
	Date    = "unknown"
)

func BuildInfo() map[string]string {
	return map[string]string{
		"version": Version,
		"commit":  Commit,
		"date":    Date,
		"go":      runtime.Version(),
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
	}
}

func VersionString() string {
	return fmt.Sprintf("%s (%s, go %s)", Version, Commit, runtime.Version())
}
