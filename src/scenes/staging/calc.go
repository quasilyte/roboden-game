package staging

import (
	"github.com/quasilyte/gmath"
)

func mergeAgents(x, y *colonyAgentNode) *agentStats {
	list := tier2agentMergeRecipeList
	if x.stats.tier == 2 {
		list = tier3agentMergeRecipeList
	}
	for _, recipe := range list {
		if recipe.Match1(x) && recipe.Match2(y) {
			return recipe.result
		}
		if recipe.Match1(y) && recipe.Match2(x) {
			return recipe.result
		}
	}
	return nil
}

func agentCloningEnergyCost() float64 {
	return 30.0
}

func agentCloningCost(core *colonyCoreNode, cloner, a *colonyAgentNode) float64 {
	multiplier := 0.85
	if cloner.faction == greenFactionTag {
		multiplier = 0.6
	}
	return a.stats.cost * multiplier
}

func resourceScore(core *colonyCoreNode, source *essenceSourceNode) float64 {
	if source.stats == redOilSource && !core.hasRedMiner {
		return 0
	}
	if source.stats.regenDelay != 0 && source.percengage < 0.15 {
		return 0
	}
	dist := core.pos.DistanceTo(source.pos)
	maxDist := 1.5 + (core.GetResourcePriority() * 0.5)
	if dist > core.realRadius*maxDist || source.resource == 0 {
		return 0
	}
	distScore := 4.0 - gmath.ClampMax(dist/200, 4.0)
	percentagePenalty := 0.0
	if source.percengage < 0.1 {
		percentagePenalty += 0.55
	}
	if dist > core.realRadius*1.2 {
		percentagePenalty += 0.6
	}
	multiplier := 1.0 + (source.stats.value * 0.4)
	if source.stats.regenDelay == 0 {
		multiplier += 0.3
	}
	return gmath.ClampMin((distScore+(source.percengage*3)-percentagePenalty)*multiplier, 0.01)
}
