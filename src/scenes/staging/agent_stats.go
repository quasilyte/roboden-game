package staging

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
)

type agentMergeRecipe struct {
	agent1kind    gamedata.ColonyAgentKind
	agent1faction factionTag
	agent2kind    gamedata.ColonyAgentKind
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

func (r *agentMergeRecipe) match(kind gamedata.ColonyAgentKind, faction factionTag, s recipeSubject) bool {
	if s.kind != kind {
		return false
	}
	if faction == neutralFactionTag {
		return true
	}
	return s.faction == faction
}

type recipeSubject struct {
	kind    gamedata.ColonyAgentKind
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
	kinds := []gamedata.ColonyAgentKind{
		gamedata.AgentWorker,
		gamedata.AgentMilitia,
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

type agentStats struct {
	Kind   gamedata.ColonyAgentKind
	Image  resource.ImageID
	Tier   int
	Cost   float64
	Upkeep int

	Size gamedata.UnitSize

	Speed float64

	MaxHealth float64

	CanGather  bool
	CanPatrol  bool
	MaxPayload int

	DiodeOffset float64

	SupportReload float64
	SupportRange  float64

	Weapon *gamedata.WeaponStats
}

var gunpointAgentStats = &agentStats{
	Kind:      gamedata.AgentGunpoint,
	Image:     assets.ImageGunpointAgent,
	Size:      gamedata.SizeLarge,
	Cost:      12,
	Upkeep:    18,
	MaxHealth: 85,
	CanPatrol: true,
	Weapon: initWeaponStats(&gamedata.WeaponStats{
		AttackRange:     240,
		Reload:          2.2,
		AttackSound:     assets.AudioGunpointShot,
		ProjectileImage: assets.ImageGunpointProjectile,
		ImpactArea:      10,
		ProjectileSpeed: 280,
		Damage:          gamedata.DamageValue{Health: 2},
		MaxTargets:      1,
		BurstSize:       3,
		BurstDelay:      0.1,
		TargetFlags:     gamedata.TargetGround,
		FireOffset:      gmath.Vec{Y: 4},
	}),
}

var workerAgentStats = &agentStats{
	Kind:        gamedata.AgentWorker,
	Image:       assets.ImageWorkerAgent,
	Size:        gamedata.SizeSmall,
	DiodeOffset: 5,
	Tier:        1,
	Cost:        8,
	Upkeep:      2,
	CanGather:   true,
	MaxPayload:  1,
	Speed:       80,
	MaxHealth:   12,
}

var redminerAgentStats = &agentStats{
	Kind:        gamedata.AgentRedminer,
	Image:       assets.ImageRedminerAgent,
	Size:        gamedata.SizeMedium,
	DiodeOffset: 6,
	Tier:        2,
	Cost:        15,
	Upkeep:      3,
	CanGather:   true,
	MaxPayload:  1,
	Speed:       75,
	MaxHealth:   18,
}

var generatorAgentStats = &agentStats{
	Kind:        gamedata.AgentGenerator,
	Image:       assets.ImageGeneratorAgent,
	Size:        gamedata.SizeMedium,
	DiodeOffset: 10,
	Tier:        2,
	Cost:        15,
	Upkeep:      2,
	CanGather:   true,
	MaxPayload:  1,
	Speed:       90,
	MaxHealth:   20,
}

var repairAgentStats = &agentStats{
	Kind:          gamedata.AgentRepair,
	Image:         assets.ImageRepairAgent,
	Size:          gamedata.SizeMedium,
	DiodeOffset:   5,
	Tier:          2,
	Cost:          20,
	Upkeep:        5,
	CanGather:     true,
	MaxPayload:    1,
	Speed:         100,
	MaxHealth:     18,
	SupportReload: 8.0,
	SupportRange:  450,
}

var rechargeAgentStats = &agentStats{
	Kind:          gamedata.AgentRecharger,
	Image:         assets.ImageRechargerAgent,
	Size:          gamedata.SizeMedium,
	DiodeOffset:   9,
	Tier:          2,
	Cost:          15,
	Upkeep:        4,
	CanGather:     true,
	MaxPayload:    1,
	Speed:         90,
	MaxHealth:     16,
	SupportReload: 7,
	SupportRange:  400,
}

var refresherAgentStats = &agentStats{
	Kind:          gamedata.AgentRefresher,
	Image:         assets.ImageRefresherAgent,
	Size:          gamedata.SizeLarge,
	DiodeOffset:   7,
	Tier:          3,
	Cost:          40,
	Upkeep:        10,
	CanGather:     true,
	MaxPayload:    1,
	Speed:         100,
	MaxHealth:     24,
	SupportReload: rechargeAgentStats.SupportReload,
	SupportRange:  rechargeAgentStats.SupportRange,
}

var servoAgentStats = &agentStats{
	Kind:          gamedata.AgentServo,
	Image:         assets.ImageServoAgent,
	Size:          gamedata.SizeMedium,
	DiodeOffset:   -4,
	Tier:          2,
	Cost:          30,
	Upkeep:        7,
	CanGather:     true,
	MaxPayload:    1,
	Speed:         125,
	MaxHealth:     18,
	SupportReload: 8,
	SupportRange:  310,
}

var freighterAgentStats = &agentStats{
	Kind:        gamedata.AgentFreighter,
	Image:       assets.ImageFreighterAgent,
	Size:        gamedata.SizeMedium,
	DiodeOffset: 1,
	Tier:        2,
	Cost:        15,
	Upkeep:      3,
	CanGather:   true,
	MaxPayload:  3,
	Speed:       70,
	MaxHealth:   25,
}

var militiaAgentStats = &agentStats{
	Kind:        gamedata.AgentMilitia,
	Image:       assets.ImageMilitiaAgent,
	Size:        gamedata.SizeSmall,
	DiodeOffset: 5,
	Tier:        1,
	Cost:        10,
	Upkeep:      4,
	CanPatrol:   true,
	Speed:       75,
	MaxHealth:   12,
	Weapon: initWeaponStats(&gamedata.WeaponStats{
		AttackRange:     130,
		Reload:          2.5,
		AttackSound:     assets.AudioMilitiaShot,
		ProjectileImage: assets.ImageMilitiaProjectile,
		ImpactArea:      10,
		ProjectileSpeed: 180,
		Damage:          gamedata.DamageValue{Health: 2, Morale: 2},
		MaxTargets:      1,
		BurstSize:       1,
		TargetFlags:     gamedata.TargetFlying | gamedata.TargetGround,
	}),
}

var cripplerAgentStats = &agentStats{
	Kind:        gamedata.AgentCrippler,
	Image:       assets.ImageCripplerAgent,
	Size:        gamedata.SizeMedium,
	DiodeOffset: 5,
	Tier:        1,
	Cost:        15,
	Upkeep:      4,
	CanPatrol:   true,
	Speed:       65,
	MaxHealth:   15,
	Weapon: initWeaponStats(&gamedata.WeaponStats{
		AttackRange:     240,
		Reload:          3.2,
		AttackSound:     assets.AudioCripplerShot,
		ProjectileImage: assets.ImageCripplerProjectile,
		ImpactArea:      10,
		ProjectileSpeed: 250,
		Damage:          gamedata.DamageValue{Health: 1, Slow: 2},
		MaxTargets:      2,
		BurstSize:       1,
		TargetFlags:     gamedata.TargetFlying | gamedata.TargetGround,
	}),
}

var flamerAgentStats = &agentStats{
	Kind:        gamedata.AgentFlamer,
	Image:       assets.ImageFlamerAgent,
	Size:        gamedata.SizeLarge,
	DiodeOffset: 7,
	Tier:        3,
	Cost:        30,
	Upkeep:      8,
	CanPatrol:   true,
	Speed:       105,
	MaxHealth:   40,
	Weapon: initWeaponStats(&gamedata.WeaponStats{
		AttackRange:     115,
		Reload:          1.1,
		AttackSound:     assets.AudioFlamerShot,
		ProjectileImage: assets.ImageFlamerProjectile,
		Explosion:       gamedata.ProjectileExplosionNormal,
		ImpactArea:      18,
		ProjectileSpeed: 160,
		Damage:          gamedata.DamageValue{Health: 5},
		MaxTargets:      2,
		BurstSize:       1,
		TargetFlags:     gamedata.TargetFlying | gamedata.TargetGround,
	}),
}

var prismAgentStats = &agentStats{
	Kind:        gamedata.AgentPrism,
	Image:       assets.ImagePrismAgent,
	Size:        gamedata.SizeMedium,
	DiodeOffset: 1,
	Tier:        2,
	Cost:        24,
	Upkeep:      12,
	CanPatrol:   true,
	Speed:       70,
	MaxHealth:   28,
	Weapon: initWeaponStats(&gamedata.WeaponStats{
		AttackRange:     200,
		Reload:          3.7,
		AttackSound:     assets.AudioPrismShot,
		ImpactArea:      8,
		ProjectileSpeed: 220,
		Damage:          gamedata.DamageValue{Health: 4},
		MaxTargets:      1,
		BurstSize:       1,
		TargetFlags:     gamedata.TargetFlying | gamedata.TargetGround,
	}),
}

var fighterAgentStats = &agentStats{
	Kind:        gamedata.AgentFighter,
	Image:       assets.ImageFighterAgent,
	Size:        gamedata.SizeMedium,
	DiodeOffset: 1,
	Tier:        2,
	Cost:        20,
	Upkeep:      7,
	CanPatrol:   true,
	Speed:       90,
	MaxHealth:   26,
	Weapon: initWeaponStats(&gamedata.WeaponStats{
		AttackRange:     180,
		Reload:          2,
		AttackSound:     assets.AudioFighterBeam,
		ProjectileImage: assets.ImageFighterProjectile,
		ImpactArea:      8,
		ProjectileSpeed: 220,
		Damage:          gamedata.DamageValue{Health: 4},
		MaxTargets:      1,
		BurstSize:       1,
		TargetFlags:     gamedata.TargetFlying | gamedata.TargetGround,
	}),
}

var antiAirAgentStats = &agentStats{
	Kind:        gamedata.AgentAntiAir,
	Image:       assets.ImageAntiAirAgent,
	Size:        gamedata.SizeMedium,
	DiodeOffset: 1,
	Tier:        2,
	Cost:        22,
	Upkeep:      8,
	CanPatrol:   true,
	Speed:       80,
	MaxHealth:   22,
	Weapon: initWeaponStats(&gamedata.WeaponStats{
		AttackRange:     250,
		Reload:          2.4,
		AttackSound:     assets.AudioAntiAirMissiles,
		ProjectileImage: assets.ImageAntiAirMissile,
		ImpactArea:      18,
		ProjectileSpeed: 250,
		Damage:          gamedata.DamageValue{Health: 2},
		MaxTargets:      1,
		BurstSize:       4,
		BurstDelay:      0.1,
		Explosion:       gamedata.ProjectileExplosionNormal,
		ArcPower:        2,
		TargetFlags:     gamedata.TargetFlying,
		FireOffset:      gmath.Vec{Y: -8},
	}),
}

var mortarAgentStats = &agentStats{
	Kind:        gamedata.AgentMortar,
	Image:       assets.ImageMortarAgent,
	Size:        gamedata.SizeMedium,
	DiodeOffset: 1,
	Tier:        2,
	Cost:        18,
	Upkeep:      6,
	CanPatrol:   true,
	Speed:       70,
	MaxHealth:   28,
	Weapon: initWeaponStats(&gamedata.WeaponStats{
		AttackRange:     350,
		Reload:          3.6,
		AttackSound:     assets.AudioMortarShot,
		ProjectileImage: assets.ImageMortarProjectile,
		ImpactArea:      14,
		ProjectileSpeed: 180,
		Damage:          gamedata.DamageValue{Health: 9},
		MaxTargets:      1,
		BurstSize:       1,
		Explosion:       gamedata.ProjectileExplosionNormal,
		ArcPower:        2.5,
		TargetFlags:     gamedata.TargetGround,
		RoundProjectile: true,
	}),
}

var destroyerAgentStats = &agentStats{
	Kind:        gamedata.AgentDestroyer,
	Image:       assets.ImageDestroyerAgent,
	Size:        gamedata.SizeLarge,
	DiodeOffset: 0,
	Tier:        3,
	Cost:        45,
	Upkeep:      20,
	CanPatrol:   true,
	Speed:       85,
	MaxHealth:   35,
	Weapon: initWeaponStats(&gamedata.WeaponStats{
		AttackRange: 210,
		Reload:      1.9,
		AttackSound: assets.AudioDestroyerBeam,
		Damage:      gamedata.DamageValue{Health: 6},
		MaxTargets:  1,
		BurstSize:   1,
		TargetFlags: gamedata.TargetFlying | gamedata.TargetGround,
	}),
}

var repellerAgentStats = &agentStats{
	Kind:        gamedata.AgentRepeller,
	Image:       assets.ImageRepellerAgent,
	Size:        gamedata.SizeMedium,
	DiodeOffset: 8,
	Tier:        2,
	Cost:        15,
	Upkeep:      4,
	CanGather:   true,
	MaxPayload:  1,
	CanPatrol:   true,
	Speed:       105,
	MaxHealth:   22,
	Weapon: initWeaponStats(&gamedata.WeaponStats{
		AttackRange:     160,
		Reload:          2.4,
		AttackSound:     assets.AudioRepellerBeam,
		ProjectileImage: assets.ImageRepellerProjectile,
		ImpactArea:      10,
		ProjectileSpeed: 200,
		Damage:          gamedata.DamageValue{Health: 1, Morale: 4},
		MaxTargets:      2,
		BurstSize:       1,
		TargetFlags:     gamedata.TargetFlying | gamedata.TargetGround,
	}),
}
