//go:build linux || darwin || windows

package userdevice

func GetInfo() Info {
	return Info{
		IsMobile: false,
	}
}
