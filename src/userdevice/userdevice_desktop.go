//go:build (linux || darwin || windows) && !steam

package userdevice

func GetInfo() Info {
	return Info{
		IsMobile: false,
	}
}

func GetSteamInfo(config SteamAppConfig) (SteamInfo, error) {
	return SteamInfo{}, nil
}
