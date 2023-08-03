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
	AllianceColorOffset gmath.Vec

	MobilityRating  int
	DefenseRating   int
	CapacityRating  int
	UnitLimitRating int

	ScoreCost int

	FlightHeight  float64
	DefaultHeight float64

	Speed          float64
	JumpDist       float64
	DroneLimit     int
	ResourcesLimit float64
	MaxHealth      float64
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
}

var DenCoreStats = &ColonyCoreStats{
	Name:                "den",
	Image:               assets.ImageDenCore,
	Shadow:              assets.ImageDenShadow,
	ShadowOffsetY:       10,
	HatchOffsetY:        -22,
	AllianceColorOffset: gmath.Vec{Y: 27},

	FlightHeight:  50,
	DefaultHeight: 0,

	Speed:          18,
	JumpDist:       350,
	DroneLimit:     160,
	ResourcesLimit: 500,
	MaxHealth:      120,

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
	AllianceColorOffset: gmath.Vec{X: 7, Y: 30},
	ScoreCost:           ArkCoreCost,

	FlightHeight:  30,
	DefaultHeight: 30,

	Speed:          26,
	JumpDist:       600,
	DroneLimit:     80,
	ResourcesLimit: 250,
	MaxHealth:      80,

	MobilityRating:  10,
	DefenseRating:   8,
	CapacityRating:  5,
	UnitLimitRating: 7,
}
