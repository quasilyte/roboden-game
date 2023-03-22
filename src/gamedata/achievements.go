package gamedata

type Achievement struct {
	Name string
	Mode string
}

const (
	ModeClassic = "classic"
)

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
