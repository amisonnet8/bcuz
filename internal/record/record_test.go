package record

import (
	"strings"
	"testing"
	"time"
)

var jst = time.FixedZone("JST", 9*60*60)

func TestMarshal(t *testing.T) {
	tests := []struct {
		name string
		r    Record
		want string
	}{
		{
			name: "docs/reference/history.md",
			r: Record{
				Time: time.Date(2026, 10, 12, 6, 35, 12, 345_000_000, jst),
				Why:  "xxテスト実施",
				Cmd:  []string{"make", "test"},
				Dir:  "backend",
				Exit: 0,
				MS:   8312,
			},
			want: `{"v":1,"ts":"2026-10-11T21:35:12.345Z","why":"xxテスト実施","cmd":["make","test"],"dir":"backend","exit":0,"ms":8312}` + "\n",
		},
		{
			name: "html characters stay as they are",
			r: Record{
				Time: time.Date(2026, 10, 12, 6, 35, 12, 0, jst),
				Why:  "a<b && c>d",
				Cmd:  []string{"sh", "-c", "grep -c '<x>' *.go | sort"},
				Dir:  ".",
				Exit: 1,
				MS:   0,
			},
			want: `{"v":1,"ts":"2026-10-11T21:35:12.000Z","why":"a<b && c>d","cmd":["sh","-c","grep -c '<x>' *.go | sort"],"dir":".","exit":1,"ms":0}` + "\n",
		},
		{
			name: "newline and tab in the reason are escaped, not lost",
			r: Record{
				Time: time.Date(2026, 10, 12, 6, 35, 12, 0, time.UTC),
				Why:  "a\nb\tc",
				Cmd:  []string{"ls"},
				Dir:  ".",
				Exit: 130,
				MS:   65200,
			},
			want: `{"v":1,"ts":"2026-10-12T06:35:12.000Z","why":"a\nb\tc","cmd":["ls"],"dir":".","exit":130,"ms":65200}` + "\n",
		},
		{
			name: "sub-millisecond part is cut, a time with an offset is written in UTC",
			r: Record{
				Time: time.Date(2026, 1, 2, 3, 4, 5, 999_999_999, time.FixedZone("", -5*60*60)),
				Why:  "w",
				Cmd:  []string{"x"},
				Dir:  "a/b",
				Exit: 127,
				MS:   0,
			},
			want: `{"v":1,"ts":"2026-01-02T08:04:05.999Z","why":"w","cmd":["x"],"dir":"a/b","exit":127,"ms":0}` + "\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.r)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Errorf("Marshal =\n%s want\n%s", got, tt.want)
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	const ok = `{"v":1,"ts":"2026-10-12T06:35:12.345+09:00","why":"xxテスト実施","cmd":["make","test"],"dir":"backend","exit":0,"ms":8312}`

	t.Run("a valid line, with or without the newline", func(t *testing.T) {
		for _, in := range []string{ok, ok + "\n"} {
			r, err := Unmarshal([]byte(in))
			if err != nil {
				t.Fatal(err)
			}
			if !r.Time.Equal(time.Date(2026, 10, 12, 6, 35, 12, 345_000_000, jst)) {
				t.Errorf("Time = %v", r.Time)
			}
			if r.Why != "xxテスト実施" || strings.Join(r.Cmd, " ") != "make test" ||
				r.Dir != "backend" || r.Exit != 0 || r.MS != 8312 {
				t.Errorf("Record = %+v", r)
			}
		}
	})

	t.Run("unknown fields are ignored", func(t *testing.T) {
		in := `{"v":1,"ts":"2026-10-12T06:35:12.345+09:00","why":"w","cmd":["x"],"dir":".","exit":0,"ms":1,"host":"h"}`
		if _, err := Unmarshal([]byte(in)); err != nil {
			t.Error(err)
		}
	})

	bad := []struct{ name, in string }{
		{"not json", `not json`},
		{"truncated", `{"v":1,"ts":"2026-10-12T06:35:12.345+09:00"`},
		{"empty", ``},
		{"missing v", `{"ts":"2026-10-12T06:35:12.345+09:00","why":"w","cmd":["x"],"dir":".","exit":0,"ms":1}`},
		{"missing ts", `{"v":1,"why":"w","cmd":["x"],"dir":".","exit":0,"ms":1}`},
		{"missing why", `{"v":1,"ts":"2026-10-12T06:35:12.345+09:00","cmd":["x"],"dir":".","exit":0,"ms":1}`},
		{"missing cmd", `{"v":1,"ts":"2026-10-12T06:35:12.345+09:00","why":"w","dir":".","exit":0,"ms":1}`},
		{"missing dir", `{"v":1,"ts":"2026-10-12T06:35:12.345+09:00","why":"w","cmd":["x"],"exit":0,"ms":1}`},
		{"missing exit", `{"v":1,"ts":"2026-10-12T06:35:12.345+09:00","why":"w","cmd":["x"],"dir":".","ms":1}`},
		{"missing ms", `{"v":1,"ts":"2026-10-12T06:35:12.345+09:00","why":"w","cmd":["x"],"dir":".","exit":0}`},
		{"unknown version", `{"v":2,"ts":"2026-10-12T06:35:12.345+09:00","why":"w","cmd":["x"],"dir":".","exit":0,"ms":1}`},
		{"invalid ts", `{"v":1,"ts":"yesterday","why":"w","cmd":["x"],"dir":".","exit":0,"ms":1}`},
		{"wrong type", `{"v":1,"ts":"2026-10-12T06:35:12.345+09:00","why":"w","cmd":"x","dir":".","exit":0,"ms":1}`},
	}
	for _, tt := range bad {
		t.Run(tt.name, func(t *testing.T) {
			if r, err := Unmarshal([]byte(tt.in)); err == nil {
				t.Errorf("Unmarshal succeeded: %+v", r)
			}
		})
	}
}

// An old line, with a time zone offset in ts, is read as the same instant and
// written again in UTC: the records of earlier versions stay readable.
func TestOldOffsetLineIsWrittenInUTC(t *testing.T) {
	in := `{"v":1,"ts":"2026-10-12T06:35:12.345+09:00","why":"w","cmd":["x"],"dir":".","exit":0,"ms":8312}`
	want := `{"v":1,"ts":"2026-10-11T21:35:12.345Z","why":"w","cmd":["x"],"dir":".","exit":0,"ms":8312}` + "\n"
	r, err := Unmarshal([]byte(in))
	if err != nil {
		t.Fatal(err)
	}
	out, err := Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != want {
		t.Errorf("got %s want %s", out, want)
	}
}

// A line that is read and written again is the same line.
func TestRoundTrip(t *testing.T) {
	lines := []string{
		`{"v":1,"ts":"2026-10-11T21:35:12.345Z","why":"a<b>&c","cmd":["x","y z"],"dir":"a/b","exit":0,"ms":8312}`,
		`{"v":1,"ts":"2026-10-12T06:35:12.000Z","why":"改行\nあり","cmd":["x"],"dir":".","exit":130,"ms":0}`,
		`{"v":1,"ts":"2026-10-12T11:05:12.100Z","why":"w","cmd":["x"],"dir":".","exit":127,"ms":1}`,
	}
	for _, in := range lines {
		r, err := Unmarshal([]byte(in))
		if err != nil {
			t.Fatal(err)
		}
		out, err := Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		if string(out) != in+"\n" {
			t.Errorf("round trip:\n got %s want %s\n", out, in)
		}
	}
}
