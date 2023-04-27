package serverapi

import (
	"github.com/quasilyte/roboden-game/gamedata"
)

type LeaderboardEntry struct {
	Rank       int    `json:"rank"`
	Difficulty int    `json:"difficulty"`
	Score      int    `json:"score"`
	PlayerName string `json:"player_name"`
	Drones     string `json:"drones"`
}

type GameReplay struct {
	Results GameResults `json:"results"`

	Config ReplayLevelConfig `json:"config"`

	Actions []PlayerAction `json:"actions"`
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

type ExecutionMode int

const (
	ExecuteNormal ExecutionMode = iota
	ExecuteSimulation
	ExecuteReplay
)

type ReplayLevelConfig struct {
	Resources int `json:"resources"`

	GameMode gamedata.Mode `json:"mode"`

	ExtraUI      bool `json:"extra_ui"`
	FogOfWar     bool `json:"fog_of_war"`
	InfiniteMode bool `json:"infinite_mode"`

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

	WorldSize int `json:"world_size"`

	DifficultyScore int `json:"difficulty"`

	DronePointsAllocated int      `json:"points_allocated"`
	Tier2Recipes         []string `json:"tier2_recipes"`

	TurretDesign string `json:"turret_design"`
}

type LevelConfig struct {
	ReplayLevelConfig

	ExecMode ExecutionMode

	AttackActionAvailable      bool
	BuildTurretActionAvailable bool
	RadiusActionAvailable      bool
	EliteResources             bool
	EnemyBoss                  bool

	Tutorial    *gamedata.TutorialData
	SecondBase  bool
	ExtraDrones []*gamedata.AgentStats
}

func (options *LevelConfig) Clone() LevelConfig {
	cloned := *options

	cloned.Tier2Recipes = make([]string, len(options.Tier2Recipes))
	copy(cloned.Tier2Recipes, options.Tier2Recipes)

	return cloned
}
