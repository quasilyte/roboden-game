package gamedata

import (
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/serverapi"
)

func CalcDifficultyScore(config serverapi.ReplayLevelConfig, pointsAllocated int) int {
	score := 100

	switch config.RawGameMode {
	case "classic":
		if config.NumCreepBases != 0 {
			score += (config.CreepDifficulty - 1) * 10
			if config.SuperCreeps {
				score += 45
			}
		} else {
			score += (config.CreepDifficulty - 1) * 5
			if config.SuperCreeps {
				score += 35
			}
		}
		score += 10 - (config.Resources * 5)
		score += (config.NumCreepBases - 2) * 15
		score += (config.BossDifficulty - 1) * 15
		score += (config.CreepSpawnRate - 1) * 10
		score += (config.InitialCreeps - 1) * 10
		score -= config.StartingResources * 4
	case "arena", "inf_arena":
		score += 10 - (config.Resources * 5)
		score += (config.CreepDifficulty - 1) * 15
		score += config.InitialCreeps * 5
		score += (config.ArenaProgression - 1) * 10
		score -= config.StartingResources * 2
	}

	score += 10 - (config.OilRegenRate * 5)

	if !config.ExtraUI {
		score += 5
	}
	if config.FogOfWar {
		score += 5
	}

	score += 5 - (config.Teleporters * 5)

	score += 20 - pointsAllocated

	return gmath.ClampMin(score, 1)
}
