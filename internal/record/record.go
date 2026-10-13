// Package record owns the format of .bcuz/history.jsonl: the shape of one
// line, how it is written and read, where .bcuz/ is looked for, and how a
// line is appended (docs/reference/history.md). Other packages do not build
// JSON by themselves.
package record

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Version is the format version written to the "v" field.
const Version = 1

// tsLayout is RFC 3339 with milliseconds. Records are written in UTC, so the
// zone is always "Z"; a reader accepts any offset.
const tsLayout = "2006-01-02T15:04:05.000Z07:00"

// Record is one execution: one line of history.jsonl.
type Record struct {
	Time time.Time // when the command was started ("ts")
	Why  string    // the reason ("why")
	Cmd  []string  // the command and its arguments ("cmd")
	Dir  string    // where it ran, relative to the .bcuz/ directory's parent ("dir")
	Exit int       // the exit code ("exit")
	MS   int64     // milliseconds from start to end ("ms")
}

// line is the written form. A struct, not a map: the order of the fields is
// fixed.
type line struct {
	V    int      `json:"v"`
	TS   string   `json:"ts"`
	Why  string   `json:"why"`
	Cmd  []string `json:"cmd"`
	Dir  string   `json:"dir"`
	Exit int      `json:"exit"`
	MS   int64    `json:"ms"`
}

// parsedLine is the read form. A pointer tells a missing field from a zero
// value; unknown fields are ignored by encoding/json.
type parsedLine struct {
	V    *int      `json:"v"`
	TS   *string   `json:"ts"`
	Why  *string   `json:"why"`
	Cmd  *[]string `json:"cmd"`
	Dir  *string   `json:"dir"`
	Exit *int      `json:"exit"`
	MS   *int64    `json:"ms"`
}

// Marshal returns the line for r, including the final "\n".
func Marshal(r Record) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	// Keep <, > and & as they are: reasons and commands are read by people.
	enc.SetEscapeHTML(false)
	err := enc.Encode(line{
		V:    Version,
		TS:   r.Time.UTC().Format(tsLayout),
		Why:  r.Why,
		Cmd:  r.Cmd,
		Dir:  r.Dir,
		Exit: r.Exit,
		MS:   r.MS,
	})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Unmarshal reads one line (with or without the final "\n"). A line that is
// not JSON, lacks one of the seven fields, or has an unknown "v" is an error:
// the caller skips it as an unreadable line.
func Unmarshal(b []byte) (Record, error) {
	var p parsedLine
	if err := json.Unmarshal(b, &p); err != nil {
		return Record{}, err
	}
	switch {
	case p.V == nil:
		return Record{}, errors.New(`missing "v"`)
	case p.TS == nil:
		return Record{}, errors.New(`missing "ts"`)
	case p.Why == nil:
		return Record{}, errors.New(`missing "why"`)
	case p.Cmd == nil:
		return Record{}, errors.New(`missing "cmd"`)
	case p.Dir == nil:
		return Record{}, errors.New(`missing "dir"`)
	case p.Exit == nil:
		return Record{}, errors.New(`missing "exit"`)
	case p.MS == nil:
		return Record{}, errors.New(`missing "ms"`)
	}
	if *p.V != Version {
		return Record{}, fmt.Errorf("unknown version %d", *p.V)
	}
	t, err := time.Parse(time.RFC3339, *p.TS)
	if err != nil {
		return Record{}, fmt.Errorf(`invalid "ts": %w`, err)
	}
	return Record{
		Time: t,
		Why:  *p.Why,
		Cmd:  *p.Cmd,
		Dir:  *p.Dir,
		Exit: *p.Exit,
		MS:   *p.MS,
	}, nil
}
