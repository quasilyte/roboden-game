package staging

//go:generate stringer -type=colonyPriority -trimprefix=priority
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
	Value4   int
	TimeCost float64
}

const (
	actionNone colonyActionKind = iota
	actionRecycleAgent
	actionGenerateEvo
	actionConvertEvo
	actionMineEssence
	actionGrabArtifact
	actionMineSulfurEssence
	actionSendCourier
	actionCloneAgent
	actionProduceAgent
	actionSetPatrol
	actionAttachToCommander
	actionDefenceGarrison
	actionDefencePatrol
	actionDefencePanic
	actionMergeAgents
	actionRepairBase
	actionRepairTurret
	actionBuildBuilding
	actionGetReinforcements
	actionCaptureBuilding
)
