package gamedata

import (
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
)

var TurretStatsList = []*AgentStats{
	GunpointAgentStats,
	TetherBeaconAgentStats,
	BeamTowerAgentStats,
	HarvesterAgentStats,
	SiegeAgentStats,
	RefineryAgentStats,
	SentinelpointAgentStats,
}

var SiegeAgentWeapon = InitWeaponStats(&WeaponStats{
	AttackRange:               800,
	Reload:                    10,
	AttackSound:               assets.AudioSiegeRocket1,
	ProjectileImage:           assets.ImageSiegeRocket,
	TrailEffect:               ProjectileTrailSmoke,
	Explosion:                 ProjectileExplosionLarge,
	ImpactArea:                26,
	ProjectileSpeed:           300,
	Damage:                    DamageValue{Health: 20},
	BuildingDamageBonus:       0.25,
	AttackRangeMarkMultiplier: 1.5,
	MaxTargets:                1,
	TargetFlags:               TargetGround,
	ArcPower:                  2.2,
	Accuracy:                  0.9,
	AlwaysExplodes:            true,
	ProjectileFireSound:       true,
})

var SiegeAgentStats = InitDroneStats(&AgentStats{
	ScoreCost:    SiegeTurretCost,
	Kind:         AgentSiege,
	IsFlying:     false,
	IsTurret:     true,
	IsBuilding:   true,
	Image:        assets.ImageSiegeAgent,
	PreviewImage: assets.ImageSiegeAgentIcon,
	Size:         SizeLarge,
	Upkeep:       15,
	MaxHealth:    65,
})

var SentinelpointAgentStats = InitDroneStats(&AgentStats{
	ScoreCost:    SentinelpointTurretCost,
	Kind:         AgentSentinelpoint,
	IsFlying:     false,
	IsTurret:     true,
	IsBuilding:   true,
	NoAutoAttack: true,
	Image:        assets.ImageSentinelpointAgent,
	PreviewImage: assets.ImageSentinelpointAgentIcon,
	Size:         SizeLarge,
	Upkeep:       25,
	MaxHealth:    80,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:     300,
		Reload:          1.85,
		AttackSound:     assets.AudioSentinelpointShot,
		ProjectileImage: assets.ImageSentinelpointProjectile,
		Explosion:       ProjectileExplosionSentinelGun,
		ImpactArea:      10,
		ProjectileSpeed: 400,
		Damage:          DamageValue{Health: 4, Disarm: 0.1},
		FireOffsets:     []gmath.Vec{{X: -4, Y: 1}, {X: 4, Y: 1}},
		BurstSize:       2,
		AttacksPerBurst: 2,
		MaxTargets:      6,
		TargetMaxDist:   80,
		Accuracy:        0.8,
		ArcPower:        0.6,
		TargetFlags:     TargetGround | TargetFlying,
	}),
})

var GunpointAgentStats = InitDroneStats(&AgentStats{
	Kind:            AgentGunpoint,
	IsFlying:        false,
	IsTurret:        true,
	IsBuilding:      true,
	Image:           assets.ImageGunpointAgent,
	Size:            SizeLarge,
	Upkeep:          12,
	MaxHealth:       100,
	DamageReduction: 0.15,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:     280,
		Reload:          2,
		AttackSound:     assets.AudioGunpointShot,
		ProjectileImage: assets.ImageGunpointProjectile,
		ImpactArea:      12,
		ProjectileSpeed: 280,
		Damage:          DamageValue{Health: 4},
		MaxTargets:      1,
		BurstSize:       3,
		BurstDelay:      0.1,
		TargetFlags:     TargetGround,
		FireOffsets:     []gmath.Vec{{Y: 6}},
	}),
})

var HarvesterAgentStats = InitDroneStats(&AgentStats{
	ScoreCost:  HarvesterTurretCost,
	Kind:       AgentHarvester,
	IsFlying:   false,
	IsTurret:   true,
	IsBuilding: false,
	Image:      assets.ImageHarvesterAgent,
	Size:       SizeLarge,
	Upkeep:     14,
	MaxHealth:  60,
	Speed:      8,
})

var BeamTowerAgentStats = InitDroneStats(&AgentStats{
	ScoreCost:  BeamTowerTurretCost,
	Kind:       AgentBeamTower,
	IsFlying:   false,
	IsTurret:   true,
	IsBuilding: true,
	Image:      assets.ImageBeamtowerAgent,
	Size:       SizeLarge,
	Upkeep:     22,
	MaxHealth:  50,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange: 380,
		Reload:      3.1,
		AttackSound: assets.AudioBeamTowerShot,
		Damage:      DamageValue{Health: 15},
		MaxTargets:  1,
		BurstSize:   1,
		TargetFlags: TargetFlying,
	}),
	FireOffset:     -16,
	BeamOpaqueTime: 0.1,
	BeamSlideSpeed: 0.4,
})

var TetherBeaconAgentStats = InitDroneStats(&AgentStats{
	Kind:            AgentTetherBeacon,
	IsFlying:        false,
	IsTurret:        true,
	IsBuilding:      true,
	Image:           assets.ImageTetherBeaconAgent,
	Size:            SizeLarge,
	Upkeep:          8,
	MaxHealth:       75,
	DamageReduction: 0.1,
	SupportReload:   10,
	SupportRange:    500,
	BeamSlideSpeed:  0.4,
	HasSupport:      true,
})

var RefineryAgentStats = InitDroneStats(&AgentStats{
	ScoreCost:       RefineryTurretCost,
	Kind:            AgentRefinery,
	IsFlying:        false,
	IsTurret:        true,
	IsBuilding:      true,
	Image:           assets.ImageRefineryAgent,
	Size:            SizeLarge,
	Upkeep:          5,
	MaxHealth:       110,
	DamageReduction: 0.2,
})
