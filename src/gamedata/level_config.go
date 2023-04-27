package gamedata

import (
	"github.com/quasilyte/roboden-game/serverapi"
)

type ExecutionMode int

const (
	ExecuteNormal ExecutionMode = iota
	ExecuteSimulation
	ExecuteReplay
)

type LevelConfig struct {
	serverapi.ReplayLevelConfig

	GameMode Mode

	ExecMode ExecutionMode

	AttackActionAvailable      bool
	BuildTurretActionAvailable bool
	RadiusActionAvailable      bool
	EliteResources             bool
	EnemyBoss                  bool

	Tutorial    *TutorialData
	SecondBase  bool
	ExtraDrones []*AgentStats
}

func (options *LevelConfig) Clone() LevelConfig {
	cloned := *options

	cloned.Tier2Recipes = make([]string, len(options.Tier2Recipes))
	copy(cloned.Tier2Recipes, options.Tier2Recipes)

	return cloned
}
