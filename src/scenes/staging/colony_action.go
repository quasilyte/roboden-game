package staging

type colonyPriority int

const (
	priorityResources colonyPriority = iota
	priorityGrowth
	priorityEvolution
	prioritySecurity
)

type colonyActionKind int

type colonyAction struct {
	Kind     colonyActionKind
	Value    any
	Value2   any
	Value3   float64
	TimeCost float64
}

const (
	actionNone colonyActionKind = iota
	actionRecycleAgent
	actionGenerateEvo
	actionMineEssence
	actionSendCourier
	actionCloneAgent
	actionProduceAgent
	actionSetPatrol
	actionDefenceGarrison
	actionDefencePatrol
	actionDefencePanic
	actionMergeAgents
	actionRepairBase
	actionRepairTurret
	actionBuildBuilding
	actionGetReinforcements
)
