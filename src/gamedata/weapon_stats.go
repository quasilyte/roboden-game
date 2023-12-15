package gamedata

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/gmath"
)

type DamageFlags uint64

const (
	DmgflagNoFlash DamageFlags = 1 << iota
	DmgflagStun
	DmgflagStunImproved
	DmgflagAggro
	DmgflagMark
	DmgflagUnblockable
)

type DamageValue struct {
	Health float64
	Morale float64
	Disarm float64
	Energy float64
	Slow   float64
	Flags  DamageFlags
}

func (dmg DamageValue) HasFlag(mask DamageFlags) bool {
	return dmg.Flags&mask != 0
}

type WeaponStats struct {
	MaxTargets                int
	TargetMaxDist             float64 // If 0, no limits are applied
	ProjectileImage           resource.ImageID
	ProjectileSpeed           float64
	ProjectileRotateSpeed     float64
	ProjectileFireSound       bool
	ImpactArea                float64
	ImpactAreaSqr             float64 // A precomputed ImpactArea*ImpactArea value
	AttackRange               float64
	AttackRangeSqr            float64 // A precomputed AttackRange*AttackRange value
	AttackRangeMarkMultiplier float64
	Damage                    DamageValue
	Explosion                 ProjectileExplosionKind
	TrailEffect               ProjectileTrailEffect
	AlwaysExplodes            bool
	BurstSize                 int
	AttacksPerBurst           int
	BurstDelay                float64
	Reload                    float64
	EnergyCost                float64
	AttackSound               resource.AudioID
	FireOffsets               []gmath.Vec
	ArcPower                  float64
	Accuracy                  float64
	TargetFlags               TargetKind

	GroundDamageBonus        float64
	FlyingDamageBonus        float64
	BuildingDamageBonus      float64
	GroundTargetDamageMult   float64
	FlyingTargetDamageMult   float64
	BuildingTargetDamageMult float64

	RoundProjectile bool
	RandArc         bool
}

type ProjectileTrailEffect int

const (
	ProjectileTrailNone ProjectileTrailEffect = iota
	ProjectileTrailSmoke
	ProjectileTrailRoomba
	ProjectileTrailFire
	ProjectileTrailEnergySpear
	ProjectileTrailIonMortar
	ProjectileTrailSuperIonMortar
	ProjectileTrailTankColonyWeapon1
	ProjectileTrailMagma
	ProjectileTrailGrenade
)

type ProjectileExplosionKind int

const (
	ProjectileExplosionNone ProjectileExplosionKind = iota
	ProjectileExplosionNormal
	ProjectileExplosionLarge
	ProjectileExplosionIonBlast
	ProjectileExplosionSuperIonBlast
	ProjectileExplosionAbomb
	ProjectileExplosionMagma
	ProjectileExplosionBigVertical
	ProjectileExplosionPurple
	ProjectileExplosionHeavyCrawlerLaser
	ProjectileExplosionFighterLaser
	ProjectileExplosionScoutIon
	ProjectileExplosionCommanderLaser
	ProjectileExplosionShocker
	ProjectileExplosionCripplerBlaster
	ProjectileExplosionStealthLaser
	ProjectileExplosionGreenZap
	ProjectileExplosionPurpleZap
	ProjectileExplosionRoombaShot
	ProjectileExplosionScarab
	ProjectileExplosionServant
	ProjectileExplosionTankColonyBlaster
	ProjectileExplosionSentinelGun
)

type TargetKind int

const (
	TargetFlying TargetKind = 1 << iota
	TargetGround
)

const TargetAny = TargetFlying | TargetGround

func InitDroneStats(stats *AgentStats) *AgentStats {
	stats.SupportRangeSqr = stats.SupportRange * stats.SupportRange
	if !stats.NoAutoAttack {
		stats.NoAutoAttack = stats.Weapon == nil
	}
	return stats
}

func InitWeaponStats(stats *WeaponStats) *WeaponStats {
	if stats.Accuracy == 0 {
		stats.Accuracy = 1.0
	}
	if stats.AttacksPerBurst == 0 {
		stats.AttacksPerBurst = 1
	}
	if stats.AttackRangeMarkMultiplier == 0 {
		stats.AttackRangeMarkMultiplier = 1
	}
	stats.ImpactAreaSqr = stats.ImpactArea * stats.ImpactArea
	stats.AttackRangeSqr = stats.AttackRange * stats.AttackRange
	stats.FlyingTargetDamageMult = 1 + stats.FlyingDamageBonus
	stats.GroundTargetDamageMult = 1 + stats.GroundDamageBonus
	stats.BuildingTargetDamageMult = 1 + stats.BuildingDamageBonus
	return stats
}
