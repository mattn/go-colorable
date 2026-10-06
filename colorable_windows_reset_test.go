//go:build windows && !appengine
// +build windows,!appengine

package colorable

import (
	"io"
	"testing"
)

func BenchmarkWriterEmptyReset(b *testing.B) {
	w := &writer{out: io.Discard, oldattr: 7, curattr: 7}
	p := []byte("\x1b[m")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w.Write(p)
	}
}

func TestWriterEmptyReset(t *testing.T) {
	for _, attr := range []word{7, foregroundRed, backgroundBlue | foregroundGreen | commonLvbUnderscore} {
		w := &writer{out: io.Discard, oldattr: 7, curattr: attr}
		for i := 0; i < 2; i++ {
			w.Write([]byte("\x1b[m"))
			if w.curattr != 7 {
				t.Fatalf("reset %x: got %x", attr, w.curattr)
			}
		}
	}
}
