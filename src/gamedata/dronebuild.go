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
