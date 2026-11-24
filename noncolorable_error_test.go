package colorable

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"
)

type nonColorableWriteFunc func([]byte) (int, error)

func (f nonColorableWriteFunc) Write(p []byte) (int, error) { return f(p) }

func TestNonColorableWriteClosedFile(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "closed-output")
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	n, err := NewNonColorable(file).Write([]byte("hello"))
	if n != 0 || !errors.Is(err, os.ErrClosed) {
		t.Fatalf("got (%d, %v), want (0, os.ErrClosed)", n, err)
	}
}

func TestNonColorableWritePartial(t *testing.T) {
	writeErr := errors.New("output failed")
	for _, tc := range []struct {
		name       string
		text       string
		limit      int
		err        error
		wantN      int
		wantOutput string
		wantCalls  int
	}{
		{"Plain", "hello", 2, writeErr, 2, "he", 1},
		{"BeforeEscape", "hello\x1b[31mworld", 2, writeErr, 2, "he", 1},
		{"AfterEscape", "\x1b[31mhello", 2, writeErr, 7, "he", 1},
		{"BetweenEscapes", "a\x1b[31mbc\x1b[0md", 2, writeErr, 7, "ab", 2},
		{"EmptyAfterEscape", "\x1b[31mhello", 0, writeErr, 5, "", 1},
		{"PlainShortWrite", "hello", 2, nil, 2, "he", 1},
		{"BeforeEscapeShortWrite", "hello\x1b[31mworld", 2, nil, 2, "he", 1},
		{"AfterEscapeShortWrite", "\x1b[31mhello", 2, nil, 7, "he", 1},
		{"LaterShortWrite", "hello\x1b[31mworld", 7, nil, 12, "hellowo", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			calls := 0
			out := nonColorableWriteFunc(func(p []byte) (int, error) {
				calls++
				n := len(p)
				if remaining := tc.limit - output.Len(); n > remaining {
					n = remaining
				}
				output.Write(p[:n])
				if n < len(p) {
					return n, tc.err
				}
				return n, nil
			})
			wantErr := tc.err
			if wantErr == nil {
				wantErr = io.ErrShortWrite
			}
			n, err := NewNonColorable(out).Write([]byte(tc.text))
			if n != tc.wantN || !errors.Is(err, wantErr) {
				t.Fatalf("got (%d, %v), want (%d, %v)", n, err, tc.wantN, wantErr)
			}
			if output.String() != tc.wantOutput || calls != tc.wantCalls {
				t.Fatalf("output=%q calls=%d, want %q and %d", output.String(), calls, tc.wantOutput, tc.wantCalls)
			}
		})
	}
}

func TestNonColorableWriteFullSpanError(t *testing.T) {
	writeErr := errors.New("output failed")
	for _, text := range []string{"hello", "hello\x1b[0mworld"} {
		t.Run(text, func(t *testing.T) {
			var output bytes.Buffer
			calls := 0
			out := nonColorableWriteFunc(func(p []byte) (int, error) {
				calls++
				output.Write(p)
				return len(p), writeErr
			})
			n, err := NewNonColorable(out).Write([]byte(text))
			if n != len("hello") || !errors.Is(err, writeErr) {
				t.Fatalf("got (%d, %v), want (5, %v)", n, err, writeErr)
			}
			if output.String() != "hello" || calls != 1 {
				t.Fatalf("output=%q calls=%d, want hello and 1", output.String(), calls)
			}
		})
	}
}
