package gamedata

import "fmt"

// Merge usage:
//
// yellow worker +++
// yellow militia +++
// red worker +++
// red militia +++
// green worker ++++
// green militia +++
// blue worker ++++
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
// repair: blue worker + green militia
// prism: yellow militia + blue militia
// cloner: red worker + blue militia
// scavenger: red worker + yellow militia
// courier: red worker + green militia
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
var Tier2agentMergeRecipes = []AgentMergeRecipe{
	{
		Drone1: RecipeSubject{RedFactionTag, AgentWorker},
		Drone2: RecipeSubject{BlueFactionTag, AgentMilitia},
		Result: ClonerAgentStats,
	},
	{
		Drone1: RecipeSubject{RedFactionTag, AgentMilitia},
		Drone2: RecipeSubject{GreenFactionTag, AgentMilitia},
		Result: FighterAgentStats,
	},
	{
		Drone1: RecipeSubject{BlueFactionTag, AgentWorker},
		Drone2: RecipeSubject{GreenFactionTag, AgentMilitia},
		Result: RepairAgentStats,
	},
	{
		Drone1: RecipeSubject{YellowFactionTag, AgentWorker},
		Drone2: RecipeSubject{GreenFactionTag, AgentWorker},
		Result: FreighterAgentStats,
	},
	{
		Drone1: RecipeSubject{YellowFactionTag, AgentMilitia},
		Drone2: RecipeSubject{GreenFactionTag, AgentMilitia},
		Result: CripplerAgentStats,
	},
	{
		Drone1: RecipeSubject{RedFactionTag, AgentWorker},
		Drone2: RecipeSubject{YellowFactionTag, AgentWorker},
		Result: RedminerAgentStats,
	},
	{
		Drone1: RecipeSubject{YellowFactionTag, AgentWorker},
		Drone2: RecipeSubject{BlueFactionTag, AgentWorker},
		Result: ServoAgentStats,
	},
	{
		Drone1: RecipeSubject{RedFactionTag, AgentWorker},
		Drone2: RecipeSubject{YellowFactionTag, AgentMilitia},
		Result: ScavengerAgentStats,
	},
	{
		Drone1: RecipeSubject{RedFactionTag, AgentWorker},
		Drone2: RecipeSubject{GreenFactionTag, AgentMilitia},
		Result: CourierAgentStats,
	},
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
		Drone1: RecipeSubject{RedFactionTag, AgentWorker},
		Drone2: RecipeSubject{BlueFactionTag, AgentWorker},
		Result: RepellerAgentStats,
	},
	{
		Drone1: RecipeSubject{GreenFactionTag, AgentWorker},
		Drone2: RecipeSubject{YellowFactionTag, AgentMilitia},
		Result: GeneratorAgentStats,
	},
}

var Tier3agentMergeRecipes = []AgentMergeRecipe{
	{
		Drone1:  RecipeSubject{Kind: AgentRepeller},
		Drone2:  RecipeSubject{Kind: AgentGenerator},
		EvoCost: 8,
		Result:  StormbringerAgentStats,
	},
	{
		Drone1:  RecipeSubject{Kind: AgentFreighter},
		Drone2:  RecipeSubject{Kind: AgentCourier},
		EvoCost: 8,
		Result:  TruckerAgentStats,
	},
	{
		Drone1:  RecipeSubject{Kind: AgentFighter},
		Drone2:  RecipeSubject{Kind: AgentFighter},
		EvoCost: 11,
		Result:  DestroyerAgentStats,
	},
	{
		Drone1:  RecipeSubject{Kind: AgentScavenger},
		Drone2:  RecipeSubject{Kind: AgentCrippler},
		EvoCost: 10,
		Result:  MarauderAgentStats,
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

func FindRecipe(stats *AgentStats) AgentMergeRecipe {
	var slice []AgentMergeRecipe
	if stats.Tier == 2 {
		slice = Tier2agentMergeRecipes
	} else {
		slice = Tier3agentMergeRecipes
	}
	for _, r := range slice {
		if r.Result == stats {
			return r
		}
	}
	panic(fmt.Sprintf("requested a non-existing recipe: %s", stats.Kind.String()))
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

// var RecipesIndex = map[RecipeSubject][]AgentMergeRecipe{}

// func init() {
// 	factions := []FactionTag{
// 		YellowFactionTag,
// 		RedFactionTag,
// 		BlueFactionTag,
// 		GreenFactionTag,
// 	}
// 	kinds := []ColonyAgentKind{
// 		AgentWorker,
// 		AgentMilitia,
// 	}
// 	for _, f := range factions {
// 		for _, k := range kinds {
// 			subject := RecipeSubject{Kind: k, Faction: f}
// 			for _, recipe := range Tier2agentMergeRecipeList {
// 				if !recipe.Match1(subject) && !recipe.Match2(subject) {
// 					continue
// 				}
// 				RecipesIndex[subject] = append(RecipesIndex[subject], recipe)
// 			}
// 		}
// 	}
// }
