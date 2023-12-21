package gamedata

type GameModeInfo struct {
	ScoreCost int
}

var GameModeInfoMap = map[string]GameModeInfo{
	"blitz":     {ScoreCost: 0},
	"classic":   {ScoreCost: ClassicModeCost},
	"arena":     {ScoreCost: ArenaModeCost},
	"reverse":   {ScoreCost: ReverseModeCost},
	"inf_arena": {ScoreCost: InfArenaModeCost},
}
