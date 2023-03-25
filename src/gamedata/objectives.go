package gamedata

type GameObjective int

const (
	ObjectiveBoss GameObjective = iota
	ObjectiveBuildBase
	ObjectiveAcquireDestroyer
	ObjectiveDestroyCreepBases
	ObjectiveAcquireSuperElite
)

func (o GameObjective) String() string {
	switch o {
	case ObjectiveBoss:
		return "boss"
	case ObjectiveBuildBase:
		return "build_base"
	case ObjectiveAcquireDestroyer:
		return "acquire_destroyer"
	case ObjectiveDestroyCreepBases:
		return "destroy_creep_bases"
	case ObjectiveAcquireSuperElite:
		return "super_elite"
	default:
		return ""
	}
}
