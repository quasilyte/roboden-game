package gamedata

const (
	ClassicModePoints int = 20
)

func BlitzModeSetupTime(numPlayers int) float64 {
	if numPlayers == 1 {
		return 5 * 60
	}
	return 3 * 60
}
