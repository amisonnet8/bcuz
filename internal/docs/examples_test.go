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
// (optionally with TZ=UTC in front) and "echo $?" are understood, and
// arguments are split at spaces (no quotes).

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
		var restore func()
		if zone == "UTC" {
			old := time.Local
			time.Local = time.UTC
			restore = func() { time.Local = old }
		}
		code, stdout, stderr := runCLI(strings.Fields(args)...)
		if restore != nil {
			restore()
		}
		out.WriteString(stdout + stderr)
		lastCode = code
	}
	return out.String()
}

// tokyoZone makes Asia/Tokyo the zone of the viewer for the test: the
// records of docs/examples/history.jsonl were written there.
func tokyoZone(t *testing.T) {
	t.Helper()
	old := time.Local
	time.Local = time.FixedZone("Asia/Tokyo", 9*3600)
	t.Cleanup(func() { time.Local = old })
}

// historyExampleWorkspace is a workspace whose history is
// docs/examples/history.jsonl.
func historyExampleWorkspace(t *testing.T) {
	t.Helper()
	work := t.TempDir()
	if err := os.Mkdir(filepath.Join(work, ".bcuz"), 0o750); err != nil {
		t.Fatal(err)
	}
	text := readRepoFile(t, "docs/examples/history.jsonl")
	if err := os.WriteFile(filepath.Join(work, ".bcuz", "history.jsonl"), []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(work)
}

// Every block of history-filter.md is what -H prints for
// docs/examples/history.jsonl. The records were written in Asia/Tokyo, which
// is the zone of the viewer unless the block says TZ=UTC.
func TestHistoryFilterExamples(t *testing.T) {
	md := readRepoFile(t, "docs/examples/history-filter.md")
	blocks := fencedBlocks(md, "")
	if len(blocks) < 7 {
		t.Fatalf("found %d blocks in history-filter.md, want at least 7", len(blocks))
	}
	for i, block := range blocks {
		t.Run(fmt.Sprintf("block %d", i+1), func(t *testing.T) {
			tokyoZone(t)
			historyExampleWorkspace(t)
			if got := runTranscript(t, block); got != block {
				t.Errorf("block does not match.\n--- got\n%s--- want\n%s", got, block)
			}
		})
	}
}

// The same records seen from UTC: the dates of the first and last records
// of the examples fall on the day before, and the filter follows the
// viewer's day.
func TestHistoryFilterExamplesInUTC(t *testing.T) {
	old := time.Local
	time.Local = time.UTC
	t.Cleanup(func() { time.Local = old })
	historyExampleWorkspace(t)

	// 10-10 18:02:44+09:00 is 09:02:44Z; 10-11 09:15:30+09:00 is 10-11 00:15:30Z.
	code, stdout, stderr := runCLI("-H", "--date", "2026-10-10")
	want := "" +
		"2026-10-10 09:02:44    0   35.0s  .  リリース前の全テスト  $ make test\n" +
		"2026-10-10 08:55:01    0   950ms  .  タグを確認  $ git tag --list 'v0.*'\n"
	if code != 0 || stderr != "" || stdout != want {
		t.Errorf("code %d, stdout %q, stderr %q; want 0, %q", code, stdout, stderr, want)
	}

	// The first line of the example's "all" block, in UTC.
	_, all, _ := runCLI("-H", "--since", "2026-10-11", "--until", "2026-10-11")
	if !strings.HasPrefix(all, "2026-10-11 21:52:10") || !strings.HasSuffix(all, "2026-10-11 00:15:30    0    4.0s  frontend  依存を入れ直す  $ npm ci\n") {
		t.Errorf("11th in UTC: %q", all)
	}
}
