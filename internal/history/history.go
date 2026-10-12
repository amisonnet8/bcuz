// Package history reads history.jsonl and shows it: the filter by date, the
// order, and the one-line form of -H (docs/reference/cli.md). It reads what
// package record parses and never writes.
package history

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/amisonnet8/bcuz/internal/record"
)

// dateLayout is YYYY-MM-DD.
const dateLayout = "2006-01-02"

// timeLayout is the date and time column, in the viewer's time zone.
const timeLayout = "2006-01-02 15:04:05"

// ParseDate checks that s is a real date written as YYYY-MM-DD.
func ParseDate(s string) error {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return err
	}
	// Be strict about the written form: it must read back the same.
	if t.Format(dateLayout) != s {
		return errors.New("not in YYYY-MM-DD form")
	}
	return nil
}

// Read reads every record in r. Lines that cannot be read are skipped and
// their numbers (1-based) are returned. A last line that does not end with a
// newline is a write in progress: it is skipped without a mention.
func Read(r io.Reader) (recs []record.Record, skipped []int, err error) {
	br := bufio.NewReader(r)
	for n := 1; ; n++ {
		b, readErr := br.ReadBytes('\n')
		if readErr != nil {
			if readErr == io.EOF {
				return recs, skipped, nil
			}
			return recs, skipped, readErr
		}
		rec, perr := record.Unmarshal(b)
		if perr != nil {
			skipped = append(skipped, n)
			continue
		}
		recs = append(recs, rec)
	}
}

// Filter limits records by the date they were started on, in the time zone
// time.Local. An empty field means no limit. Dates are YYYY-MM-DD.
type Filter struct {
	Date  string
	Since string
	Until string
}

// Match reports whether r passes the filter. Dates in YYYY-MM-DD compare as
// text in the right order.
func (f Filter) Match(r record.Record) bool {
	day := r.Time.In(time.Local).Format(dateLayout)
	if f.Date != "" && day != f.Date {
		return false
	}
	if f.Since != "" && day < f.Since {
		return false
	}
	if f.Until != "" && day > f.Until {
		return false
	}
	return true
}

// Select returns the records that pass f, newest first. Records with the same
// start time keep the later line first.
func Select(recs []record.Record, f Filter) []record.Record {
	var out []record.Record
	for _, r := range recs {
		if f.Match(r) {
			out = append(out, r)
		}
	}
	// Reverse first, then a stable sort: equal times end with the later line first.
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Time.After(out[j].Time)
	})
	return out
}

// FormatLine returns the one-line form of r, without a newline.
func FormatLine(r record.Record) string {
	args := make([]string, len(r.Cmd))
	for i, a := range r.Cmd {
		args[i] = QuoteArg(a)
	}
	return fmt.Sprintf("%s  %3d  %6s  %s  %s  $ %s",
		r.Time.In(time.Local).Format(timeLayout),
		r.Exit,
		FormatDuration(r.MS),
		r.Dir,
		oneLine(r.Why),
		strings.Join(args, " "))
}

// oneLine replaces each line break and tab with one space.
func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	return strings.Map(func(c rune) rune {
		if c == '\n' || c == '\r' || c == '\t' {
			return ' '
		}
		return c
	}, s)
}

// FormatDuration writes ms as 123ms, 8.3s, 2m05s or 1h02m.
func FormatDuration(ms int64) string {
	switch {
	case ms < 1000:
		return fmt.Sprintf("%dms", ms)
	case ms < 60000:
		return fmt.Sprintf("%d.%ds", ms/1000, ms%1000/100)
	case ms < 3600000:
		s := ms / 1000
		return fmt.Sprintf("%dm%02ds", s/60, s%60)
	default:
		m := ms / 60000
		return fmt.Sprintf("%dh%02dm", m/60, m%60)
	}
}

// QuoteArg writes s so that a shell reads it back as the same argument.
// Arguments made only of ASCII letters, digits and -_./=:,+@% stay as they
// are.
func QuoteArg(s string) string {
	if s != "" && strings.IndexFunc(s, func(c rune) bool { return !plain(c) }) < 0 {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func plain(c rune) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		return true
	}
	return strings.ContainsRune("-_./=:,+@%", c)
}
