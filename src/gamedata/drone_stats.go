package gamedata

import (
	"fmt"
	"image/color"
	"math"

	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
)

type UnitSize int

const (
	SizeSmall UnitSize = iota
	SizeMedium
	SizeLarge
)

//go:generate stringer -type=ColonyAgentKind -trimprefix=Agent
type ColonyAgentKind uint8

const (
	agentFirst ColonyAgentKind = iota

	AgentWorker
	AgentScout

	// Tier2
	AgentFreighter
	AgentRedminer
	AgentCrippler
	AgentFighter
	AgentScavenger
	AgentCourier
	AgentPrism
	AgentServo
	AgentRepeller
	AgentDisintegrator
	AgentRepair
	AgentCloner
	AgentRecharger
	AgentGenerator
	AgentMortar
	AgentAntiAir
	AgentDefender
	AgentKamikaze
	AgentSkirmisher
	AgentScarab
	AgentRoomba
	AgentCommander
	AgentTargeter
	AgentFirebug

	// Tier3
	AgentGuardian
	AgentStormbringer
	AgentDestroyer
	AgentBomber
	AgentMarauder
	AgentTrucker
	AgentDevourer

	AgentKindNum

	// Buildings (not real agents/drones)
	AgentGunpoint
	AgentTetherBeacon
	AgentBeamTower
	AgentHarvester
	AgentSiege

	// Neutral buildings
	AgentDroneFactory
	AgentPowerPlant
	AgentRepulseTower

	// Other units
	AgentRelict
	AgentMegaRoomba

	agentLast
)

var DroneKindByName = map[string]ColonyAgentKind{}

func init() {
	for k := ColonyAgentKind(agentFirst); k < agentLast; k++ {
		DroneKindByName[k.String()] = k
	}

	type topEntry struct {
		unit  *AgentStats
		score float64
	}
	type topScores struct {
		dpsTop     []topEntry
		rangeTop   []topEntry
		defenseTop []topEntry
		upkeepTop  []topEntry
	}

	findMax := func(list []topEntry) float64 {
		v := 0.0
		for _, x := range list {
			if x.score > v {
				v = x.score
			}
		}
		return v
	}
	findMin := func(list []topEntry) float64 {
		v := math.MaxFloat64
		for _, x := range list {
			if x.score < v {
				v = x.score
			}
		}
		return v
	}
	calcRatings := func(list []topEntry) []int {
		var ratings []int
		max := findMax(list)
		min := findMin(list)
		for _, x := range list {
			if x.score == 0 {
				ratings = append(ratings, 0)
				continue
			}
			if gmath.EqualApprox(x.score, min) {
				ratings = append(ratings, 1)
				continue
			}
			r := int(math.Ceil(10 * ((x.score - min) / (max - min))))
			ratings = append(ratings, r)
		}
		return ratings
	}

	recipesToStatsList := func(list []AgentMergeRecipe) []*AgentStats {
		var result []*AgentStats
		for _, x := range list {
			result = append(result, x.Result)
		}
		return result
	}

	collectTop := func(list []*AgentStats) topScores {
		var result topScores
		for _, stats := range list {
			if stats.Weapon != nil {
				damage := stats.Weapon.Damage.Health * float64(stats.Weapon.BurstSize)
				switch stats.Kind {
				case AgentPrism:
					damage += (PrismDamagePerReflection * PrismMaxReflections) + PrismDamagePerMax
				}
				dps := damage / stats.Weapon.Reload
				result.dpsTop = append(result.dpsTop, topEntry{unit: stats, score: dps})

				attackRange := stats.Weapon.AttackRange
				result.rangeTop = append(result.rangeTop, topEntry{unit: stats, score: attackRange})
			}

			defense := stats.MaxHealth + (stats.SelfRepair * 5)
			result.defenseTop = append(result.defenseTop, topEntry{unit: stats, score: defense})

			upkeep := stats.Upkeep
			result.upkeepTop = append(result.upkeepTop, topEntry{unit: stats, score: float64(upkeep)})
		}
		return result
	}

	getStats := func(stats *AgentStats, global bool) *DroneDocs {
		if global {
			return &stats.GlobalDocs
		}
		return &stats.Docs
	}
	fillDoc := func(top topScores, global bool) {
		for i, dpsRating := range calcRatings(top.dpsTop) {
			getStats(top.dpsTop[i].unit, global).DamageRating = dpsRating
		}
		for i, rangeRating := range calcRatings(top.rangeTop) {
			getStats(top.rangeTop[i].unit, global).AttackRangeRating = rangeRating
		}
		for i, defenseRating := range calcRatings(top.defenseTop) {
			getStats(top.defenseTop[i].unit, global).DefenseRating = defenseRating
		}
		for i, upkeepRating := range calcRatings(top.upkeepTop) {
			getStats(top.upkeepTop[i].unit, global).UpkeepRating = upkeepRating
		}
	}

	tier2top := collectTop(recipesToStatsList(Tier2agentMergeRecipes))
	fillDoc(tier2top, false)

	allStats := AllDroneStats()
	globalTop := collectTop(allStats)
	fillDoc(globalTop, true)
}

func AllDroneStats() []*AgentStats {
	var drones []*AgentStats
	drones = append(drones, WorkerAgentStats, ScoutAgentStats)
	for _, recipe := range Tier2agentMergeRecipes {
		drones = append(drones, recipe.Result)
	}
	for _, recipe := range Tier3agentMergeRecipes {
		drones = append(drones, recipe.Result)
	}
	return drones
}

func FindTurretByName(turretName string) *AgentStats {
	for _, stats := range TurretStatsList {
		if stats.Kind.String() == turretName {
			return stats
		}
	}
	panic(fmt.Sprintf("requested a non-existing turret: %s", turretName))
}

type AgentStats struct {
	Kind         ColonyAgentKind
	Image        resource.ImageID
	PreviewImage resource.ImageID
	AnimSpeed    float64
	Tier         int
	PointCost    int
	ScoreCost    int
	Upkeep       int

	Cost       float64
	PowerScore float64

	Size UnitSize

	Speed float64

	MaxHealth float64

	EnergyRegenRateBonus float64

	CanGather  bool
	CanPatrol  bool
	CanCloak   bool
	HasSupport bool
	IsFlying   bool
	IsTurret   bool
	IsBuilding bool
	IsNeutral  bool
	MaxPayload int

	SelfRepair float64

	DiodeOffset float64
	FireOffset  float64

	SupportReload   float64
	SupportRange    float64
	SupportRangeSqr float64

	Weapon *WeaponStats

	BeamWidth      float64
	BeamColor      color.RGBA
	BeamSlideSpeed float64
	BeamOpaqueTime float64
	BeamShift      float64
	BeamTexture    *ge.Texture
	BeamExplosion  resource.ImageID

	Docs       DroneDocs
	GlobalDocs DroneDocs
}

type DroneDocs struct {
	DamageRating      int
	AttackRangeRating int
	DefenseRating     int
	UpkeepRating      int
}

var WorkerAgentStats = InitDroneStats(&AgentStats{
	Kind:        AgentWorker,
	IsFlying:    true,
	Image:       assets.ImageWorkerAgent,
	Size:        SizeSmall,
	DiodeOffset: 5,
	Tier:        1,
	Cost:        8,
	Upkeep:      2,
	CanGather:   true,
	MaxPayload:  1,
	Speed:       80,
	MaxHealth:   12,
})

var TargeterAgentStats = InitDroneStats(&AgentStats{
	ScoreCost:   TargeterDronecost,
	Kind:        AgentTargeter,
	IsFlying:    true,
	Image:       assets.ImageTargeterAgent,
	Size:        SizeMedium,
	PointCost:   3,
	DiodeOffset: 4,
	Tier:        2,
	Cost:        25,
	PowerScore:  15,
	Upkeep:      13,
	CanPatrol:   true,
	Speed:       50,
	MaxHealth:   20,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange: 320,
		Reload:      3.4,
		EnergyCost:  1.5,
		AttackSound: assets.AudioTargeterShot,
		Damage:      DamageValue{Health: 2, Flags: DmgflagMark},
		MaxTargets:  1,
		BurstSize:   1,
		TargetFlags: TargetFlying | TargetGround,
	}),
	BeamOpaqueTime: 0.1,
	BeamSlideSpeed: 2.0,
	BeamExplosion:  assets.ImageTargeterShotExplosion,
})

var CommanderAgentStats = InitDroneStats(&AgentStats{
	ScoreCost:   CommanderDroneCost,
	Kind:        AgentCommander,
	IsFlying:    true,
	Image:       assets.ImageCommanderAgent,
	Size:        SizeMedium,
	PointCost:   2,
	DiodeOffset: 5,
	Tier:        2,
	Cost:        28,
	PowerScore:  10,
	Upkeep:      9,
	CanPatrol:   true,
	Speed:       50,
	MaxHealth:   30,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:               180,
		AttackRangeMarkMultiplier: 2,
		Reload:                    2.2,
		AttackSound:               assets.AudioCommanderShot,
		ProjectileImage:           assets.ImageCommanderProjectile,
		ImpactArea:                14,
		ProjectileSpeed:           240,
		Damage:                    DamageValue{Health: 3},
		MaxTargets:                1,
		BurstSize:                 1,
		TargetFlags:               TargetFlying | TargetGround,
		Explosion:                 ProjectileExplosionCommanderLaser,
	}),
})

var ScoutAgentStats = InitDroneStats(&AgentStats{
	Kind:        AgentScout,
	IsFlying:    true,
	Image:       assets.ImageScoutAgent,
	Size:        SizeSmall,
	DiodeOffset: 5,
	Tier:        1,
	Cost:        10,
	PowerScore:  8,
	Upkeep:      4,
	CanPatrol:   true,
	Speed:       75,
	MaxHealth:   12,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:               130,
		AttackRangeMarkMultiplier: 2,
		Reload:                    2.5,
		AttackSound:               assets.AudioScoutShot,
		ProjectileImage:           assets.ImageScoutProjectile,
		ImpactArea:                10,
		ProjectileSpeed:           180,
		Damage:                    DamageValue{Health: 2, Disarm: 0.2},
		MaxTargets:                1,
		BurstSize:                 1,
		TargetFlags:               TargetFlying | TargetGround,
		Explosion:                 ProjectileExplosionScoutIon,
		RoundProjectile:           true,
	}),
})

var TruckerAgentStats = InitDroneStats(&AgentStats{
	Kind:                 AgentTrucker,
	IsFlying:             true,
	Image:                assets.ImageTruckerAgent,
	Size:                 SizeLarge,
	DiodeOffset:          4,
	Tier:                 3,
	Cost:                 40,
	PowerScore:           15,
	Upkeep:               4,
	CanGather:            true,
	MaxPayload:           3,
	Speed:                85,
	MaxHealth:            45,
	EnergyRegenRateBonus: 0.5,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRangeMarkMultiplier: 1.75,
		AttackRange:               200,
		Reload:                    2.6,
		AttackSound:               assets.AudioCourierShot,
		ProjectileImage:           assets.ImageCourierProjectile,
		ImpactArea:                15,
		ProjectileSpeed:           170,
		Damage:                    DamageValue{Health: 2, Slow: 1, Morale: 0.2},
		MaxTargets:                2,
		BurstSize:                 1,
		ProjectileRotateSpeed:     24,
		TargetFlags:               TargetFlying,
	}),
})

var CourierAgentStats = InitDroneStats(&AgentStats{
	Kind:        AgentCourier,
	IsFlying:    true,
	Image:       assets.ImageCourierAgent,
	Size:        SizeMedium,
	DiodeOffset: 5,
	Tier:        2,
	Cost:        20,
	PowerScore:  9,
	PointCost:   2,
	Upkeep:      4,
	CanGather:   true,
	MaxPayload:  1,
	Speed:       80,
	MaxHealth:   30,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRangeMarkMultiplier: 1.75,
		AttackRange:               140,
		Reload:                    3.2,
		AttackSound:               assets.AudioCourierShot,
		ProjectileImage:           assets.ImageCourierProjectile,
		ImpactArea:                10,
		ProjectileSpeed:           170,
		Damage:                    DamageValue{Health: 2, Slow: 1, Morale: 0.2},
		MaxTargets:                1,
		BurstSize:                 1,
		ProjectileRotateSpeed:     24,
		TargetFlags:               TargetFlying,
	}),
	BeamSlideSpeed: 2.2,
})

var RedminerAgentStats = InitDroneStats(&AgentStats{
	Kind:                 AgentRedminer,
	IsFlying:             true,
	Image:                assets.ImageRedminerAgent,
	Size:                 SizeMedium,
	DiodeOffset:          6,
	Tier:                 2,
	PointCost:            2,
	Cost:                 16,
	Upkeep:               6,
	CanGather:            true,
	MaxPayload:           1,
	Speed:                75,
	MaxHealth:            20,
	EnergyRegenRateBonus: 0.2,
})

var GeneratorAgentStats = InitDroneStats(&AgentStats{
	Kind:                 AgentGenerator,
	IsFlying:             true,
	Image:                assets.ImageGeneratorAgent,
	Size:                 SizeMedium,
	DiodeOffset:          8,
	Tier:                 2,
	PointCost:            1,
	Cost:                 16,
	Upkeep:               2,
	CanGather:            true,
	MaxPayload:           1,
	Speed:                90,
	MaxHealth:            26,
	EnergyRegenRateBonus: 2,
})

var ClonerAgentStats = InitDroneStats(&AgentStats{
	Kind:        AgentCloner,
	IsFlying:    true,
	Image:       assets.ImageClonerAgent,
	Size:        SizeMedium,
	DiodeOffset: 5,
	Tier:        2,
	PointCost:   4,
	Cost:        26,
	Upkeep:      10,
	CanGather:   true,
	MaxPayload:  1,
	Speed:       90,
	MaxHealth:   16,
})

var RepairAgentStats = InitDroneStats(&AgentStats{
	Kind:           AgentRepair,
	IsFlying:       true,
	Image:          assets.ImageRepairAgent,
	Size:           SizeMedium,
	DiodeOffset:    5,
	FireOffset:     -2,
	Tier:           2,
	PointCost:      4,
	Cost:           26,
	PowerScore:     5,
	Upkeep:         18,
	CanGather:      true,
	HasSupport:     true,
	MaxPayload:     1,
	Speed:          100,
	MaxHealth:      18,
	SupportReload:  8.0,
	SupportRange:   250,
	BeamOpaqueTime: 0.2,
	BeamSlideSpeed: 0.6,
})

var RechargerAgentStats = InitDroneStats(&AgentStats{
	Kind:                 AgentRecharger,
	IsFlying:             true,
	Image:                assets.ImageRechargerAgent,
	Size:                 SizeMedium,
	DiodeOffset:          9,
	Tier:                 2,
	PointCost:            2,
	Cost:                 15,
	Upkeep:               6,
	CanGather:            true,
	HasSupport:           true,
	MaxPayload:           1,
	Speed:                90,
	MaxHealth:            16,
	EnergyRegenRateBonus: 0.2,
	SupportReload:        7,
	SupportRange:         350,
	BeamOpaqueTime:       0.2,
	BeamSlideSpeed:       0.8,
})

var GuardianAgentStats = InitDroneStats(&AgentStats{
	Kind:        AgentGuardian,
	IsFlying:    true,
	Image:       assets.ImageGuardianAgent,
	Size:        SizeLarge,
	DiodeOffset: -4,
	Tier:        3,
	Cost:        50,
	PowerScore:  35,
	Upkeep:      18,
	CanPatrol:   true,
	Speed:       55,
	MaxHealth:   50,
	SelfRepair:  0.75,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRangeMarkMultiplier: 1.5,
		AttackRange:               260,
		Reload:                    3.2,
		EnergyCost:                2,
		AttackSound:               assets.AudioDefenderShot,
		Damage:                    DamageValue{Health: 3, Flags: DmgflagAggro},
		MaxTargets:                2,
		BurstSize:                 1,
		TargetFlags:               TargetFlying | TargetGround,
	}),
	BeamOpaqueTime: 0.1,
	BeamSlideSpeed: -1.6,
})

var ServoAgentStats = InitDroneStats(&AgentStats{
	Kind:          AgentServo,
	IsFlying:      true,
	Image:         assets.ImageServoAgent,
	Size:          SizeMedium,
	DiodeOffset:   -4,
	Tier:          2,
	PointCost:     2,
	Cost:          26,
	Upkeep:        7,
	CanGather:     true,
	MaxPayload:    1,
	Speed:         125,
	MaxHealth:     18,
	SupportReload: 8,
	SupportRange:  310,
})

var FreighterAgentStats = InitDroneStats(&AgentStats{
	Kind:                 AgentFreighter,
	IsFlying:             true,
	Image:                assets.ImageFreighterAgent,
	Size:                 SizeMedium,
	DiodeOffset:          1,
	Tier:                 2,
	PointCost:            1,
	Cost:                 18,
	Upkeep:               0,
	CanGather:            true,
	MaxPayload:           3,
	Speed:                70,
	MaxHealth:            28,
	EnergyRegenRateBonus: 0.5,
})

var CripplerAgentStats = InitDroneStats(&AgentStats{
	Kind:        AgentCrippler,
	IsFlying:    true,
	Image:       assets.ImageCripplerAgent,
	Size:        SizeMedium,
	DiodeOffset: 5,
	Tier:        2,
	PointCost:   2,
	Cost:        16,
	PowerScore:  9,
	Upkeep:      4,
	CanPatrol:   true,
	CanCloak:    true,
	Speed:       65,
	MaxHealth:   18,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRangeMarkMultiplier: 1.5,
		AttackRange:               255,
		Reload:                    2.7,
		AttackSound:               assets.AudioCripplerShot,
		ProjectileImage:           assets.ImageCripplerProjectile,
		ImpactArea:                10,
		ProjectileSpeed:           250,
		Damage:                    DamageValue{Health: 1, Slow: 2},
		MaxTargets:                6,
		BurstSize:                 1,
		TargetFlags:               TargetFlying | TargetGround,
		Explosion:                 ProjectileExplosionCripplerBlaster,
	}),
})

var StormbringerAgentStats = InitDroneStats(&AgentStats{
	Kind:        AgentStormbringer,
	IsFlying:    true,
	Image:       assets.ImageStormbringerAgent,
	Size:        SizeLarge,
	DiodeOffset: 7,
	Tier:        3,
	Cost:        50,
	PowerScore:  40,
	Upkeep:      18,
	CanPatrol:   true,
	CanGather:   true,
	Speed:       100,
	MaxHealth:   40,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRangeMarkMultiplier: 1.5,
		AttackRange:               170,
		Reload:                    2.6,
		EnergyCost:                2,
		AttackSound:               assets.AudioStormbringerShot,
		ProjectileImage:           assets.ImageStormbringerProjectile,
		ImpactArea:                18,
		ProjectileSpeed:           200,
		ProjectileRotateSpeed:     4,
		Damage:                    DamageValue{Health: 1, Disarm: 0.2},
		MaxTargets:                2,
		BurstSize:                 4,
		BurstDelay:                0.03,
		TargetFlags:               TargetFlying | TargetGround,
		Explosion:                 ProjectileExplosionShocker,
	}),
})

const (
	PrismMaxReflections      = 4
	PrismDamagePerReflection = 2
	PrismDamagePerMax        = 2
)

var PrismAgentStats = InitDroneStats(&AgentStats{
	ScoreCost:   PrismDroneCost,
	IsFlying:    true,
	Kind:        AgentPrism,
	Image:       assets.ImagePrismAgent,
	Size:        SizeMedium,
	DiodeOffset: 1,
	Tier:        2,
	PointCost:   3,
	Cost:        26,
	PowerScore:  26,
	Upkeep:      12,
	CanPatrol:   true,
	Speed:       70,
	MaxHealth:   34,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRangeMarkMultiplier: 1.25,
		AttackRange:               230,
		Reload:                    3.7,
		EnergyCost:                3,
		AttackSound:               assets.AudioPrismShot,
		ImpactArea:                8,
		ProjectileSpeed:           220,
		Damage:                    DamageValue{Health: 4},
		MaxTargets:                1,
		BurstSize:                 1,
		TargetFlags:               TargetFlying | TargetGround,
		BuildingDamageBonus:       -0.25,
	}),
	BeamExplosion: assets.ImagePrismShotExplosion,
})

var RoombaAgentStats = InitDroneStats(&AgentStats{
	ScoreCost:   RoombaDroneCost,
	Kind:        AgentRoomba,
	IsFlying:    false,
	Image:       assets.ImageRoombaAgent,
	Size:        SizeMedium,
	DiodeOffset: 2,
	Tier:        2,
	PointCost:   2,
	Cost:        20,
	PowerScore:  25,
	Upkeep:      9,
	Speed:       40,
	MaxHealth:   55,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:         200,
		Reload:              2,
		AttackSound:         assets.AudioRoombaShot,
		ProjectileImage:     assets.ImageRoombaProjectile,
		ImpactArea:          10,
		ProjectileSpeed:     400,
		Damage:              DamageValue{Health: 4},
		MaxTargets:          1,
		BurstSize:           2,
		BurstDelay:          0.65,
		TargetFlags:         TargetFlying | TargetGround,
		Accuracy:            0.8,
		ProjectileFireSound: true,
		Explosion:           ProjectileExplosionRoombaShot,
		TrailEffect:         ProjectileTrailRoomba,
	}),
})

var FighterAgentStats = InitDroneStats(&AgentStats{
	Kind:        AgentFighter,
	IsFlying:    true,
	Image:       assets.ImageFighterAgent,
	Size:        SizeMedium,
	DiodeOffset: 1,
	Tier:        2,
	PointCost:   4,
	Cost:        22,
	PowerScore:  22,
	Upkeep:      9,
	CanPatrol:   true,
	Speed:       90,
	MaxHealth:   28,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRangeMarkMultiplier: 2,
		AttackRange:               195,
		Reload:                    1.9,
		EnergyCost:                1.5,
		AttackSound:               assets.AudioFighterBeam,
		ProjectileImage:           assets.ImageFighterProjectile,
		ImpactArea:                10,
		ProjectileSpeed:           250,
		Damage:                    DamageValue{Health: 5},
		Explosion:                 ProjectileExplosionFighterLaser,
		MaxTargets:                1,
		BurstSize:                 1,
		TargetFlags:               TargetFlying | TargetGround,
		BuildingDamageBonus:       -0.4,
	}),
})

var SkirmisherAgentStats = InitDroneStats(&AgentStats{
	ScoreCost:   SkirmisherDroneCost,
	IsFlying:    true,
	Kind:        AgentSkirmisher,
	Image:       assets.ImageSkirmisherAgent,
	Size:        SizeMedium,
	DiodeOffset: 3,
	Tier:        2,
	PointCost:   2,
	Cost:        25,
	PowerScore:  20,
	Upkeep:      8,
	CanPatrol:   true,
	Speed:       80,
	MaxHealth:   22,
	SelfRepair:  0.5,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRangeMarkMultiplier: 1.5,
		AttackRange:               160,
		Reload:                    2,
		EnergyCost:                1,
		AttackSound:               assets.AudioSkirmisherShot,
		ProjectileImage:           assets.ImageSkirmisherProjectile,
		ImpactArea:                15,
		ProjectileSpeed:           340,
		Damage:                    DamageValue{Health: 2},
		Explosion:                 ProjectileExplosionGreenZap,
		MaxTargets:                1,
		BurstSize:                 4,
		AttacksPerBurst:           4,
		BurstDelay:                0.3,
		TargetFlags:               TargetFlying | TargetGround,
		FlyingDamageBonus:         -0.5,
		ArcPower:                  1.2,
		RandArc:                   true,
	}),
})

var DefenderAgentStats = InitDroneStats(&AgentStats{
	ScoreCost:   DefenderDroneCost,
	IsFlying:    true,
	Kind:        AgentDefender,
	Image:       assets.ImageDefenderAgent,
	Size:        SizeMedium,
	DiodeOffset: 6,
	Tier:        2,
	PointCost:   2,
	Cost:        20,
	PowerScore:  12,
	Upkeep:      4,
	CanPatrol:   true,
	Speed:       55,
	MaxHealth:   35,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRangeMarkMultiplier: 1.25,
		AttackRange:               240,
		Reload:                    3.5,
		EnergyCost:                2,
		AttackSound:               assets.AudioDefenderShot,
		Damage:                    DamageValue{Health: 3, Flags: DmgflagAggro},
		MaxTargets:                2,
		BurstSize:                 1,
		TargetFlags:               TargetFlying | TargetGround,
	}),
	BeamOpaqueTime: 0.1,
	BeamSlideSpeed: -1.6,
})

var FirebugAgentStats = InitDroneStats(&AgentStats{
	ScoreCost:   FirebugDroneCost,
	IsFlying:    true,
	Kind:        AgentFirebug,
	Image:       assets.ImageFirebugAgent,
	AnimSpeed:   0.1,
	Size:        SizeMedium,
	DiodeOffset: -1,
	Tier:        2,
	PointCost:   2,
	Cost:        20,
	PowerScore:  16,
	Upkeep:      6,
	CanPatrol:   true,
	Speed:       85,
	MaxHealth:   30,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange: 55,
		Reload:      2.4,
		AttackSound: assets.AudioFirebugShot,
		Damage:      DamageValue{Health: 13},
		MaxTargets:  1,
		BurstSize:   1,
		TargetFlags: TargetFlying | TargetGround,
	}),
	BeamOpaqueTime: 0.15,
	BeamSlideSpeed: 2.2,
})

var ScavengerAgentStats = InitDroneStats(&AgentStats{
	Kind:          AgentScavenger,
	IsFlying:      true,
	Image:         assets.ImageScavengerAgent,
	Size:          SizeMedium,
	DiodeOffset:   -5,
	Tier:          2,
	PointCost:     2,
	Cost:          18,
	PowerScore:    14,
	Upkeep:        6,
	CanPatrol:     true,
	HasSupport:    true,
	Speed:         100,
	MaxHealth:     22,
	SupportReload: 16,
	MaxPayload:    2,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRangeMarkMultiplier: 1.5,
		AttackRange:               160,
		Reload:                    2.5,
		AttackSound:               assets.AudioScavengerShot,
		ProjectileImage:           assets.ImageScavengerProjectile,
		ImpactArea:                8,
		ProjectileSpeed:           250,
		Damage:                    DamageValue{Health: 2},
		MaxTargets:                1,
		BurstSize:                 2,
		BurstDelay:                0.12,
		Accuracy:                  0.9,
		TargetFlags:               TargetFlying | TargetGround,
	}),
})

var ScarabAgentStats = InitDroneStats(&AgentStats{
	ScoreCost:   ScarabDroneCost,
	IsFlying:    true,
	Kind:        AgentScarab,
	Image:       assets.ImageScarabAgent,
	Size:        SizeMedium,
	DiodeOffset: -5,
	Tier:        2,
	PointCost:   3,
	Cost:        20,
	PowerScore:  9,
	Upkeep:      8,
	CanGather:   true,
	CanPatrol:   true,
	Speed:       65,
	MaxHealth:   14,
	MaxPayload:  1,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRangeMarkMultiplier: 1.5,
		AttackRange:               150,
		Reload:                    2.5,
		AttackSound:               assets.AudioScarabShot,
		ProjectileImage:           assets.ImageScarabProjectile,
		ImpactArea:                8,
		ProjectileSpeed:           350,
		Damage:                    DamageValue{Health: 1.5},
		MaxTargets:                1,
		AttacksPerBurst:           2,
		ArcPower:                  1,
		RandArc:                   true,
		RoundProjectile:           true,
		BurstSize:                 2,
		Accuracy:                  0.95,
		TargetFlags:               TargetGround,
		Explosion:                 ProjectileExplosionScarab,
		BuildingDamageBonus:       0.25,
	}),
})

const (
	// +1 burst size per level (+7)
	// +5 max hp per level (+35)
	DevourerMaxLevel = 7
)

var DevourerAgentStats = InitDroneStats(&AgentStats{
	Kind:          AgentDevourer,
	IsFlying:      true,
	Image:         assets.ImageDevourerAgent,
	Size:          SizeLarge,
	DiodeOffset:   7,
	Tier:          3,
	Cost:          60,
	PowerScore:    55,
	Upkeep:        20,
	CanPatrol:     true,
	HasSupport:    true,
	Speed:         75,
	MaxHealth:     35,
	MaxPayload:    1,
	SupportReload: 25,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRangeMarkMultiplier: 1.5,
		AttackRange:               200,
		Reload:                    2.0,
		AttackSound:               assets.AudioScarabShot,
		ProjectileImage:           assets.ImageScarabProjectile,
		ProjectileFireSound:       true,
		ImpactArea:                10,
		ProjectileSpeed:           350,
		Damage:                    DamageValue{Health: 2},
		Accuracy:                  0.95,
		MaxTargets:                1,
		AttacksPerBurst:           3,
		ArcPower:                  1,
		RandArc:                   true,
		RoundProjectile:           true,
		BurstDelay:                0.25,
		BurstSize:                 3,
		TargetFlags:               TargetFlying | TargetGround,
		Explosion:                 ProjectileExplosionScarab,
		BuildingDamageBonus:       0.25,
	}),
})

var AntiAirAgentStats = InitDroneStats(&AgentStats{
	ScoreCost:   AntiAirDroneCost,
	IsFlying:    true,
	Kind:        AgentAntiAir,
	Image:       assets.ImageAntiAirAgent,
	Size:        SizeMedium,
	DiodeOffset: 1,
	Tier:        2,
	PointCost:   3,
	Cost:        28,
	PowerScore:  26,
	Upkeep:      11,
	CanPatrol:   true,
	Speed:       80,
	MaxHealth:   18,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRangeMarkMultiplier: 1.5,
		AttackRange:               270,
		Reload:                    2.4,
		AttackSound:               assets.AudioAntiAirMissiles,
		ProjectileImage:           assets.ImageAntiAirMissile,
		ImpactArea:                18,
		ProjectileSpeed:           250,
		Damage:                    DamageValue{Health: 2},
		MaxTargets:                1,
		BurstSize:                 4,
		BurstDelay:                0.1,
		Explosion:                 ProjectileExplosionNormal,
		TrailEffect:               ProjectileTrailSmoke,
		ArcPower:                  2,
		Accuracy:                  0.9,
		TargetFlags:               TargetFlying,
		FireOffsets:               []gmath.Vec{{Y: -8}},
	}),
})

var MortarAgentStats = InitDroneStats(&AgentStats{
	ScoreCost:   MortarDroneCost,
	IsFlying:    true,
	Kind:        AgentMortar,
	Image:       assets.ImageMortarAgent,
	Size:        SizeMedium,
	DiodeOffset: 1,
	Tier:        2,
	PointCost:   2,
	Cost:        18,
	PowerScore:  14,
	Upkeep:      7,
	CanPatrol:   true,
	Speed:       70,
	MaxHealth:   30,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRangeMarkMultiplier: 1.75,
		AttackRange:               370,
		Reload:                    3.6,
		AttackSound:               assets.AudioMortarShot,
		ProjectileImage:           assets.ImageMortarProjectile,
		ImpactArea:                18,
		ProjectileSpeed:           180,
		Damage:                    DamageValue{Health: 13},
		MaxTargets:                1,
		BurstSize:                 1,
		Explosion:                 ProjectileExplosionNormal,
		ArcPower:                  2.5,
		Accuracy:                  0.9,
		TargetFlags:               TargetGround,
		RoundProjectile:           true,
	}),
})

var DestroyerAgentStats = InitDroneStats(&AgentStats{
	Kind:        AgentDestroyer,
	IsFlying:    true,
	Image:       assets.ImageDestroyerAgent,
	Size:        SizeLarge,
	DiodeOffset: 0,
	Tier:        3,
	Cost:        60,
	PowerScore:  50,
	Upkeep:      22,
	CanPatrol:   true,
	Speed:       85,
	MaxHealth:   45,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRangeMarkMultiplier: 1.25,
		AttackRange:               220,
		Reload:                    1.7,
		EnergyCost:                5.0,
		AttackSound:               assets.AudioDestroyerBeam,
		Damage:                    DamageValue{Health: 7},
		MaxTargets:                1,
		BurstSize:                 1,
		TargetFlags:               TargetFlying | TargetGround,
		BuildingDamageBonus:       -0.5,
	}),
})

var BomberAgentStats = InitDroneStats(&AgentStats{
	Kind:        AgentBomber,
	IsFlying:    true,
	Image:       assets.ImageBomberAgent,
	Size:        SizeLarge,
	DiodeOffset: 6,
	Tier:        3,
	Cost:        50,
	PowerScore:  20,
	Upkeep:      14,
	CanPatrol:   true,
	Speed:       65,
	MaxHealth:   70,
})

var MarauderAgentStats = InitDroneStats(&AgentStats{
	Kind:          AgentMarauder,
	IsFlying:      true,
	Image:         assets.ImageMarauderAgent,
	Size:          SizeLarge,
	DiodeOffset:   0,
	Tier:          3,
	Cost:          55,
	PowerScore:    30,
	Upkeep:        18,
	CanPatrol:     true,
	CanCloak:      true,
	HasSupport:    true,
	Speed:         100,
	SupportReload: 14,
	MaxHealth:     30,
	MaxPayload:    3,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRangeMarkMultiplier: 1.25,
		AttackRange:               255,
		Reload:                    2.45,
		ProjectileImage:           assets.ImageMarauderProjectile,
		ImpactArea:                16,
		ProjectileSpeed:           300,
		AttackSound:               assets.AudioMarauderShot,
		Damage:                    DamageValue{Health: 3.5, Slow: 2},
		BurstSize:                 1,
		MaxTargets:                3,
		Accuracy:                  0.9,
		TargetFlags:               TargetFlying | TargetGround,
		Explosion:                 ProjectileExplosionCripplerBlaster,
	}),
})

var RepellerAgentStats = InitDroneStats(&AgentStats{
	Kind:        AgentRepeller,
	IsFlying:    true,
	Image:       assets.ImageRepellerAgent,
	Size:        SizeMedium,
	DiodeOffset: 3,
	Tier:        2,
	PointCost:   3,
	Cost:        28,
	PowerScore:  14,
	Upkeep:      9,
	CanGather:   true,
	MaxPayload:  1,
	CanPatrol:   true,
	Speed:       105,
	MaxHealth:   24,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRangeMarkMultiplier: 1.25,
		AttackRange:               170,
		Reload:                    2.2,
		EnergyCost:                1,
		AttackSound:               assets.AudioRepellerBeam,
		ProjectileImage:           assets.ImageRepellerProjectile,
		ImpactArea:                10,
		ProjectileSpeed:           200,
		Damage:                    DamageValue{Health: 2, Disarm: 0.5},
		MaxTargets:                2,
		BurstSize:                 1,
		TargetFlags:               TargetFlying | TargetGround,
		Explosion:                 ProjectileExplosionShocker,
	}),
})

var KamikazeAgentStats = InitDroneStats(&AgentStats{
	ScoreCost:   KamikazeDroneCost,
	IsFlying:    true,
	Kind:        AgentKamikaze,
	Image:       assets.ImageKamikazeAgent,
	Size:        SizeMedium,
	DiodeOffset: 4,
	Tier:        2,
	PointCost:   1,
	Cost:        14,
	PowerScore:  12,
	Upkeep:      3,
	CanGather:   true,
	MaxPayload:  1,
	CanPatrol:   true,
	Speed:       100,
	MaxHealth:   20,
})

var DisintegratorAgentStats = InitDroneStats(&AgentStats{
	ScoreCost:     DisintegratorDroneCost,
	IsFlying:      true,
	Kind:          AgentDisintegrator,
	Image:         assets.ImageDisintegratorAgent,
	Size:          SizeMedium,
	DiodeOffset:   4,
	Tier:          2,
	PointCost:     3,
	Cost:          18,
	PowerScore:    16,
	Upkeep:        7,
	CanGather:     true,
	HasSupport:    true,
	MaxPayload:    1,
	CanPatrol:     true,
	Speed:         80,
	MaxHealth:     20,
	SupportReload: 8,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRangeMarkMultiplier: 1.25,
		AttackRange:               220,
		AttackSound:               assets.AudioDisintegratorShot,
		ProjectileImage:           assets.ImageDisintegratorProjectile,
		ImpactArea:                18,
		ProjectileSpeed:           210,
		ProjectileRotateSpeed:     26,
		Reload:                    10, // Approx, for balance calculations
		Damage:                    DamageValue{Health: 16},
		MaxTargets:                1,
		BurstSize:                 1,
		Explosion:                 ProjectileExplosionPurple,
		TargetFlags:               TargetFlying | TargetGround,
		GroundDamageBonus:         -0.5,
	}),
})
