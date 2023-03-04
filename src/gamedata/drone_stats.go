package gamedata

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
)

type UnitSize int

const (
	SizeSmall UnitSize = iota
	SizeMedium
	SizeLarge
)

type ColonyAgentKind uint8

const (
	AgentWorker ColonyAgentKind = iota
	AgentMilitia

	// Tier2
	AgentFreighter
	AgentRedminer
	AgentCrippler
	AgentFighter
	AgentPrism
	AgentServo
	AgentRepeller
	AgentRepair
	AgentRecharger
	AgentGenerator
	AgentMortar
	AgentAntiAir

	// Tier3
	AgentRefresher
	AgentFlamer
	AgentDestroyer

	AgentKindNum

	// Buildings (not real agents/drones)
	AgentGunpoint
)

type AgentStats struct {
	Kind   ColonyAgentKind
	Image  resource.ImageID
	Tier   int
	Cost   float64
	Upkeep int

	Size UnitSize

	Speed float64

	MaxHealth float64

	CanGather  bool
	CanPatrol  bool
	MaxPayload int

	DiodeOffset float64

	SupportReload float64
	SupportRange  float64

	Weapon *WeaponStats
}

var GunpointAgentStats = &AgentStats{
	Kind:      AgentGunpoint,
	Image:     assets.ImageGunpointAgent,
	Size:      SizeLarge,
	Cost:      12,
	Upkeep:    18,
	MaxHealth: 85,
	CanPatrol: true,
	Weapon: initWeaponStats(&WeaponStats{
		AttackRange:     240,
		Reload:          2.2,
		AttackSound:     assets.AudioGunpointShot,
		ProjectileImage: assets.ImageGunpointProjectile,
		ImpactArea:      10,
		ProjectileSpeed: 280,
		Damage:          DamageValue{Health: 2},
		MaxTargets:      1,
		BurstSize:       3,
		BurstDelay:      0.1,
		TargetFlags:     TargetGround,
		FireOffset:      gmath.Vec{Y: 4},
	}),
}

var WorkerAgentStats = &AgentStats{
	Kind:        AgentWorker,
	Image:       assets.ImageWorkerAgent,
	Size:        SizeSmall,
	DiodeOffset: 5,
	Tier:        1,
	Cost:        8,
	Upkeep:      2,
	CanGather:   true,
	MaxPayload:  1,
	Speed:       80,
	MaxHealth:   12,
}

var RedminerAgentStats = &AgentStats{
	Kind:        AgentRedminer,
	Image:       assets.ImageRedminerAgent,
	Size:        SizeMedium,
	DiodeOffset: 6,
	Tier:        2,
	Cost:        15,
	Upkeep:      3,
	CanGather:   true,
	MaxPayload:  1,
	Speed:       75,
	MaxHealth:   18,
}

var GeneratorAgentStats = &AgentStats{
	Kind:        AgentGenerator,
	Image:       assets.ImageGeneratorAgent,
	Size:        SizeMedium,
	DiodeOffset: 10,
	Tier:        2,
	Cost:        15,
	Upkeep:      2,
	CanGather:   true,
	MaxPayload:  1,
	Speed:       90,
	MaxHealth:   20,
}

var RepairAgentStats = &AgentStats{
	Kind:          AgentRepair,
	Image:         assets.ImageRepairAgent,
	Size:          SizeMedium,
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

var RechargeAgentStats = &AgentStats{
	Kind:          AgentRecharger,
	Image:         assets.ImageRechargerAgent,
	Size:          SizeMedium,
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

var RefresherAgentStats = &AgentStats{
	Kind:          AgentRefresher,
	Image:         assets.ImageRefresherAgent,
	Size:          SizeLarge,
	DiodeOffset:   7,
	Tier:          3,
	Cost:          40,
	Upkeep:        10,
	CanGather:     true,
	MaxPayload:    1,
	Speed:         100,
	MaxHealth:     24,
	SupportReload: RechargeAgentStats.SupportReload,
	SupportRange:  RechargeAgentStats.SupportRange,
}

var ServoAgentStats = &AgentStats{
	Kind:          AgentServo,
	Image:         assets.ImageServoAgent,
	Size:          SizeMedium,
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

var FreighterAgentStats = &AgentStats{
	Kind:        AgentFreighter,
	Image:       assets.ImageFreighterAgent,
	Size:        SizeMedium,
	DiodeOffset: 1,
	Tier:        2,
	Cost:        15,
	Upkeep:      3,
	CanGather:   true,
	MaxPayload:  3,
	Speed:       70,
	MaxHealth:   25,
}

var MilitiaAgentStats = &AgentStats{
	Kind:        AgentMilitia,
	Image:       assets.ImageMilitiaAgent,
	Size:        SizeSmall,
	DiodeOffset: 5,
	Tier:        1,
	Cost:        10,
	Upkeep:      4,
	CanPatrol:   true,
	Speed:       75,
	MaxHealth:   12,
	Weapon: initWeaponStats(&WeaponStats{
		AttackRange:     130,
		Reload:          2.5,
		AttackSound:     assets.AudioMilitiaShot,
		ProjectileImage: assets.ImageMilitiaProjectile,
		ImpactArea:      10,
		ProjectileSpeed: 180,
		Damage:          DamageValue{Health: 2, Morale: 2},
		MaxTargets:      1,
		BurstSize:       1,
		TargetFlags:     TargetFlying | TargetGround,
	}),
}

var CripplerAgentStats = &AgentStats{
	Kind:        AgentCrippler,
	Image:       assets.ImageCripplerAgent,
	Size:        SizeMedium,
	DiodeOffset: 5,
	Tier:        1,
	Cost:        15,
	Upkeep:      4,
	CanPatrol:   true,
	Speed:       65,
	MaxHealth:   15,
	Weapon: initWeaponStats(&WeaponStats{
		AttackRange:     240,
		Reload:          3.2,
		AttackSound:     assets.AudioCripplerShot,
		ProjectileImage: assets.ImageCripplerProjectile,
		ImpactArea:      10,
		ProjectileSpeed: 250,
		Damage:          DamageValue{Health: 1, Slow: 2},
		MaxTargets:      2,
		BurstSize:       1,
		TargetFlags:     TargetFlying | TargetGround,
	}),
}

var FlamerAgentStats = &AgentStats{
	Kind:        AgentFlamer,
	Image:       assets.ImageFlamerAgent,
	Size:        SizeLarge,
	DiodeOffset: 7,
	Tier:        3,
	Cost:        30,
	Upkeep:      8,
	CanPatrol:   true,
	Speed:       105,
	MaxHealth:   40,
	Weapon: initWeaponStats(&WeaponStats{
		AttackRange:     115,
		Reload:          1.1,
		AttackSound:     assets.AudioFlamerShot,
		ProjectileImage: assets.ImageFlamerProjectile,
		Explosion:       ProjectileExplosionNormal,
		ImpactArea:      18,
		ProjectileSpeed: 160,
		Damage:          DamageValue{Health: 5},
		MaxTargets:      2,
		BurstSize:       1,
		TargetFlags:     TargetFlying | TargetGround,
	}),
}

var PrismAgentStats = &AgentStats{
	Kind:        AgentPrism,
	Image:       assets.ImagePrismAgent,
	Size:        SizeMedium,
	DiodeOffset: 1,
	Tier:        2,
	Cost:        24,
	Upkeep:      12,
	CanPatrol:   true,
	Speed:       70,
	MaxHealth:   28,
	Weapon: initWeaponStats(&WeaponStats{
		AttackRange:     200,
		Reload:          3.7,
		AttackSound:     assets.AudioPrismShot,
		ImpactArea:      8,
		ProjectileSpeed: 220,
		Damage:          DamageValue{Health: 4},
		MaxTargets:      1,
		BurstSize:       1,
		TargetFlags:     TargetFlying | TargetGround,
	}),
}

var FighterAgentStats = &AgentStats{
	Kind:        AgentFighter,
	Image:       assets.ImageFighterAgent,
	Size:        SizeMedium,
	DiodeOffset: 1,
	Tier:        2,
	Cost:        20,
	Upkeep:      7,
	CanPatrol:   true,
	Speed:       90,
	MaxHealth:   26,
	Weapon: initWeaponStats(&WeaponStats{
		AttackRange:     180,
		Reload:          2,
		AttackSound:     assets.AudioFighterBeam,
		ProjectileImage: assets.ImageFighterProjectile,
		ImpactArea:      8,
		ProjectileSpeed: 220,
		Damage:          DamageValue{Health: 4},
		MaxTargets:      1,
		BurstSize:       1,
		TargetFlags:     TargetFlying | TargetGround,
	}),
}

var AntiAirAgentStats = &AgentStats{
	Kind:        AgentAntiAir,
	Image:       assets.ImageAntiAirAgent,
	Size:        SizeMedium,
	DiodeOffset: 1,
	Tier:        2,
	Cost:        22,
	Upkeep:      8,
	CanPatrol:   true,
	Speed:       80,
	MaxHealth:   22,
	Weapon: initWeaponStats(&WeaponStats{
		AttackRange:     250,
		Reload:          2.4,
		AttackSound:     assets.AudioAntiAirMissiles,
		ProjectileImage: assets.ImageAntiAirMissile,
		ImpactArea:      18,
		ProjectileSpeed: 250,
		Damage:          DamageValue{Health: 2},
		MaxTargets:      1,
		BurstSize:       4,
		BurstDelay:      0.1,
		Explosion:       ProjectileExplosionNormal,
		ArcPower:        2,
		TargetFlags:     TargetFlying,
		FireOffset:      gmath.Vec{Y: -8},
	}),
}

var MortarAgentStats = &AgentStats{
	Kind:        AgentMortar,
	Image:       assets.ImageMortarAgent,
	Size:        SizeMedium,
	DiodeOffset: 1,
	Tier:        2,
	Cost:        18,
	Upkeep:      6,
	CanPatrol:   true,
	Speed:       70,
	MaxHealth:   28,
	Weapon: initWeaponStats(&WeaponStats{
		AttackRange:     350,
		Reload:          3.6,
		AttackSound:     assets.AudioMortarShot,
		ProjectileImage: assets.ImageMortarProjectile,
		ImpactArea:      14,
		ProjectileSpeed: 180,
		Damage:          DamageValue{Health: 9},
		MaxTargets:      1,
		BurstSize:       1,
		Explosion:       ProjectileExplosionNormal,
		ArcPower:        2.5,
		TargetFlags:     TargetGround,
		RoundProjectile: true,
	}),
}

var DestroyerAgentStats = &AgentStats{
	Kind:        AgentDestroyer,
	Image:       assets.ImageDestroyerAgent,
	Size:        SizeLarge,
	DiodeOffset: 0,
	Tier:        3,
	Cost:        45,
	Upkeep:      20,
	CanPatrol:   true,
	Speed:       85,
	MaxHealth:   35,
	Weapon: initWeaponStats(&WeaponStats{
		AttackRange: 210,
		Reload:      1.9,
		AttackSound: assets.AudioDestroyerBeam,
		Damage:      DamageValue{Health: 6},
		MaxTargets:  1,
		BurstSize:   1,
		TargetFlags: TargetFlying | TargetGround,
	}),
}

var RepellerAgentStats = &AgentStats{
	Kind:        AgentRepeller,
	Image:       assets.ImageRepellerAgent,
	Size:        SizeMedium,
	DiodeOffset: 8,
	Tier:        2,
	Cost:        15,
	Upkeep:      4,
	CanGather:   true,
	MaxPayload:  1,
	CanPatrol:   true,
	Speed:       105,
	MaxHealth:   22,
	Weapon: initWeaponStats(&WeaponStats{
		AttackRange:     160,
		Reload:          2.4,
		AttackSound:     assets.AudioRepellerBeam,
		ProjectileImage: assets.ImageRepellerProjectile,
		ImpactArea:      10,
		ProjectileSpeed: 200,
		Damage:          DamageValue{Health: 1, Morale: 4},
		MaxTargets:      2,
		BurstSize:       1,
		TargetFlags:     TargetFlying | TargetGround,
	}),
}
