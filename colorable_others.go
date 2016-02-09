// +build !windows

package colorable

import (
	"io"
	"os"
)

func NewColorable(file *os.File) io.Writer {
	return file
}

func NewColorableStdout() io.Writer {
	return os.Stdout
}

func NewColorableStderr() io.Writer {
	return os.Stderr
}
