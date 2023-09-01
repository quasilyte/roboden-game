//go:build (linux || darwin || windows) && !steam

package userdevice

func GetInfo() (Info, error) {
	info := Info{
		Kind: DeviceDesktop,
		Steam: SteamInfo{
			Initialized: false,
			Enabled:     false,
		},
	}
	return info, nil
}
