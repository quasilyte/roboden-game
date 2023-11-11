package gamedata

import (
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
)

var ArtifactsList = []*AgentStats{
	DroneFactoryAgentStats,
	PowerPlantAgentStats,
	RepulseTowerAgentStats,
	MegaRoombaAgentStats,
}

var MegaRoombaAgentStats = InitDroneStats(&AgentStats{
	Kind:            AgentMegaRoomba,
	IsFlying:        false,
	IsTurret:        true,
	IsNeutral:       true,
	Image:           assets.ImageMegaRoombaAgent,
	Size:            SizeMedium,
	Tier:            3,
	Speed:           30,
	MaxHealth:       170,
	DamageReduction: 0.4,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:         260,
		Reload:              4.2,
		AttackSound:         assets.AudioMegaRoombaShot1,
		ProjectileImage:     assets.ImageMegaRoombaProjectile,
		ImpactArea:          16,
		ProjectileSpeed:     370,
		Damage:              DamageValue{Health: 9, Slow: 2.5},
		FireOffsets:         []gmath.Vec{{X: -8, Y: -1}, {X: 8, Y: -1}, {Y: 1}},
		MaxTargets:          1,
		BurstSize:           3,
		AttacksPerBurst:     2,
		BurstDelay:          0.6,
		Accuracy:            0.85,
		TargetFlags:         TargetFlying | TargetGround,
		GroundDamageBonus:   -0.5,
		ProjectileFireSound: true,
	}),
})

var RepulseTowerAgentStats = InitDroneStats(&AgentStats{
	Kind:       AgentRepulseTower,
	IsFlying:   false,
	IsTurret:   true,
	IsBuilding: true,
	IsNeutral:  true,
	Image:      assets.ImageRepulseTower,
	Size:       SizeLarge,
	Upkeep:     0,
	MaxHealth:  140,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange: 440,
		Reload:      3.5,
		MaxTargets:  4,
		AttackSound: assets.AudioRepulseTowerAttack,
		Damage:      DamageValue{Health: 4, Morale: 0.9},
		TargetFlags: TargetGround | TargetFlying,
	}),
	FireOffset:     -14,
	BeamOpaqueTime: 0.1,
	BeamSlideSpeed: 1.5,
	BeamShift:      14,
})

var PowerPlantAgentStats = InitDroneStats(&AgentStats{
	Kind:            AgentPowerPlant,
	IsFlying:        false,
	IsTurret:        true,
	IsBuilding:      true,
	IsNeutral:       true,
	Image:           assets.ImagePowerPlantAgent,
	Size:            SizeLarge,
	Upkeep:          0,
	MaxHealth:       140,
	DamageReduction: 0.35,
})

var DroneFactoryAgentStats = InitDroneStats(&AgentStats{
	Kind:            AgentDroneFactory,
	IsFlying:        false,
	IsTurret:        true,
	IsBuilding:      true,
	IsNeutral:       true,
	Image:           assets.ImageRelictFactoryAgent,
	Size:            SizeLarge,
	Upkeep:          0,
	MaxHealth:       200,
	DamageReduction: 0.3,
})

var RelictAgentStats = InitDroneStats(&AgentStats{
	Kind:        AgentRelict,
	IsFlying:    true,
	IsNeutral:   true,
	Image:       assets.ImageRelictAgent,
	Size:        SizeMedium,
	DiodeOffset: 1,
	Tier:        2,
	Speed:       65,
	MaxHealth:   30,
	SelfRepair:  0.25,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRangeMarkMultiplier: 1.3,
		AttackRange:               220,
		Reload:                    2.5,
		AttackSound:               assets.AudioRelictAgentShot,
		Damage:                    DamageValue{Health: 6},
		MaxTargets:                1,
		TargetFlags:               TargetFlying | TargetGround,
		BuildingDamageBonus:       -0.25,
	}),
	BeamOpaqueTime: 0.15,
	BeamSlideSpeed: 4.2,
})
