//go:build android

package userdevice

func GetInfo() (Info, error) {
	info := Info{
		Kind: DeviceMobile,
		Steam: SteamInfo{
			Initialized: false,
			Enabled:     false,
		},
	}
	return info, nil
}
