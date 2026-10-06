//go:build windows && !appengine
// +build windows,!appengine

package colorable

import (
	"os"
	"testing"
)

func BenchmarkNewColorableConsole(b *testing.B) {
	for _, mode := range []struct {
		name  string
		value uintptr
	}{{"Legacy", 0}, {"VT", cENABLE_VIRTUAL_TERMINAL_PROCESSING}} {
		b.Run(mode.name, func(b *testing.B) {
			h, _, err := procCreateConsoleScreenBuffer.Call(uintptr(genericRead|genericWrite), 0, 0, uintptr(consoleTextmodeBuffer), 0)
			if h == ^uintptr(0) || h == 0 {
				b.Skipf("requires a console: %v", err)
			}
			f := os.NewFile(h, "benchmark console")
			defer f.Close()
			if r, _, err := procSetConsoleMode.Call(h, mode.value); r == 0 {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				NewColorable(f)
			}
		})
	}
}

func TestNewColorableFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "output")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if NewColorable(f) != f {
		t.Fatal("regular file must pass through")
	}
}
