package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestCompletionArguments(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStderr string
	}{
		{"no shell", []string{"--completion"}, 125, msgMissingShell + "\n"},
		{"unknown shell", []string{"--completion", "tcsh"}, 125, fmt.Sprintf(fmtUnknownShell, "tcsh") + "\n"},
		{"shell names are lowercase", []string{"--completion", "Bash"}, 125, fmt.Sprintf(fmtUnknownShell, "Bash") + "\n"},
		{"pwsh is not a name", []string{"--completion", "pwsh"}, 125, fmt.Sprintf(fmtUnknownShell, "pwsh") + "\n"},
		{"two shells", []string{"--completion", "bash", "zsh"}, 125, fmt.Sprintf(fmtUnknownArgument, "zsh") + "\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, stdout, stderr := runMain(t, tt.args...)
			if code != tt.wantCode || stdout != "" || stderr != tt.wantStderr {
				t.Errorf("got (%d, %q, %q), want (%d, \"\", %q)", code, stdout, stderr, tt.wantCode, tt.wantStderr)
			}
		})
	}
}

// Each shell gets its own script, and only that.
func TestCompletionPrintsTheScriptOfTheShell(t *testing.T) {
	marks := map[string]string{
		"bash":       "complete -F _bcuz bcuz",
		"zsh":        "#compdef bcuz",
		"fish":       "complete -c bcuz",
		"powershell": "Register-ArgumentCompleter -Native -CommandName bcuz",
	}
	for shell := range marks {
		code, stdout, stderr := runMain(t, "--completion", shell)
		if code != 0 || stderr != "" {
			t.Errorf("%s: got (%d, %q)", shell, code, stderr)
		}
		for other, otherMark := range marks {
			if got := strings.Contains(stdout, otherMark); got != (other == shell) {
				t.Errorf("--completion %s: has the mark of %s = %v", shell, other, got)
			}
		}
	}
}

// fakeCommand is a command on the PATH of the shells, with a name no other
// command starts with. Each shell is told that its arguments are alpha and
// beta, to see that the arguments of a command are left to its completion.
const fakeCommand = "bcuzfakecmd"

// completionCase is a line as typed, with the cursor at its end, and what the
// shells offer for it. The words are separated by one space; none has a space.
type completionCase struct {
	line string
	want []string
	// zsh is what the stubbed _normal of the zsh test prints, where the line
	// is left to the completion of the command ("_normal:CURRENT:words").
	zsh []string
	// needsCommandCompletion is set where bash needs bash-completion
	// (_command_offset) to use the completion of the command.
	needsCommandCompletion bool
}

// The table of docs/reference/cli.md ("--completion").
var completionCases = []completionCase{
	{line: "bcuz -", want: []string{"-H", "--history", "--init", "-h", "--help", "--version", "--completion", "--"}},
	{line: "bcuz --c", want: []string{"--completion"}},
	{line: "bcuz ", want: nil},
	{line: "bcuz wh", want: nil},
	{line: "bcuz --completion ", want: []string{"bash", "zsh", "fish", "powershell"}},
	{line: "bcuz --completion b", want: []string{"bash"}},
	{line: "bcuz --completion bash ", want: nil},
	{line: "bcuz -H ", want: []string{"-", "--date", "--since", "--until"}},
	{line: "bcuz --history --s", want: []string{"--since"}},
	{line: "bcuz -H --date ", want: nil},
	{line: "bcuz -H --date 2026-10-10 ", want: []string{"-", "--since", "--until"}},
	{line: "bcuz -H - --", want: []string{"--date", "--since", "--until"}},
	{line: "bcuz --init ", want: nil},
	{line: "bcuz -h ", want: nil},
	{line: "bcuz -x ", want: nil},
	{line: "bcuz -- ", want: nil},
	{line: "bcuz -- -x", want: nil},
	{line: "bcuz why bcuzfake", want: []string{fakeCommand}, zsh: []string{"_normal:1:bcuzfake"}},
	{line: "bcuz -- -x bcuzfake", want: []string{fakeCommand}, zsh: []string{"_normal:1:bcuzfake"}},
	{line: "bcuz why bcuzfakecmd a", want: []string{"alpha"}, zsh: []string{"_normal:2:bcuzfakecmd a"}, needsCommandCompletion: true},
	{line: "bcuz -- -x bcuzfakecmd ", want: []string{"alpha", "beta"}, zsh: []string{"_normal:2:bcuzfakecmd "}, needsCommandCompletion: true},
}

// shellEnv is an environment with the fake command first on the PATH, and the
// directory the shell runs in. The directory is empty, so that no file is
// offered where a shell falls back to files.
func shellEnv(t *testing.T) (env []string, dir string) {
	t.Helper()
	bin := t.TempDir()
	name := fakeCommand
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	if err := os.WriteFile(filepath.Join(bin, name), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	env = append(os.Environ(), "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"TERM=dumb", "POWERSHELL_UPDATECHECK=Off", "POWERSHELL_TELEMETRY_OPTOUT=1", "DOTNET_CLI_TELEMETRY_OPTOUT=1")
	return env, t.TempDir()
}

// runShell runs a shell with a script on its standard input and returns what
// it printed. A shell that is not installed skips the test.
func runShell(t *testing.T, name, script string, args ...string) string {
	t.Helper()
	path, err := exec.LookPath(name)
	if err != nil {
		t.Skipf("%s is not installed", name)
	}
	if name == "bash" {
		path = gitBash(path)
	}
	env, dir := shellEnv(t)
	cmd := exec.Command(path, args...)
	cmd.Env, cmd.Dir = env, dir
	cmd.Stdin = strings.NewReader(script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s: %v\n%s\n%s", name, err, stdout.String(), stderr.String())
	}
	if stderr.Len() > 0 {
		t.Logf("%s said on standard error:\n%s", name, stderr.String())
	}
	return stdout.String()
}

// gitBash is the bash to use on Windows, where the one the PATH finds first may
// be the launcher of the Windows Subsystem for Linux. Git for Windows has a
// bash of its own. Elsewhere it is the bash that was found.
func gitBash(found string) string {
	if runtime.GOOS != "windows" || !strings.Contains(strings.ToLower(found), `\system32\`) {
		return found
	}
	own := filepath.Join(os.Getenv("ProgramFiles"), "Git", "bin", "bash.exe")
	if _, err := os.Stat(own); err == nil {
		return own
	}
	return found
}

func completionScript(t *testing.T, shell string) string {
	t.Helper()
	code, stdout, stderr := runMain(t, "--completion", shell)
	if code != 0 {
		t.Fatalf("--completion %s: %d %s", shell, code, stderr)
	}
	return stdout
}

// parseAnswers reads the output of the cases: a line "### n" starts the answer
// of case n, and each line after it is a candidate (before a tab, if it has a
// description). On Windows the shells may offer the fake command with its
// .exe, which is taken off.
func parseAnswers(out string, count int) [][]string {
	answers := make([][]string, count)
	current := -1
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if rest, ok := strings.CutPrefix(line, "### "); ok {
			_, _ = fmt.Sscanf(rest, "%d", &current)
			continue
		}
		if current < 0 || line == "" {
			continue
		}
		value, _, _ := strings.Cut(line, "\t")
		if runtime.GOOS == "windows" {
			value = strings.TrimSuffix(value, ".exe")
		}
		answers[current] = append(answers[current], value)
	}
	return answers
}

func checkAnswers(t *testing.T, cases []completionCase, answers [][]string, want func(completionCase) []string) {
	t.Helper()
	for i, tc := range cases {
		got, w := slices.Clone(answers[i]), slices.Clone(want(tc))
		slices.Sort(got)
		slices.Sort(w)
		if !slices.Equal(got, w) {
			t.Errorf("%q:\n got  %q\n want %q", tc.line, got, w)
		}
	}
}

func wantAll(tc completionCase) []string { return tc.want }

func quoteSh(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }

// bashCases calls _bcuz as bash would, for each case, after setup.
func bashCases(t *testing.T, setup string, cases []completionCase) string {
	t.Helper()
	var b strings.Builder
	b.WriteString(setup)
	b.WriteString(completionScript(t, "bash"))
	for i, tc := range cases {
		words := strings.Split(tc.line, " ") // the last is empty when the line ends with a space
		quoted := make([]string, len(words))
		for j, w := range words {
			quoted[j] = quoteSh(w)
		}
		fmt.Fprintf(&b, "echo '### %d'\nCOMP_WORDS=(%s)\nCOMP_CWORD=%d\nCOMP_LINE=%s\nCOMP_POINT=${#COMP_LINE}\n_bcuz\nprintf '%%s\\n' \"${COMPREPLY[@]}\"\n",
			i, strings.Join(quoted, " "), len(words)-1, quoteSh(tc.line))
	}
	return runShell(t, "bash", b.String(), "--norc", "--noprofile")
}

// Without bash-completion, the command after the reason is completed from the
// command names; its arguments are files, and the directory has none.
func TestCompletionInBash(t *testing.T) {
	var cases []completionCase
	for _, tc := range completionCases {
		if !tc.needsCommandCompletion {
			cases = append(cases, tc)
		}
	}
	out := bashCases(t, "", cases)
	checkAnswers(t, cases, parseAnswers(out, len(cases)), wantAll)
}

// With bash-completion, the command after the reason, and its arguments, are
// left to _command_offset.
func TestCompletionInBashWithBashCompletion(t *testing.T) {
	const file = "/usr/share/bash-completion/bash_completion"
	if _, err := os.Stat(file); err != nil {
		t.Skip("bash-completion is not installed")
	}
	setup := ". " + file + "\ncomplete -W 'alpha beta' " + fakeCommand + "\n"
	out := bashCases(t, setup, completionCases)
	checkAnswers(t, completionCases, parseAnswers(out, len(completionCases)), wantAll)
}

// zsh needs its completion system to offer anything, which a script cannot
// start without a terminal. So _bcuz is called as the system would call it
// (with words and CURRENT set), compadd prints what it is given, and _normal
// prints the words it is left with.
func TestCompletionInZsh(t *testing.T) {
	script := completionScript(t, "zsh")
	var b strings.Builder
	b.WriteString("compdef() { :; }\n")
	// Like the real compadd, keep only the words that start with the word typed.
	b.WriteString("compadd() { [[ $1 == -- ]] && shift; local c p=${words[CURRENT]}; for c; do [[ ${c[1,${#p}]} == \"$p\" ]] && print -r -- \"$c\"; done; }\n")
	b.WriteString("_normal() { print -r -- \"_normal:$CURRENT:${(j: :)words}\"; }\n")
	b.WriteString(script)
	for i, tc := range completionCases {
		words := strings.Split(tc.line, " ")
		quoted := make([]string, len(words))
		for j, w := range words {
			quoted[j] = quoteSh(w)
		}
		fmt.Fprintf(&b, "echo '### %d'\n(words=(%s); CURRENT=%d; _bcuz)\n", i, strings.Join(quoted, " "), len(words))
	}
	out := runShell(t, "zsh", b.String(), "-f")
	checkAnswers(t, completionCases, parseAnswers(out, len(completionCases)), func(tc completionCase) []string {
		if tc.zsh != nil {
			return tc.zsh
		}
		return tc.want
	})
}

func TestCompletionInFish(t *testing.T) {
	quote := func(s string) string {
		return "'" + strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), "'", `\'`) + "'"
	}
	var b strings.Builder
	b.WriteString(completionScript(t, "fish"))
	b.WriteString("complete -c " + fakeCommand + " -f -a 'alpha beta'\n")
	for i, tc := range completionCases {
		fmt.Fprintf(&b, "echo '### %d'\ncomplete -C %s\n", i, quote(tc.line))
	}
	out := runShell(t, "fish", b.String(), "--no-config")
	checkAnswers(t, completionCases, parseAnswers(out, len(completionCases)), wantAll)
}

func TestCompletionInPowerShell(t *testing.T) {
	quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
	var b strings.Builder
	b.WriteString(completionScript(t, "powershell"))
	b.WriteString("\nRegister-ArgumentCompleter -Native -CommandName " + fakeCommand + " -ScriptBlock {\n" +
		"    param($word)\n" +
		"    'alpha', 'beta' | Where-Object { $_.StartsWith($word) } |\n" +
		"        ForEach-Object { [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_) }\n" +
		"}\n")
	for i, tc := range completionCases {
		fmt.Fprintf(&b, "'### %d'\n$line = %s\n(TabExpansion2 $line $line.Length).CompletionMatches | ForEach-Object { $_.CompletionText }\n",
			i, quote(tc.line))
	}
	out := runShell(t, "pwsh", b.String(), "-NoLogo", "-NoProfile", "-NonInteractive", "-Command", "-")
	checkAnswers(t, completionCases, parseAnswers(out, len(completionCases)), wantAll)
}
