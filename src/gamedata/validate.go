package gamedata

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

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
	case "classic", "arena", "inf_arena", "reverse", "blitz":
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
	case "classic", "arena", "reverse", "blitz":
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
	if len(replay.Debug.Checkpoints) > 48 {
		return false
	}
	if len(replay.Actions) > 6000 {
		return false
	}
	if (time.Second * time.Duration(replay.Results.Time)) > 8*time.Hour {
		return false
	}

	if replay.Results.Time < 0 || replay.Results.Ticks < 0 || replay.Results.Score < 0 {
		return false
	}

	switch replay.Config.RawGameMode {
	case "blitz", "classic", "arena", "inf_arena", "reverse":
		// OK.
	default:
		return false
	}

	if replay.Config.RawGameMode != "reverse" {
		if replay.Config.EliteFleet {
			return false
		}
		if replay.Config.DronesPower != 1 {
			return false
		}
	}
	switch replay.Config.RawGameMode {
	case "reverse":
		if replay.Config.FogOfWar {
			return false
		}
	case "blitz":
		if replay.Config.NumCreepBases < 2 || replay.Config.NumCreepBases > 4 {
			return false
		}
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
		{cfg.CreepDifficulty, 0, 13},
		{cfg.DronesPower, 0, 7},
		{cfg.TechProgressRate, 0, 8},
		{cfg.CreepSpawnRate, 0, 5},
		{cfg.BossDifficulty, 0, 3},
		{cfg.ArenaProgression, 0, 7},
		{cfg.GameSpeed, 0, 3},
		{cfg.Teleporters, 0, 2},
		{cfg.WorldSize, 0, 3},
		{cfg.WorldShape, 0, 2},
		{cfg.Resources, 0, 4},
		{cfg.OilRegenRate, 0, 3},
		{cfg.Terrain, 0, 2},
		{cfg.InterfaceMode, 0, 2},
		{cfg.Environment, 0, 2},
		{cfg.CreepProductionRate, 0, 10},
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

func CleanUsername(s string) string {
	if IsValidUsername(s) {
		return s
	}

	replacements := []byte{
		'-': '_',
		':', '_',
	}

	needSpace := false
	prevOK := false

	result := make([]byte, 0, len(s))
	for _, r := range s {
		if r > utf8.RuneSelf {
			needSpace = true
			prevOK = false
			continue
		}
		if needSpace && !prevOK {
			needSpace = false
			result = append(result, ' ')
		}
		prevOK = true
		b := byte(r)
		if int(b) < len(replacements) && replacements[b] != 0 {
			b = replacements[b]
		}
		if !isValidChar(b) {
			continue
		}
		result = append(result, b)
	}

	trimmed := strings.TrimSpace(string(result))

	// Remove duplicated spaces.
	parts := strings.Fields(trimmed)
	return strings.Join(parts, " ")
}
