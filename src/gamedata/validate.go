package gamedata

import (
	"time"

	"github.com/quasilyte/roboden-game/serverapi"
)

func IsSendableReplay(r serverapi.GameReplay) bool {
	if r.Results.Score <= 0 {
		return false
	}
	switch r.Config.RawGameMode {
	case "classic", "arena":
		// There is no point in running a non-victory game replay
		// for a mode that can be won.
		if !r.Results.Victory {
			return false
		}
	case "inf_arena":
		// Infinite arena can't be won.
		if r.Results.Victory {
			return false
		}
	}
	return true
}

func IsValidReplay(replay serverapi.GameReplay) bool {
	if replay.GameVersion < 0 {
		return false
	}
	if len(replay.Actions) > 2000 {
		return false
	}
	if (time.Second * time.Duration(replay.Results.Time)) > 8*time.Hour {
		return false
	}

	if replay.Results.Time < 0 || replay.Results.Ticks < 0 || replay.Results.Score < 0 {
		return false
	}

	switch replay.Config.RawGameMode {
	case "classic", "arena", "inf_arena":
		// OK.
	default:
		return false
	}

	cfg := &replay.Config

	pointsAllocated := 0
	for _, droneName := range cfg.Tier2Recipes {
		recipe := findRecipeByName(droneName)
		if recipe.Result == nil {
			return false
		}
		pointsAllocated += recipe.Result.PointCost
	}
	if pointsAllocated > 20 {
		return false
	}

	difficultyScore := CalcDifficultyScore(replay.Config, pointsAllocated)
	if difficultyScore != replay.Config.DifficultyScore {
		return false
	}

	type optionValidator struct {
		actual int
		min    int
		max    int
	}
	toValidate := [...]optionValidator{
		{cfg.InitialCreeps, 0, 2},
		{cfg.NumCreepBases, 0, 4},
		{cfg.CreepDifficulty, 0, 7},
		{cfg.CreepSpawnRate, 0, 3},
		{cfg.BossDifficulty, 0, 3},
		{cfg.ArenaProgression, 0, 6},
		{cfg.StartingResources, 0, 2},
		{cfg.GameSpeed, 0, 2},
		{cfg.Teleporters, 0, 2},
		{cfg.WorldSize, 0, 3},
		{cfg.Resources, 0, 4},
		{cfg.OilRegenRate, 0, 3},
		{cfg.Terrain, 0, 2},
	}
	for _, o := range toValidate {
		if o.actual < o.min || o.actual > o.max {
			return false
		}
	}

	return true
}
