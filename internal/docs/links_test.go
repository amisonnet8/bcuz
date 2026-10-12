package docs

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode"
)

var linkRE = regexp.MustCompile(`\[[^\]]*\]\(([^)\s]+)\)`)

// markdownFiles lists README.md and every .md under docs/, relative to the
// repository root with forward slashes.
func markdownFiles(t *testing.T) []string {
	t.Helper()
	files := []string{"README.md", "README_ja.md"}
	root := filepath.Join(repoRoot, "docs")
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
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
	return files
}

// isExternal reports whether target is not a path inside the repository.
func isExternal(target string) bool {
	return strings.Contains(target, "://") ||
		strings.HasPrefix(target, "mailto:") ||
		strings.HasPrefix(target, "#")
}

// relativeLinks returns the in-repository link targets of md, with any
// #fragment removed. Links inside code are not links.
func relativeLinks(md string) []string {
	var targets []string
	for _, m := range linkRE.FindAllStringSubmatch(stripCode(md), -1) {
		if isExternal(m[1]) {
			continue
		}
		target, _, _ := strings.Cut(m[1], "#")
		targets = append(targets, target)
	}
	return targets
}

// missingLinks returns the relative links of md (the content of the file at
// slash-separated path file under root) that do not point at an existing file.
func missingLinks(root, file, md string) []string {
	var missing []string
	for _, target := range relativeLinks(md) {
		path := filepath.Join(root, filepath.Dir(filepath.FromSlash(file)), filepath.FromSlash(target))
		if _, err := os.Stat(path); err != nil {
			missing = append(missing, target)
		}
	}
	return missing
}

// devLinks returns the links of md (the content of file) that point into dev/.
func devLinks(file, md string) []string {
	var found []string
	for _, target := range relativeLinks(md) {
		resolved := filepath.ToSlash(filepath.Join(filepath.Dir(file), filepath.FromSlash(target)))
		if resolved == "dev" || strings.HasPrefix(resolved, "dev/") {
			found = append(found, target)
		}
	}
	return found
}

func TestRelativeLinksExist(t *testing.T) {
	for _, file := range markdownFiles(t) {
		for _, target := range missingLinks(repoRoot, file, readRepoFile(t, file)) {
			t.Errorf("%s: broken link %q", file, target)
		}
	}
}

// docs/ is shown to users, so it must not point into dev/
// (.claude/rules/documentation.md).
func TestDocsDoNotLinkToDev(t *testing.T) {
	for _, file := range markdownFiles(t) {
		if !strings.HasPrefix(file, "docs/") {
			continue
		}
		for _, target := range devLinks(file, readRepoFile(t, file)) {
			t.Errorf("%s: links to dev/: %q", file, target)
		}
	}
}

// The checks above pass on a healthy repository, so these cases make sure
// they can fail.
func TestLinkChecksDetectProblems(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "docs", "reference"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "reference", "ok.md"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	md := "[ok](ok.md) [frag](ok.md#top) [gone](gone.md) [web](https://example.com/x.md) " +
		"[self](#top) `[code](code.md)`\n\n```\n[fenced](fenced.md)\n```\n"
	got := missingLinks(root, "docs/reference/cli.md", md)
	if len(got) != 1 || got[0] != "gone.md" {
		t.Errorf("missingLinks = %q, want [gone.md]", got)
	}

	dev := devLinks("docs/reference/cli.md", "[a](../../dev/roadmap.md) [b](../examples/x.md) [c](https://e.com/dev/x)")
	if len(dev) != 1 || dev[0] != "../../dev/roadmap.md" {
		t.Errorf("devLinks = %q, want [../../dev/roadmap.md]", dev)
	}
}

var (
	htmlAnchorRE = regexp.MustCompile(`href="#([^"]+)"`)
	mdAnchorRE   = regexp.MustCompile(`\]\(#([^)\s]+)\)`)
)

// anchorOf is the anchor GitHub makes of a heading: lower case, spaces become
// hyphens, and everything but letters, digits, "-" and "_" is dropped.
func anchorOf(heading string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(heading)) {
		switch {
		case r == ' ':
			b.WriteByte('-')
		case r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		}
	}
	return b.String()
}

// anchorsOf returns the anchors of the headings of md (outside code blocks).
func anchorsOf(md string) map[string]bool {
	anchors := map[string]bool{}
	inFence := false
	for _, line := range strings.Split(md, "\n") {
		switch {
		case strings.HasPrefix(line, "```"):
			inFence = !inFence
		case !inFence && strings.HasPrefix(line, "#"):
			anchors[anchorOf(strings.ReplaceAll(strings.TrimLeft(line, "#"), "`", ""))] = true
		}
	}
	return anchors
}

// missingAnchors returns the in-page links (#anchor) of md that match no
// heading of md.
func missingAnchors(md string) []string {
	anchors := anchorsOf(md)
	var missing []string
	for _, re := range []*regexp.Regexp{htmlAnchorRE, mdAnchorRE} {
		for _, m := range re.FindAllStringSubmatch(stripCode(md), -1) {
			if !anchors[m[1]] {
				missing = append(missing, m[1])
			}
		}
	}
	return missing
}

// The links of the table of contents at the top of a README point at
// headings that exist.
func TestAnchorLinksExist(t *testing.T) {
	for _, file := range markdownFiles(t) {
		for _, a := range missingAnchors(readRepoFile(t, file)) {
			t.Errorf("%s: no heading for the link #%s", file, a)
		}
	}
}

func TestAnchorChecksDetectProblems(t *testing.T) {
	for in, want := range map[string]string{
		"At a glance":        "at-a-glance",
		"Exit codes":         "exit-codes",
		"`--init`":           "--init",
		"ひと目で":               "ひと目で",
		"Good to know (now)": "good-to-know-now",
	} {
		if got := anchorOf(strings.ReplaceAll(in, "`", "")); got != want {
			t.Errorf("anchorOf(%q) = %q, want %q", in, got, want)
		}
	}
	md := "# T\n\n## At a glance\n\n<a href=\"#at-a-glance\">ok</a> [x](#at-a-glance) <a href=\"#gone\">bad</a> [y](#gone2)\n\n```\n## Hidden\n```\n"
	got := missingAnchors(md)
	if len(got) != 2 || got[0] != "gone" || got[1] != "gone2" {
		t.Errorf("missingAnchors = %q, want [gone gone2]", got)
	}
}
