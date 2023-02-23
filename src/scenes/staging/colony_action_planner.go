package staging

import (
	"math"

	"github.com/quasilyte/gmath"
)

type colonyActionPlanner struct {
	colony *colonyCoreNode

	numTier1Agents    int
	numTier2Agents    int
	numPatrolAgents   int
	numGarrisonAgents int

	leadingFaction             factionTag
	leadingFactionAgents       float64
	leadingFactionCombatAgents float64

	world *worldState

	priorityPicker *gmath.RandPicker[colonyPriority]
}

func newColonyActionPlanner(colony *colonyCoreNode, rand *gmath.Rand) *colonyActionPlanner {
	return &colonyActionPlanner{
		colony:         colony,
		world:          colony.world,
		priorityPicker: gmath.NewRandPicker[colonyPriority](rand),
	}
}

func (p *colonyActionPlanner) PickAction() colonyAction {
	p.leadingFaction = p.colony.factionWeights.MaxKey()
	p.numPatrolAgents = 0
	p.numTier2Agents = 0
	leadingFactionAgents := 0
	leadingFactionCombatAgents := 0

	p.colony.agents.Update()

	p.colony.agents.Each(func(a *colonyAgentNode) {
		switch a.stats.tier {
		case 1:
			p.numTier1Agents++
		case 2:
			p.numTier2Agents++
		}
		if !a.stats.canPatrol {
			if a.faction == p.leadingFaction {
				leadingFactionAgents++
			}
			return
		}
		if a.faction == p.leadingFaction {
			leadingFactionCombatAgents++
		}
		if a.mode == agentModePatrol {
			p.numPatrolAgents++
		}
		if a.mode == agentModeStandby {
			p.numGarrisonAgents++
		}
	})

	p.leadingFactionAgents = float64(leadingFactionAgents) / float64(len(p.colony.agents.workers))
	p.leadingFactionCombatAgents = float64(leadingFactionCombatAgents) / float64(len(p.colony.agents.fighters))

	p.priorityPicker.Reset()
	p.priorityPicker.AddOption(priorityResources, p.colony.GetResourcePriority())
	p.priorityPicker.AddOption(priorityGrowth, p.colony.GetGrowthPriority())
	p.priorityPicker.AddOption(prioritySecurity, p.colony.GetSecurityPriority())
	p.priorityPicker.AddOption(priorityEvolution, p.colony.GetEvolutionPriority())

	actionKind := p.priorityPicker.Pick()
	switch actionKind {
	case priorityResources:
		return p.pickGatherAction()
	case priorityGrowth:
		return p.pickGrowthAction()
	case prioritySecurity:
		return p.pickSecurityAction()
	case priorityEvolution:
		return p.pickEvolutionAction()
	}

	panic("unreachable")
}

func (p *colonyActionPlanner) pickGatherAction() colonyAction {
	if len(p.world.essenceSources) == 0 {
		return colonyAction{}
	}
	if p.colony.agents.NumAvailableWorkers() == 0 {
		return colonyAction{}
	}
	baseNeedsResources := p.colony.resources <= maxVisualResources
	if !baseNeedsResources {
		return colonyAction{}
	}

	var bestSource *essenceSourceNode
	bestScore := 0.0
	for _, source := range p.world.essenceSources {
		score := resourceScore(p.colony, source) * p.world.rand.FloatRange(0.8, 1.4)
		if score > bestScore {
			bestScore = score
			bestSource = source
		}
	}
	if bestSource != nil {
		return colonyAction{
			Kind:     actionMineEssence,
			Value:    bestSource,
			TimeCost: 0.75,
		}
	}

	return colonyAction{}
}

func (p *colonyActionPlanner) combatUnitProbability() float64 {
	minCombatUnits := int(p.colony.GetSecurityPriority() * 20)
	if p.colony.GetSecurityPriority() > 0.1 && len(p.colony.agents.fighters) < minCombatUnits {
		return 0.9
	}
	wantedCombatAgentRatio := p.colony.GetSecurityPriority() * 0.8
	currentCombatAgentRatio := float64(len(p.colony.agents.fighters)) / float64(len(p.colony.agents.workers))
	if currentCombatAgentRatio < wantedCombatAgentRatio {
		return 0.75
	}
	return 0
}

func (p *colonyActionPlanner) pickCloner() *colonyAgentNode {
	var bestCandidate *colonyAgentNode
	p.colony.agents.Find(searchWorkers|searchOnlyAvailable, func(a *colonyAgentNode) bool {
		if a.energy < agentCloningEnergyCost() || a.energyBill != 0 {
			return true
		}
		bestCandidate = a
		return a.faction == greenFactionTag
	})
	return bestCandidate
}

func (p *colonyActionPlanner) pickUnitToClone(cloner *colonyAgentNode, combat bool) *colonyAgentNode {
	var agentCountTable [agentKindNum]uint8
	var agentKindThreshold uint8
	searchFlags := searchFighters | searchOnlyAvailable | searchRandomized
	if !combat {
		searchFlags = searchWorkers | searchOnlyAvailable | searchRandomized
		p.colony.agents.Find(searchWorkers, func(a *colonyAgentNode) bool {
			agentCountTable[a.stats.kind]++
			return false
		})
		agentKindThreshold = uint8(gmath.Clamp(p.colony.NumAgents()/5, 5, math.MaxUint8))
	}

	bestScore := 0.0
	var bestCandidate *colonyAgentNode
	p.colony.agents.Find(searchFlags, func(a *colonyAgentNode) bool {
		if a == cloner {
			return false // Self is not a cloning target
		}
		if agentCloningCost(p.colony, cloner, a)*1.5 > p.colony.resources {
			return false // Not enough resources
		}
		if !combat && a.stats.tier > 1 && agentCountTable[a.stats.kind] > agentKindThreshold {
			return false // Don't need more of those
		}
		// Try to use weighted priorities with randomization.
		// Higher tiers are good, but it's also good to clone the units that
		// are not as numerous as some others.
		scoreMultiplier := gmath.ClampMin(1.0/float64(agentCountTable[a.stats.kind]), 0.1)
		score := ((1.25 * float64(a.stats.tier)) * scoreMultiplier) * p.world.rand.FloatRange(0.7, 1.3)
		if score > bestScore {
			bestScore = score
			bestCandidate = a
		}
		return false
	})
	return bestCandidate
}

func (p *colonyActionPlanner) maybeCloneAgent(combatUnit bool) colonyAction {
	// TODO: prefer a green cloner.
	cloner := p.pickCloner()
	if cloner == nil {
		return colonyAction{}
	}
	cloneTarget := p.pickUnitToClone(cloner, combatUnit)
	if cloneTarget == nil {
		return colonyAction{}
	}
	if cloner != nil {
		return colonyAction{
			Kind:     actionCloneAgent,
			Value:    cloneTarget,
			Value2:   cloner,
			TimeCost: 0.8,
		}
	}
	return colonyAction{}
}

func (p *colonyActionPlanner) pickGrowthAction() colonyAction {
	canRepair := p.colony.agents.NumAvailableWorkers() != 0 &&
		p.colony.health < p.colony.maxHealth &&
		p.colony.resources > 30
	if canRepair && p.world.rand.Chance(0.25) {
		return colonyAction{
			Kind:     actionRepairBase,
			TimeCost: 0.4,
		}
	}

	canBuild := p.colony.agents.NumAvailableWorkers() != 0 &&
		len(p.world.coreConstructions) != 0 &&
		p.colony.resources > 30
	if canBuild && p.world.rand.Chance(0.55) {
		var construction *colonyCoreConstructionNode
		closest := 0.0
		for _, c := range p.world.coreConstructions {
			if c.attention > 2 {
				continue
			}
			dist := c.pos.DistanceTo(p.colony.pos)
			if dist > p.colony.realRadius*1.75 {
				continue
			}
			if construction == nil || closest > dist {
				closest = dist
				construction = c
			}
		}
		if construction != nil {
			return colonyAction{
				Kind:     actionBuildBase,
				Value:    construction,
				TimeCost: 0.35,
			}
		}
	}

	combatUnit := p.world.rand.Chance(p.combatUnitProbability())

	if !combatUnit && p.colony.NumAgents() > p.colony.calcUnitLimit() {
		return colonyAction{}
	}

	tryCloning := p.colony.agents.NumAvailableWorkers() >= 2 &&
		p.leadingFactionCombatAgents >= 0.2 &&
		p.leadingFactionAgents >= 0.3
	if tryCloning {
		action := p.maybeCloneAgent(combatUnit)
		if action.Kind != actionNone {
			return action
		}
	}

	stats := workerAgentStats
	if combatUnit {
		stats = militiaAgentStats
	}
	if p.colony.resources >= stats.cost {
		return colonyAction{
			Kind:     actionProduceAgent,
			Value:    stats,
			TimeCost: 0.6,
		}
	} else {
		p.colony.resourceShortage++
	}

	return colonyAction{}
}

func (p *colonyActionPlanner) pickSecurityAction() colonyAction {
	if p.colony.NumAgents() == 0 {
		// Need to call reinforcements.
		for _, c := range p.world.colonies {
			if c.NumAgents() < 10 || c.agents.NumAvailableWorkers() < 6 || c.agents.NumAvailableFighters() < 2 {
				continue
			}
			dist := c.pos.DistanceTo(p.colony.pos)
			if dist > c.realRadius*3 {
				continue
			}
			return colonyAction{
				Kind:     actionGetReinforcements,
				Value:    c,
				TimeCost: 1,
			}
		}
	}

	// Are there any intruders?
	intrusionDist := p.colony.PatrolRadius() * 0.85
	numAttackers := 0
	var closestAttacker *creepNode
	closestAttackerDist := float64(math.MaxFloat64)
	for _, creep := range p.world.creeps {
		dist := creep.pos.DistanceTo(p.colony.pos)
		if dist >= intrusionDist {
			continue
		}
		if dist < closestAttackerDist {
			closestAttackerDist = dist
			closestAttacker = creep
		}
		numAttackers++
		if numAttackers > 5 {
			break
		}
	}
	if numAttackers == 0 {
		numPatrolWanted := int(p.colony.PatrolRadius() / 40)
		if p.numGarrisonAgents != 0 && p.numPatrolAgents < numPatrolWanted {
			return colonyAction{Kind: actionSetPatrol, TimeCost: 0.25}
		}
		return colonyAction{}
	}
	if numAttackers <= 5 {
		if numAttackers*3 < p.numGarrisonAgents {
			return colonyAction{Kind: actionDefenceGarrison, Value: closestAttacker, TimeCost: 0.5}
		}
	}
	return colonyAction{Kind: actionDefencePatrol, Value: closestAttacker, TimeCost: 0.5}
	// return colonyAction{Kind: actionDefencePanic, TimeCost: 0.5}
}

func (p *colonyActionPlanner) tryMergingAction() colonyAction {
	var list []agentMergeRecipe
	if p.colony.evoPoints >= blueEvoThreshold && p.numTier2Agents >= 2 && (p.numTier1Agents < 2 || p.world.rand.Bool()) {
		list = tier3agentMergeRecipeList
	} else {
		list = tier2agentMergeRecipeList
	}

	recipe := gmath.RandElem(p.world.rand, list)
	if recipe.evoCost > p.colony.evoPoints {
		recipe = gmath.RandElem(p.world.rand, tier2agentMergeRecipeList)
	}

	firstAgent := p.colony.agents.Find(searchWorkers|searchFighters|searchOnlyAvailable|searchRandomized, func(a *colonyAgentNode) bool {
		return recipe.Match1(a.AsRecipeSubject())
	})
	if firstAgent == nil {
		return colonyAction{}
	}
	secondAgent := p.colony.agents.Find(searchWorkers|searchFighters|searchOnlyAvailable|searchRandomized, func(a *colonyAgentNode) bool {
		return a != firstAgent && recipe.Match2(a.AsRecipeSubject())
	})
	if secondAgent == nil {
		return colonyAction{}
	}
	return colonyAction{
		Kind:     actionMergeAgents,
		Value:    firstAgent,
		Value2:   secondAgent,
		Value3:   recipe.evoCost,
		TimeCost: 1.2,
	}
}

func (p *colonyActionPlanner) pickEvolutionAction() colonyAction {
	// Maybe find merging candidates.
	if p.world.rand.Chance(0.85) {
		numAttempts := 1
		if p.world.rand.Chance(p.colony.GetEvolutionPriority() * 1.1) {
			numAttempts++
		}
		for i := 0; i < numAttempts; i++ {
			action := p.tryMergingAction()
			if action.Kind != actionNone {
				return action
			}
		}
	}

	if p.colony.agents.TotalNum() >= 15 && p.colony.factionWeights.GetWeight(neutralFactionTag) < 0.6 {
		// Are there any drones to recycle?
		recycleOther := p.world.rand.Chance(0.35)
		toRecycle := p.colony.agents.Find(searchWorkers|searchFighters|searchRandomized, func(a *colonyAgentNode) bool {
			switch a.mode {
			case agentModeStandby, agentModeCharging:
				// OK
			default:
				return false
			}
			switch a.stats.kind {
			case agentWorker, agentMilitia:
				// OK
			default:
				return false
			}
			if a.faction == neutralFactionTag {
				return true
			}
			if recycleOther && a.faction != p.leadingFaction {
				if a.stats.canPatrol {
					return p.leadingFactionCombatAgents < 0.35
				}
				return p.leadingFactionAgents < 0.25
			}
			return false
		})
		if toRecycle != nil {
			// Try not to recyle units that may be needed later for merging.
			if toRecycle.faction != neutralFactionTag && p.canUseInRecipe(toRecycle) {
				return colonyAction{}
			}
			return colonyAction{
				Kind:     actionRecycleAgent,
				Value:    toRecycle,
				TimeCost: 0.8,
			}
		}
	}

	if p.numTier2Agents > 0 {
		evoPointsChance := gmath.Clamp(p.colony.GetEvolutionPriority()-0.1, 0, 0.6)
		if p.world.rand.Chance(evoPointsChance) {
			return colonyAction{
				Kind:     actionGenerateEvo,
				TimeCost: 0.3,
			}
		}
	}

	return colonyAction{}
}

func (p *colonyActionPlanner) canUseInRecipe(x *colonyAgentNode) bool {
	recipes := recipesIndex[x.AsRecipeSubject()]
	mergeCandidate := p.colony.agents.Find(searchWorkers|searchFighters|searchRandomized, func(y *colonyAgentNode) bool {
		if x == y {
			return false
		}
		for _, recipe := range recipes {
			if recipe.Match(x, y) {
				return true
			}
		}
		return false
	})
	return mergeCandidate != nil
}
