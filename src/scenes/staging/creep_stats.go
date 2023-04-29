package staging

import (
	"image/color"

	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
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

	weapon        *gamedata.WeaponStats
	specialWeapon *gamedata.WeaponStats

	beamColor      color.RGBA
	beamWidth      float64
	beamSlideSpeed float64
	beamOpaqueTime float64
	beamTexture    *ge.Texture

	disarmable    bool
	canBeRepelled bool
}

var turretCreepStats = &creepStats{
	kind:      creepTurret,
	image:     assets.ImageTurretCreep,
	speed:     0,
	maxHealth: 120,
	weapon: gamedata.InitWeaponStats(&gamedata.WeaponStats{
		MaxTargets:      1,
		BurstSize:       1,
		AttackSound:     assets.AudioMissile,
		AttackRange:     290,
		ImpactArea:      18,
		ProjectileSpeed: 360,
		Damage:          gamedata.DamageValue{Health: 10},
		ProjectileImage: assets.ImageMissile,
		Explosion:       gamedata.ProjectileExplosionNormal,
		TrailEffect:     gamedata.ProjectileTrailSmoke,
		Reload:          3.5,
		FireOffset:      gmath.Vec{Y: -8},
		TargetFlags:     gamedata.TargetFlying | gamedata.TargetGround,
	}),
	size:          40,
	canBeRepelled: false,
	disarmable:    false,
}

var baseCreepStats = &creepStats{
	kind:          creepBase,
	image:         assets.ImageCreepBase,
	speed:         0,
	maxHealth:     170,
	size:          60,
	disarmable:    false,
	canBeRepelled: false,
}

var crawlerBaseCreepStats = &creepStats{
	kind:          creepCrawlerBase,
	image:         assets.ImageCrawlerCreepBase,
	speed:         0,
	maxHealth:     140,
	size:          60,
	disarmable:    false,
	canBeRepelled: false,
}

var crawlerBaseConstructionCreepStats = &creepStats{
	kind:          creepCrawlerBaseConstruction,
	image:         assets.ImageCrawlerCreepBase,
	speed:         0,
	maxHealth:     35,
	size:          40,
	disarmable:    false,
	canBeRepelled: false,
}

var turretConstructionCreepStats = &creepStats{
	kind:          creepTurretConstruction,
	image:         assets.ImageTurretCreep,
	speed:         0,
	maxHealth:     35,
	size:          40,
	disarmable:    false,
	canBeRepelled: false,
}

var wandererCreepStats = &creepStats{
	kind:        creepPrimitiveWanderer,
	image:       assets.ImagePrimitiveCreep,
	shadowImage: assets.ImageSmallShadow,
	tier:        1,
	speed:       40,
	maxHealth:   14,
	weapon: gamedata.InitWeaponStats(&gamedata.WeaponStats{
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
	disarmable:    true,
	canBeRepelled: true,
}

var servantCreepStats = &creepStats{
	kind:        creepServant,
	image:       assets.ImageServantCreep,
	shadowImage: assets.ImageMediumShadow,
	tier:        2,
	speed:       70,
	maxHealth:   65,
	animSpeed:   0.15,
	weapon: gamedata.InitWeaponStats(&gamedata.WeaponStats{
		MaxTargets:      1,
		BurstSize:       1,
		AttackSound:     assets.AudioServantShot,
		AttackRange:     240,
		ImpactArea:      10,
		ProjectileSpeed: 340,
		Damage:          gamedata.DamageValue{Health: 4, Energy: 20},
		ProjectileImage: assets.ImageServantProjectile,
		Reload:          3.2,
		TargetFlags:     gamedata.TargetFlying | gamedata.TargetGround,
	}),
	disarmable:    true,
	canBeRepelled: false,
}

var crawlerCreepStats = &creepStats{
	kind:      creepCrawler,
	image:     assets.ImageCrawlerCreep,
	animSpeed: 0.09,
	speed:     44,
	maxHealth: 18,
	weapon: gamedata.InitWeaponStats(&gamedata.WeaponStats{
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
	size:          24,
	disarmable:    true,
	canBeRepelled: false,
}

var eliteCrawlerCreepStats = &creepStats{
	kind:      creepCrawler,
	image:     assets.ImageEliteCrawlerCreep,
	animSpeed: 0.09,
	speed:     40,
	maxHealth: 28,
	weapon: gamedata.InitWeaponStats(&gamedata.WeaponStats{
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
	size:          24,
	disarmable:    true,
	canBeRepelled: false,
}

var heavyCrawlerCreepStats = &creepStats{
	kind:      creepCrawler,
	image:     assets.ImageHeavyCrawlerCreep,
	animSpeed: 0.15,
	speed:     30,
	maxHealth: 50,
	weapon: gamedata.InitWeaponStats(&gamedata.WeaponStats{
		MaxTargets:      1,
		BurstSize:       5,
		BurstDelay:      0.1,
		AttackSound:     assets.AudioHeavyCrawlerShot,
		AttackRange:     220,
		ImpactArea:      12,
		ProjectileSpeed: 280,
		Damage:          gamedata.DamageValue{Health: 2},
		ProjectileImage: assets.ImageHeavyCrawlerProjectile,
		Reload:          2.5,
		TargetFlags:     gamedata.TargetFlying | gamedata.TargetGround,
		FireOffset:      gmath.Vec{Y: -2},
		Explosion:       gamedata.ProjectileExplosionHeavyCrawlerLaser,
		ArcPower:        1.5,
		Accuracy:        0.85,
	}),
	size:          24,
	disarmable:    true,
	canBeRepelled: false,
}

var howitzerCreepStats = &creepStats{
	kind:      creepHowitzer,
	image:     assets.ImageHowitzerCreep,
	animSpeed: 0.2,
	speed:     10,
	maxHealth: 340,
	weapon: gamedata.InitWeaponStats(&gamedata.WeaponStats{
		MaxTargets:          3,
		BurstSize:           4,
		BurstDelay:          0.2,
		Accuracy:            0.7,
		AttackSound:         assets.AudioHowitzerLaserShot,
		AttackRange:         300,
		ImpactArea:          14,
		ProjectileSpeed:     480,
		Damage:              gamedata.DamageValue{Health: 2},
		ProjectileImage:     assets.ImageHowitzerLaserProjectile,
		Reload:              1.5,
		TargetFlags:         gamedata.TargetFlying,
		FireOffset:          gmath.Vec{Y: -2},
		ProjectileFireSound: true,
	}),
	specialWeapon: gamedata.InitWeaponStats(&gamedata.WeaponStats{
		MaxTargets:          1,
		BurstSize:           1,
		AttackSound:         assets.AudioHowitzerShot,
		AttackRange:         800,
		ImpactArea:          26,
		ProjectileSpeed:     150,
		Damage:              gamedata.DamageValue{Health: 20},
		ProjectileImage:     assets.ImageHowitzerProjectile,
		Reload:              10,
		TargetFlags:         gamedata.TargetGround,
		Explosion:           gamedata.ProjectileExplosionBigVertical,
		TrailEffect:         gamedata.ProjectileTrailSmoke,
		AlwaysExplodes:      true,
		ArcPower:            3,
		Accuracy:            0.4,
		ProjectileFireSound: true,
	}),
	size:          32,
	disarmable:    false,
	canBeRepelled: false,
}

var stealthCrawlerCreepStats = &creepStats{
	kind:      creepCrawler,
	image:     assets.ImageStealthCrawlerCreep,
	animSpeed: 0.09,
	speed:     50,
	maxHealth: 25,
	weapon: gamedata.InitWeaponStats(&gamedata.WeaponStats{
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
	size:          24,
	disarmable:    true,
	canBeRepelled: false,
}

var assaultCreepStats = &creepStats{
	kind:        creepAssault,
	image:       assets.ImageCreepTier3,
	animSpeed:   0.2,
	shadowImage: assets.ImageBigShadow,
	tier:        3,
	speed:       30,
	maxHealth:   80,
	weapon: gamedata.InitWeaponStats(&gamedata.WeaponStats{
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
	disarmable:    true,
	canBeRepelled: true,
}

var dominatorCreepStats = &creepStats{
	kind:        creepDominator,
	image:       assets.ImageCreepDominator,
	shadowImage: assets.ImageBigShadow,
	tier:        3,
	speed:       35,
	maxHealth:   200,
	weapon: gamedata.InitWeaponStats(&gamedata.WeaponStats{
		MaxTargets:  1,
		BurstSize:   1,
		AttackSound: assets.AudioDominatorShot,
		AttackRange: 265,
		Damage:      gamedata.DamageValue{Health: 8, Morale: 4},
		Reload:      1.65,
		TargetFlags: gamedata.TargetFlying | gamedata.TargetGround,
	}),
	beamColor:     dominatorBeamColorCenter,
	beamWidth:     1,
	disarmable:    false,
	canBeRepelled: false,
}

var builderCreepStats = &creepStats{
	kind:          creepBuilder,
	image:         assets.ImageBuilderCreep,
	animSpeed:     0.1,
	shadowImage:   assets.ImageBigShadow,
	tier:          3,
	speed:         40,
	maxHealth:     150,
	canBeRepelled: false,
	disarmable:    false,
}

var uberBossCreepStats = &creepStats{
	kind:        creepUberBoss,
	image:       assets.ImageUberBoss,
	animSpeed:   0.5,
	shadowImage: assets.ImageUberBossShadow,
	speed:       10,
	maxHealth:   600,
	weapon: gamedata.InitWeaponStats(&gamedata.WeaponStats{
		MaxTargets:  5,
		BurstSize:   1,
		AttackSound: assets.AudioRailgun,
		AttackRange: 220,
		Damage:      gamedata.DamageValue{Health: 9},
		Reload:      2.8,
		TargetFlags: gamedata.TargetFlying | gamedata.TargetGround,
	}),
	beamSlideSpeed: 2,
	disarmable:     false,
	canBeRepelled:  false,
}

var stunnerCreepStats = &creepStats{
	kind:        creepStunner,
	image:       assets.ImageCreepTier2,
	shadowImage: assets.ImageMediumShadow,
	tier:        2,
	speed:       55,
	maxHealth:   35,
	weapon: gamedata.InitWeaponStats(&gamedata.WeaponStats{
		MaxTargets:  3,
		BurstSize:   1,
		AttackSound: assets.AudioStunBeam,
		AttackRange: 230,
		Damage:      gamedata.DamageValue{Health: 2, Energy: 50},
		Reload:      2.6,
		TargetFlags: gamedata.TargetFlying | gamedata.TargetGround,
	}),
	beamSlideSpeed: 0.8,
	disarmable:     true,
	canBeRepelled:  true,
}
