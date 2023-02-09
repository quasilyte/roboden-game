package staging

import (
	"github.com/quasilyte/colony-game/assets"
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/gmath"
)

type creepStats struct {
	kind        creepKind
	image       resource.ImageID
	shadowImage resource.ImageID

	speed float64

	maxHealth float64

	maxTargets            int
	attackRange           float64
	projectileArea        float64
	projectileRotateSpeed float64
	projectileSpeed       float64
	projectileDamage      damageValue
	projectileImage       resource.ImageID
	projectileExplosion   projectileExplosionKind
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
	weaponReload:        3.5,
}

var baseCreepStats = &creepStats{
	kind:      creepBase,
	image:     assets.ImageCreepBase,
	speed:     0,
	maxHealth: 70,
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

var uberBossCreepStats = &creepStats{
	kind:             creepUberBoss,
	image:            assets.ImageUberBoss,
	shadowImage:      assets.ImageUberBossShadow,
	speed:            10,
	maxHealth:        300,
	maxTargets:       5,
	attackSound:      assets.AudioRailgun,
	attackRange:      250,
	projectileDamage: damageValue{health: 8},
	weaponReload:     2.2,
}

// var wandererStunnerCreepStats = &creepStats{
// 	kind:                  creepPrimitiveWandererStunner,
// 	image:                 assets.ImagePrimitiveCreep2,
// 	speed:                 45,
// 	maxHealth:             25,
// 	maxTargets:            2,
// 	attackSound:           assets.AudioStunBeam,
// 	attackRange:           230,
// 	projectileArea:        15,
// 	projectileSpeed:       340,
// 	projectileDamage:      damageValue{energy: 50},
// 	projectileImage:       assets.ImageWandererStunnerProjectile,
// 	projectileRotateSpeed: 20,
// }
