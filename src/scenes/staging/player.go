package staging

import (
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/serverapi"
)

type player interface {
	Init()
	Update(computedDelta, delta float64)
	HandleInput()
	GetState() *playerState
}

func isHumanPlayer(p player) bool {
	_, ok := p.(*humanPlayer)
	return ok
}

type playerState struct {
	id int

	colonies []*colonyCoreNode

	selectedColony *colonyCoreNode

	replay []serverapi.PlayerAction

	camera *cameraManager

	hasRoombas bool
}

func newPlayerState() *playerState {
	pstate := &playerState{
		colonies: make([]*colonyCoreNode, 0, 1),
	}

	return pstate
}

func (pstate *playerState) Init(world *worldState) {
	pstate.hasRoombas = xslices.Contains(world.tier2recipes, gamedata.FindRecipe(gamedata.RoombaAgentStats))
}

func getUnitColony(u any) *colonyCoreNode {
	switch u := u.(type) {
	case *colonyAgentNode:
		return u.colonyCore
	case *colonyCoreNode:
		return u
	default:
		return nil
	}
}

const maxCreepGroupsPerSide = 3

type creepsPlayerState struct {
	techLevel float64

	maxSideCost int
	attackSides [4]*creepsCombinedGroup
}

type creepsCombinedGroup struct {
	groups    [maxCreepGroupsPerSide]arenaWaveGroup
	totalCost int
}

func newCreepsPlayerState() *creepsPlayerState {
	state := &creepsPlayerState{}
	for i := range state.attackSides {
		cg := &creepsCombinedGroup{}
		state.attackSides[i] = cg
		for j := range cg.groups {
			cg.groups[j].side = i
		}
	}
	state.RecalcMaxCost()
	return state
}

func (state *creepsPlayerState) RecalcMaxCost() {
	const maxCost = maxArenaGroupBudget * maxCreepGroupsPerSide
	const maxCostTechRequired = 2.0
	const multiplier = 1.0 / maxCostTechRequired
	cost := ((state.techLevel * multiplier) * maxCost) + 10
	state.maxSideCost = int(gmath.ClampMax(cost, maxCost))
}

func (state *creepsPlayerState) ResetGroups() {
	for _, cg := range state.attackSides {
		for j := range cg.groups {
			g := &cg.groups[j]
			g.units = g.units[:0]
			g.totalCost = 0
		}
		cg.totalCost = 0
	}
}

func (state *creepsPlayerState) AddUnits(world *worldState, side int, info creepOptionInfo) bool {
	unitsAdded := 0
	numUnits := numCreepsPerCard(state, info)
	for i := 0; i < numUnits; i++ {
		super := false
		switch {
		case state.techLevel >= 2:
			super = true
		case state.techLevel > 1:
			super = world.rand.Chance(1.0 - state.techLevel)
		}
		if !state.addUnit(side, info.stats, super) {
			break
		}
		unitsAdded++
	}

	return unitsAdded > 0
}

func (state *creepsPlayerState) addUnit(side int, stats *creepStats, super bool) bool {
	var dst *arenaWaveGroup
	cg := state.attackSides[side]
	if cg.totalCost >= state.maxSideCost {
		return false
	}
	for i := range cg.groups {
		g := &cg.groups[i]
		if g.totalCost < maxArenaGroupBudget {
			dst = g
			break
		}
	}
	if dst == nil {
		return false
	}

	cost := creepCost(stats, false)
	dst.totalCost += cost
	cg.totalCost += cost
	dst.units = append(dst.units, arenaWaveUnit{
		stats: stats,
		super: super,
	})

	return true
}
