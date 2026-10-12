package docs

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The English file X.md is the source of truth, and X_ja.md is its Japanese
// translation (.claude/rules/documentation.md). These tests keep the pair in
// the same shape; what the words say is for a person to read.

const jaSuffix = "_ja"

// jaName returns the Japanese counterpart of an English file name.
func jaName(en string) string {
	ext := filepath.Ext(en)
	return strings.TrimSuffix(en, ext) + jaSuffix + ext
}

// translatedFiles lists the English files that have a Japanese pair: the
// README.md, and under docs/ every .md and the .jsonl of the examples.
func translatedFiles(t *testing.T) []string {
	t.Helper()
	var files []string
	root := filepath.Join(repoRoot, "docs")
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		ext := filepath.Ext(path)
		if (ext == ".md" || ext == ".jsonl") && !strings.HasSuffix(strings.TrimSuffix(path, ext), jaSuffix) {
			rel, err := filepath.Rel(repoRoot, path)
			if err != nil {
				return err
			}
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no documents found")
	}
	// The README at the top is a pair like the others.
	return append([]string{"README.md"}, files...)
}

func exists(t *testing.T, rel string) bool {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(repoRoot, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return len(matches) == 1
}

// Every English file has a Japanese one, and every Japanese file an English one.
func TestEveryDocumentHasATranslation(t *testing.T) {
	for _, en := range translatedFiles(t) {
		if !exists(t, jaName(en)) {
			t.Errorf("%s has no Japanese translation %s", en, jaName(en))
		}
	}

	root := filepath.Join(repoRoot, "docs")
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		ext := filepath.Ext(path)
		base := strings.TrimSuffix(path, ext)
		if strings.HasSuffix(base, jaSuffix) {
			en := strings.TrimSuffix(base, jaSuffix) + ext
			rel, _ := filepath.Rel(repoRoot, en)
			if !exists(t, filepath.ToSlash(rel)) {
				t.Errorf("%s has no English original %s", path, rel)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// outline is the shape of a document: the levels of its headings, its fenced
// blocks (language and content), and the number of table rows and list items.
type outline struct {
	headings []int
	blocks   []block
	rows     int
	items    int
}

func outlineOf(md string) outline {
	var o outline
	o.blocks = allBlocks(md)
	inFence := false
	for _, line := range strings.Split(md, "\n") {
		switch {
		case strings.HasPrefix(line, "```"):
			inFence = !inFence
		case inFence:
		case strings.HasPrefix(line, "#"):
			o.headings = append(o.headings, len(line)-len(strings.TrimLeft(line, "#")))
		case strings.HasPrefix(line, "|"):
			o.rows++
		case strings.HasPrefix(line, "- ") || regexp.MustCompile(`^\d+\. `).MatchString(line):
			o.items++
		}
	}
	return o
}

func isASCII(s string) bool {
	for _, r := range s {
		if r > 127 {
			return false
		}
	}
	return true
}

// The two languages have the same headings, tables, lists and code blocks. A
// code block of the Japanese file that has no Japanese in it (a message, a
// usage text, a command) is the same character for character in English.
func TestTranslationsHaveTheSameStructure(t *testing.T) {
	for _, en := range translatedFiles(t) {
		if !strings.HasSuffix(en, ".md") {
			continue
		}
		ja := jaName(en)
		if !exists(t, ja) {
			continue // reported by TestEveryDocumentHasATranslation
		}
		t.Run(en, func(t *testing.T) {
			e, j := outlineOf(readRepoFile(t, en)), outlineOf(readRepoFile(t, ja))
			if len(e.headings) != len(j.headings) {
				t.Fatalf("%d headings in %s, %d in %s", len(e.headings), en, len(j.headings), ja)
			}
			for i := range e.headings {
				if e.headings[i] != j.headings[i] {
					t.Errorf("heading %d is level %d in %s and level %d in %s", i+1, e.headings[i], en, j.headings[i], ja)
				}
			}
			if e.rows != j.rows {
				t.Errorf("%d table rows in %s, %d in %s", e.rows, en, j.rows, ja)
			}
			if e.items != j.items {
				t.Errorf("%d list items in %s, %d in %s", e.items, en, j.items, ja)
			}
			if len(e.blocks) != len(j.blocks) {
				t.Fatalf("%d code blocks in %s, %d in %s", len(e.blocks), en, len(j.blocks), ja)
			}
			for i := range e.blocks {
				if e.blocks[i].lang != j.blocks[i].lang {
					t.Errorf("code block %d: language %q in %s, %q in %s", i+1, e.blocks[i].lang, en, j.blocks[i].lang, ja)
				}
				if isASCII(j.blocks[i].text) && e.blocks[i].text != j.blocks[i].text {
					t.Errorf("code block %d differs between %s and %s\n en: %q\n ja: %q", i+1, en, ja, e.blocks[i].text, j.blocks[i].text)
				}
			}
		})
	}
}

// The first line of each file switches to the other language.
func TestLanguageSwitch(t *testing.T) {
	for _, en := range translatedFiles(t) {
		if !strings.HasSuffix(en, ".md") {
			continue
		}
		ja := jaName(en)
		wantEn := "*[日本語](" + filepath.Base(ja) + ") | **English***"
		wantJa := "*[English](" + filepath.Base(en) + ") | **日本語***"
		for _, c := range []struct{ file, want string }{{en, wantEn}, {ja, wantJa}} {
			lines := strings.SplitN(readRepoFile(t, c.file), "\n", 3)
			if len(lines) < 3 || lines[0] != c.want || lines[1] != "" || !strings.HasPrefix(lines[2], "# ") {
				t.Errorf("%s: the file must start with %q, an empty line and the title", c.file, c.want)
			}
		}
	}
}

// A file links to the files of its own language. The switch on the first
// line is the only link from English to Japanese and back.
func TestLinksStayInTheLanguage(t *testing.T) {
	for _, en := range translatedFiles(t) {
		if !strings.HasSuffix(en, ".md") {
			continue
		}
		ja := jaName(en)

		body := func(md string) string {
			_, rest, _ := strings.Cut(md, "\n")
			return rest // without the switch
		}
		for _, target := range relativeLinks(body(readRepoFile(t, en))) {
			if strings.Contains(target, jaSuffix+".") {
				t.Errorf("%s links to the Japanese file %s", en, target)
			}
		}
		for _, target := range relativeLinks(body(readRepoFile(t, ja))) {
			ext := filepath.Ext(target)
			if (ext == ".md" || ext == ".jsonl") && !strings.HasSuffix(strings.TrimSuffix(target, ext), jaSuffix) {
				t.Errorf("%s links to the English file %s", ja, target)
			}
		}
	}
}
