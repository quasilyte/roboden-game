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
}

var GunpointAgentStats = InitDroneStats(&AgentStats{
	Kind:       AgentGunpoint,
	IsFlying:   false,
	IsTurret:   true,
	IsBuilding: true,
	Image:      assets.ImageGunpointAgent,
	Size:       SizeLarge,
	Upkeep:     12,
	MaxHealth:  100,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:     280,
		Reload:          2,
		AttackSound:     assets.AudioGunpointShot,
		ProjectileImage: assets.ImageGunpointProjectile,
		ImpactArea:      10,
		ProjectileSpeed: 280,
		Damage:          DamageValue{Health: 4},
		MaxTargets:      1,
		BurstSize:       3,
		BurstDelay:      0.1,
		TargetFlags:     TargetGround,
		FireOffset:      gmath.Vec{Y: 6},
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
	Upkeep:     14,
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
	Kind:           AgentTetherBeacon,
	IsFlying:       false,
	IsTurret:       true,
	IsBuilding:     true,
	Image:          assets.ImageTetherBeaconAgent,
	Size:           SizeLarge,
	Upkeep:         8,
	MaxHealth:      75,
	SupportReload:  10,
	SupportRange:   450,
	BeamSlideSpeed: 0.4,
	HasSupport:     true,
})
