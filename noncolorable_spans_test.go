package colorable

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func BenchmarkNonColorableSpans(b *testing.B) {
	for _, tc := range []struct{ name, text string }{
		{"Short", "hello world\n"}, {"Large", strings.Repeat("x", 4096)},
		{"Mixed", strings.Repeat("x", 4096) + "\x1b[0m"},
		{"Dense", strings.Repeat("hello\x1b[31mworld\x1b[0m", 256)},
	} {
		b.Run(tc.name, func(b *testing.B) {
			w := NewNonColorable(io.Discard)
			data := []byte(tc.text)
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				w.Write(data)
			}
		})
	}
}

func FuzzNonColorableSpans(f *testing.F) {
	for _, s := range []string{"", "hello\xff", "\x1b", "\x1b[", "a\x1b[31mb\x1b[0mc", "\x1b]0;title\a", "a\x1b!b", "\x1b[123@xyz"} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		var got, want bytes.Buffer
		n, e := NewNonColorable(&got).Write(data)
		rn, re := (&referenceNonColorable{out: &want}).Write(data)
		if n != rn || e != re || !bytes.Equal(got.Bytes(), want.Bytes()) {
			t.Fatalf("input=%q got=(%q,%d,%v) want=(%q,%d,%v)", data, got.Bytes(), n, e, want.Bytes(), rn, re)
		}
	})
}

// Keep the original buffered implementation as a differential fuzz oracle.
type referenceNonColorable struct{ out io.Writer }

func (w *referenceNonColorable) Write(data []byte) (n int, err error) {
	er := bytes.NewReader(data)
	var plaintext bytes.Buffer
loop:
	for {
		c1, err := er.ReadByte()
		if err != nil {
			plaintext.WriteTo(w.out)
			break loop
		}
		if c1 != 0x1b {
			plaintext.WriteByte(c1)
			continue
		}
		_, err = plaintext.WriteTo(w.out)
		if err != nil {
			break loop
		}
		c2, err := er.ReadByte()
		if err != nil {
			break loop
		}
		if c2 != 0x5b {
			continue
		}

		for {
			c, err := er.ReadByte()
			if err != nil {
				break loop
			}
			if ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || c == '@' {
				break
			}
		}
	}

	return len(data), nil
}
