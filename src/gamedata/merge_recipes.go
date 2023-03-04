package gamedata

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
var Tier2agentMergeRecipeList = []AgentMergeRecipe{
	{
		Agent1kind:    AgentWorker,
		Agent1faction: GreenFactionTag,
		Agent2kind:    AgentMilitia,
		Agent2faction: RedFactionTag,
		Result:        MortarAgentStats,
	},
	{
		Agent1kind:    AgentMilitia,
		Agent1faction: RedFactionTag,
		Agent2kind:    AgentMilitia,
		Agent2faction: BlueFactionTag,
		Result:        AntiAirAgentStats,
	},
	{
		Agent1kind:    AgentMilitia,
		Agent1faction: YellowFactionTag,
		Agent2kind:    AgentMilitia,
		Agent2faction: BlueFactionTag,
		Result:        PrismAgentStats,
	},
	{
		Agent1kind:    AgentWorker,
		Agent1faction: BlueFactionTag,
		Agent2kind:    AgentWorker,
		Agent2faction: GreenFactionTag,
		Result:        RechargeAgentStats,
	},
	{
		Agent1kind:    AgentWorker,
		Agent1faction: YellowFactionTag,
		Agent2kind:    AgentWorker,
		Agent2faction: GreenFactionTag,
		Result:        FreighterAgentStats,
	},
	{
		Agent1kind:    AgentWorker,
		Agent1faction: RedFactionTag,
		Agent2kind:    AgentWorker,
		Agent2faction: YellowFactionTag,
		Result:        RedminerAgentStats,
	},
	{
		Agent1kind:    AgentMilitia,
		Agent1faction: RedFactionTag,
		Agent2kind:    AgentMilitia,
		Agent2faction: GreenFactionTag,
		Result:        FighterAgentStats,
	},
	{
		Agent1kind:    AgentWorker,
		Agent1faction: YellowFactionTag,
		Agent2kind:    AgentWorker,
		Agent2faction: BlueFactionTag,
		Result:        ServoAgentStats,
	},
	{
		Agent1kind:    AgentMilitia,
		Agent1faction: YellowFactionTag,
		Agent2kind:    AgentMilitia,
		Agent2faction: GreenFactionTag,
		Result:        CripplerAgentStats,
	},
	{
		Agent1kind:    AgentWorker,
		Agent1faction: RedFactionTag,
		Agent2kind:    AgentWorker,
		Agent2faction: BlueFactionTag,
		Result:        RepellerAgentStats,
	},
	{
		Agent1kind:    AgentWorker,
		Agent1faction: GreenFactionTag,
		Agent2kind:    AgentMilitia,
		Agent2faction: YellowFactionTag,
		Result:        GeneratorAgentStats,
	},
	{
		Agent1kind:    AgentWorker,
		Agent1faction: BlueFactionTag,
		Agent2kind:    AgentMilitia,
		Agent2faction: GreenFactionTag,
		Result:        RepairAgentStats,
	},
}

var Tier3agentMergeRecipeList = []AgentMergeRecipe{
	{
		Agent1kind: AgentRepeller,
		Agent2kind: AgentFreighter,
		EvoCost:    5,
		Result:     FlamerAgentStats,
	},
	{
		Agent1kind: AgentFighter,
		Agent2kind: AgentFighter,
		EvoCost:    11,
		Result:     DestroyerAgentStats,
	},

	{
		Agent1kind: AgentRecharger,
		Agent2kind: AgentRepair,
		EvoCost:    7,
		Result:     RefresherAgentStats,
	},
}

type AgentMergeRecipe struct {
	Agent1kind    ColonyAgentKind
	Agent1faction FactionTag
	Agent2kind    ColonyAgentKind
	Agent2faction FactionTag
	EvoCost       float64
	Result        *AgentStats
}

func (r *AgentMergeRecipe) Match(x, y RecipeSubject) bool {
	if r.Match1(x) && r.Match2(y) {
		return true
	}
	if r.Match1(y) && r.Match2(x) {
		return true
	}
	return false
}

func (r *AgentMergeRecipe) Match1(s RecipeSubject) bool {
	return r.match(r.Agent1kind, r.Agent1faction, s)
}

func (r *AgentMergeRecipe) Match2(s RecipeSubject) bool {
	return r.match(r.Agent2kind, r.Agent2faction, s)
}

func (r *AgentMergeRecipe) match(kind ColonyAgentKind, faction FactionTag, s RecipeSubject) bool {
	if s.Kind != kind {
		return false
	}
	if faction == NeutralFactionTag {
		return true
	}
	return s.Faction == faction
}

type RecipeSubject struct {
	Kind    ColonyAgentKind
	Faction FactionTag
}

var RecipesIndex = map[RecipeSubject][]AgentMergeRecipe{}

func init() {
	factions := []FactionTag{
		YellowFactionTag,
		RedFactionTag,
		BlueFactionTag,
		GreenFactionTag,
	}
	kinds := []ColonyAgentKind{
		AgentWorker,
		AgentMilitia,
	}
	for _, f := range factions {
		for _, k := range kinds {
			subject := RecipeSubject{Kind: k, Faction: f}
			for _, recipe := range Tier2agentMergeRecipeList {
				if !recipe.Match1(subject) && !recipe.Match2(subject) {
					continue
				}
				RecipesIndex[subject] = append(RecipesIndex[subject], recipe)
			}
		}
	}
}
