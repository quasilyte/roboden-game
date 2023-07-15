package gamedata

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/roboden-game/assets"
)

type Achievement struct {
	Name      string
	Mode      Mode
	Icon      resource.ImageID
	OnlyElite bool
}

type Mode int

const (
	ModeClassic Mode = iota
	ModeArena
	ModeInfArena
	ModeReverse

	ModeTutorial

	ModeAny

	ModeUnknown
)

func (m Mode) String() string {
	switch m {
	case ModeClassic:
		return "classic"
	case ModeArena:
		return "arena"
	case ModeInfArena:
		return "inf_arena"
	case ModeReverse:
		return "reverse"
	case ModeTutorial:
		return "tutorial"
	default:
		return "unknown"
	}
}

var AchievementList = []*Achievement{
	// Any mode achievements.
	{
		Name:      "t3engineer",
		Mode:      ModeAny,
		Icon:      assets.ImageAchievementT3Engineer,
		OnlyElite: true,
	},
	{
		Name: "trample",
		Mode: ModeAny,
		Icon: assets.ImageAchievementTrample,
	},
	{
		Name: "nopeeking",
		Mode: ModeAny,
		Icon: assets.ImageAchievementNoPeeking,
	},
	{
		Name: "nonstop",
		Mode: ModeAny,
		Icon: assets.ImageAchievementNonstop,
	},

	// Classic mode achievements.
	{
		Name:      "impossible",
		Mode:      ModeClassic,
		Icon:      assets.ImageAchievementImpossible,
		OnlyElite: true,
	},
	{
		Name: "cheapbuild10",
		Mode: ModeClassic,
		Icon: assets.ImageAchievementCheapBuild10,
	},
	{
		Name: "hightension",
		Mode: ModeClassic,
		Icon: assets.ImageAchievementHighTension,
	},
	{
		Name: "solobase",
		Mode: ModeClassic,
		Icon: assets.ImageAchievementSoloBase,
	},
	{
		Name: "uiless",
		Mode: ModeClassic,
		Icon: assets.ImageAchievementUILess,
	},
	{
		Name: "powerof3",
		Mode: ModeClassic,
		Icon: assets.ImageAchievementPowerOf3,
	},
	{
		Name: "tinyradius",
		Mode: ModeClassic,
		Icon: assets.ImageAchievementTinyRadius,
	},
	{
		Name: "t1army",
		Mode: ModeClassic,
		Icon: assets.ImageAchievementT1Army,
	},
	{
		Name: "groundwin",
		Mode: ModeClassic,
		Icon: assets.ImageAchievementGroundWin,
	},
	{
		Name: "speedrunning",
		Mode: ModeClassic,
		Icon: assets.ImageAchievementSpeedrunning,
	},
	{
		Name: "victorydrag",
		Mode: ModeClassic,
		Icon: assets.ImageAchievementVictoryDrag,
	},
	{
		Name: "t3less",
		Mode: ModeClassic,
		Icon: assets.ImageAchievementT3Less,
	},
	{
		Name: "turretdamage",
		Mode: ModeClassic,
		Icon: assets.ImageAchievementTurretDamage,
	},

	// Arena mode achievements.
	{
		Name: "antidominator",
		Mode: ModeArena,
		Icon: assets.ImageAchievementAntiDominator,
	},

	// Infinite arena mode achievements.
	{
		Name: "infinite",
		Mode: ModeInfArena,
		Icon: assets.ImageAchievementInfinite,
	},

	// Reverse mode achievements.
	{
		Name: "colonyhunter",
		Mode: ModeReverse,
		Icon: assets.ImageAchievementColonyHunter,
	},
	{
		Name: "groundcontrol",
		Mode: ModeReverse,
		Icon: assets.ImageAchievementGroundControl,
	},
	{
		Name: "atomicfinisher",
		Mode: ModeReverse,
		Icon: assets.ImageAchievementAtomicFinisher,
	},

	// Other achievements.
	{
		Name: "secret",
		Mode: ModeUnknown,
		Icon: assets.ImageAchievementSecret,
	},
	{
		Name: "terminal",
		Mode: ModeUnknown,
		Icon: assets.ImageAchievementTerminal,
	},
}
