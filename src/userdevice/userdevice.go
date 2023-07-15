package userdevice

type Info struct {
	IsMobile bool
}

type SteamInfo struct {
	Initialized bool

	// Whether this game is built with -steam tag.
	Enabled bool
}

type SteamAppConfig struct {
	SteamAppID int
}
