package gamedata

import (
	"fmt"

	"github.com/quasilyte/roboden-game/serverapi"
)

type ExecutionMode int

const (
	ExecuteNormal ExecutionMode = iota
	ExecuteDemo
	ExecuteSimulation
	ExecuteReplay
)

type PlayerKind int

const (
	PlayerNone PlayerKind = iota
	PlayerHuman
	PlayerComputer
)

func (pk PlayerKind) String() string {
	switch pk {
	case PlayerNone:
		return "none"
	case PlayerHuman:
		return "human"
	case PlayerComputer:
		return "computer"
	default:
		return "?"
	}
}

func MakeLevelConfig(mode ExecutionMode, config serverapi.ReplayLevelConfig) LevelConfig {
	enemyBoss := config.RawGameMode == "classic" ||
		config.RawGameMode == "reverse"
	return LevelConfig{
		ReplayLevelConfig:          config,
		ExecMode:                   mode,
		AttackActionAvailable:      true,
		BuildTurretActionAvailable: true,
		RadiusActionAvailable:      true,
		EliteResources:             true,
		EnemyBoss:                  enemyBoss,
	}
}

type LevelConfig struct {
	serverapi.ReplayLevelConfig

	Players []PlayerKind

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

func (config *LevelConfig) Finalize() {
	switch config.RawGameMode {
	case "inf_arena":
		config.GameMode = ModeInfArena
	case "arena":
		config.GameMode = ModeArena
	case "classic":
		config.GameMode = ModeClassic
	case "reverse":
		config.GameMode = ModeReverse
	default:
		panic(fmt.Sprintf("unexpected game mode: %q", config.RawGameMode))
	}

	if config.GameMode == ModeReverse {
		switch config.PlayersMode {
		case serverapi.PmodeSinglePlayer:
			config.Players = []PlayerKind{PlayerHuman, PlayerComputer}
		case serverapi.PmodeTwoPlayers:
			config.Players = []PlayerKind{PlayerHuman, PlayerHuman}
		default:
			panic(fmt.Sprintf("unexpected mode: %d", config.PlayersMode))
		}
	} else {
		switch config.PlayersMode {
		case serverapi.PmodeSinglePlayer:
			config.Players = []PlayerKind{PlayerHuman}
		case serverapi.PmodeSingleBot:
			config.Players = []PlayerKind{PlayerComputer}
		case serverapi.PmodePlayerAndBot:
			config.Players = []PlayerKind{PlayerHuman, PlayerComputer}
		case serverapi.PmodeTwoPlayers:
			config.Players = []PlayerKind{PlayerHuman, PlayerHuman}
		case serverapi.PmodeTwoBots:
			config.Players = []PlayerKind{PlayerComputer, PlayerComputer}
		default:
			panic(fmt.Sprintf("unexpected mode: %d", config.PlayersMode))
		}
	}
}

func (config *LevelConfig) Clone() LevelConfig {
	cloned := *config

	cloned.Tier2Recipes = make([]string, len(config.Tier2Recipes))
	copy(cloned.Tier2Recipes, config.Tier2Recipes)

	return cloned
}
