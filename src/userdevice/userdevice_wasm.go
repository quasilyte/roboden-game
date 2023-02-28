//go:build wasm

package userdevice

import (
	"syscall/js"
)

func GetInfo() Info {
	var result Info
	result.IsMobile = js.Global().Call("matchMedia", "(hover: none)").Get("matches").Bool()
	return result
}
