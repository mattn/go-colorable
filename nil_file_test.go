package colorable

import (
	"io"
	"testing"
)

func TestNewColorableNil(t *testing.T) {
	w := NewColorable(nil)
	if w != io.Discard {
		// on some platforms may wrap discard - just ensure Write works
		if _, err := w.Write([]byte("x")); err != nil {
			t.Fatal(err)
		}
	}
	// must not panic
	if _, err := NewColorable(nil).Write([]byte("ok")); err != nil {
		t.Fatal(err)
	}
}
