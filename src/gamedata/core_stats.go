package gamedata

import (
	"fmt"

	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
)

type ColonyCoreStats struct {
	Name string

	Image               resource.ImageID
	Shadow              resource.ImageID
	ShadowOffsetY       float64
	HatchOffsetY        float64
	DiodeOffset         gmath.Vec
	AllianceColorOffset gmath.Vec

	SelectorImage      resource.ImageID
	AllianceColorImage resource.ImageID

	MobilityRating  int
	AttackRating    int
	DefenseRating   int
	CapacityRating  int
	UnitLimitRating int

	ScoreCost int

	FlightHeight  float64
	DefaultHeight float64

	Speed               float64
	JumpDist            float64
	DroneLimit          int
	StartingDrones      int
	DroneLimitScaling   float64
	MinDrones           int
	ResourcesLimit      float64
	MaxHealth           float64
	DroneProductionCost float64

	DamageReduction float64
}

func (stats *ColonyCoreStats) FlyingImageID() resource.ImageID {
	return resource.ImageID(stats.Image) + 1
}

func (stats *ColonyCoreStats) SelectorImageID() resource.ImageID {
	if stats.SelectorImage != 0 {
		return stats.SelectorImage
	}
	return resource.ImageID(stats.Image) + 2
}

func (stats *ColonyCoreStats) AllianceColorImageID() resource.ImageID {
	if stats.AllianceColorImage != 0 {
		return stats.SelectorImage
	}
	return resource.ImageID(stats.Image) + 3
}

func FindCoreByName(coreName string) *ColonyCoreStats {
	for _, stats := range CoreStatsList {
		if stats.Name == coreName {
			return stats
		}
	}
	panic(fmt.Sprintf("requested a non-existing core: %s", coreName))
}

var CoreStatsList = []*ColonyCoreStats{
	DenCoreStats,
	HiveCoreStats,
	ArkCoreStats,
	TankCoreStats,
}

var DenCoreStats = &ColonyCoreStats{
	Name:                "den",
	Image:               assets.ImageDenCore,
	Shadow:              assets.ImageDenShadow,
	ShadowOffsetY:       10,
	HatchOffsetY:        -22,
	DiodeOffset:         gmath.Vec{X: -16, Y: -27},
	AllianceColorOffset: gmath.Vec{Y: 27},

	FlightHeight:  50,
	DefaultHeight: 0,

	Speed:               22,
	JumpDist:            600,
	DroneLimit:          160,
	DroneLimitScaling:   1.15,
	StartingDrones:      15,
	ResourcesLimit:      550,
	DroneProductionCost: 1,
	MaxHealth:           140,

	MobilityRating:  5,
	AttackRating:    2,
	DefenseRating:   8,
	CapacityRating:  9,
	UnitLimitRating: 9,
}

var ArkCoreStats = &ColonyCoreStats{
	Name:                "ark",
	Image:               assets.ImageArkCore,
	Shadow:              assets.ImageArkShadow,
	ShadowOffsetY:       20,
	HatchOffsetY:        -20,
	DiodeOffset:         gmath.Vec{X: -16, Y: -29},
	AllianceColorOffset: gmath.Vec{X: 7, Y: 30},
	ScoreCost:           ArkCoreCost,

	FlightHeight:  30,
	DefaultHeight: 30,

	Speed:               32,
	JumpDist:            750,
	DroneLimit:          75,
	DroneLimitScaling:   0.95,
	StartingDrones:      10,
	ResourcesLimit:      250,
	DroneProductionCost: 1,
	MaxHealth:           85,

	MobilityRating:  10,
	AttackRating:    0,
	DefenseRating:   6,
	CapacityRating:  5,
	UnitLimitRating: 6,
}

var TankCoreWeapon1 = InitWeaponStats(&WeaponStats{
	AttackRangeMarkMultiplier: 1.5,
	AttackRange:               260,
	Reload:                    3.0,
	AttackSound:               assets.AudioTankColonyBlasterShot,
	ProjectileImage:           assets.ImageTankColonyProjectile1,
	ImpactArea:                15,
	ProjectileSpeed:           270,
	Damage:                    DamageValue{Health: 12},
	Explosion:                 ProjectileExplosionTankColonyBlaster,
	MaxTargets:                1,
	BurstSize:                 1,
	TargetFlags:               TargetFlying,
	TrailEffect:               ProjectileTrailTankColonyWeapon1,
	ProjectileFireSound:       true,
})

var TankCoreStats = &ColonyCoreStats{
	Name:                "tank",
	Image:               assets.ImageTankCore,
	HatchOffsetY:        -11,
	DiodeOffset:         gmath.Vec{X: -8, Y: -20},
	AllianceColorOffset: gmath.Vec{X: 17, Y: 36},
	ScoreCost:           TankCoreCost,

	DefaultHeight: 0,

	Speed:               32,
	JumpDist:            650,
	DroneLimit:          40,
	DroneLimitScaling:   0.7,
	StartingDrones:      5,
	ResourcesLimit:      200,
	DroneProductionCost: 1,
	MaxHealth:           75,

	MobilityRating:  6,
	AttackRating:    8,
	DefenseRating:   5,
	CapacityRating:  4,
	UnitLimitRating: 3,
}

var HiveMortarWeapon = InitWeaponStats(&WeaponStats{
	AttackRangeMarkMultiplier: 2.2,
	AttackRange:               300,
	Reload:                    3.3,
	AttackSound:               assets.AudioHiveMortarShot1,
	ProjectileFireSound:       true,
	ProjectileImage:           assets.ImageHiveMortarProjectile,
	AlwaysExplodes:            true,
	ImpactArea:                20,
	ProjectileSpeed:           240,
	Damage:                    DamageValue{Health: 8},
	BuildingDamageBonus:       -0.5,
	MaxTargets:                1,
	AttacksPerBurst:           1,
	BurstDelay:                0.25,
	BurstSize:                 4,
	FireOffsets:               []gmath.Vec{{X: -18, Y: 5}, {X: -18, Y: 5}, {X: 18, Y: 5}, {X: 18, Y: 5}},
	Explosion:                 ProjectileExplosionPurpleBurst,
	ArcPower:                  3.4,
	Accuracy:                  0.9,
	TargetFlags:               TargetGround,
	TrailEffect:               ProjectileHiveMortarTrailFire,
})

var HiveCoreStats = &ColonyCoreStats{
	Name:                "hive",
	Image:               assets.ImageHiveCore,
	SelectorImage:       assets.ImageHiveCoreSelector,
	AllianceColorImage:  assets.ImageHiveCoreAllianceColor,
	HatchOffsetY:        -22,
	DiodeOffset:         gmath.Vec{X: 9, Y: -28},
	AllianceColorOffset: gmath.Vec{Y: 27},

	DefaultHeight: 0,

	Speed:               0,
	JumpDist:            0,
	DroneLimit:          180,
	DroneLimitScaling:   1.2,
	StartingDrones:      20,
	ResourcesLimit:      600,
	DroneProductionCost: 0.75,
	MaxHealth:           200,
	DamageReduction:     0.65,

	MobilityRating:  0,
	AttackRating:    10,
	DefenseRating:   10,
	CapacityRating:  10,
	UnitLimitRating: 10,
}
