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

	ModeTutorial

	ModeAny
)

func (m Mode) String() string {
	switch m {
	case ModeClassic:
		return "classic"
	case ModeArena:
		return "arena"
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
	{
		Name: "infinite",
		Mode: ModeArena,
		Icon: assets.ImageAchievementInfinite,
	},
}
