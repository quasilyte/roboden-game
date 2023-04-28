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

func gzipCompress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	b.Grow(len(data) / 3)

	w := gzip.NewWriter(&b)

	if _, err := w.Write(data); err != nil {
		return nil, err
	}

	if err := w.Flush(); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
