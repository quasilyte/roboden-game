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
	GameVersion int    `json:"game_version"`
	GameCommit  string `json:"game_commit"`

	LevelGenChecksum int `json:"level_gen_checksum"`

	Results GameResults `json:"results"`

	Config ReplayLevelConfig `json:"config"`

	Debug ReplayDebugInfo `json:"debug"`

	Actions [][]PlayerAction `json:"actions"`
}

type ReplayDebugInfo struct {
	PlayerName string `json:"player_name"`

	NumPauses      int `json:"num_pauses"`
	NumFastForward int `json:"num_fastforward"`

	Checkpoints []int `json:"checkpoints"`
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
	Resources   int  `json:"resources"`
	GoldEnabled bool `json:"gold_enabled"`

	RawGameMode string `json:"mode"`

	PlayersMode   int `json:"players_mode"`
	InterfaceMode int `json:"ui_mode"`

	Relicts           bool `json:"relicts"`
	FogOfWar          bool `json:"fog_of_war"`
	SuperCreeps       bool `json:"super_creps"`
	CreepFortress     bool `json:"creep_fortress"`
	CoordinatorCreeps bool `json:"coordinator_creeps"`
	AtomicBomb        bool `json:"atomic_bomb"`
	IonMortars        bool `json:"ion_mortars"`

	InitialCreeps         int  `json:"initial_creeps"`
	NumCreepBases         int  `json:"num_creep_bases"`
	CreepDifficulty       int  `json:"creep_difficulty"`
	DronesPower           int  `json:"drones_power"`
	CreepSpawnRate        int  `json:"creep_spawn_rate"`
	TechProgressRate      int  `json:"tech_progress_rate"`
	ReverseSuperCreepRate int  `json:"reverse_super_creep_rate"`
	BossDifficulty        int  `json:"boss_difficulty"`
	ArenaProgression      int  `json:"arena_progression"`
	GameSpeed             int  `json:"game_speed"`
	StartingResources     bool `json:"starting_resources"`

	Teleporters int `json:"teleporters"`

	Seed int64 `json:"seed"`

	WorldSize    int `json:"world_size"`
	OilRegenRate int `json:"oil_regen_rage"`
	Terrain      int `json:"terrain"`
	Environment  int `json:"environment"`

	DifficultyScore int `json:"difficulty"`

	DronePointsAllocated int      `json:"points_allocated"`
	Tier2Recipes         []string `json:"tier2_recipes"`

	TurretDesign string `json:"turret_design"`
	CoreDesign   string `json:"core_design"`
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
