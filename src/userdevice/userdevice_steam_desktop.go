//go:build (linux || darwin || windows) && steam && !android

package userdevice

import (
	"errors"

	"github.com/hajimehoshi/go-steamworks"
)

func GetInfo() (Info, error) {
	var info Info

	info.Steam.Enabled = true
	if !steamworks.Init() {
		return info, errors.New("steamworks.Init() failed")
	}
	info.Steam.Initialized = true

	info.Kind = DeviceDesktop
	if steamworks.SteamUtils().IsSteamRunningOnSteamDeck() {
		info.Kind = DeviceSteamDeck
	}

	return info, nil
}
