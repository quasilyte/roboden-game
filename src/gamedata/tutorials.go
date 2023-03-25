package gamedata

type TutorialData struct {
	ID          int
	ScoreReward int

	Seed int64

	Tier2Drones   []AgentMergeRecipe
	NumEnemyBases int

	CanBuildTurrets bool
	CanAttack       bool
	Boss            bool

	RedCrystals   bool
	InitialCreeps int
	Resources     int
	WorldSize     int

	Objective GameObjective
}

var Tutorials = []*TutorialData{
	{
		ID:            0,
		ScoreReward:   200,
		NumEnemyBases: 0,
		Objective:     ObjectiveBuildBase,
		Seed:          0xF0F1000 + 6,
	},

	{
		ID:            1,
		ScoreReward:   250,
		NumEnemyBases: 0,
		Tier2Drones: []AgentMergeRecipe{
			FindRecipe(ClonerAgentStats),
			FindRecipe(FighterAgentStats),
			FindRecipe(FreighterAgentStats),
			FindRecipe(ServoAgentStats),
		},
		Objective: ObjectiveAcquireDestroyer,
		Seed:      0xF0F2000 + 9,
	},

	{
		ID:            2,
		ScoreReward:   400,
		NumEnemyBases: 1,
		CanAttack:     true,
		Tier2Drones: []AgentMergeRecipe{
			FindRecipe(ClonerAgentStats),
			FindRecipe(FighterAgentStats),
			FindRecipe(RepairAgentStats),
			FindRecipe(FreighterAgentStats),
			FindRecipe(ServoAgentStats),
		},
		Objective: ObjectiveDestroyCreepBases,
		Seed:      0xF0F3000 + 0,
	},

	{
		ID:            3,
		ScoreReward:   450,
		NumEnemyBases: 1,
		WorldSize:     1,
		Resources:     2,
		RedCrystals:   true,
		Tier2Drones: []AgentMergeRecipe{
			FindRecipe(ClonerAgentStats),
			FindRecipe(FighterAgentStats),
			FindRecipe(RepairAgentStats),
			FindRecipe(FreighterAgentStats),
			FindRecipe(CripplerAgentStats),
			FindRecipe(RedminerAgentStats),
			FindRecipe(ServoAgentStats),
		},
		Objective: ObjectiveAcquireSuperElite,
		Seed:      0xF0F4000 + 1,
	},

	{
		ID:            4,
		ScoreReward:   700,
		WorldSize:     1,
		Resources:     2,
		InitialCreeps: 1,
		RedCrystals:   true,
		Boss:          true,
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
		Seed:      0xF0F5000 + 3,
	},
}
