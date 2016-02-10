// +build !windows

package colorable

import (
	"io"
	"os"
)

func NewColorableStdout() io.WriteCloser {
	return os.Stdout
}

func NewColorableStderr() io.WriteCloser {
	return os.Stderr
}
