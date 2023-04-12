//go:build linux || darwin || windows

package assets

import (
	"io"
	"os"
)

func openfile(path string) (io.ReadCloser, error) {
	return os.Open(path)
}
