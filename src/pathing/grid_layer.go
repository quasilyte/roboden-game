package pathing

import "unsafe"

type GridLayer uint32

func MakeGridLayer(v0, v1, v2, v3 uint8) GridLayer {
	merged := uint32(v0) | uint32(v1)<<8 | uint32(v2)<<16 | uint32(v3)<<24
	return GridLayer(merged)
}

func (l GridLayer) Get(tag uint8) uint8 {
	return uint8(l >> (uint32(tag) * 8))
}

func (l GridLayer) getFast(tag uint8) uint8 {
	return *(*uint8)(unsafe.Add(unsafe.Pointer(&l), tag))
}
