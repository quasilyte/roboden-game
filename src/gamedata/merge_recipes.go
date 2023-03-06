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
		Drone1: RecipeSubject{GreenFactionTag, AgentWorker},
		Drone2: RecipeSubject{RedFactionTag, AgentMilitia},
		Result: MortarAgentStats,
	},
	{
		Drone1: RecipeSubject{RedFactionTag, AgentMilitia},
		Drone2: RecipeSubject{BlueFactionTag, AgentMilitia},
		Result: AntiAirAgentStats,
	},
	{
		Drone1: RecipeSubject{YellowFactionTag, AgentMilitia},
		Drone2: RecipeSubject{BlueFactionTag, AgentMilitia},
		Result: PrismAgentStats,
	},
	{
		Drone1: RecipeSubject{BlueFactionTag, AgentWorker},
		Drone2: RecipeSubject{GreenFactionTag, AgentWorker},
		Result: RechargeAgentStats,
	},
	{
		Drone1: RecipeSubject{YellowFactionTag, AgentWorker},
		Drone2: RecipeSubject{GreenFactionTag, AgentWorker},
		Result: FreighterAgentStats,
	},
	{
		Drone1: RecipeSubject{RedFactionTag, AgentWorker},
		Drone2: RecipeSubject{YellowFactionTag, AgentWorker},
		Result: RedminerAgentStats,
	},
	{
		Drone1: RecipeSubject{RedFactionTag, AgentMilitia},
		Drone2: RecipeSubject{GreenFactionTag, AgentMilitia},
		Result: FighterAgentStats,
	},
	{
		Drone1: RecipeSubject{YellowFactionTag, AgentWorker},
		Drone2: RecipeSubject{BlueFactionTag, AgentWorker},
		Result: ServoAgentStats,
	},
	{
		Drone1: RecipeSubject{YellowFactionTag, AgentMilitia},
		Drone2: RecipeSubject{GreenFactionTag, AgentMilitia},
		Result: CripplerAgentStats,
	},
	{
		Drone1: RecipeSubject{RedFactionTag, AgentWorker},
		Drone2: RecipeSubject{BlueFactionTag, AgentWorker},
		Result: RepellerAgentStats,
	},
	{
		Drone1: RecipeSubject{GreenFactionTag, AgentWorker},
		Drone2: RecipeSubject{YellowFactionTag, AgentMilitia},
		Result: GeneratorAgentStats,
	},
	{
		Drone1: RecipeSubject{BlueFactionTag, AgentWorker},
		Drone2: RecipeSubject{GreenFactionTag, AgentMilitia},
		Result: RepairAgentStats,
	},
}

var Tier3agentMergeRecipeList = []AgentMergeRecipe{
	{
		Drone1:  RecipeSubject{Kind: AgentRepeller},
		Drone2:  RecipeSubject{Kind: AgentFreighter},
		EvoCost: 5,
		Result:  FlamerAgentStats,
	},
	{
		Drone1:  RecipeSubject{Kind: AgentFighter},
		Drone2:  RecipeSubject{Kind: AgentFighter},
		EvoCost: 11,
		Result:  DestroyerAgentStats,
	},

	{
		Drone1:  RecipeSubject{Kind: AgentRecharger},
		Drone2:  RecipeSubject{Kind: AgentRepair},
		EvoCost: 7,
		Result:  RefresherAgentStats,
	},
}

type AgentMergeRecipe struct {
	Drone1  RecipeSubject
	Drone2  RecipeSubject
	EvoCost float64
	Result  *AgentStats
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
	return r.match(r.Drone1.Kind, r.Drone1.Faction, s)
}

func (r *AgentMergeRecipe) Match2(s RecipeSubject) bool {
	return r.match(r.Drone2.Kind, r.Drone2.Faction, s)
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
	Faction FactionTag
	Kind    ColonyAgentKind
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
