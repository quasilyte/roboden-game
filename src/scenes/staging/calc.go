package staging

import (
	"math"

	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/gamedata"
)

func calcScore(world *worldState) int {
	score := world.config.DifficultyScore * 10
	crystalsCollected := gmath.Percentage(world.result.RedCrystalsCollected, world.numRedCrystals)
	score += crystalsCollected * 3
	multiplier := 1.0 - (0.000347222 * (world.result.TimePlayed.Seconds() / 5))
	if multiplier < 0 {
		multiplier = 0.001
	}
	return int(math.Round(float64(score) * multiplier))
}

func mergeAgents(world *worldState, x, y *colonyAgentNode) *gamedata.AgentStats {
	list := world.tier2recipes
	if x.stats.Tier == 2 {
		list = gamedata.Tier3agentMergeRecipes
	}
	for _, recipe := range list {
		if recipe.Match(x.AsRecipeSubject(), y.AsRecipeSubject()) {
			return recipe.Result
		}
	}
	return nil
}

func agentCloningEnergyCost() float64 {
	return 30.0
}

func agentCloningCost(core *colonyCoreNode, cloner, a *colonyAgentNode) float64 {
	multiplier := 0.85
	return a.stats.Cost * multiplier
}

func resourceScore(core *colonyCoreNode, source *essenceSourceNode) float64 {
	if source.stats.regenDelay != 0 && source.percengage < 0.15 {
		return 0
	}
	if core.failedResource == source {
		return 0
	}
	dist := core.pos.DistanceTo(source.pos)
	maxDist := 1.5 + (core.GetResourcePriority() * 0.5)
	if dist > core.realRadius*maxDist || source.resource == 0 {
		return 0
	}
	distScore := 8.0 - gmath.ClampMax(dist/120, 8.0)
	multiplier := 1.0 + (source.stats.value * 0.1)
	if source.stats.regenDelay == 0 {
		multiplier += 0.3
	}
	if dist > core.realRadius*1.2 {
		multiplier -= 0.5
	}
	if source.percengage <= 0.25 {
		multiplier += 0.35
	} else if source.percengage <= 0.5 {
		multiplier += 0.1
	}
	return gmath.ClampMin(distScore*multiplier, 0.01)
}
