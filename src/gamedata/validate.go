package gamedata

import (
	"fmt"
	"time"

	"github.com/quasilyte/roboden-game/serverapi"
)

func Validate() {
	recipes := map[[2]RecipeSubject]string{}
	for _, r := range Tier2agentMergeRecipes {
		k1 := [2]RecipeSubject{r.Drone1, r.Drone2}
		k2 := [2]RecipeSubject{r.Drone2, r.Drone1}
		var resultName string
		if name, ok := recipes[k1]; ok {
			resultName = name
		}
		if name, ok := recipes[k2]; ok {
			resultName = name
		}
		if resultName != "" {
			panic(fmt.Sprintf("%s and %s recipe conflict", r.Result.Kind, resultName))
		}
		recipes[k1] = r.Result.Kind.String()
		recipes[k2] = r.Result.Kind.String()
	}
}

func IsRunnableReplay(r serverapi.GameReplay) bool {
	switch r.Config.RawGameMode {
	case "classic", "arena", "inf_arena", "reverse":
		return true
	default:
		return false
	}
}

func IsSendableReplay(r serverapi.GameReplay) bool {
	if !IsRunnableReplay(r) {
		return false
	}
	if GetSeedKind(r.Config.Seed, r.Config.RawGameMode) != SeedNormal {
		return false
	}
	if r.Results.Score <= 0 {
		return false
	}
	if r.Config.PlayersMode != serverapi.PmodeSinglePlayer {
		return false
	}
	switch r.Config.RawGameMode {
	case "classic", "arena", "reverse":
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
	case "classic", "arena", "inf_arena", "reverse":
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
		{cfg.NumCreepBases, 0, 5},
		{cfg.CreepDifficulty, 0, 11},
		{cfg.DronesPower, 0, 6},
		{cfg.TechProgressRate, 0, 7},
		{cfg.CreepSpawnRate, 0, 3},
		{cfg.BossDifficulty, 0, 3},
		{cfg.ArenaProgression, 0, 7},
		{cfg.StartingResources, 0, 2},
		{cfg.GameSpeed, 0, 3},
		{cfg.Teleporters, 0, 2},
		{cfg.WorldSize, 0, 3},
		{cfg.Resources, 0, 4},
		{cfg.OilRegenRate, 0, 3},
		{cfg.Terrain, 0, 2},
		{cfg.InterfaceMode, 0, 2},
		{cfg.PlayersMode, serverapi.PmodeSinglePlayer, serverapi.PmodeTwoBots},
	}
	for _, o := range toValidate {
		if o.actual < o.min || o.actual > o.max {
			return false
		}
	}

	return true
}

func isValidChar(ch byte) bool {
	isLetter := func(ch byte) bool {
		return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
	}
	isDigit := func(ch byte) bool {
		return ch >= '0' && ch <= '9'
	}
	return isLetter(ch) || isDigit(ch) || ch == ' ' || ch == '_'
}

func IsValidUsername(s string) bool {
	nonSpace := 0
	if len(s) > serverapi.MaxNameLength {
		return false
	}
	for i := 0; i < len(s); i++ {
		ch := s[i]
		isValid := isValidChar(ch)
		if !isValid {
			return false
		}
		if ch != ' ' {
			nonSpace++
		}
	}
	return nonSpace != 0
}
