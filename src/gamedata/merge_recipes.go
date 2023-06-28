package gamedata

import "fmt"

// Merge usage:
//
// yellow worker +++++
// yellow scout  ++++++
// red worker    ++++++
// red scout     +++++++
// green worker  +++++
// green scout   ++++++
// blue worker   +++++
// blue scout    ++++++
//
// Used:
// mortar: green worker + red scout
// antiair: red scout + blue scout
// recharger: green worker + blue worker
// freighter: yellow worker + green worker
// redminer: yellow worker + red worker
// fighter: red scout + green scout
// servo: yellow worker + blue worker
// crippler: yellow scout + green scout
// repeller: red worker + blue worker
// generator: green worker + yellow scout
// repair: blue worker + green scout
// prism: yellow scout + blue scout
// cloner: red worker + blue scout
// scavenger: red worker + yellow scout
// courier: red worker + green scout
// disintegrator: yellow worker + blue scout
// defender: yellow scout + red scout
// kamikaze: blue worker + blue scout [! a non-standard combination]
// roomba: red scout + red scout [! a non-standard combination]
// skirmisher: green scout + blue scout
// scarab: yellow worker + red scout
// commander: red worker + yellow scout
// targeter: green worker + green scout [! a non-standard combination]
//
// Unused:
// yellow worker + green scout
// red worker + green worker
// red worker + green scout
// green worker + blue scout
// blue worker + red scout
// blue worker + green scout
// blue worker + yellow scout
var Tier2agentMergeRecipes = []AgentMergeRecipe{
	{
		Drone1: RecipeSubject{RedFactionTag, AgentWorker},
		Drone2: RecipeSubject{BlueFactionTag, AgentScout},
		Result: ClonerAgentStats,
	},
	{
		Drone1: RecipeSubject{RedFactionTag, AgentScout},
		Drone2: RecipeSubject{GreenFactionTag, AgentScout},
		Result: FighterAgentStats,
	},
	{
		Drone1: RecipeSubject{BlueFactionTag, AgentWorker},
		Drone2: RecipeSubject{GreenFactionTag, AgentScout},
		Result: RepairAgentStats,
	},
	{
		Drone1: RecipeSubject{BlueFactionTag, AgentWorker},
		Drone2: RecipeSubject{GreenFactionTag, AgentWorker},
		Result: RechargerAgentStats,
	},
	{
		Drone1: RecipeSubject{YellowFactionTag, AgentScout},
		Drone2: RecipeSubject{GreenFactionTag, AgentScout},
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
		Drone2: RecipeSubject{YellowFactionTag, AgentScout},
		Result: ScavengerAgentStats,
	},
	{
		Drone1: RecipeSubject{RedFactionTag, AgentWorker},
		Drone2: RecipeSubject{GreenFactionTag, AgentScout},
		Result: CourierAgentStats,
	},
	{
		Drone1: RecipeSubject{YellowFactionTag, AgentWorker},
		Drone2: RecipeSubject{GreenFactionTag, AgentWorker},
		Result: FreighterAgentStats,
	},
	{
		Drone1: RecipeSubject{RedFactionTag, AgentWorker},
		Drone2: RecipeSubject{BlueFactionTag, AgentWorker},
		Result: RepellerAgentStats,
	},
	{
		Drone1: RecipeSubject{GreenFactionTag, AgentWorker},
		Drone2: RecipeSubject{YellowFactionTag, AgentScout},
		Result: GeneratorAgentStats,
	},
	{
		Drone1: RecipeSubject{RedFactionTag, AgentScout},
		Drone2: RecipeSubject{RedFactionTag, AgentScout},
		Result: RoombaAgentStats,
	},
	{
		Drone1: RecipeSubject{GreenFactionTag, AgentWorker},
		Drone2: RecipeSubject{RedFactionTag, AgentScout},
		Result: MortarAgentStats,
	},
	{
		Drone1: RecipeSubject{RedFactionTag, AgentScout},
		Drone2: RecipeSubject{BlueFactionTag, AgentScout},
		Result: AntiAirAgentStats,
	},
	{
		Drone1: RecipeSubject{YellowFactionTag, AgentWorker},
		Drone2: RecipeSubject{BlueFactionTag, AgentScout},
		Result: DisintegratorAgentStats,
	},
	{
		Drone1: RecipeSubject{RedFactionTag, AgentWorker},
		Drone2: RecipeSubject{YellowFactionTag, AgentScout},
		Result: CommanderAgentStats,
	},
	{
		Drone1: RecipeSubject{YellowFactionTag, AgentScout},
		Drone2: RecipeSubject{BlueFactionTag, AgentScout},
		Result: PrismAgentStats,
	},
	{
		Drone1: RecipeSubject{GreenFactionTag, AgentWorker},
		Drone2: RecipeSubject{GreenFactionTag, AgentScout},
		Result: TargeterAgentStats,
	},
	{
		Drone1: RecipeSubject{YellowFactionTag, AgentScout},
		Drone2: RecipeSubject{RedFactionTag, AgentScout},
		Result: DefenderAgentStats,
	},
	{
		Drone1: RecipeSubject{BlueFactionTag, AgentWorker},
		Drone2: RecipeSubject{BlueFactionTag, AgentScout},
		Result: KamikazeAgentStats,
	},
	{
		Drone1: RecipeSubject{GreenFactionTag, AgentScout},
		Drone2: RecipeSubject{BlueFactionTag, AgentScout},
		Result: SkirmisherAgentStats,
	},
	{
		Drone1: RecipeSubject{YellowFactionTag, AgentWorker},
		Drone2: RecipeSubject{RedFactionTag, AgentScout},
		Result: ScarabAgentStats,
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
		Drone1:  RecipeSubject{Kind: AgentScarab},
		Drone2:  RecipeSubject{Kind: AgentScarab},
		EvoCost: 11,
		Result:  DevourerAgentStats,
	},
	{
		Drone1:  RecipeSubject{Kind: AgentScavenger},
		Drone2:  RecipeSubject{Kind: AgentCrippler},
		EvoCost: 10,
		Result:  MarauderAgentStats,
	},
	{
		Drone1:  RecipeSubject{Kind: AgentSkirmisher},
		Drone2:  RecipeSubject{Kind: AgentDefender},
		EvoCost: 8,
		Result:  GuardianAgentStats,
	},
}

type AgentMergeRecipe struct {
	Drone1  RecipeSubject
	Drone2  RecipeSubject
	EvoCost float64
	Result  *AgentStats
}

func FindRecipeByName(droneName string) AgentMergeRecipe {
	r := findRecipeByName(droneName)
	if r.Result != nil {
		return r
	}
	panic(fmt.Sprintf("requested a non-existing recipe: %s", droneName))
}

func findRecipeByName(droneName string) AgentMergeRecipe {
	for _, r := range Tier2agentMergeRecipes {
		if r.Result.Kind.String() == droneName {
			return r
		}
	}
	for _, r := range Tier3agentMergeRecipes {
		if r.Result.Kind.String() == droneName {
			return r
		}
	}
	return AgentMergeRecipe{}
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
