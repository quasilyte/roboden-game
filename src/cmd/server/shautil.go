package main

import (
	"crypto/sha1"
)

func sha1encode(data []byte) string {
	checksum := sha1.Sum(data)
	return string(checksum[:])
}
