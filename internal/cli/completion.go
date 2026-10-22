package cli

import (
	"embed"
	"fmt"
	"io"
)

// The scripts are fixed text: they do not run bcuz while completing, because
// its options are few and fixed (docs/design/decisions.md). The tests run
// each of them in its real shell to see that they offer the same.
//
//go:embed completions
var completionScripts embed.FS

// scriptFiles are the shells that have a script, and the files that hold them.
var scriptFiles = map[string]string{
	"bash":       "completions/bcuz.bash",
	"zsh":        "completions/bcuz.zsh",
	"fish":       "completions/bcuz.fish",
	"powershell": "completions/bcuz.ps1",
}

// printCompletion prints the completion script of the shell in args
// (--completion).
func printCompletion(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return fail(stderr, msgMissingShell)
	}
	file, ok := scriptFiles[args[0]]
	if !ok {
		return fail(stderr, fmt.Sprintf(fmtUnknownShell, args[0]))
	}
	if len(args) > 1 {
		return fail(stderr, fmt.Sprintf(fmtUnknownArgument, args[1]))
	}
	script, err := completionScripts.ReadFile(file)
	if err != nil {
		panic(err) // the files are embedded: only a broken build gets here
	}
	write(stdout, string(script))
	return 0
}
