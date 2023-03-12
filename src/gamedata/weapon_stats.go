package gamedata

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/gmath"
)

type DamageValue struct {
	Health float64
	Morale float64
	Disarm float64
	Energy float64
	Slow   float64
}

type WeaponStats struct {
	MaxTargets            int
	ProjectileImage       resource.ImageID
	ProjectileSpeed       float64
	ProjectileRotateSpeed float64
	ImpactArea            float64
	ImpactAreaSqr         float64 // A precomputed ImpactArea*ImpactArea value
	AttackRange           float64
	AttackRangeSqr        float64 // A precomputed AttackRange*AttackRange value
	Damage                DamageValue
	Explosion             ProjectileExplosionKind
	BurstSize             int
	BurstDelay            float64
	Reload                float64
	AttackSound           resource.AudioID
	FireOffset            gmath.Vec
	ArcPower              float64
	TargetFlags           TargetKind
	RoundProjectile       bool
}

type ProjectileExplosionKind int

const (
	ProjectileExplosionNone ProjectileExplosionKind = iota
	ProjectileExplosionNormal
)

type TargetKind int

const (
	TargetFlying TargetKind = 1 << iota
	TargetGround
)

func initWeaponStats(stats *WeaponStats) *WeaponStats {
	stats.ImpactAreaSqr = stats.ImpactArea * stats.ImpactArea
	stats.AttackRangeSqr = stats.AttackRange * stats.AttackRange
	return stats
}
