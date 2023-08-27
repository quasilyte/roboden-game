package gamedata

import (
	"github.com/quasilyte/roboden-game/assets"
)

var ArtifactsList = []*AgentStats{
	DroneFactoryAgentStats,
	PowerPlantAgentStats,
	RepulseTowerAgentStats,
}

var RepulseTowerAgentStats = InitDroneStats(&AgentStats{
	Kind:       AgentRepulseTower,
	IsFlying:   false,
	IsTurret:   true,
	IsBuilding: true,
	IsNeutral:  true,
	Image:      assets.ImageRepulseTower,
	Size:       SizeLarge,
	Upkeep:     0,
	MaxHealth:  130,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange: 440,
		Reload:      3.6,
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
	Kind:       AgentPowerPlant,
	IsFlying:   false,
	IsTurret:   true,
	IsBuilding: true,
	IsNeutral:  true,
	Image:      assets.ImagePowerPlantAgent,
	Size:       SizeLarge,
	Upkeep:     0,
	MaxHealth:  140,
})

var DroneFactoryAgentStats = InitDroneStats(&AgentStats{
	Kind:       AgentDroneFactory,
	IsFlying:   false,
	IsTurret:   true,
	IsBuilding: true,
	IsNeutral:  true,
	Image:      assets.ImageRelictFactoryAgent,
	Size:       SizeLarge,
	Upkeep:     0,
	MaxHealth:  170,
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
	MaxHealth:   25,
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
	BeamSlideSpeed: 4.6,
})
