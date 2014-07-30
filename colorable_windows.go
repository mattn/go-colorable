package colorable

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

const (
	foregroundBlue      = 0x1
	foregroundGreen     = 0x2
	foregroundRed       = 0x4
	foregroundIntensity = 0x8
	foregroundMask      = (foregroundRed | foregroundBlue | foregroundGreen | foregroundIntensity)
	backgroundBlue      = 0x10
	backgroundGreen     = 0x20
	backgroundRed       = 0x40
	backgroundIntensity = 0x80
	backgroundMask      = (backgroundRed | backgroundBlue | backgroundGreen | backgroundIntensity)
)

type wchar uint16
type short int16
type dword uint32
type word uint16

type coord struct {
	x short
	y short
}

type smallRect struct {
	left   short
	top    short
	right  short
	bottom short
}

type consoleScreenBufferInfo struct {
	size              coord
	cursorPosition    coord
	attributes        word
	window            smallRect
	maximumWindowSize coord
}

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	procSetConsoleTextAttribute    = kernel32.NewProc("SetConsoleTextAttribute")
)

type Writer struct {
	out     syscall.Handle
	lastbuf bytes.Buffer
}

func NewColorableStdout() io.Writer {
	return &Writer{out: syscall.Handle(os.Stdout.Fd())}
}

func NewColorableStderr() io.Writer {
	return &Writer{out: syscall.Handle(os.Stderr.Fd())}
}

func (w *Writer) Write(data []byte) (n int, err error) {
	var csbi consoleScreenBufferInfo
	procGetConsoleScreenBufferInfo.Call(uintptr(w.out), uintptr(unsafe.Pointer(&csbi)))
	attr_old := csbi.attributes
	defer func() {
		procSetConsoleTextAttribute.Call(uintptr(w.out), uintptr(attr_old))
	}()

	er := bytes.NewBuffer(data)
loop:
	for {
		r1, _, err := procGetConsoleScreenBufferInfo.Call(uintptr(w.out), uintptr(unsafe.Pointer(&csbi)))
		if r1 == 0 {
			break loop
		}

		c1, _, err := er.ReadRune()
		if err != nil {
			break loop
		}
		if c1 != 0x1b {
			fmt.Print(string(c1))
			continue
		}
		c2, _, err := er.ReadRune()
		if err != nil {
			w.lastbuf.WriteRune(c1)
			break loop
		}
		if c2 != 0x5b {
			w.lastbuf.WriteRune(c1)
			w.lastbuf.WriteRune(c2)
			continue
		}

		var buf bytes.Buffer
		var m rune
		for {
			c, _, err := er.ReadRune()
			if err != nil {
				w.lastbuf.WriteRune(c1)
				w.lastbuf.WriteRune(c2)
				w.lastbuf.Write(buf.Bytes())
				break loop
			}
			if ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || c == '@' {
				m = c
				break
			}
			buf.Write([]byte(string(c)))
		}

		switch m {
		case 'm':
			attr := csbi.attributes
			cs := buf.String()
			if cs == "" {
				procSetConsoleTextAttribute.Call(uintptr(w.out), uintptr(attr_old))
				continue
			}
			for _, ns := range strings.Split(cs, ";") {
				if n, err = strconv.Atoi(ns); err == nil {
					switch {
					case n == 0 || n == 100:
						attr = attr_old
					case 1 <= n && n <= 5:
						attr |= foregroundIntensity
					case n == 7:
						attr = ((attr & foregroundMask) << 4) | ((attr & backgroundMask) >> 4)
					case 22 == n || n == 25 || n == 25:
						attr |= foregroundIntensity
					case n == 27:
						attr = ((attr & foregroundMask) << 4) | ((attr & backgroundMask) >> 4)
					case 30 <= n && n <= 37:
						attr = (attr & backgroundMask)
						if (n-30)&1 != 0 {
							attr |= foregroundRed
						}
						if (n-30)&2 != 0 {
							attr |= foregroundGreen
						}
						if (n-30)&4 != 0 {
							attr |= foregroundBlue
						}
					case 40 <= n && n <= 47:
						attr = (attr & foregroundMask)
						if (n-40)&1 != 0 {
							attr |= backgroundRed
						}
						if (n-40)&2 != 0 {
							attr |= backgroundGreen
						}
						if (n-40)&4 != 0 {
							attr |= backgroundBlue
						}
					}
					procSetConsoleTextAttribute.Call(uintptr(w.out), uintptr(attr))
				}
			}
		}
	}
	return len(data) - w.lastbuf.Len(), nil
}
