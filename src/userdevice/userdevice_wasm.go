//go:build wasm

package userdevice

import (
	"syscall/js"
)

func GetInfo(config ApplicationConfig) Info {
	var result Info
	result.IsMobile = js.Global().Call("matchMedia", "(hover: none)").Get("matches").Bool()
	return result
}

func GetSteamInfo(config SteamAppConfig) (SteamInfo, error) {
	return SteamInfo{}, nil
}
