package docs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot is the repository root, seen from this package's directory.
const repoRoot = "../.."

func readRepoFile(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// fencedBlockAfter returns the content of the first fenced code block that
// follows the line equal to heading. It fails the test if there is none.
func fencedBlockAfter(t *testing.T, md, heading string) string {
	t.Helper()
	lines := strings.Split(md, "\n")
	i := 0
	for ; i < len(lines); i++ {
		if lines[i] == heading {
			break
		}
	}
	if i == len(lines) {
		t.Fatalf("heading %q not found", heading)
	}
	for i++; i < len(lines) && !strings.HasPrefix(lines[i], "```"); i++ {
	}
	var block []string
	for i++; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "```") {
			return strings.Join(block, "\n") + "\n"
		}
		block = append(block, lines[i])
	}
	t.Fatalf("no closed fenced block after %q", heading)
	return ""
}

// fencedBlocks returns the content of every fenced code block whose info
// string is lang (for example "json"), in order.
func fencedBlocks(md, lang string) []string {
	var blocks []string
	var cur []string
	in, match := false, false
	for _, line := range strings.Split(md, "\n") {
		switch {
		case !in && strings.HasPrefix(line, "```"):
			in, match, cur = true, strings.TrimPrefix(line, "```") == lang, nil
		case in && strings.HasPrefix(line, "```"):
			if match {
				blocks = append(blocks, strings.Join(cur, "\n")+"\n")
			}
			in = false
		case in:
			cur = append(cur, line)
		}
	}
	return blocks
}

// block is a fenced block of a document, with its info string.
type block struct {
	lang string
	text string
}

// allBlocks returns every fenced block of md, of any language, in order.
func allBlocks(md string) []block {
	var blocks []block
	var cur []string
	var lang string
	in := false
	for _, line := range strings.Split(md, "\n") {
		switch {
		case !in && strings.HasPrefix(line, "```"):
			in, lang, cur = true, strings.TrimPrefix(line, "```"), nil
		case in && strings.HasPrefix(line, "```"):
			blocks = append(blocks, block{lang, strings.Join(cur, "\n") + "\n"})
			in = false
		case in:
			cur = append(cur, line)
		}
	}
	return blocks
}

// stripCode removes fenced code blocks and inline code, so that what looks
// like a link inside code is not checked.
func stripCode(md string) string {
	var out []string
	inFence := false
	for _, line := range strings.Split(md, "\n") {
		if strings.HasPrefix(line, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		out = append(out, stripInlineCode(line))
	}
	return strings.Join(out, "\n")
}

func stripInlineCode(line string) string {
	var b strings.Builder
	in := false
	for _, r := range line {
		switch {
		case r == '`':
			in = !in
		case !in:
			b.WriteRune(r)
		}
	}
	return b.String()
}
