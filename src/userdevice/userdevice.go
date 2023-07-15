package userdevice

type Info struct {
	IsMobile bool
}

type SteamInfo struct {
	// Whether this game is built with -steam tag.
	Enabled bool

	// SteamUserID contains a Steam user ID if the game
	// can detect the Steam environment.
	SteamUserID uint64
}

type SteamAppConfig struct {
	SteamAppID int
}
