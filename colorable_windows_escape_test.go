//go:build windows && !appengine
// +build windows,!appengine

package colorable

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func BenchmarkWriterEscapes(b *testing.B) {
	for _, tc := range []struct{ name, data string }{
		{"Short", "hello\x1b[0mworld"},
		{"LongText", strings.Repeat("x", 4096) + "\x1b[0m"},
		{"Dense", strings.Repeat("x\x1b[0m", 256)},
		{"Unknown", strings.Repeat("x\x1b[123q", 256)},
		{"SGR", "\x1b[1;4;31;44;22;24mhello\x1b[0m"},
		{"TrueColor", "\x1b[38;2;128;64;255mhello\x1b[0m"},
		{"Indexed", "\x1b[38;5;196mhello\x1b[0m"},
	} {
		b.Run(tc.name, func(b *testing.B) {
			w := &writer{out: io.Discard, oldattr: 7, curattr: 7}
			data := []byte(tc.data)
			w.Write(data)
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				w.Write(data)
			}
		})
	}
	b.Run("Fragmented", func(b *testing.B) {
		w := &writer{out: io.Discard, oldattr: 7, curattr: 7}
		a, c := []byte("hello\x1b["), []byte("0mworld")
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			w.Write(a)
			w.Write(c)
		}
	})
}

func TestWriterEscapeSlices(t *testing.T) {
	for _, tc := range []struct {
		name       string
		chunks     []string
		want, rest string
		attr       word
	}{
		{"plain", []string{"hello\xff"}, "hello\xff", "", 7},
		{"mixed", []string{"a\x1b[31mb\x1b[0mc"}, "abc", "", 7},
		{"dense", []string{strings.Repeat("a\x1b[0m", 256)}, strings.Repeat("a", 256), "", 7},
		{"fragment", []string{"a\x1b[", "3", "1mb\x1b[0", "mc"}, "abc", "", 7},
		{"pending", []string{"a\x1b[12", ";"}, "a", "\x1b[12;", 7},
		{"title", []string{"a\x1b]0;\a", "b"}, "ab", "", 7},
		{"titleFragment", []string{"a\x1b]", "0;", "\ab\x1b[0mc"}, "abc", "", 7},
		{"titleUnknown", []string{"a\x1b]9;xyz\ab"}, "a;xyz\ab", "", 7},
		{"titlePending", []string{"a\x1b]0;title"}, "a", "\x1b]0;title", 7},
		{"unknown", []string{"a\x1b>\x1b!b\x1b[123qc"}, "abc", "", 7},
		{"trailingEscape", []string{"a\x1b", "b"}, "ab", "", 7},
		{"attribute", []string{"\x1b[1;4;31;44;22;24m"}, "", "", backgroundBlue | foregroundRed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			w := &writer{out: &out, oldattr: 7, curattr: 7}
			for _, s := range tc.chunks {
				n, err := w.Write([]byte(s))
				if n != len(s) || err != nil {
					t.Fatalf("Write=(%d,%v)", n, err)
				}
			}
			if out.String() != tc.want || w.rest.String() != tc.rest || w.curattr != tc.attr {
				t.Fatalf("output=%q rest=%q attr=%x; want %q %q %x", out.String(), w.rest.String(), w.curattr, tc.want, tc.rest, tc.attr)
			}
		})
	}
}

func TestWriterEscapeSplitPoints(t *testing.T) {
	data := []byte("a\xff\x1b[0mb\x1b[123q\x1b]0;\ac\x1b[31md\x1b[0me")
	for split := 0; split <= len(data); split++ {
		// A standalone ESC is discarded by the existing writer contract.
		if split > 0 && data[split-1] == 0x1b {
			continue
		}
		var out bytes.Buffer
		w := &writer{out: &out, oldattr: 7, curattr: 7}
		w.Write(data[:split])
		w.Write(data[split:])
		if out.String() != "a\xffbcde" || w.rest.Len() != 0 || w.curattr != 7 {
			t.Fatalf("split %d: out=%q rest=%q attr=%x", split, out.String(), w.rest.String(), w.curattr)
		}
	}
}
