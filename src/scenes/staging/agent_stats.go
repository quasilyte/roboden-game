package staging

import (
	"math"

	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/roboden-game/assets"
)

var minEvoCost float64 = 0.0

func init() {
	minCost := math.MaxFloat64
	for _, recipe := range tier3agentMergeRecipeList {
		if recipe.evoCost < minCost {
			minCost = recipe.evoCost
		}
	}
	minEvoCost = minCost
}

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
	evoCost       float64
	result        *agentStats
}

func (r *agentMergeRecipe) Match(x, y *colonyAgentNode) bool {
	if r.Match1(x.AsRecipeSubject()) && r.Match2(y.AsRecipeSubject()) {
		return true
	}
	if r.Match1(y.AsRecipeSubject()) && r.Match2(x.AsRecipeSubject()) {
		return true
	}
	return false
}

func (r *agentMergeRecipe) Match1(s recipeSubject) bool {
	return r.match(r.agent1kind, r.agent1faction, s)
}

func (r *agentMergeRecipe) Match2(s recipeSubject) bool {
	return r.match(r.agent2kind, r.agent2faction, s)
}

func (r *agentMergeRecipe) match(kind colonyAgentKind, faction factionTag, s recipeSubject) bool {
	if s.kind != kind {
		return false
	}
	if faction == neutralFactionTag {
		return true
	}
	return s.faction == faction
}

type recipeSubject struct {
	kind    colonyAgentKind
	faction factionTag
}

var recipesIndex = map[recipeSubject][]agentMergeRecipe{}

func init() {
	factions := []factionTag{
		yellowFactionTag,
		redFactionTag,
		blueFactionTag,
		greenFactionTag,
	}
	kinds := []colonyAgentKind{
		agentWorker,
		agentMilitia,
	}
	for _, f := range factions {
		for _, k := range kinds {
			subject := recipeSubject{kind: k, faction: f}
			for _, recipe := range tier2agentMergeRecipeList {
				if !recipe.Match1(subject) && !recipe.Match2(subject) {
					continue
				}
				recipesIndex[subject] = append(recipesIndex[subject], recipe)
			}
		}
	}
}

// Merge usage:
//
// * militia
//   * red ++
//   * blue +
//   * yellow +
//   * green +++
// * worker
//   * red ++
//   * blue ++++
//   * yellow +++
//   * green +++
var tier2agentMergeRecipeList = []agentMergeRecipe{
	{
		agent1kind:    agentWorker,
		agent1faction: blueFactionTag,
		agent2kind:    agentWorker,
		agent2faction: blueFactionTag,
		result:        rechargeAgentStats,
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
		agent1faction: redFactionTag,
		agent2kind:    agentWorker,
		agent2faction: redFactionTag,
		result:        redminerAgentStats,
	},

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
		agent2faction: greenFactionTag,
		result:        servoAgentStats,
	},
	{
		agent1kind:    agentMilitia,
		agent1faction: greenFactionTag,
		agent2kind:    agentMilitia,
		agent2faction: greenFactionTag,
		result:        cripplerAgentStats,
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
		agent1faction: redFactionTag,
		agent2kind:    agentMilitia,
		agent2faction: greenFactionTag,
		result:        repairAgentStats,
	},
}

var tier3agentMergeRecipeList = []agentMergeRecipe{
	{
		agent1kind: agentRepeller,
		agent2kind: agentFreighter,
		evoCost:    5,
		result:     flamerAgentStats,
	},
	{
		agent1kind: agentFighter,
		agent2kind: agentFighter,
		evoCost:    11,
		result:     destroyerAgentStats,
	},

	{
		agent1kind: agentRecharger,
		agent2kind: agentRepair,
		evoCost:    7,
		result:     refresherAgentStats,
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

	attackTargets         int
	attackRange           float64
	attackDelay           float64
	projectileArea        float64
	projectileSpeed       float64
	projectileRotateSpeed float64
	projectileDamage      damageValue
	projectileImage       resource.ImageID
	projectileImpact      projectileExplosionKind
	attackSound           resource.AudioID
}

var workerAgentStats = &agentStats{
	kind:        agentWorker,
	image:       assets.ImageWorkerAgent,
	size:        sizeSmall,
	diodeOffset: 5,
	tier:        1,
	cost:        9,
	upkeep:      2,
	canGather:   true,
	maxPayload:  1,
	speed:       80,
	maxHealth:   12,
}

var redminerAgentStats = &agentStats{
	kind:        agentRedminer,
	image:       assets.ImageRedminerAgent,
	size:        sizeMedium,
	diodeOffset: 6,
	tier:        2,
	cost:        15,
	upkeep:      3,
	canGather:   true,
	maxPayload:  1,
	speed:       75,
	maxHealth:   18,
}

var generatorAgentStats = &agentStats{
	kind:        agentGenerator,
	image:       assets.ImageGeneratorAgent,
	size:        sizeMedium,
	diodeOffset: 10,
	tier:        2,
	cost:        15,
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
	cost:          20,
	upkeep:        5,
	canGather:     true,
	maxPayload:    1,
	speed:         110,
	maxHealth:     18,
	supportReload: 8.0,
	supportRange:  450,
}

var rechargeAgentStats = &agentStats{
	kind:          agentRecharger,
	image:         assets.ImageRechargerAgent,
	size:          sizeMedium,
	diodeOffset:   5,
	tier:          2,
	cost:          15,
	upkeep:        4,
	canGather:     true,
	maxPayload:    1,
	speed:         90,
	maxHealth:     16,
	supportReload: 7,
	supportRange:  400,
}

var refresherAgentStats = &agentStats{
	kind:          agentRefresher,
	image:         assets.ImageRefresherAgent,
	size:          sizeLarge,
	diodeOffset:   5,
	tier:          3,
	cost:          40,
	upkeep:        10,
	canGather:     true,
	maxPayload:    1,
	speed:         110,
	maxHealth:     24,
	supportReload: rechargeAgentStats.supportReload,
	supportRange:  rechargeAgentStats.supportRange,
}

var servoAgentStats = &agentStats{
	kind:          agentServo,
	image:         assets.ImageServoAgent,
	size:          sizeMedium,
	diodeOffset:   -4,
	tier:          2,
	cost:          30,
	upkeep:        7,
	canGather:     true,
	maxPayload:    1,
	speed:         165,
	maxHealth:     18,
	supportReload: 8,
	supportRange:  310,
}

var freighterAgentStats = &agentStats{
	kind:        agentFreighter,
	image:       assets.ImageFreighterAgent,
	size:        sizeMedium,
	diodeOffset: 1,
	tier:        2,
	cost:        15,
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
	cost:             12,
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
	attackTargets:    1,
}

var cripplerAgentStats = &agentStats{
	kind:             agentCrippler,
	image:            assets.ImageCripplerAgent,
	size:             sizeMedium,
	diodeOffset:      5,
	tier:             1,
	cost:             15,
	upkeep:           4,
	canPatrol:        true,
	speed:            55,
	maxHealth:        15,
	attackRange:      240,
	attackDelay:      3.2,
	attackSound:      assets.AudioCripplerShot,
	projectileImage:  assets.ImageCripplerProjectile,
	projectileArea:   10,
	projectileSpeed:  250,
	projectileDamage: damageValue{health: 1, slow: 2},
	attackTargets:    2,
}

var flamerAgentStats = &agentStats{
	kind:             agentFlamer,
	image:            assets.ImageFlamerAgent,
	size:             sizeLarge,
	diodeOffset:      7,
	tier:             3,
	cost:             30,
	upkeep:           8,
	canPatrol:        true,
	speed:            135,
	maxHealth:        40,
	attackRange:      115,
	attackDelay:      1.1,
	attackSound:      assets.AudioFlamerShot,
	projectileImage:  assets.ImageFlamerProjectile,
	projectileImpact: projectileExplosionNormal,
	projectileArea:   18,
	projectileSpeed:  160,
	projectileDamage: damageValue{health: 5},
	attackTargets:    2,
}

var fighterAgentStats = &agentStats{
	kind:             agentFighter,
	image:            assets.ImageFighterAgent,
	size:             sizeMedium,
	diodeOffset:      1,
	tier:             2,
	cost:             20,
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
	attackTargets:    1,
}

var destroyerAgentStats = &agentStats{
	kind:             agentDestroyer,
	image:            assets.ImageDestroyerAgent,
	size:             sizeLarge,
	diodeOffset:      0,
	tier:             3,
	cost:             45,
	upkeep:           20,
	canPatrol:        true,
	speed:            85,
	maxHealth:        35,
	attackRange:      210,
	attackDelay:      1.9,
	attackSound:      assets.AudioDestroyerBeam,
	projectileDamage: damageValue{health: 6},
	attackTargets:    1,
}

var repellerAgentStats = &agentStats{
	kind:             agentRepeller,
	image:            assets.ImageRepellerAgent,
	size:             sizeMedium,
	diodeOffset:      8,
	tier:             2,
	cost:             15,
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
	attackTargets:    2,
}
