// +build !windows

package colorable

import (
	"os"
)

func NewColorableWriter() *Writer {
	return os.Stdout
}
