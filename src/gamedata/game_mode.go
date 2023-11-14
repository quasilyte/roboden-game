package gamedata

type GameModeInfo struct {
	ScoreCost int
}

var GameModeInfoMap = map[string]GameModeInfo{
	"classic":   {ScoreCost: ClassicModeCost},
	"blitz":     {ScoreCost: BlitzModeCost},
	"arena":     {ScoreCost: ArenaModeCost},
	"inf_arena": {ScoreCost: InfArenaModeCost},
	"reverse":   {ScoreCost: ReverseModeCost},
}
