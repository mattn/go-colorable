package colorable

import (
	"bytes"
	"io"
)

// NonColorable holds writer but removes escape sequence.
type NonColorable struct {
	out io.Writer
}

// NewNonColorable returns new instance of Writer which removes escape sequence from Writer.
func NewNonColorable(w io.Writer) io.Writer {
	return &NonColorable{out: w}
}

// Write removes escape sequences and writes the remaining data to the underlying writer.
// The returned count includes escape sequences consumed before an error.
func (w *NonColorable) Write(data []byte) (n int, err error) {
	for offset := 0; offset < len(data); {
		escape := bytes.IndexByte(data[offset:], 0x1b)
		if escape < 0 {
			n, err := w.out.Write(data[offset:])
			if n < len(data)-offset && err == nil {
				err = io.ErrShortWrite
			}
			return offset + n, err
		}
		if escape > 0 {
			n, err := w.out.Write(data[offset : offset+escape])
			if n < escape && err == nil {
				err = io.ErrShortWrite
			}
			if err != nil {
				return offset + n, err
			}
		}
		offset += escape + 1
		if offset == len(data) {
			break
		}
		c2 := data[offset]
		offset++
		if c2 != 0x5b {
			continue
		}

		for offset < len(data) {
			c := data[offset]
			offset++
			if ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || c == '@' {
				break
			}
		}
	}

	return len(data), nil
}
