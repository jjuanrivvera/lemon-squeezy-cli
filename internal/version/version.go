// Package version holds build metadata injected via -ldflags at build time.
package version

import "fmt"

// These are set via -X ldflags (see Makefile / .goreleaser.yaml). Defaults apply to
// `go run` / un-stamped builds.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// Info is the structured build metadata for `version --json`.
type Info struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

// Get returns the current build info.
func Get() Info { return Info{Version: Version, Commit: Commit, Date: Date} }

// String renders a one-line human summary.
func (i Info) String() string {
	return fmt.Sprintf("lsqueezy %s (commit %s, built %s)", i.Version, i.Commit, i.Date)
}
