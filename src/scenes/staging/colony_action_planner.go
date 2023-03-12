package staging

import (
	"math"

	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/gamedata"
)

type colonyActionPlanner struct {
	colony *colonyCoreNode

	numTier1Agents    int
	numTier2Agents    int
	numPatrolAgents   int
	numGarrisonAgents int

	agentCountTable [gamedata.AgentKindNum]uint8
	mergetab        mergeTable

	leadingFaction             gamedata.FactionTag
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
	p.agentCountTable = [gamedata.AgentKindNum]uint8{}
	leadingFactionAgents := 0
	leadingFactionCombatAgents := 0

	p.colony.agents.Update()

	p.colony.agents.Each(func(a *colonyAgentNode) {
		p.agentCountTable[a.stats.Kind]++
		switch a.stats.Tier {
		case 1:
			p.numTier1Agents++
		case 2:
			p.numTier2Agents++
		}
		if !a.stats.CanPatrol {
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
		return p.pickResourcesAction()
	case priorityGrowth:
		return p.pickGrowthAction()
	case prioritySecurity:
		return p.pickSecurityAction()
	case priorityEvolution:
		return p.pickEvolutionAction()
	}

	panic("unreachable")
}

func (p *colonyActionPlanner) trySendingCourier() colonyAction {
	// Try to find a colony for a trading route.
	maxTradingDist := (p.colony.PatrolRadius() * 1.75) + 200
	potentialTargets := &p.world.tmpColonySlice
	(*potentialTargets) = (*potentialTargets)[:0]
	for _, colony := range p.world.colonies {
		if colony == p.colony {
			continue
		}
		if colony.mode != colonyModeNormal {
			continue
		}
		if colony.pos.DistanceTo(p.colony.pos) > maxTradingDist {
			continue
		}
		*potentialTargets = append(*potentialTargets, colony)
	}
	if len(*potentialTargets) == 0 {
		return colonyAction{}
	}

	selectedColony := gmath.RandElem(p.world.rand, *potentialTargets)
	if p.colony.resources >= maxVisualResources && selectedColony.resources >= maxVisualResources {
		return colonyAction{}
	}
	courier := p.colony.agents.Find(searchWorkers|searchOnlyAvailable|searchRandomized, func(a *colonyAgentNode) bool {
		switch a.stats.Kind {
		case gamedata.AgentCourier, gamedata.AgentTrucker:
			return a.energy >= 50 && a.energyBill < 20
		default:
			return false
		}
	})
	if courier == nil {
		return colonyAction{}
	}
	return colonyAction{
		Kind:     actionSendCourier,
		Value:    selectedColony,
		Value2:   courier,
		TimeCost: 0.2,
	}
}

func (p *colonyActionPlanner) pickResourcesAction() colonyAction {
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

	if len(p.world.colonies) >= 2 && p.colony.agents.hasCourier {
		a := p.trySendingCourier()
		if a.Kind != actionNone {
			return a
		}
	}

	var bestSource *essenceSourceNode
	var bestRedOilSource *essenceSourceNode
	bestScore := 0.0
	bestRedOilScore := 0.0
	for _, source := range p.world.essenceSources {
		score := resourceScore(p.colony, source) * p.world.rand.FloatRange(0.8, 1.4)
		if source.stats == redOilSource {
			if !p.colony.agents.hasRedMiner {
				continue
			}
			if score != 0 && score > bestRedOilScore {
				bestRedOilScore = score
				bestRedOilSource = source
			}
		} else {
			if score != 0 && score > bestScore {
				bestScore = score
				bestSource = source
			}
		}
	}

	if bestRedOilSource != nil && p.world.rand.Chance(0.25) {
		return colonyAction{
			Kind:     actionMineEssence,
			Value:    bestRedOilSource,
			TimeCost: 0.35,
		}
	}

	if bestSource != nil {
		return colonyAction{
			Kind:     actionMineEssence,
			Value:    bestSource,
			TimeCost: 0.7,
		}
	}

	return colonyAction{}
}

func (p *colonyActionPlanner) combatUnitProbability() float64 {
	if len(p.colony.agents.workers) < 2 {
		return 0
	}
	minCombatUnits := int(p.colony.GetSecurityPriority() * 20)
	if p.colony.GetSecurityPriority() > 0.1 && len(p.colony.agents.fighters) < minCombatUnits {
		return 0.9
	}
	wantedCombatAgentRatio := p.colony.GetSecurityPriority() * 0.8
	currentCombatAgentRatio := float64(len(p.colony.agents.fighters)) / float64(len(p.colony.agents.workers))
	if currentCombatAgentRatio < wantedCombatAgentRatio {
		return 0.75
	}
	return 0.2
}

func (p *colonyActionPlanner) pickCloner() *colonyAgentNode {
	cloner := p.colony.agents.Find(searchWorkers|searchOnlyAvailable, func(a *colonyAgentNode) bool {
		if a.stats.Kind != gamedata.AgentCloner {
			return false
		}
		return a.energy >= agentCloningEnergyCost() && a.energyBill == 0
	})
	return cloner
}

func (p *colonyActionPlanner) pickUnitToClone(cloner *colonyAgentNode, combat bool) *colonyAgentNode {
	agentKindThreshold := uint8(gmath.Clamp(p.colony.NumAgents()/5, 5, math.MaxUint8))
	searchFlags := searchFighters | searchOnlyAvailable | searchRandomized
	if !combat {
		searchFlags = searchWorkers | searchOnlyAvailable | searchRandomized
	}

	bestScore := 0.0
	var bestCandidate *colonyAgentNode
	p.colony.agents.Find(searchFlags, func(a *colonyAgentNode) bool {
		if a == cloner {
			return false // Self is not a cloning target
		}
		if a.rank > 0 {
			return false // Can't clone elite units
		}
		if agentCloningCost(p.colony, cloner, a)*1.5 > p.colony.resources {
			return false // Not enough resources
		}
		if a.stats.Tier > 1 && p.agentCountTable[a.stats.Kind] > agentKindThreshold {
			return false // Don't need more of those
		}
		// Try to use weighted priorities with randomization.
		// Higher tiers are good, but it's also good to clone the units that
		// are not as numerous as some others.
		scoreMultiplier := gmath.ClampMin(1.0/float64(p.agentCountTable[a.stats.Kind]), 0.1)
		score := ((1.25 * float64(a.stats.Tier)) * scoreMultiplier) * p.world.rand.FloatRange(0.7, 1.3)
		if a.faction != gamedata.NeutralFactionTag {
			score *= 1.1
		}
		if score > bestScore {
			bestScore = score
			bestCandidate = a
		}
		return false
	})
	return bestCandidate
}

func (p *colonyActionPlanner) maybeCloneAgent(combatUnit bool) colonyAction {
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
			TimeCost: 0.6,
		}
	}
	return colonyAction{}
}

func (p *colonyActionPlanner) pickGrowthAction() colonyAction {
	canRepairColony := p.colony.agents.NumAvailableWorkers() != 0 &&
		p.colony.health < p.colony.maxHealth &&
		p.colony.resources > 30
	if canRepairColony && p.world.rand.Chance(0.25) {
		return colonyAction{
			Kind:     actionRepairBase,
			TimeCost: 0.4,
		}
	}
	canRepairTurret := p.colony.agents.NumAvailableWorkers() != 0 &&
		p.colony.resources > 40 &&
		len(p.colony.turrets) > 0
	if canRepairTurret && p.world.rand.Chance(0.3) {
		for _, turret := range p.colony.turrets {
			if turret.health >= turret.maxHealth*0.9 {
				continue
			}
			distSqr := turret.pos.DistanceSquaredTo(p.colony.pos)
			if distSqr > p.colony.realRadiusSqr*1.8 {
				continue
			}
			return colonyAction{
				Kind:     actionRepairTurret,
				Value:    turret,
				TimeCost: 0.3,
			}
		}
	}

	canBuild := p.colony.agents.NumAvailableWorkers() != 0 &&
		len(p.world.constructions) != 0 &&
		p.colony.resources > 30
	if canBuild && p.world.rand.Chance(0.55) {
		var construction *constructionNode
		closest := 0.0
		for _, c := range p.world.constructions {
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
				Kind:     actionBuildBuilding,
				Value:    construction,
				TimeCost: 0.35,
			}
		}
	}

	combatUnitChance := p.combatUnitProbability()
	combatUnit := false
	if combatUnitChance != 0 {
		combatUnit = p.world.rand.Chance(combatUnitChance)
	}

	softUnitLimit := p.colony.calcUnitLimit()
	if combatUnit {
		softUnitLimit = int(float64(softUnitLimit) * 1.3)
	}
	if p.colony.NumAgents() > softUnitLimit {
		return colonyAction{}
	}

	tryCloning := p.colony.cloningDelay == 0 &&
		p.colony.agents.hasCloner &&
		p.colony.agents.NumAvailableWorkers() >= 2 &&
		p.leadingFactionCombatAgents >= 0.2 &&
		p.leadingFactionAgents >= 0.3
	if tryCloning {
		action := p.maybeCloneAgent(combatUnit)
		if action.Kind != actionNone {
			return action
		}
	}

	stats := gamedata.WorkerAgentStats
	if combatUnit {
		stats = gamedata.MilitiaAgentStats
	}
	if p.colony.resources >= stats.Cost {
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
}

func (p *colonyActionPlanner) pickMergeRecipe(list []gamedata.AgentMergeRecipe, tier int) gamedata.AgentMergeRecipe {
	if tier == 3 {
		return randIterate(p.world.rand, list, func(recipe gamedata.AgentMergeRecipe) bool {
			return p.agentCountTable[recipe.Result.Kind] < 10
		})
	}

	bestScore := 0.0
	var result gamedata.AgentMergeRecipe
	preferCombatUnits := p.colony.GetResourcePriority() < p.colony.GetSecurityPriority()
	for _, recipe := range list {
		if p.agentCountTable[recipe.Result.Kind] >= 10 {
			continue
		}
		if !p.mergetab.CanProduce(recipe) {
			continue
		}
		count := p.agentCountTable[recipe.Result.Kind]
		score := 12.0 - float64(count)
		score *= p.world.rand.FloatRange(0.7, 1.4)
		if count == 0 {
			score *= 1.2
		}
		switch {
		case recipe.Result.CanPatrol && preferCombatUnits:
			score *= 1.15
		case recipe.Result.CanGather && !preferCombatUnits:
			score *= 1.1
		}
		if recipe.Result.Tier == 2 && recipe.Drone1.Faction == recipe.Drone2.Faction {
			score *= 0.9
		}
		if score > bestScore {
			bestScore = score
			result = recipe
		}
	}
	return result
}

func (p *colonyActionPlanner) tryMergingAction() colonyAction {
	var list []gamedata.AgentMergeRecipe
	tier := 0
	if p.colony.evoPoints >= blueEvoThreshold && p.numTier2Agents >= 2 && (p.numTier1Agents < 2 || p.world.rand.Bool()) {
		list = gamedata.Tier3agentMergeRecipes
		tier = 3
	} else {
		list = p.world.tier2recipes
		tier = 2
		p.mergetab.Update(p.colony.agents)
	}

	recipe := p.pickMergeRecipe(list, tier)
	if recipe.Result == nil {
		return colonyAction{}
	}
	if recipe.EvoCost > p.colony.evoPoints {
		// This happens only when tier3 can't be produced due to the evo points shortage.
		list = p.world.tier2recipes
		tier = 2
		p.mergetab.Update(p.colony.agents)
		recipe = p.pickMergeRecipe(list, tier)
		if recipe.Result == nil {
			return colonyAction{}
		}
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
		Value3:   recipe.EvoCost,
		TimeCost: 0.9,
	}
}

func (p *colonyActionPlanner) pickEvolutionAction() colonyAction {
	// Maybe find merging candidates.
	if p.world.rand.Chance(0.85) {
		action := p.tryMergingAction()
		if action.Kind != actionNone {
			return action
		}
	}

	canRecycle := p.colony.agents.TotalNum() > 15 &&
		len(p.colony.agents.workers) > 5 &&
		p.colony.GetEvolutionPriority() >= 0.1 &&
		p.colony.factionWeights.GetWeight(gamedata.NeutralFactionTag) < 0.5 &&
		p.colony.resources > 20
	if canRecycle {
		// Are there any drones to recycle?
		recycleOther := p.world.rand.Chance(0.35)
		toRecycle := p.colony.agents.Find(searchWorkers|searchFighters|searchRandomized, func(a *colonyAgentNode) bool {
			switch a.mode {
			case agentModeStandby, agentModeCharging:
				// OK
			default:
				return false
			}
			switch a.stats.Kind {
			case gamedata.AgentWorker, gamedata.AgentMilitia:
				// OK
			default:
				return false
			}
			if a.faction == gamedata.NeutralFactionTag {
				return true
			}
			if recycleOther && a.faction != p.leadingFaction {
				if a.stats.CanPatrol {
					return p.leadingFactionCombatAgents < 0.35
				}
				return p.leadingFactionAgents < 0.25
			}
			return false
		})
		if toRecycle != nil {
			// Try not to recyle units that may be needed later for merging.
			if toRecycle.faction != gamedata.NeutralFactionTag && p.canUseInRecipe(toRecycle) {
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
	recipes := p.world.tier2recipeIndex[x.AsRecipeSubject()]
	mergeCandidate := p.colony.agents.Find(searchWorkers|searchFighters|searchRandomized, func(y *colonyAgentNode) bool {
		if x == y {
			return false
		}
		for _, recipe := range recipes {
			if recipe.Match(x.AsRecipeSubject(), y.AsRecipeSubject()) {
				return true
			}
		}
		return false
	})
	return mergeCandidate != nil
}
