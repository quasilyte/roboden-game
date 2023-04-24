//go:build wasm

package assets

import (
	"bytes"
	"github.com/quasilyte/roboden-game/httpfetch"
	"io"
)

func openfile(path string) (io.ReadCloser, error) {
	data, err := httpfetch.GetBytes(path)
	if err != nil {
		return nil, err
	}
	return &nopCloser{bytes.NewReader(data)}, nil
}

type nopCloser struct {
	io.ReadSeeker
}

func (*nopCloser) Close() error { return nil }
