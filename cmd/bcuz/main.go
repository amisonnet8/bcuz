// Command bcuz runs a command with a reason and records it.
package main

import (
	"os"

	"github.com/amisonnet8/bcuz/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
