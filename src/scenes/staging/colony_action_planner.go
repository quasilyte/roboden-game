package staging

import (
	"math"

	"github.com/quasilyte/gmath"
)

type colonyActionPlanner struct {
	colony *colonyCoreNode

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
	p.colony.availableAgents = p.colony.availableAgents[:0]
	p.colony.availableCombatAgents = p.colony.availableCombatAgents[:0]
	p.colony.availableUniversalAgents = p.colony.availableUniversalAgents[:0]
	leadingFactionAgents := 0
	leadingFactionCombatAgents := 0
	for _, a := range p.colony.agents {
		if a.mode == agentModeStandby {
			p.colony.availableAgents = append(p.colony.availableAgents, a)
		}
		if a.faction == p.leadingFaction {
			leadingFactionAgents++
		}
	}
	for _, a := range p.colony.combatAgents {
		if a.mode == agentModePatrol {
			p.numPatrolAgents++
		}
		if a.mode == agentModeStandby {
			p.numGarrisonAgents++
		}
		if a.mode == agentModePatrol || a.mode == agentModeStandby {
			if a.stats.canGather {
				p.colony.availableUniversalAgents = append(p.colony.availableUniversalAgents, a)
			} else {
				p.colony.availableCombatAgents = append(p.colony.availableCombatAgents, a)
			}
		}
		if a.faction == p.leadingFaction {
			leadingFactionCombatAgents++
		}
	}
	p.leadingFactionAgents = float64(leadingFactionAgents) / float64(len(p.colony.agents))
	p.leadingFactionCombatAgents = float64(leadingFactionCombatAgents) / float64(len(p.colony.combatAgents))

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
	if p.colony.GetSecurityPriority() > 0.1 && len(p.colony.combatAgents) < minCombatUnits {
		return 0.9
	}
	wantedCombatAgentRatio := p.colony.GetSecurityPriority() * 0.8
	currentCombatAgentRatio := float64(len(p.colony.combatAgents)) / float64(len(p.colony.agents))
	if currentCombatAgentRatio < wantedCombatAgentRatio {
		return 0.75
	}
	return 0
}

func (p *colonyActionPlanner) pickCloner() *colonyAgentNode {
	return randFind(p.world.rand, p.colony.availableAgents, func(a *colonyAgentNode) bool {
		return a.energy > agentCloningEnergyCost() && a.energyBill == 0
	})
}

func (p *colonyActionPlanner) pickUnitToClone(cloner *colonyAgentNode, combat bool) *colonyAgentNode {
	var agents = p.colony.availableAgents
	if combat {
		agents = p.colony.availableCombatAgents
	}
	return randFind(p.world.rand, agents, func(a *colonyAgentNode) bool {
		return a != cloner && agentCloningCost(p.colony, cloner, a)*1.5 < p.colony.resources.Essence
	})
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
			TimeCost: 0.8,
		}
	}
	return colonyAction{}
}

func (p *colonyActionPlanner) pickGrowthAction() colonyAction {
	canBuild := len(p.colony.availableAgents) != 0 &&
		len(p.world.coreConstructions) != 0 &&
		p.colony.resources.Essence > 15
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

	tryCloning := len(p.colony.availableAgents) >= 2 &&
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
	if p.colony.resources.Essence >= stats.cost {
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
			if c.NumAgents() < 10 || len(c.availableAgents) < 6 || len(c.availableCombatAgents) < 3 {
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
	recipe := gmath.RandElem(p.world.rand, agentMergeRecipeList)
	firstAgent := p.colony.FindAgent(func(a *colonyAgentNode) bool {
		return a.mode == agentModeStandby && recipe.Match1(a)
	})
	if firstAgent == nil {
		return colonyAction{}
	}
	secondAgent := p.colony.FindAgent(func(a *colonyAgentNode) bool {
		return a.mode == agentModeStandby && a != firstAgent && recipe.Match2(a)
	})
	if secondAgent == nil {
		return colonyAction{}
	}
	return colonyAction{
		Kind:     actionMergeAgents,
		Value:    firstAgent,
		Value2:   secondAgent,
		TimeCost: 1.2,
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

	if len(p.colony.agents) > 5 && p.colony.factionWeights.GetWeight(neutralFactionTag) < 0.6 {
		// Are there any drones to recycle?
		recycleOther := p.world.rand.Chance(0.35)
		toRecycle := p.colony.FindAgent(func(a *colonyAgentNode) bool {
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
			return colonyAction{
				Kind:     actionRecycleAgent,
				Value:    toRecycle,
				TimeCost: 0.8,
			}
		}
	}

	return colonyAction{}
}
