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

// cliReferences are the two languages of the command reference, with the
// headings that the tests look for.
var cliReferences = []struct {
	file        string
	helpHeading string
	versionHead string
}{
	{"docs/reference/cli.md", "## `-h`, `--help`", "## `--version`"},
	{"docs/reference/cli_ja.md", "## `-h`・`--help`", "## `--version`"},
}

// The usage text of bcuz -h is the block in docs/reference/cli.md (and in
// its Japanese translation), character for character.
func TestUsageMatchesCLIReference(t *testing.T) {
	for _, ref := range cliReferences {
		t.Run(ref.file, func(t *testing.T) {
			md := readRepoFile(t, ref.file)
			want := fencedBlockAfter(t, md, ref.helpHeading)

			for _, opt := range []string{"-h", "--help"} {
				code, stdout, stderr := runCLI(opt)
				if code != 0 || stderr != "" {
					t.Errorf("bcuz %s: exit code %d, stderr %q", opt, code, stderr)
				}
				if stdout != want {
					t.Errorf("bcuz %s: output differs from %s\n got: %q\nwant: %q", opt, ref.file, stdout, want)
				}
			}
		})
	}
}

// bcuz --version prints "bcuz <version>" in the form the reference shows.
func TestVersionMatchesCLIReference(t *testing.T) {
	shape := regexp.MustCompile(`^bcuz v\d+\.\d+\.\d+\n$`)
	for _, ref := range cliReferences {
		example := fencedBlockAfter(t, readRepoFile(t, ref.file), ref.versionHead)
		if !shape.MatchString(example) {
			t.Fatalf("version example in %s has an unexpected form: %q", ref.file, example)
		}
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
