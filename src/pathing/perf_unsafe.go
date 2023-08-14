package pathing

import (
	"unsafe"
)

type goslice struct {
	data unsafe.Pointer
	len  int
	cap  int
}

func readByte(b []byte, index uint) byte {
	return *(*byte)(unsafe.Add((*goslice)(unsafe.Pointer(&b)).data, index))
}
