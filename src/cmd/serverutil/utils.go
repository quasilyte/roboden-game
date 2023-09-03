package main

import (
	"bytes"
	"compress/gzip"
	"io"
)

func gzipUncompress(data []byte) (resData []byte, err error) {
	buf := bytes.NewBuffer(data)
	r, err := gzip.NewReader(buf)
	if err != nil {
		return
	}
	return io.ReadAll(r)
}
