//go:build windows && !appengine
// +build windows,!appengine

package colorable

import (
	syscall "golang.org/x/sys/windows"
	"os"
	"strings"
	"testing"
	"unicode/utf16"
	"unsafe"
)

func BenchmarkWriterConsoleAPIs(b *testing.B) {
	h, _, err := procCreateConsoleScreenBuffer.Call(uintptr(genericRead|genericWrite), 0, 0, uintptr(consoleTextmodeBuffer), 0)
	if h == ^uintptr(0) || h == 0 {
		b.Skipf("requires a console: %v", err)
	}
	f := os.NewFile(h, "benchmark console")
	defer f.Close()
	for _, tc := range []struct{ name, data string }{
		{"Color", "\x1b[31m\x1b[0m"}, {"Cursor", "\x1b[1G"},
		{"Erase", "\x1b[2K"}, {"Visibility", "\x1b[?25h"},
	} {
		b.Run(tc.name, func(b *testing.B) {
			w := &writer{out: f, handle: syscall.Handle(h), oldattr: 7, curattr: 7}
			data := []byte(tc.data)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				w.Write(data)
			}
		})
	}
}

func TestCoordParam(t *testing.T) {
	for _, x := range []short{-32768, -1, 0, 1, 32767} {
		for _, y := range []short{-32768, -1, 0, 1, 32767} {
			got := (coord{x: x, y: y}).param()
			if short(uint16(got)) != x || short(uint16(got>>16)) != y || uint64(got)>>32 != 0 {
				t.Fatalf("(%d,%d) encoded as %x", x, y, got)
			}
		}
	}
}

func testConsoleAPIs(t *testing.T, w *writer) {
	w.Write([]byte("\x1b[2;3H\x1b[s\x1b[4;5H\x1b[uabcdef\x1b[2;3H\x1b[3X"))
	var chars [6]uint16
	var read uint32
	readChars := kernel32.NewProc("ReadConsoleOutputCharacterW")
	if r, _, err := readChars.Call(uintptr(w.handle), uintptr(unsafe.Pointer(&chars[0])), 6, uintptr(1<<16|2), uintptr(unsafe.Pointer(&read))); r == 0 {
		t.Fatal(err)
	}
	if got := string(utf16.Decode(chars[:])); got != "   def" {
		t.Fatalf("erased text=%q", got)
	}
	for _, v := range []struct {
		seq     string
		visible int32
	}{{"\x1b[?25l", 0}, {"\x1b[?25h", 1}} {
		w.Write([]byte(v.seq))
		var ci consoleCursorInfo
		if r, _, err := procGetConsoleCursorInfo.Call(uintptr(w.handle), uintptr(unsafe.Pointer(&ci))); r == 0 {
			t.Fatal(err)
		}
		if ci.visible != v.visible {
			t.Fatalf("visibility=%d want %d", ci.visible, v.visible)
		}
	}
	w.Write([]byte("\x1b[2K\x1b[1G\x1b7\x1b[2;5H\x1b8"))
	var info consoleScreenBufferInfo
	if r, _, err := procGetConsoleScreenBufferInfo.Call(uintptr(w.handle), uintptr(unsafe.Pointer(&info))); r == 0 {
		t.Fatal(err)
	}
	if info.cursorPosition.x != 0 || info.cursorPosition.y != 1 {
		t.Fatalf("restored cursor=%+v", info.cursorPosition)
	}
	w.Write([]byte("\x1b]0;go-colorable test\a"))
}

// Include actual console text output to distinguish parser savings from I/O.
func BenchmarkWriterConsoleText(b *testing.B) {
	h, _, err := procCreateConsoleScreenBuffer.Call(uintptr(genericRead|genericWrite), 0, 0, uintptr(consoleTextmodeBuffer), 0)
	if h == ^uintptr(0) || h == 0 {
		b.Skipf("requires a console: %v", err)
	}
	f := os.NewFile(h, "benchmark console")
	defer f.Close()
	if r, _, err := procSetConsoleMode.Call(h, 3); r == 0 {
		b.Fatal(err)
	}
	for _, tc := range []struct{ name, data string }{
		{"PlainShort", "hello world\n"},
		{"Plain4K", strings.Repeat("x", 4096)},
		{"Mixed4K", strings.Repeat("x", 4096) + "\x1b[0m"},
		{"ColoredLine", "\x1b[31merror: something failed\x1b[0m\n"},
		{"DenseColored", strings.Repeat("\x1b[31mred\x1b[32mgreen\x1b[0m\n", 32)},
	} {
		b.Run(tc.name, func(b *testing.B) {
			w := &writer{out: f, handle: syscall.Handle(h), oldattr: 7, curattr: 7}
			data := []byte(tc.data)
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				w.Write(data)
			}
		})
	}
}
