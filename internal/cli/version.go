package cli

import (
	"runtime/debug"
	"strings"
)

// develVersion is shown when the build has no module version.
const develVersion = "(devel)"

// pseudoPrefix starts the pseudo-version of a build that has no tag.
const pseudoPrefix = "v0.0.0-"

// readBuildInfo is replaced by the tests.
var readBuildInfo = debug.ReadBuildInfo

// ReleaseVersion is the tag that the release build set (main.version, from
// -ldflags "-X main.version=v0.1.0" in .goreleaser.yaml). Empty otherwise.
var ReleaseVersion string

// version returns the version of the build, such as "v0.1.0": the one the
// release build set, or else the module version that Go recorded (go install
// …@v0.1.0 records v0.1.0).
func version() string {
	if ReleaseVersion != "" {
		return ReleaseVersion
	}
	info, ok := readBuildInfo()
	if !ok || info.Main.Version == "" {
		return develVersion
	}
	// Since Go 1.24 a build inside a git checkout gets a pseudo-version such
	// as v0.0.0-20261011232240-19408631a724+dirty. Only a tag is a version.
	if strings.HasPrefix(info.Main.Version, pseudoPrefix) {
		return develVersion
	}
	return info.Main.Version
}
