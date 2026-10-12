package record

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mkdir(t *testing.T, elem ...string) string {
	t.Helper()
	p := filepath.Join(elem...)
	if err := os.MkdirAll(p, 0o750); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestFindRoot(t *testing.T) {
	t.Run("the nearest .bcuz/ wins", func(t *testing.T) {
		tmp := t.TempDir()
		mkdir(t, tmp, "a", ".bcuz")
		mkdir(t, tmp, "a", "b", ".bcuz")
		deep := mkdir(t, tmp, "a", "b", "c")

		got, ok := FindRoot(deep)
		if want := filepath.Join(tmp, "a", "b"); !ok || got != want {
			t.Errorf("FindRoot = %q, %v; want %q", got, ok, want)
		}
	})

	t.Run("the directory itself", func(t *testing.T) {
		tmp := t.TempDir()
		mkdir(t, tmp, ".bcuz")
		got, ok := FindRoot(tmp)
		if !ok || got != tmp {
			t.Errorf("FindRoot = %q, %v; want %q", got, ok, tmp)
		}
	})

	t.Run("a file named .bcuz is ignored", func(t *testing.T) {
		tmp := t.TempDir()
		mkdir(t, tmp, "a", ".bcuz")
		sub := mkdir(t, tmp, "a", "b")
		if err := os.WriteFile(filepath.Join(sub, ".bcuz"), nil, 0o600); err != nil {
			t.Fatal(err)
		}

		got, ok := FindRoot(sub)
		if want := filepath.Join(tmp, "a"); !ok || got != want {
			t.Errorf("FindRoot = %q, %v; want %q", got, ok, want)
		}
	})

	t.Run("not found stops at the top", func(t *testing.T) {
		tmp := t.TempDir()
		sub := mkdir(t, tmp, "a", "b")
		// A .bcuz/ above the temporary directory may exist on a machine; it
		// is not an error, but nothing below tmp may be found.
		got, ok := FindRoot(sub)
		if ok && strings.HasPrefix(got, tmp) {
			t.Errorf("FindRoot = %q; nothing should be found inside %q", got, tmp)
		}
	})
}

func TestRelDir(t *testing.T) {
	root := filepath.Join(t.TempDir(), "work")
	tests := []struct {
		cwd  string
		want string
	}{
		{root, "."},
		{filepath.Join(root, "backend"), "backend"},
		{filepath.Join(root, "a", "b"), "a/b"},
	}
	for _, tt := range tests {
		got, err := RelDir(root, tt.cwd)
		if err != nil || got != tt.want {
			t.Errorf("RelDir(%q) = %q, %v; want %q", tt.cwd, got, err, tt.want)
		}
	}
}

func TestInit(t *testing.T) {
	t.Run("created", func(t *testing.T) {
		tmp := t.TempDir()
		res, err := Init(tmp)
		if err != nil {
			t.Fatal(err)
		}
		// A .bcuz/ above the temporary directory is not ours to judge.
		if res.Status != Created && res.Status != CreatedShadowing {
			t.Errorf("Status = %v", res.Status)
		}
		b, err := os.ReadFile(HistoryPath(tmp))
		if err != nil || len(b) != 0 {
			t.Errorf("history.jsonl = %q, %v; want an empty file", b, err)
		}
	})

	t.Run("shadows the parent", func(t *testing.T) {
		tmp := t.TempDir()
		mkdir(t, tmp, ".bcuz")
		sub := mkdir(t, tmp, "a", "b")

		res, err := Init(sub)
		if err != nil {
			t.Fatal(err)
		}
		if res.Status != CreatedShadowing || res.Shadowed != tmp {
			t.Errorf("result = %+v; want CreatedShadowing of %q", res, tmp)
		}
	})

	t.Run("already exists: nothing changes", func(t *testing.T) {
		tmp := t.TempDir()
		dir := mkdir(t, tmp, ".bcuz")

		res, err := Init(tmp)
		if err != nil || res.Status != AlreadyExists {
			t.Fatalf("result = %+v, %v", res, err)
		}
		if _, err := os.Stat(filepath.Join(dir, FileName)); !os.IsNotExist(err) {
			t.Errorf("history.jsonl was created: %v", err)
		}

		const content = "keep\n"
		if err := os.WriteFile(filepath.Join(dir, FileName), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		if res, err = Init(tmp); err != nil || res.Status != AlreadyExists {
			t.Fatalf("result = %+v, %v", res, err)
		}
		b, err := os.ReadFile(filepath.Join(dir, FileName)) //nolint:gosec // a file this test made in t.TempDir()
		if err != nil || string(b) != content {
			t.Errorf("history.jsonl = %q, %v; want %q", b, err, content)
		}
	})

	t.Run("a file named .bcuz", func(t *testing.T) {
		tmp := t.TempDir()
		if err := os.WriteFile(filepath.Join(tmp, ".bcuz"), nil, 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := Init(tmp)
		if err == nil {
			t.Fatal("Init succeeded")
		}
		if strings.Contains(err.Error(), tmp) {
			t.Errorf("error %q has the path; want the plain reason", err)
		}
	})
}
