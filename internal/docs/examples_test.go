package docs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A transcript is a fenced block of an example: lines that start with "$ "
// are commands, and the lines that follow are what they print. Only bcuz
// (optionally with TZ=UTC in front, and "< file" at the end for standard
// input) and "echo $?" are understood, and arguments are split at spaces
// (no quotes).

// runTranscript runs the commands of block in the current directory and
// returns the transcript it makes, in the same form.
func runTranscript(t *testing.T, block string) string {
	t.Helper()
	var out strings.Builder
	lastCode := 0
	for _, line := range strings.Split(strings.TrimSuffix(block, "\n"), "\n") {
		cmd, ok := strings.CutPrefix(line, "$ ")
		if !ok {
			// Expected output: the run produces its own, compared at the end.
			continue
		}
		out.WriteString(line + "\n")

		if cmd == "echo $?" {
			fmt.Fprintf(&out, "%d\n", lastCode)
			continue
		}
		zone := ""
		if rest, ok := strings.CutPrefix(cmd, "TZ=UTC "); ok {
			cmd, zone = rest, "UTC"
		}
		args, ok := strings.CutPrefix(cmd, "bcuz ")
		if !ok {
			t.Fatalf("unsupported command: %q", line)
		}
		input := ""
		if rest, file, ok := strings.Cut(args, " < "); ok {
			b, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			args, input = rest, string(b)
		}
		var restore func()
		if zone == "UTC" {
			old := time.Local
			time.Local = time.UTC
			restore = func() { time.Local = old }
		}
		code, stdout, stderr := runCLIWithInput(input, strings.Fields(args)...)
		if restore != nil {
			restore()
		}
		out.WriteString(stdout + stderr)
		lastCode = code
	}
	return out.String()
}

// tokyoZone makes Asia/Tokyo the zone of the viewer for the test: the
// records of docs/examples/history*.jsonl were written there.
func tokyoZone(t *testing.T) {
	t.Helper()
	old := time.Local
	time.Local = time.FixedZone("Asia/Tokyo", 9*3600)
	t.Cleanup(func() { time.Local = old })
}

// historyExampleWorkspace is a workspace whose history is the given file of
// docs/examples/ (history.jsonl, or history_ja.jsonl for the Japanese text).
// The same records, moved away, are archive.jsonl next to .bcuz/.
func historyExampleWorkspace(t *testing.T, jsonl string) {
	t.Helper()
	work := t.TempDir()
	if err := os.Mkdir(filepath.Join(work, ".bcuz"), 0o750); err != nil {
		t.Fatal(err)
	}
	text := readRepoFile(t, "docs/examples/"+jsonl)
	for _, path := range []string{filepath.Join(work, ".bcuz", "history.jsonl"), filepath.Join(work, "archive.jsonl")} {
		if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(work)
}

// historyExamples are the two languages of the examples: the document, the
// records it uses, and three reasons of those records that the UTC test
// looks for.
var historyExamples = []struct {
	md, jsonl             string
	tagReason, testReason string
	npmReason             string
}{
	{"history-filter.md", "history.jsonl", "check the tags", "full tests before release", "reinstall dependencies"},
	{"history-filter_ja.md", "history_ja.jsonl", "タグを確認", "リリース前の全テスト", "依存を入れ直す"},
}

// Every block of history-filter.md is what -H prints for the records of
// history.jsonl (and the same for the Japanese pair). The records were written
// in Asia/Tokyo, which is the zone of the viewer unless the block says TZ=UTC.
func TestHistoryFilterExamples(t *testing.T) {
	for _, ex := range historyExamples {
		t.Run(ex.md, func(t *testing.T) {
			blocks := fencedBlocks(readRepoFile(t, "docs/examples/"+ex.md), "")
			if len(blocks) < 7 {
				t.Fatalf("found %d blocks in %s, want at least 7", len(blocks), ex.md)
			}
			for i, block := range blocks {
				t.Run(fmt.Sprintf("block %d", i+1), func(t *testing.T) {
					tokyoZone(t)
					historyExampleWorkspace(t, ex.jsonl)
					if got := runTranscript(t, block); got != block {
						t.Errorf("block does not match.\n--- got\n%s--- want\n%s", got, block)
					}
				})
			}
		})
	}
}

// The same records seen from UTC: the dates of the first and last records
// of the examples fall on the day before, and the filter follows the
// viewer's day.
func TestHistoryFilterExamplesInUTC(t *testing.T) {
	for _, ex := range historyExamples {
		t.Run(ex.jsonl, func(t *testing.T) {
			old := time.Local
			time.Local = time.UTC
			t.Cleanup(func() { time.Local = old })
			historyExampleWorkspace(t, ex.jsonl)

			// 10-10 18:02:44+09:00 is 09:02:44Z; 10-11 09:15:30+09:00 is 10-11 00:15:30Z.
			code, stdout, stderr := runCLI("-H", "--date", "2026-10-10")
			want := "" +
				"2026-10-10 09:02:44    0   35.0s  .  " + ex.testReason + "  $ make test\n" +
				"2026-10-10 08:55:01    0   950ms  .  " + ex.tagReason + "  $ git tag --list 'v0.*'\n"
			if code != 0 || stderr != "" || stdout != want {
				t.Errorf("code %d, stdout %q, stderr %q; want 0, %q", code, stdout, stderr, want)
			}

			_, all, _ := runCLI("-H", "--since", "2026-10-11", "--until", "2026-10-11")
			last := "2026-10-11 00:15:30    0    4.0s  frontend  " + ex.npmReason + "  $ npm ci\n"
			if !strings.HasPrefix(all, "2026-10-11 21:52:10") || !strings.HasSuffix(all, last) {
				t.Errorf("11th in UTC: %q", all)
			}
		})
	}
}
