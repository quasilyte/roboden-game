package gamedata

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/roboden-game/assets"
)

type Achievement struct {
	Name         string
	Mode         Mode
	Icon         resource.ImageID
	OnlyElite    bool
	NeedsVictory bool
}

type Mode int

const (
	ModeClassic Mode = iota
	ModeArena
	ModeInfArena
	ModeReverse
	ModeBlitz

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
	case ModeBlitz:
		return "blitz"
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
		Name:         "trample",
		Mode:         ModeAny,
		Icon:         assets.ImageAchievementTrample,
		NeedsVictory: true,
	},
	{
		Name:         "nopeeking",
		Mode:         ModeAny,
		Icon:         assets.ImageAchievementNoPeeking,
		NeedsVictory: true,
	},
	{
		Name:         "nonstop",
		Mode:         ModeAny,
		Icon:         assets.ImageAchievementNonstop,
		NeedsVictory: true,
	},
	{
		Name:         "darkness",
		Mode:         ModeAny,
		Icon:         assets.ImageAchievementDarkness,
		NeedsVictory: true,
	},
	{
		Name:         "fastforward",
		Mode:         ModeAny,
		Icon:         assets.ImageAchievementFastforward,
		NeedsVictory: true,
	},
	{
		Name:         "lucky",
		Mode:         ModeAny,
		Icon:         assets.ImageAchievementLucky,
		NeedsVictory: true,
	},

	// Classic mode achievements.
	{
		Name:         "impossible",
		Mode:         ModeClassic,
		Icon:         assets.ImageAchievementImpossible,
		OnlyElite:    true,
		NeedsVictory: true,
	},
	{
		Name:         "cheapbuild10",
		Mode:         ModeClassic,
		Icon:         assets.ImageAchievementCheapBuild10,
		NeedsVictory: true,
	},
	{
		Name:         "hightension",
		Mode:         ModeClassic,
		Icon:         assets.ImageAchievementHighTension,
		NeedsVictory: true,
	},
	{
		Name:         "solobase",
		Mode:         ModeClassic,
		Icon:         assets.ImageAchievementSoloBase,
		NeedsVictory: true,
	},
	{
		Name:         "uiless",
		Mode:         ModeClassic,
		Icon:         assets.ImageAchievementUILess,
		NeedsVictory: true,
	},
	{
		Name:         "powerof3",
		Mode:         ModeClassic,
		Icon:         assets.ImageAchievementPowerOf3,
		NeedsVictory: true,
	},
	{
		Name:         "tinyradius",
		Mode:         ModeClassic,
		Icon:         assets.ImageAchievementTinyRadius,
		NeedsVictory: true,
	},
	{
		Name:         "t1army",
		Mode:         ModeClassic,
		Icon:         assets.ImageAchievementT1Army,
		NeedsVictory: true,
	},
	{
		Name:         "groundwin",
		Mode:         ModeClassic,
		Icon:         assets.ImageAchievementGroundWin,
		NeedsVictory: true,
	},
	{
		Name:         "speedrunning",
		Mode:         ModeClassic,
		Icon:         assets.ImageAchievementSpeedrunning,
		NeedsVictory: true,
	},
	{
		Name:         "victorydrag",
		Mode:         ModeClassic,
		Icon:         assets.ImageAchievementVictoryDrag,
		NeedsVictory: true,
	},
	{
		Name:         "t3less",
		Mode:         ModeClassic,
		Icon:         assets.ImageAchievementT3Less,
		NeedsVictory: true,
	},
	{
		Name:         "turretdamage",
		Mode:         ModeClassic,
		Icon:         assets.ImageAchievementTurretDamage,
		NeedsVictory: true,
	},
	{
		Name:         "cheese",
		Mode:         ModeClassic,
		Icon:         assets.ImageAchievementCheese,
		NeedsVictory: true,
		OnlyElite:    true,
	},
	{
		Name:         "leet",
		Mode:         ModeClassic,
		Icon:         assets.ImageAchievementLeet,
		NeedsVictory: true,
	},

	// Arena mode achievements.
	{
		Name:         "antidominator",
		Mode:         ModeArena,
		Icon:         assets.ImageAchievementAntiDominator,
		NeedsVictory: true,
	},
	{
		Name:         "quicksilver",
		Mode:         ModeArena,
		Icon:         assets.ImageAchievementQuicksilver,
		NeedsVictory: true,
	},
	{
		Name:         "infernal",
		Mode:         ModeArena,
		Icon:         assets.ImageAchievementInfernal,
		NeedsVictory: true,
	},

	// Infinite arena mode achievements.
	{
		Name: "infinite",
		Mode: ModeInfArena,
		Icon: assets.ImageAchievementInfinite,
	},

	// Reverse mode achievements.
	{
		Name:         "colonyhunter",
		Mode:         ModeReverse,
		Icon:         assets.ImageAchievementColonyHunter,
		NeedsVictory: true,
	},
	{
		Name:         "groundcontrol",
		Mode:         ModeReverse,
		Icon:         assets.ImageAchievementGroundControl,
		NeedsVictory: true,
	},
	{
		Name:         "atomicfinisher",
		Mode:         ModeReverse,
		Icon:         assets.ImageAchievementAtomicFinisher,
		NeedsVictory: true,
	},
	{
		Name:         "coordinator",
		Mode:         ModeReverse,
		Icon:         assets.ImageAchievementCoordinator,
		NeedsVictory: true,
	},
	{
		Name:         "siege",
		Mode:         ModeReverse,
		Icon:         assets.ImageAchievementSiege,
		NeedsVictory: true,
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
	{
		Name: "spectator",
		Mode: ModeUnknown,
		Icon: assets.ImageAchievementSpectator,
	},
	{
		Name: "gladiator",
		Mode: ModeUnknown,
		Icon: assets.ImageAchievementGladiator,
	},
}
