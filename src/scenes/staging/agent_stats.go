package staging

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
)

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

	weapon *weaponStats
}

var gunpointAgentStats = &agentStats{
	kind:      agentGunpoint,
	image:     assets.ImageGunpointAgent,
	size:      sizeLarge,
	cost:      12,
	upkeep:    18,
	maxHealth: 75,
	canPatrol: true,
	weapon: &weaponStats{
		AttackRange:     200,
		Reload:          2.2,
		AttackSound:     assets.AudioGunpointShot,
		ProjectileImage: assets.ImageGunpointProjectile,
		ImpactArea:      10,
		ProjectileSpeed: 280,
		Damage:          damageValue{health: 2},
		MaxTargets:      1,
		BurstSize:       3,
		BurstDelay:      0.1,
		TargetFlags:     targetGround,
		FireOffset:      gmath.Vec{Y: 4},
	},
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
	speed:         135,
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
	kind:        agentMilitia,
	image:       assets.ImageMilitiaAgent,
	size:        sizeSmall,
	diodeOffset: 5,
	tier:        1,
	cost:        12,
	upkeep:      4,
	canPatrol:   true,
	speed:       75,
	maxHealth:   12,
	weapon: &weaponStats{
		AttackRange:     130,
		Reload:          2.5,
		AttackSound:     assets.AudioMilitiaShot,
		ProjectileImage: assets.ImageMilitiaProjectile,
		ImpactArea:      10,
		ProjectileSpeed: 180,
		Damage:          damageValue{health: 2, morale: 2},
		MaxTargets:      1,
		BurstSize:       1,
		TargetFlags:     targetFlying | targetGround,
	},
}

var cripplerAgentStats = &agentStats{
	kind:        agentCrippler,
	image:       assets.ImageCripplerAgent,
	size:        sizeMedium,
	diodeOffset: 5,
	tier:        1,
	cost:        15,
	upkeep:      4,
	canPatrol:   true,
	speed:       60,
	maxHealth:   15,
	weapon: &weaponStats{
		AttackRange:     240,
		Reload:          3.2,
		AttackSound:     assets.AudioCripplerShot,
		ProjectileImage: assets.ImageCripplerProjectile,
		ImpactArea:      10,
		ProjectileSpeed: 250,
		Damage:          damageValue{health: 1, slow: 2},
		MaxTargets:      2,
		BurstSize:       1,
		TargetFlags:     targetFlying | targetGround,
	},
}

var flamerAgentStats = &agentStats{
	kind:        agentFlamer,
	image:       assets.ImageFlamerAgent,
	size:        sizeLarge,
	diodeOffset: 7,
	tier:        3,
	cost:        30,
	upkeep:      8,
	canPatrol:   true,
	speed:       110,
	maxHealth:   40,
	weapon: &weaponStats{
		AttackRange:     115,
		Reload:          1.1,
		AttackSound:     assets.AudioFlamerShot,
		ProjectileImage: assets.ImageFlamerProjectile,
		Explosion:       projectileExplosionNormal,
		ImpactArea:      18,
		ProjectileSpeed: 160,
		Damage:          damageValue{health: 5},
		MaxTargets:      2,
		BurstSize:       1,
		TargetFlags:     targetFlying | targetGround,
	},
}

var prismAgentStats = &agentStats{
	kind:        agentPrism,
	image:       assets.ImagePrismAgent,
	size:        sizeMedium,
	diodeOffset: 1,
	tier:        2,
	cost:        24,
	upkeep:      12,
	canPatrol:   true,
	speed:       65,
	maxHealth:   28,
	weapon: &weaponStats{
		AttackRange:     200,
		Reload:          3.7,
		AttackSound:     assets.AudioPrismShot,
		ImpactArea:      8,
		ProjectileSpeed: 220,
		Damage:          damageValue{health: 4},
		MaxTargets:      1,
		BurstSize:       1,
		TargetFlags:     targetFlying | targetGround,
	},
}

var fighterAgentStats = &agentStats{
	kind:        agentFighter,
	image:       assets.ImageFighterAgent,
	size:        sizeMedium,
	diodeOffset: 1,
	tier:        2,
	cost:        20,
	upkeep:      7,
	canPatrol:   true,
	speed:       90,
	maxHealth:   26,
	weapon: &weaponStats{
		AttackRange:     180,
		Reload:          2,
		AttackSound:     assets.AudioFighterBeam,
		ProjectileImage: assets.ImageFighterProjectile,
		ImpactArea:      8,
		ProjectileSpeed: 220,
		Damage:          damageValue{health: 4},
		MaxTargets:      1,
		BurstSize:       1,
		TargetFlags:     targetFlying | targetGround,
	},
}

var antiAirAgentStats = &agentStats{
	kind:        agentAntiAir,
	image:       assets.ImageAntiAirAgent,
	size:        sizeMedium,
	diodeOffset: 1,
	tier:        2,
	cost:        22,
	upkeep:      8,
	canPatrol:   true,
	speed:       80,
	maxHealth:   22,
	weapon: &weaponStats{
		AttackRange:     250,
		Reload:          2.4,
		AttackSound:     assets.AudioAntiAirMissiles,
		ProjectileImage: assets.ImageAntiAirMissile,
		ImpactArea:      18,
		ProjectileSpeed: 250,
		Damage:          damageValue{health: 2},
		MaxTargets:      1,
		BurstSize:       4,
		BurstDelay:      0.1,
		Explosion:       projectileExplosionNormal,
		ArcPower:        2,
		TargetFlags:     targetFlying,
		FireOffset:      gmath.Vec{Y: -8},
	},
}

var mortarAgentStats = &agentStats{
	kind:        agentMortar,
	image:       assets.ImageMortarAgent,
	size:        sizeMedium,
	diodeOffset: 1,
	tier:        2,
	cost:        18,
	upkeep:      6,
	canPatrol:   true,
	speed:       70,
	maxHealth:   28,
	weapon: &weaponStats{
		AttackRange:     350,
		Reload:          3.6,
		AttackSound:     assets.AudioMortarShot,
		ProjectileImage: assets.ImageMortarProjectile,
		ImpactArea:      14,
		ProjectileSpeed: 180,
		Damage:          damageValue{health: 9},
		MaxTargets:      1,
		BurstSize:       1,
		Explosion:       projectileExplosionNormal,
		ArcPower:        2.5,
		TargetFlags:     targetGround,
		RoundProjectile: true,
	},
}

var destroyerAgentStats = &agentStats{
	kind:        agentDestroyer,
	image:       assets.ImageDestroyerAgent,
	size:        sizeLarge,
	diodeOffset: 0,
	tier:        3,
	cost:        45,
	upkeep:      20,
	canPatrol:   true,
	speed:       85,
	maxHealth:   35,
	weapon: &weaponStats{
		AttackRange: 210,
		Reload:      1.9,
		AttackSound: assets.AudioDestroyerBeam,
		Damage:      damageValue{health: 6},
		MaxTargets:  1,
		BurstSize:   1,
		TargetFlags: targetFlying | targetGround,
	},
}

var repellerAgentStats = &agentStats{
	kind:        agentRepeller,
	image:       assets.ImageRepellerAgent,
	size:        sizeMedium,
	diodeOffset: 8,
	tier:        2,
	cost:        15,
	upkeep:      4,
	canGather:   true,
	maxPayload:  1,
	canPatrol:   true,
	speed:       105,
	maxHealth:   22,
	weapon: &weaponStats{
		AttackRange:     160,
		Reload:          2.4,
		AttackSound:     assets.AudioRepellerBeam,
		ProjectileImage: assets.ImageRepellerProjectile,
		ImpactArea:      10,
		ProjectileSpeed: 200,
		Damage:          damageValue{health: 1, morale: 4},
		MaxTargets:      2,
		BurstSize:       1,
		TargetFlags:     targetFlying | targetGround,
	},
}
