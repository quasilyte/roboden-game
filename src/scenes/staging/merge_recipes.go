package staging

// Merge usage:
//
// yellow worker +++
// yellow militia +++
// red worker +++
// red militia +++
// green worker ++++
// green militia ++
// blue worker +++
// blue militia +++
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
// repair: red worker + blue militia
// prism: yellow militia + blue militia
//
// Unused:
// yellow worker + red militia
// yellow worker + green militia
// yellow worker + blue militia
// red worker + green worker
// red worker + green militia
// red worker + yellow militia
// green worker + blue militia
// blue worker + red militia
// blue worker + green militia
// blue worker + yellow militia
// yellow militia + red militia
// green militia + blue militia
var tier2agentMergeRecipeList = []agentMergeRecipe{
	{
		agent1kind:    agentWorker,
		agent1faction: greenFactionTag,
		agent2kind:    agentMilitia,
		agent2faction: redFactionTag,
		result:        mortarAgentStats,
	},
	{
		agent1kind:    agentMilitia,
		agent1faction: redFactionTag,
		agent2kind:    agentMilitia,
		agent2faction: blueFactionTag,
		result:        antiAirAgentStats,
	},
	{
		agent1kind:    agentMilitia,
		agent1faction: yellowFactionTag,
		agent2kind:    agentMilitia,
		agent2faction: blueFactionTag,
		result:        prismAgentStats,
	},
	{
		agent1kind:    agentWorker,
		agent1faction: blueFactionTag,
		agent2kind:    agentWorker,
		agent2faction: greenFactionTag,
		result:        rechargeAgentStats,
	},
	{
		agent1kind:    agentWorker,
		agent1faction: yellowFactionTag,
		agent2kind:    agentWorker,
		agent2faction: greenFactionTag,
		result:        freighterAgentStats,
	},
	{
		agent1kind:    agentWorker,
		agent1faction: redFactionTag,
		agent2kind:    agentWorker,
		agent2faction: yellowFactionTag,
		result:        redminerAgentStats,
	},
	{
		agent1kind:    agentMilitia,
		agent1faction: redFactionTag,
		agent2kind:    agentMilitia,
		agent2faction: greenFactionTag,
		result:        fighterAgentStats,
	},
	{
		agent1kind:    agentWorker,
		agent1faction: yellowFactionTag,
		agent2kind:    agentWorker,
		agent2faction: blueFactionTag,
		result:        servoAgentStats,
	},
	{
		agent1kind:    agentMilitia,
		agent1faction: yellowFactionTag,
		agent2kind:    agentMilitia,
		agent2faction: greenFactionTag,
		result:        cripplerAgentStats,
	},
	{
		agent1kind:    agentWorker,
		agent1faction: redFactionTag,
		agent2kind:    agentWorker,
		agent2faction: blueFactionTag,
		result:        repellerAgentStats,
	},
	{
		agent1kind:    agentWorker,
		agent1faction: greenFactionTag,
		agent2kind:    agentMilitia,
		agent2faction: yellowFactionTag,
		result:        generatorAgentStats,
	},
	{
		agent1kind:    agentWorker,
		agent1faction: redFactionTag,
		agent2kind:    agentMilitia,
		agent2faction: blueFactionTag,
		result:        repairAgentStats,
	},
}

var tier3agentMergeRecipeList = []agentMergeRecipe{
	{
		agent1kind: agentRepeller,
		agent2kind: agentFreighter,
		evoCost:    5,
		result:     flamerAgentStats,
	},
	{
		agent1kind: agentFighter,
		agent2kind: agentFighter,
		evoCost:    11,
		result:     destroyerAgentStats,
	},

	{
		agent1kind: agentRecharger,
		agent2kind: agentRepair,
		evoCost:    7,
		result:     refresherAgentStats,
	},
}
