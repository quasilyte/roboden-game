package staging

import (
	"image/color"

	"github.com/quasilyte/colony-game/assets"
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/gmath"
)

type creepStats struct {
	kind        creepKind
	image       resource.ImageID
	shadowImage resource.ImageID

	speed float64
	size  float64

	animSpeed float64

	maxHealth float64

	maxTargets            int
	attackRange           float64
	projectileArea        float64
	projectileRotateSpeed float64
	projectileSpeed       float64
	projectileDamage      damageValue
	projectileImage       resource.ImageID
	projectileExplosion   projectileExplosionKind
	beamColor             color.RGBA
	beamWidth             float64
	fireOffset            gmath.Vec
	weaponReload          float64
	attackSound           resource.AudioID
}

var turretCreepStats = &creepStats{
	kind:                creepTurret,
	image:               assets.ImageTurretCreep,
	speed:               0,
	maxHealth:           42,
	maxTargets:          1,
	attackSound:         assets.AudioMissile,
	attackRange:         260,
	projectileArea:      18,
	projectileSpeed:     360,
	projectileDamage:    damageValue{health: 10},
	projectileImage:     assets.ImageMissile,
	projectileExplosion: projectileExplosionNormal,
	fireOffset:          gmath.Vec{Y: -8},
	size:                38,
	weaponReload:        3.5,
}

var baseCreepStats = &creepStats{
	kind:      creepBase,
	image:     assets.ImageCreepBase,
	speed:     0,
	maxHealth: 110,
	size:      60,
}

var wandererCreepStats = &creepStats{
	kind:             creepPrimitiveWanderer,
	image:            assets.ImagePrimitiveCreep,
	shadowImage:      assets.ImageSmallShadow,
	speed:            40,
	maxHealth:        15,
	maxTargets:       1,
	attackSound:      assets.AudioWandererBeam,
	attackRange:      190,
	projectileArea:   10,
	projectileSpeed:  400,
	projectileDamage: damageValue{health: 5},
	projectileImage:  assets.ImageWandererProjectile,
	weaponReload:     1.8,
}

var tankCreepStats = &creepStats{
	kind:             creepTank,
	image:            assets.ImageLandCreep,
	speed:            6,
	maxHealth:        10,
	maxTargets:       1,
	attackSound:      assets.AudioTankShot,
	attackRange:      110,
	projectileArea:   10,
	projectileSpeed:  350,
	projectileDamage: damageValue{health: 3},
	projectileImage:  assets.ImageTankProjectile,
	fireOffset:       gmath.Vec{Y: -2},
	weaponReload:     2.2,
	size:             24,
}

var assaultCreepStats = &creepStats{
	kind:             creepAssault,
	image:            assets.ImageCreepTier3,
	animSpeed:        0.2,
	shadowImage:      assets.ImageBigShadow,
	speed:            30,
	maxHealth:        55,
	maxTargets:       1,
	attackSound:      assets.AudioAssaultShot,
	attackRange:      150,
	projectileArea:   10,
	projectileSpeed:  460,
	projectileDamage: damageValue{health: 2},
	projectileImage:  assets.ImageAssaultProjectile,
	weaponReload:     0.55,
}

var uberBossCreepStats = &creepStats{
	kind:             creepUberBoss,
	image:            assets.ImageUberBoss,
	shadowImage:      assets.ImageUberBossShadow,
	speed:            10,
	maxHealth:        500,
	maxTargets:       5,
	attackSound:      assets.AudioRailgun,
	attackRange:      220,
	projectileDamage: damageValue{health: 8},
	beamColor:        railgunBeamColor,
	beamWidth:        3,
	weaponReload:     2.8,
}

var stunnerCreepStats = &creepStats{
	kind:             creepStunner,
	image:            assets.ImageCreepTier2,
	shadowImage:      assets.ImageMediumShadow,
	speed:            55,
	maxHealth:        30,
	maxTargets:       3,
	attackSound:      assets.AudioStunBeam,
	attackRange:      230,
	projectileDamage: damageValue{health: 2, energy: 50},
	beamColor:        stunnerBeamColor,
	beamWidth:        2,
	weaponReload:     2.6,
}
