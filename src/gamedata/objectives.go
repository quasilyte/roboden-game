package gamedata

type GameObjective int

const (
	ObjectiveBoss GameObjective = iota
	ObjectiveBuildBase
	ObjectiveDestroyCreepBases
	ObjectiveAcquireSuperElite
	ObjectiveTrigger
)

func (o GameObjective) String() string {
	switch o {
	case ObjectiveBoss:
		return "boss"
	case ObjectiveBuildBase:
		return "build_base"
	case ObjectiveDestroyCreepBases:
		return "destroy_creep_bases"
	case ObjectiveAcquireSuperElite:
		return "super_elite"
	default:
		return ""
	}
}
