//go:build wasm

package userdevice

import (
	"syscall/js"
)

func GetInfo() (Info, error) {
	var result Info
	result.Kind = DeviceDesktop
	if js.Global().Call("matchMedia", "(hover: none)").Get("matches").Bool() {
		result.Kind = DeviceMobile
	}
	return result, nil
}
