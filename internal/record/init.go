package record

import (
	"errors"
	"os"
	"path/filepath"
)

// InitStatus says what Init did.
type InitStatus int

const (
	// Created means .bcuz/ did not exist in the directory and was made.
	Created InitStatus = iota
	// CreatedShadowing means the same as Created, and a .bcuz/ of a parent
	// directory is now hidden for this directory and below.
	CreatedShadowing
	// AlreadyExists means the directory already has .bcuz/; nothing changed.
	AlreadyExists
)

// InitResult is what Init did. Shadowed is the workspace root whose .bcuz/ is
// hidden, only when Status is CreatedShadowing.
type InitResult struct {
	Status   InitStatus
	Shadowed string
}

// Init makes .bcuz/ and an empty history.jsonl in dir. A .bcuz/ in dir is left
// as it is. A .bcuz/ in a parent does not stop it. The error is the plain
// reason (such as "permission denied"), without the path.
func Init(dir string) (InitResult, error) {
	path := filepath.Join(dir, DirName)
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		return InitResult{Status: AlreadyExists}, nil
	}

	res := InitResult{Status: Created}
	if parent, ok := FindRoot(filepath.Dir(dir)); ok {
		res = InitResult{Status: CreatedShadowing, Shadowed: parent}
	}

	//nolint:gosec // .bcuz/ is meant to be readable by others and by git
	if err := os.Mkdir(path, 0o755); err != nil {
		return InitResult{}, reason(err)
	}
	//nolint:gosec // same as above
	f, err := os.OpenFile(HistoryPath(dir), os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return InitResult{}, reason(err)
	}
	if err := f.Close(); err != nil {
		return InitResult{}, reason(err)
	}
	return res, nil
}

// reason returns the cause of a path error without the operation and path.
func reason(err error) error {
	var pe *os.PathError
	if errors.As(err, &pe) {
		return pe.Err
	}
	return err
}
