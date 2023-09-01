package userdevice

type Info struct {
	Kind DeviceKind

	Steam SteamInfo
}

func (info Info) IsDesktop() bool { return info.Kind == DeviceDesktop }

func (info Info) IsMobile() bool { return info.Kind == DeviceMobile }

func (info Info) IsSteamDeck() bool { return info.Kind == DeviceSteamDeck }

type DeviceKind int

const (
	DeviceDesktop DeviceKind = iota
	DeviceMobile
	DeviceSteamDeck
)

type SteamInfo struct {
	Initialized bool

	// Whether this game is built with -steam tag.
	Enabled bool
}
