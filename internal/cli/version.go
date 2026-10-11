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

// version returns the module version of the build, such as "v0.1.0".
func version() string {
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
