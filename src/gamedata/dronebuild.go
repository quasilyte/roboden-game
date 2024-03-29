package gamedata

import (
	"sort"

	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
)

func PickColonyDesign(designsUnlocked []string, rng *gmath.Rand) string {
	picker := gmath.NewRandPicker[string](rng)
	for _, c := range designsUnlocked {
		switch c {
		case "den":
			picker.AddOption(c, 0.3)
		case "ark":
			picker.AddOption(c, 0.3)
		case "tank":
			picker.AddOption(c, 0.3)
		case "hive":
			picker.AddOption(c, 0.25)
		}
	}
	return picker.Pick()
}

func PickTurretDesign(coreDesign string, designsUnlocked []string, rng *gmath.Rand) string {
	picker := gmath.NewRandPicker[string](rng)
	for _, t := range designsUnlocked {
		switch t {
		case "BeamTower":
			picker.AddOption(t, 0.35)
		case "Gunpoint":
			if coreDesign == "hive" {
				picker.AddOption(t, 0.05)
			} else {
				picker.AddOption(t, 0.25)
			}
		case "Harvester":
			picker.AddOption(t, 0.15)
		case "TetherBeacon":
			picker.AddOption(t, 0.2)
		case "Siege":
			if coreDesign == "hive" {
				picker.AddOption(t, 0.15)
			} else {
				picker.AddOption(t, 0.2)
			}
		case "Sentinelpoint":
			picker.AddOption(t, 0.25)
		case "Refinery":
			if coreDesign == "hive" {
				picker.AddOption(t, 0.4)
			} else {
				picker.AddOption(t, 0.15)
			}
		}
	}
	return picker.Pick()
}

func RandDroneBuild(rng *gmath.Rand, dronesUnlocked []string) []string {
	points := ClassicModePoints

	numWorkers := 0

	var result []string

	selection := make([]string, len(dronesUnlocked))
	copy(selection, dronesUnlocked)
	gmath.Shuffle(rng, selection)

	maxWorkers := rng.IntRange(3, 6)
	maxDroneCost := 4
	if rng.Chance(0.15) {
		maxDroneCost = 3
	}
	minDroneCost := 1
	if rng.Chance(0.3) {
		minDroneCost = 2
	}

	for _, d := range selection {
		if points < 2 && rng.Chance(0.3) {
			break
		}

		recipe := findRecipeByName(d)
		if recipe.Result == nil {
			continue
		}
		stats := recipe.Result

		if stats.PointCost < minDroneCost {
			continue
		}
		if stats.PointCost > maxDroneCost {
			continue
		}
		if points < stats.PointCost {
			continue
		}

		skipChance := 0.0
		switch stats.Kind {
		case AgentRoomba:
			skipChance = 0.45
		case AgentCourier, AgentScarab:
			skipChance = 0.2
		case AgentFreighter, AgentCommander:
			skipChance = 0.05
		case AgentFirebug:
			skipChance = 0.1
		}
		if skipChance != 0 && rng.Chance(skipChance) {
			continue
		}

		isWorker := stats.CanGather && !stats.CanPatrol
		if isWorker && numWorkers >= maxWorkers {
			continue
		}

		if isWorker {
			numWorkers++
		}
		points -= stats.PointCost
		result = append(result, d)
	}

	return result
}

func CreateDroneBuild(rng *gmath.Rand) []string {
	points := ClassicModePoints

	popRand := func(list []*AgentStats) (*AgentStats, []*AgentStats) {
		i := gmath.RandIndex(rng, list)
		var result *AgentStats
		if points >= list[i].PointCost {
			result = list[i]
			points -= result.PointCost
			list = xslices.RemoveAt(list, i)
		}
		return result, list
	}

	drones := make([]string, 0, 8)

	if rng.Chance(0.1) {
		minAllocated := ClassicModePoints
		if rng.Chance(0.6) {
			minAllocated = rng.IntRange(10, ClassicModePoints)
		}
		minPoints := ClassicModePoints - minAllocated

		selection := []*AgentStats{
			FighterAgentStats,
			PrismAgentStats,
			ScarabAgentStats,
			SkirmisherAgentStats,
			MortarAgentStats,
			AntiAirAgentStats,
			FirebugAgentStats,
			DefenderAgentStats,
			CommanderAgentStats,
		}
		if minAllocated >= 15 {
			selection = append(selection, TargeterAgentStats)
		}
		for round := 0; round < 7; round++ {
			if points <= minPoints {
				break
			}
			var stats *AgentStats
			stats, selection = popRand(selection)
			if stats != nil {
				drones = append(drones, stats.Kind.String())
			}
		}
		sort.Strings(drones)
		return drones
	}

	coreSupportDrones := []*AgentStats{
		RedminerAgentStats,
		RepairAgentStats,
		RechargerAgentStats,
		ClonerAgentStats,
	}
	carryDrones := []*AgentStats{
		FighterAgentStats,
		PrismAgentStats,
		ScarabAgentStats,
		SkirmisherAgentStats,
	}
	supportFireDrones := []*AgentStats{
		ScavengerAgentStats,
		CripplerAgentStats,
		RepellerAgentStats,
		DisintegratorAgentStats,
		MortarAgentStats,
		AntiAirAgentStats,
		FirebugAgentStats,
	}
	extraDrones := []*AgentStats{
		RoombaAgentStats,
		KamikazeAgentStats,
		FreighterAgentStats,
		CourierAgentStats,
		ServoAgentStats,
		GeneratorAgentStats,
		DefenderAgentStats,
		CommanderAgentStats,
		TargeterAgentStats,
	}
	for round := 0; round < 3; round++ {
		if points <= 0 {
			break
		}
		var stats *AgentStats
		numCore := rng.IntRange(1, 3)
		for i := 0; i < numCore && len(coreSupportDrones) != 0; i++ {
			stats, coreSupportDrones = popRand(coreSupportDrones)
			if stats != nil {
				drones = append(drones, stats.Kind.String())
			}
		}
		stats, carryDrones = popRand(carryDrones)
		if stats != nil {
			drones = append(drones, stats.Kind.String())
		}
		numSupportFire := rng.IntRange(0, 2)
		for i := 0; i < numSupportFire && len(supportFireDrones) != 0; i++ {
			stats, supportFireDrones = popRand(supportFireDrones)
			if stats != nil {
				drones = append(drones, stats.Kind.String())
			}
		}
		numExtra := rng.IntRange(0, 2)
		for i := 0; i < numExtra && len(extraDrones) != 0; i++ {
			stats, extraDrones = popRand(extraDrones)
			if stats != nil {
				drones = append(drones, stats.Kind.String())
			}
		}
	}

	sort.Strings(drones)
	return drones
}
