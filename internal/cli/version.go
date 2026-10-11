package cli

import "runtime/debug"

// develVersion is shown when the build has no module version.
const develVersion = "(devel)"

// readBuildInfo is replaced by the tests.
var readBuildInfo = debug.ReadBuildInfo

// version returns the module version of the build, such as "v0.1.0".
func version() string {
	info, ok := readBuildInfo()
	if !ok || info.Main.Version == "" {
		return develVersion
	}
	return info.Main.Version
}
