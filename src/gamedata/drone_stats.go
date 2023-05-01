package gamedata

import (
	"fmt"
	"image/color"

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

	// Tier3
	AgentGuardian
	AgentStormbringer
	AgentDestroyer
	AgentMarauder
	AgentTrucker
	AgentDevourer

	AgentKindNum

	// Buildings (not real agents/drones)
	AgentGunpoint
	AgentTetherBeacon
	AgentBeamTower

	agentLast
)

var DroneKindByName = map[string]ColonyAgentKind{}

func init() {
	for k := ColonyAgentKind(agentFirst); k < agentLast; k++ {
		DroneKindByName[k.String()] = k
	}
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
	Kind      ColonyAgentKind
	Image     resource.ImageID
	Tier      int
	PointCost int
	ScoreCost int
	Cost      float64
	Upkeep    int

	Size UnitSize

	Speed float64

	MaxHealth float64

	EnergyRegenRateBonus float64

	CanGather  bool
	CanPatrol  bool
	CanCloak   bool
	MaxPayload int

	SelfRepair float64

	DiodeOffset float64
	FireOffset  float64

	SupportReload float64
	SupportRange  float64

	Weapon *WeaponStats

	BeamWidth      float64
	BeamColor      color.RGBA
	BeamSlideSpeed float64
	BeamOpaqueTime float64
	BeamTexture    *ge.Texture
}

var TurretStatsList = []*AgentStats{
	GunpointAgentStats,
	TetherBeaconAgentStats,
	BeamTowerAgentStats,
}

var GunpointAgentStats = &AgentStats{
	Kind:      AgentGunpoint,
	Image:     assets.ImageGunpointAgent,
	Size:      SizeLarge,
	Upkeep:    12,
	MaxHealth: 100,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:     280,
		Reload:          2,
		AttackSound:     assets.AudioGunpointShot,
		ProjectileImage: assets.ImageGunpointProjectile,
		ImpactArea:      10,
		ProjectileSpeed: 280,
		Damage:          DamageValue{Health: 4},
		MaxTargets:      1,
		BurstSize:       3,
		BurstDelay:      0.1,
		TargetFlags:     TargetGround,
		FireOffset:      gmath.Vec{Y: 6},
	}),
}

var BeamTowerAgentStats = &AgentStats{
	ScoreCost: BeamTowerTurretCost,
	Kind:      AgentBeamTower,
	Image:     assets.ImageBeamtowerAgent,
	Size:      SizeLarge,
	Upkeep:    14,
	MaxHealth: 50,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange: 380,
		Reload:      3.1,
		AttackSound: assets.AudioBeamTowerShot,
		Damage:      DamageValue{Health: 15},
		MaxTargets:  1,
		BurstSize:   1,
		TargetFlags: TargetFlying,
		FireOffset:  gmath.Vec{Y: -16},
	}),
	BeamOpaqueTime: 0.1,
	BeamSlideSpeed: 0.4,
}

var TetherBeaconAgentStats = &AgentStats{
	Kind:           AgentTetherBeacon,
	Image:          assets.ImageTetherBeaconAgent,
	Size:           SizeLarge,
	Upkeep:         8,
	MaxHealth:      70,
	SupportReload:  4,
	SupportRange:   450,
	BeamSlideSpeed: 0.4,
}

var WorkerAgentStats = &AgentStats{
	Kind:        AgentWorker,
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
}

var ScoutAgentStats = &AgentStats{
	Kind:        AgentScout,
	Image:       assets.ImageScoutAgent,
	Size:        SizeSmall,
	DiodeOffset: 5,
	Tier:        1,
	Cost:        10,
	Upkeep:      4,
	CanPatrol:   true,
	Speed:       75,
	MaxHealth:   12,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:     130,
		Reload:          2.5,
		AttackSound:     assets.AudioScoutShot,
		ProjectileImage: assets.ImageScoutProjectile,
		ImpactArea:      10,
		ProjectileSpeed: 180,
		Damage:          DamageValue{Health: 2, Disarm: 2},
		MaxTargets:      1,
		BurstSize:       1,
		TargetFlags:     TargetFlying | TargetGround,
		Explosion:       ProjectileExplosionScoutIon,
	}),
}

var TruckerAgentStats = &AgentStats{
	Kind:                 AgentTrucker,
	Image:                assets.ImageTruckerAgent,
	Size:                 SizeLarge,
	DiodeOffset:          1,
	Tier:                 3,
	Cost:                 40,
	Upkeep:               9,
	CanGather:            true,
	MaxPayload:           3,
	Speed:                85,
	MaxHealth:            45,
	EnergyRegenRateBonus: 0.5,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:           200,
		Reload:                2.6,
		AttackSound:           assets.AudioCourierShot,
		ProjectileImage:       assets.ImageCourierProjectile,
		ImpactArea:            15,
		ProjectileSpeed:       170,
		Damage:                DamageValue{Health: 2, Slow: 1, Morale: 1},
		MaxTargets:            2,
		BurstSize:             1,
		ProjectileRotateSpeed: 24,
		TargetFlags:           TargetFlying,
	}),
}

var CourierAgentStats = &AgentStats{
	Kind:        AgentCourier,
	Image:       assets.ImageCourierAgent,
	Size:        SizeMedium,
	DiodeOffset: 5,
	Tier:        2,
	Cost:        20,
	PointCost:   2,
	Upkeep:      4,
	CanGather:   true,
	MaxPayload:  1,
	Speed:       80,
	MaxHealth:   30,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:           140,
		Reload:                3.2,
		AttackSound:           assets.AudioCourierShot,
		ProjectileImage:       assets.ImageCourierProjectile,
		ImpactArea:            10,
		ProjectileSpeed:       170,
		Damage:                DamageValue{Health: 2, Slow: 1, Morale: 1},
		MaxTargets:            1,
		BurstSize:             1,
		ProjectileRotateSpeed: 24,
		TargetFlags:           TargetFlying,
	}),
}

var RedminerAgentStats = &AgentStats{
	Kind:                 AgentRedminer,
	Image:                assets.ImageRedminerAgent,
	Size:                 SizeMedium,
	DiodeOffset:          6,
	Tier:                 2,
	PointCost:            2,
	Cost:                 16,
	Upkeep:               5,
	CanGather:            true,
	MaxPayload:           1,
	Speed:                75,
	MaxHealth:            20,
	EnergyRegenRateBonus: 0.2,
}

var GeneratorAgentStats = &AgentStats{
	Kind:                 AgentGenerator,
	Image:                assets.ImageGeneratorAgent,
	Size:                 SizeMedium,
	DiodeOffset:          10,
	Tier:                 2,
	PointCost:            1,
	Cost:                 16,
	Upkeep:               2,
	CanGather:            true,
	MaxPayload:           1,
	Speed:                90,
	MaxHealth:            26,
	EnergyRegenRateBonus: 1,
}

var ClonerAgentStats = &AgentStats{
	Kind:        AgentCloner,
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
}

var RepairAgentStats = &AgentStats{
	Kind:           AgentRepair,
	Image:          assets.ImageRepairAgent,
	Size:           SizeMedium,
	DiodeOffset:    5,
	FireOffset:     -2,
	Tier:           2,
	PointCost:      4,
	Cost:           26,
	Upkeep:         14,
	CanGather:      true,
	MaxPayload:     1,
	Speed:          100,
	MaxHealth:      18,
	SupportReload:  8.0,
	SupportRange:   450,
	BeamOpaqueTime: 0.2,
	BeamSlideSpeed: 0.6,
}

var RechargeAgentStats = &AgentStats{
	Kind:                 AgentRecharger,
	Image:                assets.ImageRechargerAgent,
	Size:                 SizeMedium,
	DiodeOffset:          9,
	Tier:                 2,
	PointCost:            2,
	Cost:                 15,
	Upkeep:               4,
	CanGather:            true,
	MaxPayload:           1,
	Speed:                90,
	MaxHealth:            16,
	EnergyRegenRateBonus: 0.2,
	SupportReload:        7,
	SupportRange:         400,
	BeamOpaqueTime:       0.2,
	BeamSlideSpeed:       0.8,
}

var GuardianAgentStats = &AgentStats{
	Kind:        AgentGuardian,
	Image:       assets.ImageGuardianAgent,
	Size:        SizeLarge,
	DiodeOffset: -4,
	Tier:        3,
	Cost:        50,
	Upkeep:      16,
	CanPatrol:   true,
	Speed:       55,
	MaxHealth:   50,
	SelfRepair:  1,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange: 260,
		Reload:      3.2,
		AttackSound: assets.AudioDefenderShot,
		Damage:      DamageValue{Health: 3, Aggro: 0.9},
		MaxTargets:  2,
		BurstSize:   1,
		TargetFlags: TargetFlying | TargetGround,
	}),
	BeamOpaqueTime: 0.1,
	BeamSlideSpeed: -1.6,
}

var ServoAgentStats = &AgentStats{
	Kind:          AgentServo,
	Image:         assets.ImageServoAgent,
	Size:          SizeMedium,
	DiodeOffset:   -4,
	Tier:          2,
	PointCost:     3,
	Cost:          26,
	Upkeep:        7,
	CanGather:     true,
	MaxPayload:    1,
	Speed:         125,
	MaxHealth:     18,
	SupportReload: 8,
	SupportRange:  310,
}

var FreighterAgentStats = &AgentStats{
	Kind:                 AgentFreighter,
	Image:                assets.ImageFreighterAgent,
	Size:                 SizeMedium,
	DiodeOffset:          1,
	Tier:                 2,
	PointCost:            1,
	Cost:                 18,
	Upkeep:               3,
	CanGather:            true,
	MaxPayload:           3,
	Speed:                70,
	MaxHealth:            28,
	EnergyRegenRateBonus: 0.5,
}

var CripplerAgentStats = &AgentStats{
	Kind:        AgentCrippler,
	Image:       assets.ImageCripplerAgent,
	Size:        SizeMedium,
	DiodeOffset: 5,
	Tier:        2,
	PointCost:   2,
	Cost:        16,
	Upkeep:      4,
	CanPatrol:   true,
	CanCloak:    true,
	Speed:       65,
	MaxHealth:   18,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:     255,
		Reload:          2.7,
		AttackSound:     assets.AudioCripplerShot,
		ProjectileImage: assets.ImageCripplerProjectile,
		ImpactArea:      10,
		ProjectileSpeed: 250,
		Damage:          DamageValue{Health: 1, Slow: 2},
		MaxTargets:      3,
		BurstSize:       1,
		TargetFlags:     TargetFlying | TargetGround,
		Explosion:       ProjectileExplosionCripplerBlaster,
	}),
}

var StormbringerAgentStats = &AgentStats{
	Kind:                 AgentStormbringer,
	Image:                assets.ImageStormbringerAgent,
	Size:                 SizeLarge,
	DiodeOffset:          7,
	Tier:                 3,
	Cost:                 50,
	Upkeep:               9,
	CanPatrol:            true,
	CanGather:            true,
	Speed:                100,
	MaxHealth:            40,
	EnergyRegenRateBonus: 1,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:           170,
		Reload:                2.6,
		AttackSound:           assets.AudioStormbringerShot,
		ProjectileImage:       assets.ImageStormbringerProjectile,
		ImpactArea:            18,
		ProjectileSpeed:       200,
		ProjectileRotateSpeed: 4,
		Damage:                DamageValue{Health: 1, Disarm: 2},
		MaxTargets:            2,
		BurstSize:             4,
		BurstDelay:            0.03,
		TargetFlags:           TargetFlying | TargetGround,
		Explosion:             ProjectileExplosionShocker,
	}),
}

var PrismAgentStats = &AgentStats{
	ScoreCost:   PrismDroneCost,
	Kind:        AgentPrism,
	Image:       assets.ImagePrismAgent,
	Size:        SizeMedium,
	DiodeOffset: 1,
	Tier:        2,
	PointCost:   3,
	Cost:        26,
	Upkeep:      11,
	CanPatrol:   true,
	Speed:       70,
	MaxHealth:   30,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:     200,
		Reload:          3.7,
		AttackSound:     assets.AudioPrismShot,
		ImpactArea:      8,
		ProjectileSpeed: 220,
		Damage:          DamageValue{Health: 4},
		MaxTargets:      1,
		BurstSize:       1,
		TargetFlags:     TargetFlying | TargetGround,
	}),
}

var FighterAgentStats = &AgentStats{
	Kind:        AgentFighter,
	Image:       assets.ImageFighterAgent,
	Size:        SizeMedium,
	DiodeOffset: 1,
	Tier:        2,
	PointCost:   3,
	Cost:        22,
	Upkeep:      7,
	CanPatrol:   true,
	Speed:       90,
	MaxHealth:   28,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:     195,
		Reload:          1.9,
		AttackSound:     assets.AudioFighterBeam,
		ProjectileImage: assets.ImageFighterProjectile,
		ImpactArea:      10,
		ProjectileSpeed: 250,
		Damage:          DamageValue{Health: 5},
		Explosion:       ProjectileExplosionFighterLaser,
		MaxTargets:      1,
		BurstSize:       1,
		TargetFlags:     TargetFlying | TargetGround,
	}),
}

var SkirmisherAgentStats = &AgentStats{
	ScoreCost:   SkirmisherDroneCost,
	Kind:        AgentSkirmisher,
	Image:       assets.ImageSkirmisherAgent,
	Size:        SizeMedium,
	DiodeOffset: 3,
	Tier:        2,
	PointCost:   2,
	Cost:        25,
	Upkeep:      8,
	CanPatrol:   true,
	Speed:       80,
	MaxHealth:   20,
	SelfRepair:  0.75,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:       180,
		Reload:            2,
		AttackSound:       assets.AudioSkirmisherShot,
		ProjectileImage:   assets.ImageSkirmisherProjectile,
		ImpactArea:        15,
		ProjectileSpeed:   340,
		Damage:            DamageValue{Health: 2},
		Explosion:         ProjectileExplosionGreenZap,
		MaxTargets:        1,
		BurstSize:         1,
		AttacksPerBurst:   4,
		BurstDelay:        0.3,
		TargetFlags:       TargetFlying | TargetGround,
		FlyingDamageBonus: -0.5,
		ArcPower:          1.2,
		RandArc:           true,
	}),
}

var DefenderAgentStats = &AgentStats{
	ScoreCost:   DefenderDroneCost,
	Kind:        AgentDefender,
	Image:       assets.ImageDefenderAgent,
	Size:        SizeMedium,
	DiodeOffset: 6,
	Tier:        2,
	PointCost:   3,
	Cost:        20,
	Upkeep:      4,
	CanPatrol:   true,
	Speed:       55,
	MaxHealth:   35,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange: 240,
		Reload:      3.5,
		AttackSound: assets.AudioDefenderShot,
		Damage:      DamageValue{Health: 3, Aggro: 0.8},
		MaxTargets:  2,
		BurstSize:   1,
		TargetFlags: TargetFlying | TargetGround,
	}),
	BeamOpaqueTime: 0.1,
	BeamSlideSpeed: -1.6,
}

var ScavengerAgentStats = &AgentStats{
	Kind:          AgentScavenger,
	Image:         assets.ImageScavengerAgent,
	Size:          SizeMedium,
	DiodeOffset:   -5,
	Tier:          2,
	PointCost:     2,
	Cost:          18,
	Upkeep:        6,
	CanPatrol:     true,
	Speed:         100,
	MaxHealth:     22,
	SupportReload: 16,
	MaxPayload:    2,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:     160,
		Reload:          2.5,
		AttackSound:     assets.AudioScavengerShot,
		ProjectileImage: assets.ImageScavengerProjectile,
		ImpactArea:      8,
		ProjectileSpeed: 250,
		Damage:          DamageValue{Health: 2},
		MaxTargets:      1,
		BurstSize:       2,
		BurstDelay:      0.12,
		TargetFlags:     TargetFlying | TargetGround,
	}),
}

var ScarabAgentStats = &AgentStats{
	ScoreCost:   ScarabDroneCost,
	Kind:        AgentScarab,
	Image:       assets.ImageScarabAgent,
	Size:        SizeMedium,
	DiodeOffset: -5,
	Tier:        2,
	PointCost:   3,
	Cost:        20,
	Upkeep:      8,
	CanGather:   true,
	CanPatrol:   true,
	Speed:       65,
	MaxHealth:   14,
	MaxPayload:  1,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:     150,
		Reload:          2.5,
		AttackSound:     assets.AudioScarabShot,
		ProjectileImage: assets.ImageScarabProjectile,
		ImpactArea:      8,
		ProjectileSpeed: 350,
		Damage:          DamageValue{Health: 1},
		MaxTargets:      1,
		AttacksPerBurst: 2,
		ArcPower:        1,
		RandArc:         true,
		RoundProjectile: true,
		BurstSize:       2,
		TargetFlags:     TargetGround,
	}),
}

const (
	// +1 burst size per level (+6)
	// +5 max hp per level (+30)
	DevourerMaxLevel = 6
)

var DevourerAgentStats = &AgentStats{
	Kind:          AgentDevourer,
	Image:         assets.ImageDevourerAgent,
	Size:          SizeLarge,
	DiodeOffset:   7,
	Tier:          3,
	Cost:          60,
	Upkeep:        20,
	CanPatrol:     true,
	Speed:         75,
	MaxHealth:     30,
	MaxPayload:    1,
	SupportReload: 25,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:         200,
		Reload:              2.0,
		AttackSound:         assets.AudioScarabShot,
		ProjectileImage:     assets.ImageScarabProjectile,
		ProjectileFireSound: true,
		ImpactArea:          10,
		ProjectileSpeed:     350,
		Damage:              DamageValue{Health: 1},
		MaxTargets:          1,
		AttacksPerBurst:     3,
		ArcPower:            1,
		RandArc:             true,
		RoundProjectile:     true,
		BurstDelay:          0.25,
		BurstSize:           3,
		TargetFlags:         TargetFlying | TargetGround,
	}),
}

var AntiAirAgentStats = &AgentStats{
	ScoreCost:   AntiAirDroneCost,
	Kind:        AgentAntiAir,
	Image:       assets.ImageAntiAirAgent,
	Size:        SizeMedium,
	DiodeOffset: 1,
	Tier:        2,
	PointCost:   2,
	Cost:        24,
	Upkeep:      8,
	CanPatrol:   true,
	Speed:       80,
	MaxHealth:   20,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:     260,
		Reload:          2.35,
		AttackSound:     assets.AudioAntiAirMissiles,
		ProjectileImage: assets.ImageAntiAirMissile,
		ImpactArea:      18,
		ProjectileSpeed: 250,
		Damage:          DamageValue{Health: 2},
		MaxTargets:      1,
		BurstSize:       4,
		BurstDelay:      0.1,
		Explosion:       ProjectileExplosionNormal,
		TrailEffect:     ProjectileTrailSmoke,
		ArcPower:        2,
		Accuracy:        0.95,
		TargetFlags:     TargetFlying,
		FireOffset:      gmath.Vec{Y: -8},
	}),
}

var MortarAgentStats = &AgentStats{
	ScoreCost:   MortarDroneCost,
	Kind:        AgentMortar,
	Image:       assets.ImageMortarAgent,
	Size:        SizeMedium,
	DiodeOffset: 1,
	Tier:        2,
	PointCost:   1,
	Cost:        18,
	Upkeep:      6,
	CanPatrol:   true,
	Speed:       70,
	MaxHealth:   30,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:     370,
		Reload:          3.2,
		AttackSound:     assets.AudioMortarShot,
		ProjectileImage: assets.ImageMortarProjectile,
		ImpactArea:      18,
		ProjectileSpeed: 180,
		Damage:          DamageValue{Health: 13},
		MaxTargets:      1,
		BurstSize:       1,
		Explosion:       ProjectileExplosionNormal,
		ArcPower:        2.5,
		Accuracy:        0.9,
		TargetFlags:     TargetGround,
		RoundProjectile: true,
	}),
}

var DestroyerAgentStats = &AgentStats{
	Kind:        AgentDestroyer,
	Image:       assets.ImageDestroyerAgent,
	Size:        SizeLarge,
	DiodeOffset: 0,
	Tier:        3,
	Cost:        60,
	Upkeep:      20,
	CanPatrol:   true,
	Speed:       85,
	MaxHealth:   45,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange: 210,
		Reload:      1.8,
		AttackSound: assets.AudioDestroyerBeam,
		Damage:      DamageValue{Health: 7},
		MaxTargets:  1,
		BurstSize:   1,
		TargetFlags: TargetFlying | TargetGround,
	}),
}

var MarauderAgentStats = &AgentStats{
	Kind:          AgentMarauder,
	Image:         assets.ImageMarauderAgent,
	Size:          SizeLarge,
	DiodeOffset:   0,
	Tier:          3,
	Cost:          40,
	Upkeep:        12,
	CanPatrol:     true,
	CanCloak:      true,
	Speed:         100,
	SupportReload: 14,
	MaxHealth:     30,
	MaxPayload:    3,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:     255,
		Reload:          2.3,
		ProjectileImage: assets.ImageMarauderProjectile,
		ImpactArea:      16,
		ProjectileSpeed: 300,
		AttackSound:     assets.AudioMarauderShot,
		Damage:          DamageValue{Health: 4, Slow: 2},
		BurstSize:       1,
		MaxTargets:      3,
		TargetFlags:     TargetFlying | TargetGround,
		Explosion:       ProjectileExplosionCripplerBlaster,
	}),
}

var RepellerAgentStats = &AgentStats{
	Kind:        AgentRepeller,
	Image:       assets.ImageRepellerAgent,
	Size:        SizeMedium,
	DiodeOffset: 8,
	Tier:        2,
	PointCost:   3,
	Cost:        16,
	Upkeep:      4,
	CanGather:   true,
	MaxPayload:  1,
	CanPatrol:   true,
	Speed:       105,
	MaxHealth:   22,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:     160,
		Reload:          2.2,
		AttackSound:     assets.AudioRepellerBeam,
		ProjectileImage: assets.ImageRepellerProjectile,
		ImpactArea:      10,
		ProjectileSpeed: 200,
		Damage:          DamageValue{Health: 2, Disarm: 4},
		MaxTargets:      2,
		BurstSize:       1,
		TargetFlags:     TargetFlying | TargetGround,
		Explosion:       ProjectileExplosionShocker,
	}),
}

var KamikazeAgentStats = &AgentStats{
	ScoreCost:   KamikazeDroneCost,
	Kind:        AgentKamikaze,
	Image:       assets.ImageKamikazeAgent,
	Size:        SizeMedium,
	DiodeOffset: 4,
	Tier:        2,
	PointCost:   1,
	Cost:        14,
	Upkeep:      4,
	CanGather:   true,
	MaxPayload:  1,
	CanPatrol:   true,
	Speed:       100,
	MaxHealth:   20,
}

var DisintegratorAgentStats = &AgentStats{
	ScoreCost:     DisintegratorDroneCost,
	Kind:          AgentDisintegrator,
	Image:         assets.ImageDisintegratorAgent,
	Size:          SizeMedium,
	DiodeOffset:   4,
	Tier:          2,
	PointCost:     3,
	Cost:          22,
	Upkeep:        12,
	CanGather:     true,
	MaxPayload:    1,
	CanPatrol:     true,
	Speed:         80,
	MaxHealth:     20,
	SupportReload: 12,
	Weapon: InitWeaponStats(&WeaponStats{
		AttackRange:           220,
		AttackSound:           assets.AudioDisintegratorShot,
		ProjectileImage:       assets.ImageDisintegratorProjectile,
		ImpactArea:            18,
		ProjectileSpeed:       210,
		ProjectileRotateSpeed: 26,
		Damage:                DamageValue{Health: 15},
		MaxTargets:            1,
		BurstSize:             1,
		Explosion:             ProjectileExplosionPurple,
		TargetFlags:           TargetFlying | TargetGround,
		GroundDamageBonus:     -0.5,
	}),
}
