package serverapi

type LeaderboardEntry struct {
	Rank       int    `json:"rank"`
	Difficulty int    `json:"difficulty"`
	Score      int    `json:"score"`
	PlayerName string `json:"player_name"`
	Drones     string `json:"drones"`
}
