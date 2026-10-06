//go:build windows && !appengine
// +build windows,!appengine

package colorable

import (
	"os"
	"os/exec"
	stdsyscall "syscall"
	"testing"
	"unicode/utf16"
	"unsafe"
)

// Exercise real console APIs in a child with its own invisible console.
func TestWriterConsole(t *testing.T) {
	if os.Getenv("GO_COLORABLE_TEST_CONSOLE") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestWriterConsole$")
		cmd.Env = append(os.Environ(), "GO_COLORABLE_TEST_CONSOLE=1")
		cmd.SysProcAttr = &stdsyscall.SysProcAttr{CreationFlags: 0x08000000} // CREATE_NO_WINDOW
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("console test: %v\n%s", err, output)
		}
		return
	}
	h, _, err := procCreateConsoleScreenBuffer.Call(uintptr(genericRead|genericWrite), 0, 0, uintptr(consoleTextmodeBuffer), 0)
	if h == ^uintptr(0) || h == 0 {
		t.Fatalf("CreateConsoleScreenBuffer: %v", err)
	}
	f := os.NewFile(h, "test console")
	defer f.Close()
	if r, _, err := procSetConsoleMode.Call(h, 0); r == 0 {
		t.Fatalf("SetConsoleMode: %v", err)
	}
	if r, _, err := procSetConsoleTextAttribute.Call(h, 7); r == 0 {
		t.Fatalf("SetConsoleTextAttribute: %v", err)
	}
	w, ok := NewColorable(f).(*writer)
	if !ok {
		t.Fatal("expected legacy writer")
	}
	for _, s := range []string{"plain\x1b[31mred\x1b[mend\x1b[m", "\x1b[", "32mG\x1b[m"} {
		if n, err := w.Write([]byte(s)); n != len(s) || err != nil {
			t.Fatalf("Write: %d, %v", n, err)
		}
	}
	var csbi consoleScreenBufferInfo
	if r, _, err := procGetConsoleScreenBufferInfo.Call(h, uintptr(unsafe.Pointer(&csbi))); r == 0 {
		t.Fatal(err)
	}
	if csbi.attributes != 7 || csbi.cursorPosition.x != 12 || csbi.cursorPosition.y != 0 {
		t.Fatalf("console state: %+v", csbi)
	}
	var chars [12]uint16
	var read uint32
	readChars := kernel32.NewProc("ReadConsoleOutputCharacterW")
	if r, _, err := readChars.Call(h, uintptr(unsafe.Pointer(&chars[0])), 12, 0, uintptr(unsafe.Pointer(&read))); r == 0 {
		t.Fatal(err)
	}
	if got := string(utf16.Decode(chars[:])); got != "plainredendG" || read != 12 {
		t.Fatalf("text=%q count=%d", got, read)
	}
	testAbsoluteCursor(t, w)
	testConsoleAPIs(t, w)
	testConsoleTitle(t, w)
	if r, _, err := procSetConsoleMode.Call(h, cENABLE_VIRTUAL_TERMINAL_PROCESSING); r == 0 {
		t.Fatal(err)
	}
	if NewColorable(f) != f {
		t.Fatal("expected direct file for VT console")
	}
}
