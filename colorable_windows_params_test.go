//go:build windows && !appengine
// +build windows,!appengine

package colorable

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func BenchmarkWriterParameters(b *testing.B) {
	for _, data := range []string{"\x1b[0m", "\x1b[1;22;24;39;49m", "\x1b[38;2;128;64;255m", "\x1b[48;2;0;0;0m", "\x1b[" + strings.Repeat("0;", 32) + "0m"} {
		b.Run(data[2:len(data)-1], func(b *testing.B) {
			w := &writer{out: io.Discard, oldattr: 7, curattr: 7}
			p := []byte(data)
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				w.Write(p)
			}
		})
	}
}

func TestWriterParameters(t *testing.T) {
	for _, tc := range []struct {
		params string
		attr   word
	}{
		{"0", 7}, {"1;22;24;39;49", 7}, {"38;2;128;64;255", 7},
		{"48;2;0;0;0", 7}, {strings.Repeat("0;", 32) + "0", 7},
		{";31;;44;", foregroundRed | backgroundBlue}, {"+31;044", foregroundRed | backgroundBlue},
		{"99999999999999999999999999999;31", foregroundRed},
	} {
		t.Run(tc.params, func(t *testing.T) {
			var out bytes.Buffer
			w := &writer{out: &out, oldattr: 7, curattr: 7}
			w.Write([]byte("\x1b[" + tc.params + "mhello"))
			if w.curattr != tc.attr || out.String() != "hello" {
				t.Fatalf("attr=%x out=%q", w.curattr, out.String())
			}
		})
	}
}

func TestSplitParameters(t *testing.T) {
	for _, s := range []string{"", ";", "31", "1;;31;", "+31;-1", strings.Repeat("0;", 7), strings.Repeat("0;", 8), strings.Repeat("0;", 100)} {
		var storage [8]string
		got, want := splitParameters(s, storage[:]), strings.Split(s, ";")
		if len(got) != len(want) {
			t.Fatalf("%q: got %d tokens want %d", s, len(got), len(want))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("%q: token %d=%q want %q", s, i, got[i], want[i])
			}
		}
	}
}
