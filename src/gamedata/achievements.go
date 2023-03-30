package gamedata

type Achievement struct {
	Name string
	Mode Mode
}

type Mode int

const (
	ModeClassic Mode = iota
	ModeArena

	ModeTutorial
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
	{
		Name: "impossible",
		Mode: ModeClassic,
	},
	{
		Name: "cheapbuild10",
		Mode: ModeClassic,
	},
	{
		Name: "hightension",
		Mode: ModeClassic,
	},
	{
		Name: "solobase",
		Mode: ModeClassic,
	},
	{
		Name: "uiless",
		Mode: ModeClassic,
	},
	{
		Name: "powerof3",
		Mode: ModeClassic,
	},
	{
		Name: "tinyradius",
		Mode: ModeClassic,
	},
	{
		Name: "t1army",
		Mode: ModeClassic,
	},
	{
		Name: "groundwin",
		Mode: ModeClassic,
	},
	{
		Name: "speedrunning",
		Mode: ModeClassic,
	},
	{
		Name: "victorydrag",
		Mode: ModeClassic,
	},
	{
		Name: "t3engineer",
		Mode: ModeClassic,
	},
	{
		Name: "t3less",
		Mode: ModeClassic,
	},
	{
		Name: "turretdamage",
		Mode: ModeClassic,
	},
}
