package gamedata

import (
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/serverapi"
)

func CalcDifficultyScore(config serverapi.ReplayLevelConfig, pointsAllocated int) int {
	score := 100

	switch config.RawGameMode {
	case "blitz":
		score += (config.WorldSize - 2) * 30
		score += (config.NumCreepBases - 2) * (25 + config.CreepDifficulty + config.CreepProductionRate)
		if config.FogOfWar {
			if WorldShape(config.WorldShape) == WorldSquare {
				score += 20
			} else {
				score += 10
			}
		}
		if config.GrenadierCreeps {
			if config.CoreDesign != "ark" {
				score += 25
			} else {
				score += 10
			}
			if config.SuperCreeps {
				score += 10
			}
		}
		if config.SuperCreeps {
			score += 10
			score += (config.CreepDifficulty * 15) - 30
			score += (config.CreepProductionRate * 25)
		} else {
			score += (config.CreepDifficulty * 10) - 30
			score += (config.CreepProductionRate * 20)
		}
		score -= (config.Resources - 2) * 20
		if config.Environment == int(EnvInferno) {
			score += 10 - (config.OilRegenRate * 5)
		} else {
			score += 20 - (config.OilRegenRate * 10)
		}
		score += 20 - (1 * pointsAllocated)
		if config.CreepFortress {
			score += 40
		}
		if config.IonMortars {
			score += 25
		}
		if config.InterfaceMode < 2 {
			score += 5
		}
		if !config.GoldEnabled {
			score += 30
		}
		if !config.Relicts {
			score += 10
		}
		if config.CoreDesign != "ark" && config.CoreDesign != "hive" {
			score += 5 - (config.Teleporters * 5)
		}

	case "reverse":
		score -= (config.BossDifficulty - 2) * 20
		score += (3 - config.CreepDifficulty) * 15
		score += (config.DronesPower - 1) * 15
		score -= (config.InitialCreeps - 1) * 20
		score -= (config.TechProgressRate - 6) * 10
		score += (config.OilRegenRate - 2) * 5
		score += (config.Resources - 2) * 20
		score -= (config.ReverseSuperCreepRate - 3) * 15
		if config.EliteFleet {
			score += 20
			switch {
			case config.DronesPower >= 4:
				score += 20
			case config.DronesPower >= 3:
				score += 15
			case config.DronesPower >= 2:
				score += 10
			}
		}
		if config.StartingResources {
			score += 20
		}
		if !config.AtomicBomb {
			score += 15
		}
		if config.CreepFortress {
			score -= 40
		}
		if !config.Relicts {
			score -= 10
		}
		if config.IonMortars {
			score -= 20
		}
		if !config.GoldEnabled {
			score -= 35
		}

	case "classic":
		if config.FogOfWar {
			score += 5
		}
		if config.GrenadierCreeps {
			if config.CoreDesign != "ark" {
				score += 30
			} else {
				score += 10
			}
			if config.SuperCreeps {
				score += 15
			}
		}
		if config.CoordinatorCreeps {
			score += 15 * config.NumCreepBases
		}
		if config.CreepFortress {
			score += 30
		}
		if config.IonMortars {
			score += 10
		}
		if config.InterfaceMode < 2 {
			score += 5
		}
		if !config.GoldEnabled {
			score += 25
		}
		if !config.Relicts {
			score += 15
		}
		if config.NumCreepBases != 0 {
			score += (config.CreepDifficulty - 3) * 15
			if config.SuperCreeps {
				score += 50
			}
		} else {
			score += (config.CreepDifficulty - 3) * 10
			if config.SuperCreeps {
				score += 35
			}
		}
		score -= (config.Resources - 2) * 20
		score += (config.NumCreepBases - 2) * 15
		score += (config.BossDifficulty - 1) * 25
		switch config.BossDifficulty {
		case 0:
			// Extra penalty for the weakest boss.
			score -= 20
		case 1:
			// Nothing special.
		case 2:
			if config.CoordinatorCreeps {
				score += 5
			}
		case 3:
			// Extra points for the strongest boss.
			score += 15
			if config.CoordinatorCreeps {
				score += 10
			}
		}
		score += (config.CreepSpawnRate - 1) * 10
		score += (config.InitialCreeps - 1) * 15
		if config.Environment == int(EnvInferno) {
			score += 10 - (config.OilRegenRate * 5)
		} else {
			score += 20 - (config.OilRegenRate * 10)
		}
		score += 20 - (pointsAllocated)
		if config.StartingResources {
			score -= 10
		}
		if config.CoreDesign != "ark" && config.CoreDesign != "hive" {
			score += 5 - (config.Teleporters * 5)
		}

	case "arena", "inf_arena":
		if config.FogOfWar {
			score += 5
		}
		if config.GrenadierCreeps {
			grenadierScore := 40
			if config.CoreDesign == "ark" {
				grenadierScore = 15
			}
			if config.SuperCreeps {
				grenadierScore += 15
			}
			if config.RawGameMode == "inf_arena" {
				grenadierScore *= 2
			}
			score += grenadierScore
		}
		if config.CreepFortress {
			score += 35
		}
		if !config.Relicts {
			score += 20
		}
		if config.IonMortars {
			score += 15
		}
		if config.InterfaceMode == 0 {
			score += 5
		}
		if !config.GoldEnabled {
			score += 35
		}
		score += (config.ArenaProgression - 1) * 20
		if config.RawGameMode == "inf_arena" {
			score -= (config.Resources - 2) * 30
			score += config.InitialCreeps * 5
		} else {
			score -= (config.Resources - 2) * 15
			score += config.InitialCreeps * 10
		}
		score += (config.CreepDifficulty - 3) * 20
		if config.Environment == int(EnvInferno) {
			score += 10 - (config.OilRegenRate * 5)
		} else {
			score += 30 - (config.OilRegenRate * 15)
		}
		score += 40 - (2 * pointsAllocated)
		if config.StartingResources {
			score -= 5
		}
		if config.CoreDesign != "ark" && config.CoreDesign != "hive" {
			score += 5 - (config.Teleporters * 5)
		}
	}

	return gmath.ClampMin(score, 1)
}
