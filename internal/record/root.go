package record

import (
	"os"
	"path/filepath"
)

// DirName is the name of the directory that holds the history.
const DirName = ".bcuz"

// FileName is the history file inside DirName.
const FileName = "history.jsonl"

// FindRoot looks for a .bcuz directory in start and then in each parent, and
// returns the directory that contains the nearest one (the workspace). A file
// that is named .bcuz is ignored.
func FindRoot(start string) (root string, ok bool) {
	dir := start
	for {
		if fi, err := os.Stat(filepath.Join(dir, DirName)); err == nil && fi.IsDir() {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir { // the root of the file system, or of a drive
			return "", false
		}
		dir = parent
	}
}

// RelDir returns cwd relative to root, with "/" as the separator on every OS.
// The workspace itself is ".".
func RelDir(root, cwd string) (string, error) {
	rel, err := filepath.Rel(root, cwd)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
}

// HistoryPath is the history file of the workspace root.
func HistoryPath(root string) string {
	return filepath.Join(root, DirName, FileName)
}
