package serverapi

type LeaderboardEntry struct {
	Rank       int    `json:"rank"`
	Difficulty int    `json:"difficulty"`
	Score      int    `json:"score"`
	Time       int    `json:"time"`
	PlayerName string `json:"player_name"`
	Drones     string `json:"drones"`
}

type GameReplay struct {
	GameVersion int `json:"game_version"`

	Results GameResults `json:"results"`

	Config ReplayLevelConfig `json:"config"`

	Actions [][]PlayerAction `json:"actions"`
}

type GameResults struct {
	Time    int  `json:"time"`
	Ticks   int  `json:"ticks"`
	Score   int  `json:"score"`
	Victory bool `json:"victory"`
}

type PlayerAction struct {
	Tick           int              `json:"tick"`
	Pos            [2]float64       `json:"pos"`
	Kind           PlayerActionKind `json:"kind"`
	SelectedColony int              `json:"selected_colony"`
}

const (
	PmodeSinglePlayer int = iota
	PmodeSingleBot
	PmodePlayerAndBot
	PmodeTwoPlayers
	PmodeTwoBots
)

type PlayerActionKind int

const (
	ActionUnknown PlayerActionKind = iota
	ActionCard1
	ActionCard2
	ActionCard3
	ActionCard4
	ActionCard5
	ActionMove
)

type ReplayLevelConfig struct {
	Resources int `json:"resources"`

	RawGameMode string `json:"mode"`

	PlayersMode int `json:"players_mode"`

	ExtraUI      bool `json:"extra_ui"`
	FogOfWar     bool `json:"fog_of_war"`
	InfiniteMode bool `json:"infinite_mode"`
	SuperCreeps  bool `json:"super_creps"`

	InitialCreeps     int `json:"initial_creeps"`
	NumCreepBases     int `json:"num_creep_bases"`
	CreepDifficulty   int `json:"creep_difficulty"`
	CreepSpawnRate    int `json:"creep_spawn_rate"`
	BossDifficulty    int `json:"boss_difficulty"`
	ArenaProgression  int `json:"arena_progression"`
	StartingResources int `json:"starting_resources"`
	GameSpeed         int `json:"game_speed"`

	Teleporters int `json:"teleporters"`

	Seed int64 `json:"seed"`

	WorldSize    int `json:"world_size"`
	OilRegenRate int `json:"oil_regen_rage"`
	Terrain      int `json:"terrain"`

	DifficultyScore int `json:"difficulty"`

	DronePointsAllocated int      `json:"points_allocated"`
	Tier2Recipes         []string `json:"tier2_recipes"`

	TurretDesign string `json:"turret_design"`
}

type LeaderboardResp struct {
	NumSeasons int                `json:"num_seasons"`
	NumPlayers int                `json:"num_players"`
	Entries    []LeaderboardEntry `json:"entries"`
}

type SavePlayerScoreResp struct {
	Queued           bool `json:"queued"`
	CurrentHighscore int  `json:"current_highscore"`
}
