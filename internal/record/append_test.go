package record

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func newRecord(why string, ms int64) Record {
	return Record{
		Time: time.Date(2026, 10, 12, 6, 35, 12, 0, jst),
		Why:  why,
		Cmd:  []string{"make", "test"},
		Dir:  ".",
		Exit: 0,
		MS:   ms,
	}
}

// readAll reads every line of the history of root.
func readAll(t *testing.T, root string) []Record {
	t.Helper()
	f, err := os.Open(HistoryPath(root))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	var out []Record
	br := bufio.NewReader(f)
	for {
		line, err := br.ReadBytes('\n')
		if errors.Is(err, io.EOF) {
			if len(line) != 0 {
				t.Errorf("the last line has no newline: %.40q", line)
			}
			return out
		}
		if err != nil {
			t.Fatal(err)
		}
		r, uerr := Unmarshal(line)
		if uerr != nil {
			t.Fatalf("unreadable line %.60q: %v", line, uerr)
		}
		out = append(out, r)
	}
}

func TestAppend(t *testing.T) {
	root := t.TempDir()
	mkdir(t, root, DirName)

	for i, why := range []string{"first", "second", "third"} {
		if err := Append(root, newRecord(why, int64(i))); err != nil {
			t.Fatal(err)
		}
	}

	got := readAll(t, root)
	if len(got) != 3 || got[0].Why != "first" || got[1].Why != "second" || got[2].Why != "third" {
		t.Errorf("records = %+v", got)
	}
}

func TestAppendNeedsDirectory(t *testing.T) {
	if err := Append(t.TempDir(), newRecord("x", 0)); err == nil {
		t.Error("Append succeeded without .bcuz/")
	}
}

// Many writers append at once: every line must come out whole. Run it with
// qsoku race as well.
func TestAppendConcurrent(t *testing.T) {
	const (
		writers = 8
		perEach = 100
		size    = 8 << 10 // a long line: the write is not tiny
	)
	root := t.TempDir()
	mkdir(t, root, DirName)

	pad := strings.Repeat("x", size)
	var wg sync.WaitGroup
	errs := make(chan error, writers*perEach)
	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perEach; i++ {
				why := strconv.Itoa(w) + "-" + strconv.Itoa(i) + "-" + pad
				if err := Append(root, newRecord(why, int64(i))); err != nil {
					errs <- err
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}

	seen := make(map[string]bool)
	for _, r := range readAll(t, root) {
		if !strings.HasSuffix(r.Why, pad) {
			t.Fatalf("a line is damaged: %.40q", r.Why)
		}
		seen[strings.TrimSuffix(r.Why, pad)] = true
	}
	if len(seen) != writers*perEach {
		t.Errorf("distinct lines = %d, want %d", len(seen), writers*perEach)
	}
}
