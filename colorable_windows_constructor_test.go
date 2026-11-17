//go:build windows && !appengine
// +build windows,!appengine

package colorable

import (
	"os"
	"testing"

	syscall "golang.org/x/sys/windows"
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

func TestNewColorablePipe(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	if got := NewColorable(w); got != w {
		t.Fatalf("expected pipe to pass through, got %T", got)
	}
	if got := NewColorable(r); got != r {
		t.Fatalf("expected read pipe to pass through, got %T", got)
	}
}

func TestNewColorableConsoleInput(t *testing.T) {
	name, err := syscall.UTF16PtrFromString("CONIN$")
	if err != nil {
		t.Fatal(err)
	}
	h, err := syscall.CreateFile(name, syscall.GENERIC_READ|syscall.GENERIC_WRITE, syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE, nil, syscall.OPEN_EXISTING, 0, 0)
	if err != nil {
		t.Skipf("cannot open CONIN$: %v", err)
	}
	f := os.NewFile(uintptr(h), "CONIN$")
	defer f.Close()
	var mode uint32
	if err := syscall.GetConsoleMode(h, &mode); err == nil {
		_ = syscall.SetConsoleMode(h, mode&^uint32(cENABLE_VIRTUAL_TERMINAL_PROCESSING))
	}
	if got := NewColorable(f); got != f {
		t.Fatalf("expected CONIN$ to pass through, got %T", got)
	}
}

func TestNewColorableConsoleOutput(t *testing.T) {
	h, _, err := procCreateConsoleScreenBuffer.Call(uintptr(genericRead|genericWrite), 0, 0, uintptr(consoleTextmodeBuffer), 0)
	if h == ^uintptr(0) || h == 0 {
		t.Skipf("requires a console: %v", err)
	}
	f := os.NewFile(h, "test console")
	defer f.Close()
	if r, _, err := procSetConsoleMode.Call(h, 0); r == 0 {
		t.Fatal(err)
	}
	got := NewColorable(f)
	if _, ok := got.(*writer); !ok {
		t.Fatalf("expected *writer for real console screen buffer, got %T", got)
	}
}
