package staging

import (
	"image/color"

	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
	"github.com/quasilyte/roboden-game/gamedata"
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

	weapon *gamedata.WeaponStats

	beamColor color.RGBA
	beamWidth float64
}

var turretCreepStats = &creepStats{
	kind:      creepTurret,
	image:     assets.ImageTurretCreep,
	speed:     0,
	maxHealth: 42,
	weapon: initWeaponStats(&gamedata.WeaponStats{
		MaxTargets:      1,
		BurstSize:       1,
		AttackSound:     assets.AudioMissile,
		AttackRange:     260,
		ImpactArea:      18,
		ProjectileSpeed: 360,
		Damage:          gamedata.DamageValue{Health: 10},
		ProjectileImage: assets.ImageMissile,
		Explosion:       gamedata.ProjectileExplosionNormal,
		Reload:          3.5,
		FireOffset:      gmath.Vec{Y: -8},
		TargetFlags:     gamedata.TargetFlying | gamedata.TargetGround,
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
	maxHealth:   14,
	weapon: initWeaponStats(&gamedata.WeaponStats{
		MaxTargets:      1,
		BurstSize:       1,
		AttackSound:     assets.AudioWandererBeam,
		AttackRange:     190,
		ImpactArea:      10,
		ProjectileSpeed: 400,
		Damage:          gamedata.DamageValue{Health: 4},
		ProjectileImage: assets.ImageWandererProjectile,
		Reload:          2.2,
		TargetFlags:     gamedata.TargetFlying | gamedata.TargetGround,
	}),
}

var tankCreepStats = &creepStats{
	kind:      creepTank,
	image:     assets.ImageTankCreep,
	speed:     6,
	maxHealth: 12,
	weapon: initWeaponStats(&gamedata.WeaponStats{
		MaxTargets:      1,
		BurstSize:       3,
		BurstDelay:      0.12,
		AttackSound:     assets.AudioTankShot,
		AttackRange:     110,
		ImpactArea:      10,
		ProjectileSpeed: 350,
		Damage:          gamedata.DamageValue{Health: 1},
		ProjectileImage: assets.ImageTankProjectile,
		Reload:          2.2,
		FireOffset:      gmath.Vec{Y: -2},
		TargetFlags:     gamedata.TargetFlying | gamedata.TargetGround,
	}),
	size: 24,
}

var crawlerCreepStats = &creepStats{
	kind:      creepCrawler,
	image:     assets.ImageCrawlerCreep,
	animSpeed: 0.09,
	speed:     44,
	maxHealth: 16,
	weapon: initWeaponStats(&gamedata.WeaponStats{
		MaxTargets:      1,
		BurstSize:       2,
		BurstDelay:      0.12,
		AttackSound:     assets.AudioTankShot,
		AttackRange:     170,
		ImpactArea:      14,
		ProjectileSpeed: 350,
		Damage:          gamedata.DamageValue{Health: 2},
		ProjectileImage: assets.ImageTankProjectile,
		Reload:          1.7,
		TargetFlags:     gamedata.TargetFlying | gamedata.TargetGround,
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
	weapon: initWeaponStats(&gamedata.WeaponStats{
		MaxTargets:      2,
		BurstSize:       1,
		AttackSound:     assets.AudioEliteCrawlerShot,
		AttackRange:     160,
		ImpactArea:      10,
		ProjectileSpeed: 320,
		Damage:          gamedata.DamageValue{Health: 3},
		ProjectileImage: assets.ImageEliteCrawlerProjectile,
		Reload:          1.9,
		TargetFlags:     gamedata.TargetFlying | gamedata.TargetGround,
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
	weapon: initWeaponStats(&gamedata.WeaponStats{
		MaxTargets:      1,
		BurstSize:       1,
		AttackSound:     assets.AudioAssaultShot,
		AttackRange:     150,
		ImpactArea:      10,
		ProjectileSpeed: 460,
		Damage:          gamedata.DamageValue{Health: 2},
		ProjectileImage: assets.ImageAssaultProjectile,
		Reload:          0.55,
		TargetFlags:     gamedata.TargetFlying | gamedata.TargetGround,
	}),
}

var uberBossCreepStats = &creepStats{
	kind:        creepUberBoss,
	image:       assets.ImageUberBoss,
	animSpeed:   0.5,
	shadowImage: assets.ImageUberBossShadow,
	speed:       10,
	maxHealth:   500,
	weapon: initWeaponStats(&gamedata.WeaponStats{
		MaxTargets:  5,
		BurstSize:   1,
		AttackSound: assets.AudioRailgun,
		AttackRange: 220,
		Damage:      gamedata.DamageValue{Health: 8},
		Reload:      2.8,
		TargetFlags: gamedata.TargetFlying | gamedata.TargetGround,
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
	weapon: initWeaponStats(&gamedata.WeaponStats{
		MaxTargets:  3,
		BurstSize:   1,
		AttackSound: assets.AudioStunBeam,
		AttackRange: 230,
		Damage:      gamedata.DamageValue{Health: 2, Energy: 50},
		Reload:      2.6,
		TargetFlags: gamedata.TargetFlying | gamedata.TargetGround,
	}),
	beamColor: stunnerBeamColor,
	beamWidth: 2,
}
