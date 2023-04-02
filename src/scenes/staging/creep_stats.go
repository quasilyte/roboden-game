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

	disarmable bool
}

var turretCreepStats = &creepStats{
	kind:      creepTurret,
	image:     assets.ImageTurretCreep,
	speed:     0,
	maxHealth: 100,
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
	size: 40,
}

var baseCreepStats = &creepStats{
	kind:      creepBase,
	image:     assets.ImageCreepBase,
	speed:     0,
	maxHealth: 150,
	size:      60,
}

var turretConstructionCreepStats = &creepStats{
	kind:      creepTurretConstruction,
	image:     assets.ImageTurretCreep,
	speed:     0,
	maxHealth: 35,
	size:      40,
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
	disarmable: true,
}

var servantCreepStats = &creepStats{
	kind:        creepServant,
	image:       assets.ImageServantCreep,
	shadowImage: assets.ImageMediumShadow,
	tier:        2,
	speed:       70,
	maxHealth:   55,
	animSpeed:   0.15,
	weapon: initWeaponStats(&gamedata.WeaponStats{
		MaxTargets:      1,
		BurstSize:       1,
		AttackSound:     assets.AudioServantShot,
		AttackRange:     240,
		ImpactArea:      10,
		ProjectileSpeed: 340,
		Damage:          gamedata.DamageValue{Health: 3, Energy: 20},
		ProjectileImage: assets.ImageServantProjectile,
		Reload:          3.2,
		TargetFlags:     gamedata.TargetFlying | gamedata.TargetGround,
	}),
	disarmable: true,
}

var tankCreepStats = &creepStats{
	kind:      creepTank,
	image:     assets.ImageTankCreep,
	speed:     6,
	maxHealth: 18,
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
	size:       24,
	disarmable: true,
}

var crawlerCreepStats = &creepStats{
	kind:      creepCrawler,
	image:     assets.ImageCrawlerCreep,
	animSpeed: 0.09,
	speed:     44,
	maxHealth: 18,
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
	size:       24,
	disarmable: true,
}

var eliteCrawlerCreepStats = &creepStats{
	kind:      creepCrawler,
	image:     assets.ImageEliteCrawlerCreep,
	animSpeed: 0.09,
	speed:     40,
	maxHealth: 28,
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
	size:       24,
	disarmable: true,
}

var heavyCrawlerCreepStats = &creepStats{
	kind:      creepCrawler,
	image:     assets.ImageHeavyCrawlerCreep,
	animSpeed: 0.15,
	speed:     30,
	maxHealth: 42,
	weapon: initWeaponStats(&gamedata.WeaponStats{
		MaxTargets:      1,
		BurstSize:       5,
		BurstDelay:      0.1,
		AttackSound:     assets.AudioHeavyCrawlerShot,
		AttackRange:     185,
		ImpactArea:      10,
		ProjectileSpeed: 280,
		Damage:          gamedata.DamageValue{Health: 2},
		ProjectileImage: assets.ImageHeavyCrawlerProjectile,
		Reload:          2.5,
		TargetFlags:     gamedata.TargetFlying | gamedata.TargetGround,
		FireOffset:      gmath.Vec{Y: -2},
		Explosion:       gamedata.ProjectileExplosionHeavyCrawlerLaser,
		ArcPower:        1.5,
	}),
	size:       24,
	disarmable: true,
}

var stealthCrawlerCreepStats = &creepStats{
	kind:      creepCrawler,
	image:     assets.ImageStealthCrawlerCreep,
	animSpeed: 0.09,
	speed:     50,
	maxHealth: 25,
	weapon: initWeaponStats(&gamedata.WeaponStats{
		MaxTargets:          1,
		BurstSize:           3,
		BurstDelay:          0.4,
		ProjectileFireSound: true,
		AttackSound:         assets.AudioStealthCrawlerShot,
		AttackRange:         200,
		ImpactArea:          14,
		ProjectileSpeed:     320,
		Damage:              gamedata.DamageValue{Health: 3, Slow: 2},
		ProjectileImage:     assets.ImageStealthCrawlerProjectile,
		Reload:              3.5,
		TargetFlags:         gamedata.TargetFlying | gamedata.TargetGround,
		FireOffset:          gmath.Vec{Y: -2},
		Explosion:           gamedata.ProjectileExplosionStealthLaser,
	}),
	size:       24,
	disarmable: true,
}

var assaultCreepStats = &creepStats{
	kind:        creepAssault,
	image:       assets.ImageCreepTier3,
	animSpeed:   0.2,
	shadowImage: assets.ImageBigShadow,
	tier:        3,
	speed:       30,
	maxHealth:   80,
	weapon: initWeaponStats(&gamedata.WeaponStats{
		MaxTargets:      1,
		BurstSize:       1,
		AttackSound:     assets.AudioAssaultShot,
		AttackRange:     150,
		ImpactArea:      10,
		ProjectileSpeed: 460,
		Damage:          gamedata.DamageValue{Health: 3},
		ProjectileImage: assets.ImageAssaultProjectile,
		Reload:          0.7,
		TargetFlags:     gamedata.TargetFlying | gamedata.TargetGround,
	}),
	disarmable: true,
}

var dominatorCreepStats = &creepStats{
	kind:        creepDominator,
	image:       assets.ImageCreepDominator,
	shadowImage: assets.ImageBigShadow,
	tier:        3,
	speed:       35,
	maxHealth:   200,
	weapon: initWeaponStats(&gamedata.WeaponStats{
		MaxTargets:  1,
		BurstSize:   1,
		AttackSound: assets.AudioDominatorShot,
		AttackRange: 265,
		Damage:      gamedata.DamageValue{Health: 8, Morale: 4},
		Reload:      1.65,
		TargetFlags: gamedata.TargetFlying | gamedata.TargetGround,
	}),
	beamColor:  dominatorBeamColorCenter,
	beamWidth:  1,
	disarmable: false,
}

var builderCreepStats = &creepStats{
	kind:        creepBuilder,
	image:       assets.ImageBuilderCreep,
	animSpeed:   0.1,
	shadowImage: assets.ImageBigShadow,
	tier:        3,
	speed:       40,
	maxHealth:   140,
	// disarmable: true,
}

var uberBossCreepStats = &creepStats{
	kind:        creepUberBoss,
	image:       assets.ImageUberBoss,
	animSpeed:   0.5,
	shadowImage: assets.ImageUberBossShadow,
	speed:       10,
	maxHealth:   600,
	weapon: initWeaponStats(&gamedata.WeaponStats{
		MaxTargets:  5,
		BurstSize:   1,
		AttackSound: assets.AudioRailgun,
		AttackRange: 220,
		Damage:      gamedata.DamageValue{Health: 9},
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
	maxHealth:   35,
	weapon: initWeaponStats(&gamedata.WeaponStats{
		MaxTargets:  3,
		BurstSize:   1,
		AttackSound: assets.AudioStunBeam,
		AttackRange: 230,
		Damage:      gamedata.DamageValue{Health: 2, Energy: 50},
		Reload:      2.6,
		TargetFlags: gamedata.TargetFlying | gamedata.TargetGround,
	}),
	beamColor:  stunnerBeamColor,
	beamWidth:  2,
	disarmable: true,
}
