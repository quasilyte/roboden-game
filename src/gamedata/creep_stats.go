package gamedata

import (
	"image/color"

	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
)

//go:generate stringer -type=CreepKind -trimprefix=Creep
type CreepKind int

const (
	CreepPrimitiveWanderer CreepKind = iota
	CreepStunner
	CreepAssault
	CreepDominator
	CreepBuilder
	CreepTurret
	CreepTurretConstruction
	CreepCrawlerBaseConstruction
	CreepBase
	CreepCrawlerBase
	CreepCrawler
	CreepHowitzer
	CreepServant
	CreepUberBoss
	CreepWisp
	CreepWispLair
	CreepFortress
	CreepTemplar
	CreepCenturion
	CreepGrenadier
)

type CreepStats struct {
	Kind        CreepKind
	Image       resource.ImageID
	ShadowImage resource.ImageID

	Tier int

	Speed float64
	Size  float64

	AnimSpeed float64

	MaxHealth float64

	Weapon        *WeaponStats
	SuperWeapon   *WeaponStats
	SpecialWeapon *WeaponStats

	BeamColor      color.RGBA
	BeamWidth      float64
	BeamSlideSpeed float64
	BeamOpaqueTime float64
	BeamTexture    *ge.Texture
	BeamExplosion  resource.ImageID

	TargetKind      TargetKind
	Disarmable      bool
	CanBeRepelled   bool
	Flying          bool
	Building        bool
	SiegeTargetable bool

	NameTag string
}

var MagmaHazardWeapon = InitWeaponStats(&WeaponStats{
	MaxTargets:            1,
	BurstSize:             1,
	AttackRange:           250,
	ImpactArea:            40,
	ProjectileSpeed:       180,
	AttackSound:           assets.AudioMagmaShot1,
	ProjectileFireSound:   true,
	ProjectileRotateSpeed: 2,
	Damage:                DamageValue{Health: 15},
	ProjectileImage:       assets.ImageMagmaBall,
	ArcPower:              7,
	TrailEffect:           ProjectileTrailMagma,
	Explosion:             ProjectileExplosionMagma,
	AlwaysExplodes:        true,
})

var AtomicBombWeapon = InitWeaponStats(&WeaponStats{
	MaxTargets:          1,
	BurstSize:           1,
	AttackRange:         99999,
	ImpactArea:          150,
	ProjectileSpeed:     120,
	AttackSound:         assets.AudioAbombLaunch,
	ProjectileFireSound: true,
	Damage:              DamageValue{Health: 300},
	ProjectileImage:     assets.ImageAbombMissile,
	ArcPower:            15,
	TrailEffect:         ProjectileTrailFire,
	Explosion:           ProjectileExplosionAbomb,
	AlwaysExplodes:      true,
})

var IonMortarCreepStats = &CreepStats{
	Kind:      CreepTurret,
	Image:     assets.ImageIonMortarCreep,
	Speed:     0,
	MaxHealth: 150,
	Weapon: InitWeaponStats(&WeaponStats{
		MaxTargets:          1,
		BurstSize:           2,
		BurstDelay:          0.4,
		AttackSound:         assets.AudioIonMortarShot,
		ProjectileFireSound: true,
		AttackRange:         1000,
		ImpactArea:          40,
		ProjectileSpeed:     400,
		Damage:              DamageValue{Health: 8, Energy: 50},
		ProjectileImage:     assets.ImageIonMortarProjectile,
		Explosion:           ProjectileExplosionIonBlast,
		TrailEffect:         ProjectileTrailIonMortar,
		Reload:              11.0,
		FireOffsets:         []gmath.Vec{{Y: -12}},
		TargetFlags:         TargetFlying,
		AlwaysExplodes:      true,
		ArcPower:            4.0,
		Accuracy:            0.4,
		RoundProjectile:     true,
	}),
	SuperWeapon: InitWeaponStats(&WeaponStats{
		MaxTargets:          1,
		BurstSize:           4,
		BurstDelay:          0.2,
		AttackSound:         assets.AudioIonMortarShot,
		ProjectileFireSound: true,
		AttackRange:         1050,
		ImpactArea:          40,
		ProjectileSpeed:     450,
		Damage:              DamageValue{Health: 8, Energy: 50},
		ProjectileImage:     assets.ImageSuperIonMortarProjectile,
		Explosion:           ProjectileExplosionSuperIonBlast,
		TrailEffect:         ProjectileTrailSuperIonMortar,
		Reload:              12.0,
		FireOffsets:         []gmath.Vec{{Y: -12}},
		TargetFlags:         TargetFlying,
		AlwaysExplodes:      true,
		ArcPower:            4.0,
		Accuracy:            0.4,
		RoundProjectile:     true,
	}),
	Size:            28,
	CanBeRepelled:   false,
	Disarmable:      false,
	Building:        true,
	SiegeTargetable: true,
	TargetKind:      TargetGround,
}

var TurretCreepStats = &CreepStats{
	Kind:      CreepTurret,
	Image:     assets.ImageTurretCreep,
	Speed:     0,
	MaxHealth: 120,
	Weapon: InitWeaponStats(&WeaponStats{
		MaxTargets:      1,
		BurstSize:       1,
		AttackSound:     assets.AudioMissile,
		AttackRange:     290,
		ImpactArea:      18,
		ProjectileSpeed: 360,
		Damage:          DamageValue{Health: 10},
		ProjectileImage: assets.ImageMissile,
		Explosion:       ProjectileExplosionNormal,
		TrailEffect:     ProjectileTrailSmoke,
		Reload:          3.5,
		FireOffsets:     []gmath.Vec{{Y: -8}},
		TargetFlags:     TargetFlying | TargetGround,
	}),
	Size:            40,
	CanBeRepelled:   false,
	Disarmable:      false,
	Building:        true,
	SiegeTargetable: true,
	TargetKind:      TargetGround,
}

var FortressCreepStats = &CreepStats{
	Kind:      CreepFortress,
	Image:     assets.ImageFortressCreep,
	Speed:     0,
	MaxHealth: 375,
	Weapon: InitWeaponStats(&WeaponStats{
		MaxTargets:      1,
		AttackSound:     assets.AudioFortressAttack,
		AttackRange:     350,
		ImpactArea:      18,
		ProjectileSpeed: 450,
		Damage:          DamageValue{Health: 5, Energy: 10, Morale: 0.2},
		BurstSize:       5,
		AttacksPerBurst: 2,
		BurstDelay:      0.1,
		ProjectileImage: assets.ImageEnergySpear,
		TrailEffect:     ProjectileTrailEnergySpear,
		Reload:          2.7,
		FireOffsets:     []gmath.Vec{{Y: -1}},
		TargetFlags:     TargetFlying,
		ArcPower:        0.4,
		RandArc:         true,
	}),
	Size:            64,
	CanBeRepelled:   false,
	Disarmable:      false,
	Building:        true,
	SiegeTargetable: true,
	TargetKind:      TargetGround,
}

var BaseCreepStats = &CreepStats{
	Kind:            CreepBase,
	Image:           assets.ImageCreepBase,
	Speed:           0,
	MaxHealth:       170,
	Size:            60,
	Disarmable:      false,
	CanBeRepelled:   false,
	Building:        true,
	SiegeTargetable: true,
	TargetKind:      TargetGround,
}

var CrawlerBaseCreepStats = &CreepStats{
	Kind:            CreepCrawlerBase,
	Image:           assets.ImageCrawlerCreepBase,
	Speed:           0,
	MaxHealth:       140,
	Size:            60,
	Disarmable:      false,
	CanBeRepelled:   false,
	Building:        true,
	SiegeTargetable: true,
	TargetKind:      TargetGround,
}

var CrawlerBaseConstructionCreepStats = &CreepStats{
	Kind:            CreepCrawlerBaseConstruction,
	Image:           assets.ImageCrawlerCreepBase,
	Speed:           0,
	MaxHealth:       35,
	Size:            40,
	Disarmable:      false,
	CanBeRepelled:   false,
	Building:        true,
	SiegeTargetable: true,
	TargetKind:      TargetGround,
}

var TurretConstructionCreepStats = &CreepStats{
	Kind:            CreepTurretConstruction,
	Image:           assets.ImageTurretCreep,
	Speed:           0,
	MaxHealth:       35,
	Size:            40,
	Disarmable:      false,
	CanBeRepelled:   false,
	Building:        true,
	SiegeTargetable: true,
	TargetKind:      TargetGround,
}

var IonMortarConstructionCreepStats = &CreepStats{
	Kind:            CreepTurretConstruction,
	Image:           assets.ImageIonMortarCreep,
	Speed:           0,
	MaxHealth:       35,
	Size:            40,
	Disarmable:      false,
	CanBeRepelled:   false,
	Building:        true,
	SiegeTargetable: true,
	TargetKind:      TargetGround,
}

var WandererCreepStats = &CreepStats{
	NameTag:     "rogue",
	Kind:        CreepPrimitiveWanderer,
	Image:       assets.ImageCreepTier1,
	ShadowImage: assets.ImageSmallShadow,
	Tier:        1,
	Speed:       40,
	MaxHealth:   14,
	Weapon: InitWeaponStats(&WeaponStats{
		MaxTargets:      1,
		BurstSize:       1,
		AttackSound:     assets.AudioWandererBeam,
		AttackRange:     190,
		ImpactArea:      10,
		ProjectileSpeed: 400,
		Damage:          DamageValue{Health: 4},
		ProjectileImage: assets.ImageWandererProjectile,
		Reload:          2.2,
		TargetFlags:     TargetFlying | TargetGround,
	}),
	Disarmable:    true,
	CanBeRepelled: true,
	Flying:        true,
	TargetKind:    TargetFlying,
}

var WispCreepStats = &CreepStats{
	Kind:          CreepWisp,
	Image:         assets.ImageWisp,
	AnimSpeed:     0.12,
	ShadowImage:   assets.ImageMediumShadow,
	Tier:          2,
	Speed:         20,
	MaxHealth:     35,
	Disarmable:    false,
	CanBeRepelled: false,
	Flying:        true,
	TargetKind:    TargetFlying,
}

var WispLairCreepStats = &CreepStats{
	Kind:          CreepWispLair,
	Image:         assets.ImageWispLair,
	Speed:         0,
	MaxHealth:     160,
	Size:          60,
	Disarmable:    false,
	CanBeRepelled: false,
	Building:      true,
	TargetKind:    TargetGround,
}

var ServantCreepStats = &CreepStats{
	Kind:        CreepServant,
	Image:       assets.ImageServantCreep,
	ShadowImage: assets.ImageMediumShadow,
	Tier:        2,
	Speed:       70,
	MaxHealth:   70,
	AnimSpeed:   0.15,
	Weapon: InitWeaponStats(&WeaponStats{
		MaxTargets:          1,
		BurstSize:           1,
		AttackSound:         assets.AudioServantShot,
		AttackRange:         240,
		ImpactArea:          10,
		ProjectileSpeed:     340,
		BuildingDamageBonus: 0.75,
		Damage:              DamageValue{Health: 4, Energy: 20},
		ProjectileImage:     assets.ImageServantProjectile,
		Reload:              3.2,
		TargetFlags:         TargetFlying | TargetGround,
		Explosion:           ProjectileExplosionServant,
	}),
	Disarmable:    true,
	CanBeRepelled: false,
	Flying:        true,
	TargetKind:    TargetFlying,
}

var CrawlerCreepStats = &CreepStats{
	NameTag:   "crawler",
	Kind:      CreepCrawler,
	Image:     assets.ImageCrawlerCreep,
	AnimSpeed: 0.09,
	Speed:     44,
	MaxHealth: 18,
	Weapon: InitWeaponStats(&WeaponStats{
		MaxTargets:          1,
		BurstSize:           2,
		BurstDelay:          0.12,
		AttackSound:         assets.AudioTankShot,
		AttackRange:         160,
		ImpactArea:          14,
		ProjectileSpeed:     350,
		Damage:              DamageValue{Health: 2},
		ProjectileImage:     assets.ImageTankProjectile,
		Reload:              1.7,
		TargetFlags:         TargetFlying | TargetGround,
		FireOffsets:         []gmath.Vec{{Y: -2}},
		BuildingDamageBonus: -0.2,
	}),
	Size:          24,
	Disarmable:    true,
	CanBeRepelled: true,
	TargetKind:    TargetGround,
}

var EliteCrawlerCreepStats = &CreepStats{
	NameTag:   "sprayer",
	Kind:      CreepCrawler,
	Image:     assets.ImageEliteCrawlerCreep,
	AnimSpeed: 0.09,
	Speed:     40,
	MaxHealth: 28,
	Weapon: InitWeaponStats(&WeaponStats{
		MaxTargets:      6,
		BurstSize:       1,
		AttackSound:     assets.AudioEliteCrawlerShot,
		AttackRange:     160,
		ImpactArea:      10,
		ProjectileSpeed: 320,
		Damage:          DamageValue{Health: 1},
		ProjectileImage: assets.ImageEliteCrawlerProjectile,
		Reload:          1.9,
		TargetFlags:     TargetFlying | TargetGround,
		FireOffsets:     []gmath.Vec{{Y: -2}},
	}),
	Size:          24,
	Disarmable:    true,
	CanBeRepelled: true,
	TargetKind:    TargetGround,
}

var HeavyCrawlerCreepStats = &CreepStats{
	NameTag:   "assault_crawler",
	Kind:      CreepCrawler,
	Image:     assets.ImageHeavyCrawlerCreep,
	AnimSpeed: 0.16,
	Speed:     30,
	MaxHealth: 60,
	Weapon: InitWeaponStats(&WeaponStats{
		MaxTargets:          1,
		BurstSize:           5,
		BurstDelay:          0.1,
		AttackSound:         assets.AudioHeavyCrawlerShot,
		AttackRange:         260,
		ImpactArea:          12,
		ProjectileSpeed:     280,
		Damage:              DamageValue{Health: 2},
		ProjectileImage:     assets.ImageHeavyCrawlerProjectile,
		Reload:              2.4,
		TargetFlags:         TargetFlying | TargetGround,
		FireOffsets:         []gmath.Vec{{Y: -2}},
		Explosion:           ProjectileExplosionHeavyCrawlerLaser,
		ArcPower:            1.5,
		Accuracy:            0.85,
		BuildingDamageBonus: 0.25,
	}),
	Size:          24,
	Disarmable:    true,
	CanBeRepelled: true,
	TargetKind:    TargetGround,
}

var HowitzerCreepStats = &CreepStats{
	NameTag:   "howitzer",
	Kind:      CreepHowitzer,
	Image:     assets.ImageHowitzerCreep,
	AnimSpeed: 0.2,
	Speed:     10,
	MaxHealth: 260,
	Weapon: InitWeaponStats(&WeaponStats{
		MaxTargets:          3,
		BurstSize:           4,
		BurstDelay:          0.2,
		Accuracy:            0.7,
		AttackSound:         assets.AudioHowitzerLaserShot,
		AttackRange:         300,
		ImpactArea:          14,
		ProjectileSpeed:     480,
		Damage:              DamageValue{Health: 2},
		ProjectileImage:     assets.ImageHowitzerLaserProjectile,
		Reload:              1.9,
		TargetFlags:         TargetFlying,
		FireOffsets:         []gmath.Vec{{Y: -2}},
		ProjectileFireSound: true,
	}),
	SpecialWeapon: InitWeaponStats(&WeaponStats{
		MaxTargets:          1,
		BurstSize:           1,
		AttackSound:         assets.AudioHowitzerShot,
		AttackRange:         850,
		ImpactArea:          26,
		ProjectileSpeed:     150,
		Damage:              DamageValue{Health: 20},
		ProjectileImage:     assets.ImageHowitzerProjectile,
		Reload:              16,
		TargetFlags:         TargetGround,
		Explosion:           ProjectileExplosionBigVertical,
		TrailEffect:         ProjectileTrailSmoke,
		AlwaysExplodes:      true,
		ArcPower:            3,
		Accuracy:            0.4,
		ProjectileFireSound: true,
	}),
	Size:            32,
	Disarmable:      false,
	CanBeRepelled:   false,
	SiegeTargetable: true,
	TargetKind:      TargetGround,
}

var StealthCrawlerCreepStats = &CreepStats{
	NameTag:   "stealth_crawler",
	Kind:      CreepCrawler,
	Image:     assets.ImageStealthCrawlerCreep,
	AnimSpeed: 0.09,
	Speed:     70,
	MaxHealth: 25,
	Weapon: InitWeaponStats(&WeaponStats{
		MaxTargets:          1,
		BurstSize:           3,
		BurstDelay:          0.4,
		ProjectileFireSound: true,
		AttackSound:         assets.AudioStealthCrawlerShot,
		AttackRange:         200,
		ImpactArea:          14,
		ProjectileSpeed:     320,
		Damage:              DamageValue{Health: 5, Slow: 2},
		ProjectileImage:     assets.ImageStealthCrawlerProjectile,
		Reload:              4.0,
		TargetFlags:         TargetFlying | TargetGround,
		FireOffsets:         []gmath.Vec{{Y: -2}},
		Explosion:           ProjectileExplosionStealthLaser,
	}),
	Size:          24,
	Disarmable:    true,
	CanBeRepelled: false,
	TargetKind:    TargetGround,
}

var GrenadierCreepStats = &CreepStats{
	NameTag:     "grenadier",
	Kind:        CreepGrenadier,
	Image:       assets.ImageCreepGrenadier,
	AnimSpeed:   0.15,
	ShadowImage: assets.ImageMediumShadow,
	Tier:        2,
	Speed:       30,
	MaxHealth:   90,
	SpecialWeapon: InitWeaponStats(&WeaponStats{
		AttackRange:         200,
		Reload:              25.0,
		AttackSound:         assets.AudioGrenadierShot,
		ProjectileImage:     assets.ImageGrenadierProjectile,
		ProjectileFireSound: true,
		ImpactArea:          22,
		AlwaysExplodes:      true,
		ProjectileSpeed:     80,
		Damage:              DamageValue{Health: 15},
		MaxTargets:          1,
		BurstSize:           1,
		Explosion:           ProjectileExplosionNormal,
		ArcPower:            1.0,
		TargetFlags:         TargetGround,
		FireOffsets:         []gmath.Vec{{X: 0, Y: 10}},
		TrailEffect:         ProjectileTrailGrenade,
	}),
	Disarmable:    true,
	CanBeRepelled: true,
	Flying:        true,
	TargetKind:    TargetFlying,
}

var AssaultCreepStats = &CreepStats{
	NameTag:     "vanguard",
	Kind:        CreepAssault,
	Image:       assets.ImageCreepTier3,
	AnimSpeed:   0.2,
	ShadowImage: assets.ImageBigShadow,
	Tier:        3,
	Speed:       30,
	MaxHealth:   100,
	Weapon: InitWeaponStats(&WeaponStats{
		MaxTargets:          1,
		BurstSize:           1,
		AttackSound:         assets.AudioAssaultShot,
		AttackRange:         150,
		ImpactArea:          10,
		ProjectileSpeed:     460,
		Damage:              DamageValue{Health: 3},
		ProjectileImage:     assets.ImageAssaultProjectile,
		Reload:              0.7,
		TargetFlags:         TargetFlying | TargetGround,
		BuildingDamageBonus: 0.6,
	}),
	Disarmable:    true,
	CanBeRepelled: true,
	Flying:        true,
	TargetKind:    TargetFlying,
}

var DominatorCreepStats = &CreepStats{
	NameTag:     "dominator",
	Kind:        CreepDominator,
	Image:       assets.ImageCreepDominator,
	ShadowImage: assets.ImageBigShadow,
	Tier:        3,
	Speed:       35,
	MaxHealth:   175,
	Weapon: InitWeaponStats(&WeaponStats{
		MaxTargets:          1,
		BurstSize:           1,
		AttackSound:         assets.AudioDominatorShot,
		AttackRange:         280,
		Damage:              DamageValue{Health: 8, Morale: 0.8},
		Reload:              1.65,
		TargetFlags:         TargetFlying | TargetGround,
		BuildingDamageBonus: -0.4,
	}),
	BeamColor:     ge.RGB(0x7a51f2),
	BeamWidth:     1,
	Disarmable:    false,
	CanBeRepelled: false,
	Flying:        true,
	TargetKind:    TargetFlying,
}

var BuilderCreepStats = &CreepStats{
	NameTag:       "builder",
	Kind:          CreepBuilder,
	Image:         assets.ImageBuilderCreep,
	AnimSpeed:     0.1,
	ShadowImage:   assets.ImageBigShadow,
	Tier:          3,
	Speed:         40,
	MaxHealth:     190,
	CanBeRepelled: false,
	Disarmable:    false,
	Flying:        true,
	TargetKind:    TargetFlying,
}

var UberBossCreepStats = &CreepStats{
	Kind:        CreepUberBoss,
	Image:       assets.ImageUberBoss,
	ShadowImage: assets.ImageUberBossShadow,
	Speed:       10,
	MaxHealth:   600,
	Weapon: InitWeaponStats(&WeaponStats{
		MaxTargets:          5,
		BurstSize:           1,
		AttackSound:         assets.AudioRailgun,
		AttackRange:         220,
		Damage:              DamageValue{Health: 9},
		BuildingDamageBonus: -0.4,
		Reload:              2.8,
		TargetFlags:         TargetFlying | TargetGround,
	}),
	BeamSlideSpeed: 2,
	Disarmable:     false,
	CanBeRepelled:  false,
	Flying:         true, // Most of the time...
	TargetKind:     TargetFlying,
}

var TemplarCreepStats = &CreepStats{
	NameTag:     "stunner",
	Kind:        CreepTemplar,
	Image:       assets.ImageCreepTemplar,
	ShadowImage: assets.ImageMediumShadow,
	Tier:        2,
	Speed:       40,
	MaxHealth:   40,
	Weapon: InitWeaponStats(&WeaponStats{
		MaxTargets:  1,
		BurstSize:   1,
		AttackSound: assets.AudioTemplarAttack,
		AttackRange: 300,
		Damage:      DamageValue{Health: 1, Flags: DmgflagStun},
		Reload:      2.6,
		TargetFlags: TargetFlying,
	}),
	SuperWeapon: InitWeaponStats(&WeaponStats{
		MaxTargets:  1,
		BurstSize:   1,
		AttackSound: assets.AudioTemplarAttack,
		AttackRange: 300,
		Damage:      DamageValue{Health: 1, Flags: DmgflagStun | DmgflagStunImproved},
		Reload:      2.2,
		TargetFlags: TargetFlying,
	}),
	BeamExplosion:  assets.ImageStunExplosion,
	BeamSlideSpeed: 2.5,
	BeamOpaqueTime: 0.2,
	Disarmable:     true,
	CanBeRepelled:  true,
	Flying:         true,
	TargetKind:     TargetFlying,
}

var CenturionCreepStats = &CreepStats{
	NameTag:     "coordinator",
	Kind:        CreepCenturion,
	Image:       assets.ImageCreepCenturion,
	ShadowImage: assets.ImageMediumShadow,
	AnimSpeed:   0.1,
	Tier:        2,
	Speed:       50,
	MaxHealth:   55,
	Weapon: InitWeaponStats(&WeaponStats{
		MaxTargets:          1,
		BurstSize:           2,
		AttacksPerBurst:     1,
		BurstDelay:          0.10,
		ProjectileFireSound: true,
		AttackSound:         assets.AudioCenturionShot,
		AttackRange:         220,
		ImpactArea:          10,
		ProjectileSpeed:     425,
		Explosion:           ProjectileExplosionPurpleZap,
		Damage:              DamageValue{Health: 3},
		ProjectileImage:     assets.ImageCenturionProjectile,
		Reload:              2.45,
		TargetFlags:         TargetFlying | TargetGround,
		FireOffsets:         []gmath.Vec{{X: -10}, {X: 10}},
	}),
	Disarmable:    true,
	CanBeRepelled: false,
	Flying:        true,
	TargetKind:    TargetFlying,
}

var StunnerCreepStats = &CreepStats{
	NameTag:     "discharger",
	Kind:        CreepStunner,
	Image:       assets.ImageCreepTier2,
	ShadowImage: assets.ImageMediumShadow,
	Tier:        2,
	Speed:       70,
	MaxHealth:   40,
	Weapon: InitWeaponStats(&WeaponStats{
		MaxTargets:  3,
		BurstSize:   1,
		AttackSound: assets.AudioStunBeam,
		AttackRange: 250,
		Damage:      DamageValue{Health: 1, Energy: 40},
		Reload:      2.8,
		TargetFlags: TargetFlying | TargetGround,
	}),
	SuperWeapon: InitWeaponStats(&WeaponStats{
		MaxTargets:  3,
		BurstSize:   1,
		AttackSound: assets.AudioStunBeam,
		AttackRange: 250,
		Damage:      DamageValue{Slow: 2, Health: 2, Energy: 55},
		Reload:      2.6,
		TargetFlags: TargetFlying | TargetGround,
	}),
	BeamSlideSpeed: 0.8,
	Disarmable:     true,
	CanBeRepelled:  true,
	Flying:         true,
	TargetKind:     TargetFlying,
}
