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

	MobilityRating  int
	DefenseRating   int
	CapacityRating  int
	UnitLimitRating int

	ScoreCost int

	FlightHeight  float64
	DefaultHeight float64

	Speed             float64
	JumpDist          float64
	DroneLimit        int
	StartingDrones    int
	DroneLimitScaling float64
	MinDrones         int
	ResourcesLimit    float64
	MaxHealth         float64
}

func (stats *ColonyCoreStats) FlyingImageID() resource.ImageID {
	return resource.ImageID(stats.Image) + 1
}

func (stats *ColonyCoreStats) SelectorImageID() resource.ImageID {
	return resource.ImageID(stats.Image) + 2
}

func (stats *ColonyCoreStats) AllianceColorImageID() resource.ImageID {
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
	ArkCoreStats,
	TankCoreStats,
}

var DenCoreStats = &ColonyCoreStats{
	Name:                "den",
	Image:               assets.ImageDenCore,
	Shadow:              assets.ImageDenShadow,
	ShadowOffsetY:       10,
	HatchOffsetY:        -22,
	DiodeOffset:         gmath.Vec{X: -16, Y: -29},
	AllianceColorOffset: gmath.Vec{Y: 27},

	FlightHeight:  50,
	DefaultHeight: 0,

	Speed:             18,
	JumpDist:          350,
	DroneLimit:        160,
	DroneLimitScaling: 1.1,
	StartingDrones:    10,
	ResourcesLimit:    500,
	MaxHealth:         140,

	MobilityRating:  5,
	DefenseRating:   10,
	CapacityRating:  10,
	UnitLimitRating: 10,
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

	Speed:             26,
	JumpDist:          600,
	DroneLimit:        80,
	DroneLimitScaling: 1.0,
	StartingDrones:    10,
	ResourcesLimit:    250,
	MaxHealth:         85,

	MobilityRating:  10,
	DefenseRating:   8,
	CapacityRating:  5,
	UnitLimitRating: 7,
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
	ProjectileFireSound:       true,
})

var TankCoreStats = &ColonyCoreStats{
	Name:                "tank",
	Image:               assets.ImageTankCore,
	HatchOffsetY:        -11,
	DiodeOffset:         gmath.Vec{X: -8, Y: -20},
	AllianceColorOffset: gmath.Vec{X: 7, Y: 30},
	ScoreCost:           TankCoreCost,

	DefaultHeight: 0,

	Speed:             24,
	JumpDist:          900,
	DroneLimit:        40,
	DroneLimitScaling: 0.7,
	StartingDrones:    5,
	ResourcesLimit:    200,
	MaxHealth:         75,

	MobilityRating:  4,
	DefenseRating:   7,
	CapacityRating:  4,
	UnitLimitRating: 4,
}
