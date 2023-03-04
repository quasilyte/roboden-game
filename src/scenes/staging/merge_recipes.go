package staging

import "github.com/quasilyte/roboden-game/gamedata"

// Merge usage:
//
// yellow worker +++
// yellow militia +++
// red worker ++
// red militia +++
// green worker ++++
// green militia +++
// blue worker ++++
// blue militia ++
//
// Used:
// mortar: green worker + red militia
// antiair: red militia + blue militia
// recharger: green worker + blue worker
// freighter: yellow worker + green worker
// redminer: yellow worker + red worker
// fighter: red militia + green militia
// servo: yellow worker + blue worker
// crippler: yellow militia + green militia
// repeller: red worker + blue worker
// generator: green worker + yellow militia
// repair: blue worker + green militia
// prism: yellow militia + blue militia
//
// Unused:
// yellow worker + red militia
// yellow worker + green militia
// yellow worker + blue militia
// red worker + green worker
// red worker + green militia
// red worker + yellow militia
// red worker + blue militia
// green worker + blue militia
// blue worker + red militia
// blue worker + green militia
// blue worker + yellow militia
// yellow militia + red militia
// green militia + blue militia
var tier2agentMergeRecipeList = []agentMergeRecipe{
	{
		agent1kind:    gamedata.AgentWorker,
		agent1faction: greenFactionTag,
		agent2kind:    gamedata.AgentMilitia,
		agent2faction: redFactionTag,
		result:        gamedata.MortarAgentStats,
	},
	{
		agent1kind:    gamedata.AgentMilitia,
		agent1faction: redFactionTag,
		agent2kind:    gamedata.AgentMilitia,
		agent2faction: blueFactionTag,
		result:        gamedata.AntiAirAgentStats,
	},
	{
		agent1kind:    gamedata.AgentMilitia,
		agent1faction: yellowFactionTag,
		agent2kind:    gamedata.AgentMilitia,
		agent2faction: blueFactionTag,
		result:        gamedata.PrismAgentStats,
	},
	{
		agent1kind:    gamedata.AgentWorker,
		agent1faction: blueFactionTag,
		agent2kind:    gamedata.AgentWorker,
		agent2faction: greenFactionTag,
		result:        gamedata.RechargeAgentStats,
	},
	{
		agent1kind:    gamedata.AgentWorker,
		agent1faction: yellowFactionTag,
		agent2kind:    gamedata.AgentWorker,
		agent2faction: greenFactionTag,
		result:        gamedata.FreighterAgentStats,
	},
	{
		agent1kind:    gamedata.AgentWorker,
		agent1faction: redFactionTag,
		agent2kind:    gamedata.AgentWorker,
		agent2faction: yellowFactionTag,
		result:        gamedata.RedminerAgentStats,
	},
	{
		agent1kind:    gamedata.AgentMilitia,
		agent1faction: redFactionTag,
		agent2kind:    gamedata.AgentMilitia,
		agent2faction: greenFactionTag,
		result:        gamedata.FighterAgentStats,
	},
	{
		agent1kind:    gamedata.AgentWorker,
		agent1faction: yellowFactionTag,
		agent2kind:    gamedata.AgentWorker,
		agent2faction: blueFactionTag,
		result:        gamedata.ServoAgentStats,
	},
	{
		agent1kind:    gamedata.AgentMilitia,
		agent1faction: yellowFactionTag,
		agent2kind:    gamedata.AgentMilitia,
		agent2faction: greenFactionTag,
		result:        gamedata.CripplerAgentStats,
	},
	{
		agent1kind:    gamedata.AgentWorker,
		agent1faction: redFactionTag,
		agent2kind:    gamedata.AgentWorker,
		agent2faction: blueFactionTag,
		result:        gamedata.RepellerAgentStats,
	},
	{
		agent1kind:    gamedata.AgentWorker,
		agent1faction: greenFactionTag,
		agent2kind:    gamedata.AgentMilitia,
		agent2faction: yellowFactionTag,
		result:        gamedata.GeneratorAgentStats,
	},
	{
		agent1kind:    gamedata.AgentWorker,
		agent1faction: blueFactionTag,
		agent2kind:    gamedata.AgentMilitia,
		agent2faction: greenFactionTag,
		result:        gamedata.RepairAgentStats,
	},
}

var tier3agentMergeRecipeList = []agentMergeRecipe{
	{
		agent1kind: gamedata.AgentRepeller,
		agent2kind: gamedata.AgentFreighter,
		evoCost:    5,
		result:     gamedata.FlamerAgentStats,
	},
	{
		agent1kind: gamedata.AgentFighter,
		agent2kind: gamedata.AgentFighter,
		evoCost:    11,
		result:     gamedata.DestroyerAgentStats,
	},

	{
		agent1kind: gamedata.AgentRecharger,
		agent2kind: gamedata.AgentRepair,
		evoCost:    7,
		result:     gamedata.RefresherAgentStats,
	},
}
