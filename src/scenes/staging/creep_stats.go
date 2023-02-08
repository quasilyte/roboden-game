package staging

import (
	"github.com/quasilyte/colony-game/assets"
	resource "github.com/quasilyte/ebitengine-resource"
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
	weaponReload          float64
	attackSound           resource.AudioID
}

var wandererCreepStats = &creepStats{
	kind:             creepPrimitiveWanderer,
	image:            assets.ImagePrimitiveCreep,
	shadowImage:      assets.ImageSmallShadow,
	speed:            40,
	maxHealth:        15,
	maxTargets:       1,
	attackSound:      assets.AudioWandererBeam,
	attackRange:      200,
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
	attackRange:      240,
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
