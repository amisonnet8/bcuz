package record

import (
	"fmt"
	"os"
)

// Append writes r as one line to the history of the workspace root. The line
// is written with a single Write on a file opened with O_APPEND, so lines of
// concurrent bcuz processes do not mix.
func Append(root string, r Record) (err error) {
	b, err := Marshal(r)
	if err != nil {
		return err
	}
	//nolint:gosec // the history is meant to be readable by others and by git (docs/reference/history.md)
	f, err := os.OpenFile(HistoryPath(root), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()
	n, err := f.Write(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		return fmt.Errorf("short write: %d of %d bytes", n, len(b))
	}
	return nil
}
