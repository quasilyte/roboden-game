package staging

import (
	"image/color"

	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
)

type creepStats struct {
	kind        creepKind
	image       resource.ImageID
	shadowImage resource.ImageID

	tier int

	speed float64
	size  float64

	animSpeed float64

	maxHealth float64

	weapon *weaponStats

	beamColor color.RGBA
	beamWidth float64
}

var turretCreepStats = &creepStats{
	kind:      creepTurret,
	image:     assets.ImageTurretCreep,
	speed:     0,
	maxHealth: 42,
	weapon: initWeaponStats(&weaponStats{
		MaxTargets:      1,
		BurstSize:       1,
		AttackSound:     assets.AudioMissile,
		AttackRange:     260,
		ImpactArea:      18,
		ProjectileSpeed: 360,
		Damage:          damageValue{health: 10},
		ProjectileImage: assets.ImageMissile,
		Explosion:       projectileExplosionNormal,
		Reload:          3.5,
		FireOffset:      gmath.Vec{Y: -8},
		TargetFlags:     targetFlying | targetGround,
	}),
	size: 38,
}

var baseCreepStats = &creepStats{
	kind:      creepBase,
	image:     assets.ImageCreepBase,
	speed:     0,
	maxHealth: 110,
	size:      60,
}

var wandererCreepStats = &creepStats{
	kind:        creepPrimitiveWanderer,
	image:       assets.ImagePrimitiveCreep,
	shadowImage: assets.ImageSmallShadow,
	tier:        1,
	speed:       40,
	maxHealth:   15,
	weapon: initWeaponStats(&weaponStats{
		MaxTargets:      1,
		BurstSize:       1,
		AttackSound:     assets.AudioWandererBeam,
		AttackRange:     190,
		ImpactArea:      10,
		ProjectileSpeed: 400,
		Damage:          damageValue{health: 4},
		ProjectileImage: assets.ImageWandererProjectile,
		Reload:          1.8,
		TargetFlags:     targetFlying | targetGround,
	}),
}

var tankCreepStats = &creepStats{
	kind:      creepTank,
	image:     assets.ImageTankCreep,
	speed:     6,
	maxHealth: 12,
	weapon: initWeaponStats(&weaponStats{
		MaxTargets:      1,
		BurstSize:       3,
		BurstDelay:      0.12,
		AttackSound:     assets.AudioTankShot,
		AttackRange:     110,
		ImpactArea:      10,
		ProjectileSpeed: 350,
		Damage:          damageValue{health: 1},
		ProjectileImage: assets.ImageTankProjectile,
		Reload:          2.2,
		FireOffset:      gmath.Vec{Y: -2},
		TargetFlags:     targetFlying | targetGround,
	}),
	size: 24,
}

var crawlerCreepStats = &creepStats{
	kind:      creepCrawler,
	image:     assets.ImageCrawlerCreep,
	animSpeed: 0.09,
	speed:     44,
	maxHealth: 16,
	weapon: initWeaponStats(&weaponStats{
		MaxTargets:      1,
		BurstSize:       2,
		BurstDelay:      0.12,
		AttackSound:     assets.AudioTankShot,
		AttackRange:     170,
		ImpactArea:      14,
		ProjectileSpeed: 350,
		Damage:          damageValue{health: 2},
		ProjectileImage: assets.ImageTankProjectile,
		Reload:          1.7,
		TargetFlags:     targetFlying | targetGround,
		FireOffset:      gmath.Vec{Y: -2},
	}),
	size: 24,
}

var eliteCrawlerCreepStats = &creepStats{
	kind:      creepCrawler,
	image:     assets.ImageEliteCrawlerCreep,
	animSpeed: 0.09,
	speed:     40,
	maxHealth: 22,
	weapon: initWeaponStats(&weaponStats{
		MaxTargets:      2,
		BurstSize:       1,
		AttackSound:     assets.AudioEliteCrawlerShot,
		AttackRange:     160,
		ImpactArea:      10,
		ProjectileSpeed: 320,
		Damage:          damageValue{health: 3},
		ProjectileImage: assets.ImageEliteCrawlerProjectile,
		Reload:          1.9,
		TargetFlags:     targetFlying | targetGround,
		FireOffset:      gmath.Vec{Y: -2},
	}),
	size: 24,
}

var assaultCreepStats = &creepStats{
	kind:        creepAssault,
	image:       assets.ImageCreepTier3,
	animSpeed:   0.2,
	shadowImage: assets.ImageBigShadow,
	tier:        3,
	speed:       30,
	maxHealth:   55,
	weapon: initWeaponStats(&weaponStats{
		MaxTargets:      1,
		BurstSize:       1,
		AttackSound:     assets.AudioAssaultShot,
		AttackRange:     150,
		ImpactArea:      10,
		ProjectileSpeed: 460,
		Damage:          damageValue{health: 2},
		ProjectileImage: assets.ImageAssaultProjectile,
		Reload:          0.55,
		TargetFlags:     targetFlying | targetGround,
	}),
}

var uberBossCreepStats = &creepStats{
	kind:        creepUberBoss,
	image:       assets.ImageUberBoss,
	animSpeed:   0.5,
	shadowImage: assets.ImageUberBossShadow,
	speed:       10,
	maxHealth:   500,
	weapon: initWeaponStats(&weaponStats{
		MaxTargets:  5,
		BurstSize:   1,
		AttackSound: assets.AudioRailgun,
		AttackRange: 220,
		Damage:      damageValue{health: 8},
		Reload:      2.8,
		TargetFlags: targetFlying | targetGround,
	}),
	beamColor: railgunBeamColor,
	beamWidth: 3,
}

var stunnerCreepStats = &creepStats{
	kind:        creepStunner,
	image:       assets.ImageCreepTier2,
	shadowImage: assets.ImageMediumShadow,
	tier:        2,
	speed:       55,
	maxHealth:   30,
	weapon: initWeaponStats(&weaponStats{
		MaxTargets:  3,
		BurstSize:   1,
		AttackSound: assets.AudioStunBeam,
		AttackRange: 230,
		Damage:      damageValue{health: 2, energy: 50},
		Reload:      2.6,
		TargetFlags: targetFlying | targetGround,
	}),
	beamColor: stunnerBeamColor,
	beamWidth: 2,
}
