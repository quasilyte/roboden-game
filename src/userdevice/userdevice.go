package userdevice

type Info struct {
	IsMobile bool
}

type SteamInfo struct {
	Initialized bool

	// Whether this game is built with -steam tag.
	Enabled bool

	// Whether this game is running under a Steam Deck device.
	SteamDeck bool
}

type SteamAppConfig struct {
	SteamAppID int
}
