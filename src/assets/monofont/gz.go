package monofont

import (
	"bytes"
	"compress/gzip"
	"fmt"
)

func uncompress(data []byte) []byte {
	gzr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		panic(fmt.Errorf("uncompress: %v", err))
	}

	var uncompressed bytes.Buffer
	if _, err := uncompressed.ReadFrom(gzr); err != nil {
		panic(fmt.Errorf("uncompress: %v", err))
	}

	return uncompressed.Bytes()
}
