// Command bcuz runs a command with a reason and records it.
package main

import (
	"os"

	"github.com/amisonnet8/bcuz/internal/cli"
)

// version is set by the release build (-ldflags "-X main.version=…",
// .goreleaser.yaml). Empty for go install, which records the version itself.
var version string

func main() {
	cli.ReleaseVersion = version
	os.Exit(cli.Main(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
