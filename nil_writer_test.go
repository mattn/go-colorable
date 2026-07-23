package colorable

import "testing"

func TestNonColorableNilWriter(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Write panicked: %v", r)
		}
	}()
	w := NewNonColorable(nil)
	n, err := w.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if n != 5 {
		t.Fatalf("n=%d want 5", n)
	}
}
