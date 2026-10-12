package docs

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/amisonnet8/bcuz/internal/record"
)

// checkRecordLines reads every non-empty line as a record and writes it
// again: the result must be the same line (the format of
// docs/reference/history.md: seven fields, in this order).
func checkRecordLines(t *testing.T, where, text string) {
	t.Helper()
	n := 0
	for i, line := range strings.Split(text, "\n") {
		if line == "" {
			continue
		}
		n++
		r, err := record.Unmarshal([]byte(line))
		if err != nil {
			t.Errorf("%s:%d: not a record: %v", where, i+1, err)
			continue
		}
		out, err := record.Marshal(r)
		if err != nil {
			t.Errorf("%s:%d: %v", where, i+1, err)
			continue
		}
		if string(out) != line+"\n" {
			t.Errorf("%s:%d: written again differently\n got: %s want: %s\n", where, i+1, out, line)
		}
	}
	if n == 0 {
		t.Errorf("%s: no records", where)
	}
}

func TestExampleHistoryRoundTrips(t *testing.T) {
	for _, name := range []string{"history.jsonl", "history_ja.jsonl"} {
		rel := "docs/examples/" + name
		text := readRepoFile(t, rel)
		if !strings.HasSuffix(text, "\n") {
			t.Errorf("%s does not end with a newline", rel)
		}
		checkRecordLines(t, rel, text)
	}
}

// Every json block in the documents is a record in the documented format.
func TestJSONBlocksAreRecords(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(repoRoot, "docs", "examples", "*.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"history.md", "history_ja.md"} {
		files = append(files, filepath.Join(repoRoot, "docs", "reference", name))
	}
	files = append(files, filepath.Join(repoRoot, "README.md"), filepath.Join(repoRoot, "README_ja.md"))

	blocks := 0
	for _, f := range files {
		rel, err := filepath.Rel(repoRoot, f)
		if err != nil {
			t.Fatal(err)
		}
		rel = filepath.ToSlash(rel)
		for i, block := range fencedBlocks(readRepoFile(t, rel), "json") {
			blocks++
			checkRecordLines(t, rel+" json block "+string(rune('1'+i)), block)
		}
	}
	if blocks == 0 {
		t.Error("no json blocks found")
	}
}
