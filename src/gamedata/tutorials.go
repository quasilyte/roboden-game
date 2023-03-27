package gamedata

type TutorialData struct {
	ID          int
	ScoreReward int

	Seed int64

	Tier2Drones   []AgentMergeRecipe
	ExtraDrones   []*AgentStats
	NumEnemyBases int

	CanBuildTurrets bool
	CanAttack       bool
	CanChangeRadius bool
	Boss            bool

	RedCrystals   bool
	InitialCreeps int
	Resources     int
	WorldSize     int

	SecondBase bool

	Objective    GameObjective
	ObjectiveKey string
}

var Tutorials = []*TutorialData{
	{
		ID:          0,
		ScoreReward: 200,
		Objective:   ObjectiveBuildBase,
		Seed:        0xF0F1000 + 10,
	},

	{
		ID:              1,
		ScoreReward:     300,
		CanChangeRadius: true,
		Tier2Drones: []AgentMergeRecipe{
			FindRecipe(ClonerAgentStats),
			FindRecipe(FighterAgentStats),
		},
		Objective:    ObjectiveTrigger,
		ObjectiveKey: "objective.acquire_destroyer",
		Seed:         0xF0F2000 + 9,
	},

	{
		ID:              2,
		ScoreReward:     350,
		Resources:       1,
		NumEnemyBases:   1,
		CanAttack:       true,
		CanChangeRadius: true,
		SecondBase:      true,
		ExtraDrones: []*AgentStats{
			DestroyerAgentStats,
		},
		Tier2Drones: []AgentMergeRecipe{
			FindRecipe(ClonerAgentStats),
			FindRecipe(FighterAgentStats),
			FindRecipe(RepairAgentStats),
			FindRecipe(FreighterAgentStats),
			FindRecipe(ServoAgentStats),
		},
		Objective: ObjectiveDestroyCreepBases,
		Seed:      0xF0F3000 + 1,
	},

	{
		ID:              3,
		ScoreReward:     650,
		WorldSize:       1,
		Resources:       2,
		InitialCreeps:   1,
		Boss:            true,
		RedCrystals:     true,
		CanChangeRadius: true,
		CanAttack:       true,
		CanBuildTurrets: true,
		ExtraDrones: []*AgentStats{
			ServoAgentStats,
		},
		Tier2Drones: []AgentMergeRecipe{
			FindRecipe(ClonerAgentStats),
			FindRecipe(FighterAgentStats),
			FindRecipe(RepairAgentStats),
			FindRecipe(FreighterAgentStats),
			FindRecipe(CripplerAgentStats),
			FindRecipe(RedminerAgentStats),
			FindRecipe(ServoAgentStats),
		},
		Objective: ObjectiveBoss,
		Seed:      0xF0F5000 + 4,
	},
}
