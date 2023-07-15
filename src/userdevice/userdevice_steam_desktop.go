//go:build (linux || darwin || windows) && steam

package userdevice

import (
	"errors"
	"github.com/hajimehoshi/go-steamworks"
)

func GetInfo() Info {
	return Info{
		IsMobile: false,
	}
}

func GetSteamInfo(config SteamAppConfig) (SteamInfo, error) {
	info := SteamInfo{
		Enabled: true,
	}

	if config.SteamAppID == 0 {
		return info, nil
	}

	if !steamworks.Init() {
		return info, errors.New("steamworks.Init() failed")
	}
	info.SteamUserID = uint64(steamworks.SteamUser().GetSteamID())

	return info, nil
}
