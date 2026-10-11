package docs

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/amisonnet8/bcuz/internal/cli"
)

func runCLI(args ...string) (code int, stdout, stderr string) {
	var out, errOut bytes.Buffer
	code = cli.Main(args, strings.NewReader(""), &out, &errOut)
	return code, out.String(), errOut.String()
}

// The usage text of bcuz -h is the block in docs/reference/cli.md, character
// for character.
func TestUsageMatchesCLIReference(t *testing.T) {
	md := readRepoFile(t, "docs/reference/cli.md")
	want := fencedBlockAfter(t, md, "## `-h`・`--help`")

	for _, opt := range []string{"-h", "--help"} {
		code, stdout, stderr := runCLI(opt)
		if code != 0 || stderr != "" {
			t.Errorf("bcuz %s: exit code %d, stderr %q", opt, code, stderr)
		}
		if stdout != want {
			t.Errorf("bcuz %s: output differs from docs/reference/cli.md\n got: %q\nwant: %q", opt, stdout, want)
		}
	}
}

// bcuz --version prints "bcuz <version>" in the form the reference shows.
func TestVersionMatchesCLIReference(t *testing.T) {
	md := readRepoFile(t, "docs/reference/cli.md")
	example := fencedBlockAfter(t, md, "## `--version`")

	shape := regexp.MustCompile(`^bcuz v\d+\.\d+\.\d+\n$`)
	if !shape.MatchString(example) {
		t.Fatalf("version example in docs/reference/cli.md has an unexpected form: %q", example)
	}

	code, stdout, stderr := runCLI("--version")
	if code != 0 || stderr != "" {
		t.Errorf("bcuz --version: exit code %d, stderr %q", code, stderr)
	}
	// A local build has no module version and prints "(devel)".
	if !strings.HasPrefix(stdout, "bcuz ") || strings.Count(stdout, "\n") != 1 {
		t.Errorf("bcuz --version: unexpected output %q", stdout)
	}
}
