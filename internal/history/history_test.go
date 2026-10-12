package history

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/amisonnet8/bcuz/internal/record"
)

// setLocal makes time.Local the zone name for the test.
func setLocal(t *testing.T, name string, offsetSec int) {
	t.Helper()
	old := time.Local
	time.Local = time.FixedZone(name, offsetSec)
	t.Cleanup(func() { time.Local = old })
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		ms   int64
		want string
	}{
		{0, "0ms"},
		{123, "123ms"},
		{999, "999ms"},
		{1000, "1.0s"},
		{1049, "1.0s"},
		{1099, "1.0s"},
		{8312, "8.3s"},
		{35000, "35.0s"},
		{59999, "59.9s"},
		{60000, "1m00s"},
		{65200, "1m05s"},
		{125000, "2m05s"},
		{3599999, "59m59s"},
		{3600000, "1h00m"},
		{3720000, "1h02m"},
		{86400000, "24h00m"},
	}
	for _, tt := range tests {
		if got := FormatDuration(tt.ms); got != tt.want {
			t.Errorf("FormatDuration(%d) = %q, want %q", tt.ms, got, tt.want)
		}
	}
}

func TestQuoteArg(t *testing.T) {
	tests := []struct{ in, want string }{
		{"make", "make"},
		{"-m", "-m"},
		{"v0.*", "'v0.*'"},
		{"a/b_c-d.e=f:g,h+i@j%k", "a/b_c-d.e=f:g,h+i@j%k"},
		{"fix: a b", "'fix: a b'"},
		{"", "''"},
		{"it's", `'it'\''s'`},
		{"$HOME", "'$HOME'"},
		{"日本語", "'日本語'"},
		{"a\nb", "'a\nb'"},
		{"a|b", "'a|b'"},
	}
	for _, tt := range tests {
		if got := QuoteArg(tt.in); got != tt.want {
			t.Errorf("QuoteArg(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestParseDate(t *testing.T) {
	for _, ok := range []string{"2026-10-12", "2024-02-29", "0001-01-01"} {
		if err := ParseDate(ok); err != nil {
			t.Errorf("ParseDate(%q) = %v, want nil", ok, err)
		}
	}
	for _, bad := range []string{"", "2026-02-30", "2025-02-29", "2026-1-2", "20261012", "2026-10-12 ", "2026-13-01", "x", "2026-10-12T00:00:00Z"} {
		if err := ParseDate(bad); err == nil {
			t.Errorf("ParseDate(%q) = nil, want an error", bad)
		}
	}
}

func rec(ts, why string, cmd ...string) record.Record {
	tm, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		panic(err)
	}
	return record.Record{Time: tm, Why: why, Cmd: cmd, Dir: ".", Exit: 0, MS: 1200}
}

func TestFormatLine(t *testing.T) {
	setLocal(t, "JST", 9*3600)

	r := rec("2026-10-12T06:52:10.000+09:00", "コミット", "git", "commit", "-m", "fix: handle empty reason")
	want := "2026-10-12 06:52:10    0    1.2s  .  コミット  $ git commit -m 'fix: handle empty reason'"
	if got := FormatLine(r); got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}

	// Another zone: the time is shown in the viewer's zone.
	setLocal(t, "UTC", 0)
	want = "2026-10-11 21:52:10    0    1.2s  .  コミット  $ git commit -m 'fix: handle empty reason'"
	if got := FormatLine(r); got != want {
		t.Errorf("UTC: got  %q\nwant %q", got, want)
	}

	// Columns: exit code is 3 wide, duration 6 wide; the reason loses its
	// line breaks and tabs, and the command keeps them inside quotes.
	r = rec("2026-10-12T06:41:30.500+09:00", "a\nb\r\nc\td", "tail", "-f", "x y")
	r.Exit, r.MS, r.Dir = 130, 65200, "backend"
	want = "2026-10-11 21:41:30  130   1m05s  backend  a b c d  $ tail -f 'x y'"
	if got := FormatLine(r); got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestFilterAndSelect(t *testing.T) {
	setLocal(t, "JST", 9*3600)
	recs := []record.Record{
		rec("2026-10-10T00:00:00.000+09:00", "start of 10th"),
		rec("2026-10-10T23:59:59.999+09:00", "end of 10th"),
		rec("2026-10-11T00:00:00.000+09:00", "start of 11th"),
		rec("2026-10-11T23:59:59.999+09:00", "end of 11th"),
		rec("2026-10-12T00:00:00.000+09:00", "start of 12th"),
	}
	whys := func(rs []record.Record) []string {
		var out []string
		for _, r := range rs {
			out = append(out, r.Why)
		}
		return out
	}

	tests := []struct {
		name string
		f    Filter
		want []string
	}{
		{"all, newest first", Filter{}, []string{"start of 12th", "end of 11th", "start of 11th", "end of 10th", "start of 10th"}},
		{"date", Filter{Date: "2026-10-11"}, []string{"end of 11th", "start of 11th"}},
		{"since includes the day", Filter{Since: "2026-10-11"}, []string{"start of 12th", "end of 11th", "start of 11th"}},
		{"until includes the day", Filter{Until: "2026-10-11"}, []string{"end of 11th", "start of 11th", "end of 10th", "start of 10th"}},
		{"range", Filter{Since: "2026-10-11", Until: "2026-10-11"}, []string{"end of 11th", "start of 11th"}},
		{"since after until", Filter{Since: "2026-10-12", Until: "2026-10-10"}, nil},
		{"none", Filter{Date: "2026-09-01"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := whys(Select(recs, tt.f))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("the day is the viewer's day", func(t *testing.T) {
		setLocal(t, "UTC", 0)
		// 2026-10-11T00:00+09:00 is 2026-10-10T15:00Z.
		got := whys(Select(recs, Filter{Date: "2026-10-10"}))
		want := []string{"start of 11th", "end of 10th"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("equal times: the later line first", func(t *testing.T) {
		same := []record.Record{
			rec("2026-10-12T06:00:00.000+09:00", "first"),
			rec("2026-10-12T06:00:00.000+09:00", "second"),
		}
		got := whys(Select(same, Filter{}))
		if want := []string{"second", "first"}; !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("by instant, not by file order", func(t *testing.T) {
		unordered := []record.Record{
			rec("2026-10-12T06:00:00.000+09:00", "newer"),
			rec("2026-10-11T06:00:00.000+09:00", "older"),
		}
		got := whys(Select(unordered, Filter{}))
		if want := []string{"newer", "older"}; !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestRead(t *testing.T) {
	good := `{"v":1,"ts":"2026-10-12T06:35:12.345+09:00","why":"a","cmd":["make"],"dir":".","exit":0,"ms":1}`
	tests := []struct {
		name        string
		in          string
		wantWhys    []string
		wantSkipped []int
	}{
		{"empty", "", nil, nil},
		{"one line", good + "\n", []string{"a"}, nil},
		{"last line without newline is skipped silently", good + "\n" + good, []string{"a"}, nil},
		{"only a partial line", good, nil, nil},
		{"broken line is reported", good + "\nnot json\n" + good + "\n", []string{"a", "a"}, []int{2}},
		{"empty line is reported", good + "\n\n" + good + "\n", []string{"a", "a"}, []int{2}},
		{"missing field", `{"v":1,"ts":"2026-10-12T06:35:12.345+09:00","why":"a"}` + "\n" + good + "\n", []string{"a"}, []int{1}},
		{"unknown version", strings.Replace(good, `"v":1`, `"v":2`, 1) + "\n", nil, []int{1}},
		{"carriage return before newline", good + "\r\n", []string{"a"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recs, skipped, err := Read(strings.NewReader(tt.in))
			if err != nil {
				t.Fatal(err)
			}
			var whys []string
			for _, r := range recs {
				whys = append(whys, r.Why)
			}
			if !reflect.DeepEqual(whys, tt.wantWhys) {
				t.Errorf("records %v, want %v", whys, tt.wantWhys)
			}
			if !reflect.DeepEqual(skipped, tt.wantSkipped) {
				t.Errorf("skipped %v, want %v", skipped, tt.wantSkipped)
			}
		})
	}

	t.Run("a very long line is read", func(t *testing.T) {
		long := strings.Replace(good, `"why":"a"`, `"why":"`+strings.Repeat("x", 200_000)+`"`, 1)
		recs, skipped, err := Read(strings.NewReader(long + "\n"))
		if err != nil || len(skipped) != 0 || len(recs) != 1 || len(recs[0].Why) != 200_000 {
			t.Errorf("recs %d, skipped %v, err %v", len(recs), skipped, err)
		}
	})
}
