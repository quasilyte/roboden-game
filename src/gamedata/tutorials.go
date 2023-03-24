package gamedata

type TutorialData struct {
	ID          int
	ScoreReward int
}

var Tutorials = []*TutorialData{
	{
		ID:          0,
		ScoreReward: 200,
	},
	{
		ID:          1,
		ScoreReward: 200,
	},
	{
		ID:          2,
		ScoreReward: 400,
	},
	{
		ID:          3,
		ScoreReward: 350,
	},
}
