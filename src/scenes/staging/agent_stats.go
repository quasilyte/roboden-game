package staging

import (
	"github.com/quasilyte/colony-game/assets"
	resource "github.com/quasilyte/ebitengine-resource"
)

// tier 1
// * worker => basic unit, can be produced
// * militia => basic combat unit, can be produced
//
// tier 2
// * +++ fighter => advanced combat unit; ?
// * + repeller => a worker that can repell creeps; ?
// * ++ repairbot => a worker that can repair units; ?
// * + rechargebot => a worker that can recharge units; ?
// * + freighter => a worker with much higher capacity; ?
// * tetherbot => a worker that connects to another bot and improves it; ?
// * sciencebot => a worker that generates x2 evolution points; ?
// * generatorbot => a worker that decreases the effective upkeep by 15 (~3 militia equiv)
//
// tier 3
// * destroyer => powerful combat unit; fighter + fighter
// * artillery => long-range combat unit; fighter + freighter
// * engineer => a worker that can repair buildings (including the base); repairbot + repairbot
// * leech => a worker that drains energy from enemies and transfers it to allies; repeller + rechargebot
// * essencebot => a worker that generates resources passively; ?

type agentMergeRecipe struct {
	agent1kind    colonyAgentKind
	agent1faction factionTag
	agent2kind    colonyAgentKind
	agent2faction factionTag
	result        *agentStats
}

func (r *agentMergeRecipe) Match1(a *colonyAgentNode) bool {
	return r.match(r.agent1kind, r.agent1faction, a)
}

func (r *agentMergeRecipe) Match2(a *colonyAgentNode) bool {
	return r.match(r.agent2kind, r.agent2faction, a)
}

func (r *agentMergeRecipe) match(kind colonyAgentKind, faction factionTag, a *colonyAgentNode) bool {
	if a.stats.kind != kind {
		return false
	}
	if faction == neutralFactionTag {
		return true
	}
	return a.faction == faction
}

// Merge usage:
//
// * militia
//   * red ++
//   * blue +
//   * yellow +
// * worker
//   * blue ++++
//   * yellow ++
//   * green ++
//
// * repeller +
// * freighter +
var agentMergeRecipeList = []agentMergeRecipe{
	{
		agent1kind:    agentMilitia,
		agent1faction: redFactionTag,
		agent2kind:    agentMilitia,
		agent2faction: redFactionTag,
		result:        fighterAgentStats,
	},
	{
		agent1kind:    agentWorker,
		agent1faction: yellowFactionTag,
		agent2kind:    agentWorker,
		agent2faction: yellowFactionTag,
		result:        freighterAgentStats,
	},
	{
		agent1kind:    agentWorker,
		agent1faction: blueFactionTag,
		agent2kind:    agentMilitia,
		agent2faction: blueFactionTag,
		result:        repellerAgentStats,
	},
	{
		agent1kind:    agentWorker,
		agent1faction: blueFactionTag,
		agent2kind:    agentMilitia,
		agent2faction: yellowFactionTag,
		result:        generatorAgentStats,
	},
	{
		agent1kind:    agentWorker,
		agent1faction: blueFactionTag,
		agent2kind:    agentWorker,
		agent2faction: blueFactionTag,
		result:        rechargeAgentStats,
	},
	{
		agent1kind:    agentRepeller,
		agent1faction: neutralFactionTag,
		agent2kind:    agentFreighter,
		agent2faction: neutralFactionTag,
		result:        flamerAgentStats,
	},

	// Tier 3.
	{
		agent1kind:    agentWorker,
		agent1faction: redFactionTag,
		agent2kind:    agentMilitia,
		agent2faction: greenFactionTag,
		result:        repairAgentStats,
	},
}

type unitSize int

const (
	sizeSmall unitSize = iota
	sizeMedium
	sizeLarge
)

type agentStats struct {
	kind   colonyAgentKind
	image  resource.ImageID
	tier   int
	cost   float64
	upkeep int

	size unitSize

	speed float64

	maxHealth float64

	canGather  bool
	canPatrol  bool
	maxPayload int

	diodeOffset float64

	supportReload float64
	supportRange  float64

	attackRange           float64
	attackDelay           float64
	projectileArea        float64
	projectileSpeed       float64
	projectileRotateSpeed float64
	projectileDamage      damageValue
	projectileImage       resource.ImageID
	attackSound           resource.AudioID
}

var workerAgentStats = &agentStats{
	kind:        agentWorker,
	image:       assets.ImageWorkerAgent,
	size:        sizeSmall,
	diodeOffset: 5,
	tier:        1,
	cost:        8,
	upkeep:      2,
	canGather:   true,
	maxPayload:  1,
	speed:       80,
	maxHealth:   12,
}

var generatorAgentStats = &agentStats{
	kind:        agentGenerator,
	image:       assets.ImageGeneratorAgent,
	size:        sizeMedium,
	diodeOffset: 10,
	tier:        2,
	cost:        10,
	upkeep:      2,
	canGather:   true,
	maxPayload:  1,
	speed:       90,
	maxHealth:   20,
}

var repairAgentStats = &agentStats{
	kind:          agentRepair,
	image:         assets.ImageRepairAgent,
	size:          sizeMedium,
	diodeOffset:   5,
	tier:          2,
	cost:          12,
	upkeep:        5,
	canGather:     true,
	maxPayload:    1,
	speed:         110,
	maxHealth:     18,
	supportReload: 7.5,
	supportRange:  300,
}

var rechargeAgentStats = &agentStats{
	kind:          agentRecharger,
	image:         assets.ImageRechargerAgent,
	size:          sizeMedium,
	diodeOffset:   5,
	tier:          2,
	cost:          12,
	upkeep:        4,
	canGather:     true,
	maxPayload:    1,
	speed:         90,
	maxHealth:     16,
	supportReload: 7,
	supportRange:  340,
}

var freighterAgentStats = &agentStats{
	kind:        agentFreighter,
	image:       assets.ImageFreighterAgent,
	size:        sizeMedium,
	diodeOffset: 1,
	tier:        2,
	cost:        10,
	upkeep:      3,
	canGather:   true,
	maxPayload:  3,
	speed:       70,
	maxHealth:   25,
}

var militiaAgentStats = &agentStats{
	kind:             agentMilitia,
	image:            assets.ImageMilitiaAgent,
	size:             sizeSmall,
	diodeOffset:      5,
	tier:             1,
	cost:             10,
	upkeep:           4,
	canPatrol:        true,
	speed:            75,
	maxHealth:        12,
	attackRange:      130,
	attackDelay:      2.5,
	attackSound:      assets.AudioMilitiaShot,
	projectileImage:  assets.ImageMilitiaProjectile,
	projectileArea:   10,
	projectileSpeed:  180,
	projectileDamage: damageValue{health: 2, morale: 2},
}

var flamerAgentStats = &agentStats{
	kind:             agentFlamer,
	image:            assets.ImageFlamerAgent,
	size:             sizeLarge,
	diodeOffset:      7,
	tier:             3,
	cost:             22,
	upkeep:           8,
	canPatrol:        true,
	speed:            130,
	maxHealth:        35,
	attackRange:      100,
	attackDelay:      1.2,
	attackSound:      assets.AudioFlamerShot,
	projectileImage:  assets.ImageFlamerProjectile,
	projectileArea:   18,
	projectileSpeed:  160,
	projectileDamage: damageValue{health: 5},
}

var fighterAgentStats = &agentStats{
	kind:             agentFighter,
	image:            assets.ImageFighterAgent,
	size:             sizeMedium,
	diodeOffset:      8,
	tier:             2,
	cost:             15,
	upkeep:           7,
	canPatrol:        true,
	speed:            90,
	maxHealth:        21,
	attackRange:      180,
	attackDelay:      2,
	attackSound:      assets.AudioFighterBeam,
	projectileImage:  assets.ImageFighterProjectile,
	projectileArea:   8,
	projectileSpeed:  220,
	projectileDamage: damageValue{health: 4},
}

var repellerAgentStats = &agentStats{
	kind:             agentRepeller,
	image:            assets.ImageRepellerAgent,
	size:             sizeMedium,
	diodeOffset:      8,
	tier:             2,
	cost:             14,
	upkeep:           4,
	canGather:        true,
	maxPayload:       1,
	canPatrol:        true,
	speed:            115,
	maxHealth:        22,
	attackRange:      160,
	attackDelay:      2.4,
	attackSound:      assets.AudioRepellerBeam,
	projectileImage:  assets.ImageRepellerProjectile,
	projectileArea:   10,
	projectileSpeed:  200,
	projectileDamage: damageValue{health: 1, morale: 4},
}
