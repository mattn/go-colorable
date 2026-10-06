//go:build windows && !appengine
// +build windows,!appengine

package colorable

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"unicode/utf16"
	"unsafe"
)

func BenchmarkWriterTitles(b *testing.B) {
	for _, tc := range []struct{ name, title string }{
		{"Short", "go-colorable"}, {"Long", strings.Repeat("x", 4096)},
		{"RejectedShort", "\x00" + strings.Repeat("x", 64)}, {"RejectedLong", "\x00" + strings.Repeat("x", 4096)},
	} {
		b.Run(tc.name, func(b *testing.B) {
			w := &writer{out: io.Discard}
			data := []byte("\x1b]0;" + tc.title + "\a")
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				w.Write(data)
			}
		})
	}
}

func TestWriterTitleFragments(t *testing.T) {
	var out bytes.Buffer
	w := &writer{out: &out}
	for _, s := range []string{"before\x1b]2;\x00", "ignored", "\aafter\x1b]0;\amore"} {
		w.Write([]byte(s))
	}
	if out.String() != "beforeaftermore" || w.rest.Len() != 0 {
		t.Fatalf("out=%q rest=%q", out.String(), w.rest.String())
	}
}

func testConsoleTitle(t *testing.T, w *writer) {
	const want = "go-colorable 日本語"
	w.Write([]byte("\x1b]0;" + want + "\a\x1b]2;\a"))
	var title [128]uint16
	getTitle := kernel32.NewProc("GetConsoleTitleW")
	n, _, err := getTitle.Call(uintptr(unsafe.Pointer(&title[0])), uintptr(len(title)))
	if n == 0 {
		t.Fatal(err)
	}
	if got := string(utf16.Decode(title[:n])); got != want {
		t.Fatalf("title=%q want %q", got, want)
	}
}
