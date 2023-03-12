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

//go:generate stringer -type=ColonyAgentKind -trimprefix=Agent
type ColonyAgentKind uint8

const (
	AgentWorker ColonyAgentKind = iota
	AgentMilitia

	// Tier2
	AgentFreighter
	AgentRedminer
	AgentCrippler
	AgentFighter
	AgentScavenger
	AgentCourier
	AgentTrucker
	AgentPrism
	AgentServo
	AgentRepeller
	AgentRepair
	AgentCloner
	AgentRecharger
	AgentGenerator
	AgentMortar
	AgentAntiAir

	// Tier3
	AgentRefresher
	AgentStormbringer
	AgentDestroyer
	AgentMarauder

	AgentKindNum

	// Buildings (not real agents/drones)
	AgentGunpoint
)

type AgentStats struct {
	Kind      ColonyAgentKind
	Image     resource.ImageID
	Tier      int
	PointCost int
	Cost      float64
	Upkeep    int

	Size UnitSize

	Speed float64

	MaxHealth float64

	EnergyRegenRateBonus float64

	CanGather  bool
	CanPatrol  bool
	CanCloak   bool
	MaxPayload int

	DiodeOffset float64

	SupportReload float64
	SupportRange  float64

	Weapon *WeaponStats
}

var TurretStatsList = []*AgentStats{
	GunpointAgentStats,
}

var GunpointAgentStats = &AgentStats{
	Kind:      AgentGunpoint,
	Image:     assets.ImageGunpointAgent,
	Size:      SizeLarge,
	Cost:      12,
	Upkeep:    16,
	MaxHealth: 90,
	CanPatrol: true,
	Weapon: initWeaponStats(&WeaponStats{
		AttackRange:     245,
		Reload:          2.2,
		AttackSound:     assets.AudioGunpointShot,
		ProjectileImage: assets.ImageGunpointProjectile,
		ImpactArea:      10,
		ProjectileSpeed: 280,
		Damage:          DamageValue{Health: 3},
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
		Damage:          DamageValue{Health: 2, Disarm: 2},
		MaxTargets:      1,
		BurstSize:       1,
		TargetFlags:     TargetFlying | TargetGround,
	}),
}

var TruckerAgentStats = &AgentStats{
	Kind:                 AgentTrucker,
	Image:                assets.ImageTruckerAgent,
	Size:                 SizeLarge,
	DiodeOffset:          1,
	Tier:                 3,
	Cost:                 40,
	Upkeep:               9,
	CanGather:            true,
	MaxPayload:           3,
	Speed:                75,
	MaxHealth:            35,
	EnergyRegenRateBonus: 0.5,
	Weapon: initWeaponStats(&WeaponStats{
		AttackRange:           160,
		Reload:                3,
		AttackSound:           assets.AudioCourierShot,
		ProjectileImage:       assets.ImageCourierProjectile,
		ImpactArea:            10,
		ProjectileSpeed:       170,
		Damage:                DamageValue{Health: 2, Slow: 1, Morale: 1},
		MaxTargets:            2,
		BurstSize:             1,
		ProjectileRotateSpeed: 24,
		TargetFlags:           TargetFlying,
	}),
}

var CourierAgentStats = &AgentStats{
	Kind:        AgentCourier,
	Image:       assets.ImageCourierAgent,
	Size:        SizeMedium,
	DiodeOffset: 5,
	Tier:        2,
	Cost:        20,
	PointCost:   2,
	Upkeep:      4,
	CanGather:   true,
	MaxPayload:  1,
	Speed:       80,
	MaxHealth:   30,
	Weapon: initWeaponStats(&WeaponStats{
		AttackRange:           140,
		Reload:                3.2,
		AttackSound:           assets.AudioCourierShot,
		ProjectileImage:       assets.ImageCourierProjectile,
		ImpactArea:            10,
		ProjectileSpeed:       170,
		Damage:                DamageValue{Health: 2, Slow: 1, Morale: 1},
		MaxTargets:            1,
		BurstSize:             1,
		ProjectileRotateSpeed: 24,
		TargetFlags:           TargetFlying,
	}),
}

var RedminerAgentStats = &AgentStats{
	Kind:                 AgentRedminer,
	Image:                assets.ImageRedminerAgent,
	Size:                 SizeMedium,
	DiodeOffset:          6,
	Tier:                 2,
	PointCost:            2,
	Cost:                 16,
	Upkeep:               5,
	CanGather:            true,
	MaxPayload:           1,
	Speed:                75,
	MaxHealth:            20,
	EnergyRegenRateBonus: 0.2,
}

var GeneratorAgentStats = &AgentStats{
	Kind:                 AgentGenerator,
	Image:                assets.ImageGeneratorAgent,
	Size:                 SizeMedium,
	DiodeOffset:          10,
	Tier:                 2,
	PointCost:            1,
	Cost:                 16,
	Upkeep:               2,
	CanGather:            true,
	MaxPayload:           1,
	Speed:                90,
	MaxHealth:            26,
	EnergyRegenRateBonus: 1,
}

var ClonerAgentStats = &AgentStats{
	Kind:        AgentCloner,
	Image:       assets.ImageClonerAgent,
	Size:        SizeMedium,
	DiodeOffset: 5,
	Tier:        2,
	PointCost:   4,
	Cost:        26,
	Upkeep:      10,
	CanGather:   true,
	MaxPayload:  1,
	Speed:       90,
	MaxHealth:   16,
}

var RepairAgentStats = &AgentStats{
	Kind:          AgentRepair,
	Image:         assets.ImageRepairAgent,
	Size:          SizeMedium,
	DiodeOffset:   5,
	Tier:          2,
	PointCost:     4,
	Cost:          26,
	Upkeep:        7,
	CanGather:     true,
	MaxPayload:    1,
	Speed:         100,
	MaxHealth:     18,
	SupportReload: 8.0,
	SupportRange:  450,
}

var RechargeAgentStats = &AgentStats{
	Kind:                 AgentRecharger,
	Image:                assets.ImageRechargerAgent,
	Size:                 SizeMedium,
	DiodeOffset:          9,
	Tier:                 2,
	PointCost:            2,
	Cost:                 15,
	Upkeep:               4,
	CanGather:            true,
	MaxPayload:           1,
	Speed:                90,
	MaxHealth:            16,
	EnergyRegenRateBonus: 0.2,
	SupportReload:        7,
	SupportRange:         400,
}

var RefresherAgentStats = &AgentStats{
	Kind:                 AgentRefresher,
	Image:                assets.ImageRefresherAgent,
	Size:                 SizeLarge,
	DiodeOffset:          6,
	Tier:                 3,
	Cost:                 50,
	Upkeep:               14,
	CanGather:            true,
	MaxPayload:           1,
	Speed:                100,
	MaxHealth:            24,
	EnergyRegenRateBonus: 0.2,
	SupportReload:        RechargeAgentStats.SupportReload,
	SupportRange:         RechargeAgentStats.SupportRange,
}

var ServoAgentStats = &AgentStats{
	Kind:          AgentServo,
	Image:         assets.ImageServoAgent,
	Size:          SizeMedium,
	DiodeOffset:   -4,
	Tier:          2,
	PointCost:     3,
	Cost:          26,
	Upkeep:        7,
	CanGather:     true,
	MaxPayload:    1,
	Speed:         125,
	MaxHealth:     18,
	SupportReload: 8,
	SupportRange:  310,
}

var FreighterAgentStats = &AgentStats{
	Kind:                 AgentFreighter,
	Image:                assets.ImageFreighterAgent,
	Size:                 SizeMedium,
	DiodeOffset:          1,
	Tier:                 2,
	PointCost:            2,
	Cost:                 18,
	Upkeep:               3,
	CanGather:            true,
	MaxPayload:           3,
	Speed:                70,
	MaxHealth:            28,
	EnergyRegenRateBonus: 0.5,
}

var CripplerAgentStats = &AgentStats{
	Kind:        AgentCrippler,
	Image:       assets.ImageCripplerAgent,
	Size:        SizeMedium,
	DiodeOffset: 5,
	Tier:        2,
	PointCost:   2,
	Cost:        16,
	Upkeep:      4,
	CanPatrol:   true,
	CanCloak:    true,
	Speed:       65,
	MaxHealth:   18,
	Weapon: initWeaponStats(&WeaponStats{
		AttackRange:     255,
		Reload:          2.7,
		AttackSound:     assets.AudioCripplerShot,
		ProjectileImage: assets.ImageCripplerProjectile,
		ImpactArea:      10,
		ProjectileSpeed: 250,
		Damage:          DamageValue{Health: 1, Slow: 2},
		MaxTargets:      3,
		BurstSize:       1,
		TargetFlags:     TargetFlying | TargetGround,
	}),
}

var StormbringerAgentStats = &AgentStats{
	Kind:                 AgentStormbringer,
	Image:                assets.ImageStormbringerAgent,
	Size:                 SizeLarge,
	DiodeOffset:          7,
	Tier:                 3,
	Cost:                 45,
	Upkeep:               9,
	CanPatrol:            true,
	CanGather:            true,
	Speed:                100,
	MaxHealth:            34,
	EnergyRegenRateBonus: 1,
	Weapon: initWeaponStats(&WeaponStats{
		AttackRange:           150,
		Reload:                2.6,
		AttackSound:           assets.AudioStormbringerShot,
		ProjectileImage:       assets.ImageStormbringerProjectile,
		ImpactArea:            18,
		ProjectileSpeed:       200,
		ProjectileRotateSpeed: 4,
		Damage:                DamageValue{Health: 1, Disarm: 2},
		MaxTargets:            2,
		BurstSize:             4,
		BurstDelay:            0.03,
		TargetFlags:           TargetFlying | TargetGround,
	}),
}

var PrismAgentStats = &AgentStats{
	Kind:        AgentPrism,
	Image:       assets.ImagePrismAgent,
	Size:        SizeMedium,
	DiodeOffset: 1,
	Tier:        2,
	PointCost:   3,
	Cost:        26,
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
	PointCost:   3,
	Cost:        22,
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

var ScavengerAgentStats = &AgentStats{
	Kind:          AgentScavenger,
	Image:         assets.ImageScavengerAgent,
	Size:          SizeMedium,
	DiodeOffset:   -5,
	Tier:          2,
	PointCost:     2,
	Cost:          18,
	Upkeep:        6,
	CanPatrol:     true,
	Speed:         100,
	MaxHealth:     22,
	SupportReload: 16,
	MaxPayload:    2,
	Weapon: initWeaponStats(&WeaponStats{
		AttackRange:     160,
		Reload:          2.5,
		AttackSound:     assets.AudioScavengerShot,
		ProjectileImage: assets.ImageScavengerProjectile,
		ImpactArea:      8,
		ProjectileSpeed: 250,
		Damage:          DamageValue{Health: 2},
		MaxTargets:      1,
		BurstSize:       2,
		BurstDelay:      0.12,
		TargetFlags:     TargetFlying | TargetGround,
	}),
}

var AntiAirAgentStats = &AgentStats{
	Kind:        AgentAntiAir,
	Image:       assets.ImageAntiAirAgent,
	Size:        SizeMedium,
	DiodeOffset: 1,
	Tier:        2,
	PointCost:   3,
	Cost:        24,
	Upkeep:      8,
	CanPatrol:   true,
	Speed:       80,
	MaxHealth:   20,
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
	PointCost:   1,
	Cost:        18,
	Upkeep:      6,
	CanPatrol:   true,
	Speed:       70,
	MaxHealth:   30,
	Weapon: initWeaponStats(&WeaponStats{
		AttackRange:     370,
		Reload:          3.3,
		AttackSound:     assets.AudioMortarShot,
		ProjectileImage: assets.ImageMortarProjectile,
		ImpactArea:      18,
		ProjectileSpeed: 180,
		Damage:          DamageValue{Health: 10},
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
	Cost:        55,
	Upkeep:      20,
	CanPatrol:   true,
	Speed:       85,
	MaxHealth:   38,
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

var MarauderAgentStats = &AgentStats{
	Kind:          AgentMarauder,
	Image:         assets.ImageMarauderAgent,
	Size:          SizeLarge,
	DiodeOffset:   0,
	Tier:          3,
	Cost:          40,
	Upkeep:        12,
	CanPatrol:     true,
	CanCloak:      true,
	Speed:         100,
	SupportReload: 14,
	MaxHealth:     28,
	MaxPayload:    3,
	Weapon: initWeaponStats(&WeaponStats{
		AttackRange:     255,
		Reload:          2.5,
		ProjectileImage: assets.ImageMarauderProjectile,
		ImpactArea:      16,
		ProjectileSpeed: 300,
		AttackSound:     assets.AudioMarauderShot,
		Damage:          DamageValue{Health: 4, Slow: 2},
		BurstSize:       1,
		MaxTargets:      3,
		TargetFlags:     TargetFlying | TargetGround,
	}),
}

var RepellerAgentStats = &AgentStats{
	Kind:        AgentRepeller,
	Image:       assets.ImageRepellerAgent,
	Size:        SizeMedium,
	DiodeOffset: 8,
	Tier:        2,
	PointCost:   3,
	Cost:        16,
	Upkeep:      4,
	CanGather:   true,
	MaxPayload:  1,
	CanPatrol:   true,
	Speed:       105,
	MaxHealth:   22,
	Weapon: initWeaponStats(&WeaponStats{
		AttackRange:     160,
		Reload:          2.2,
		AttackSound:     assets.AudioRepellerBeam,
		ProjectileImage: assets.ImageRepellerProjectile,
		ImpactArea:      10,
		ProjectileSpeed: 200,
		Damage:          DamageValue{Health: 1, Disarm: 4},
		MaxTargets:      2,
		BurstSize:       1,
		TargetFlags:     TargetFlying | TargetGround,
	}),
}
