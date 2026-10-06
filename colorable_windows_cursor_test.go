//go:build windows && !appengine
// +build windows,!appengine

package colorable

import (
	syscall "golang.org/x/sys/windows"
	"os"
	"testing"
	"unsafe"
)

func BenchmarkWriterAbsoluteCursor(b *testing.B) {
	h, _, err := procCreateConsoleScreenBuffer.Call(uintptr(genericRead|genericWrite), 0, 0, uintptr(consoleTextmodeBuffer), 0)
	if h == ^uintptr(0) || h == 0 {
		b.Skipf("requires a console: %v", err)
	}
	f := os.NewFile(h, "benchmark console")
	defer f.Close()
	w := &writer{out: f, handle: syscall.Handle(h)}
	data := []byte("\x1b[2;3H")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Write(data)
	}
}

func testAbsoluteCursor(t *testing.T, w *writer) {
	for _, tc := range []struct {
		seq  string
		x, y short
	}{
		{"\x1b[2;3H", 2, 1}, {"\x1b[4;5f", 4, 3},
		{"\x1b[6H", 4, 5}, {"\x1b[H", 4, 0},
		{"\x1b[2;3;4H", 4, 0}, {"\x1b[999999999999999999999;3H", 4, 0},
		{"\x1b[2;3H", 2, 1}, {"\x1b[;H", 2, 1},
	} {
		w.Write([]byte(tc.seq))
		var info consoleScreenBufferInfo
		if r, _, err := procGetConsoleScreenBufferInfo.Call(uintptr(w.handle), uintptr(unsafe.Pointer(&info))); r == 0 {
			t.Fatal(err)
		}
		if info.cursorPosition.x != tc.x || info.cursorPosition.y != tc.y {
			t.Fatalf("%q: got %+v want %d,%d", tc.seq, info.cursorPosition, tc.x, tc.y)
		}
	}
}
