package gamedata

type UnitSize int

const (
	SizeSmall UnitSize = iota
	SizeMedium
	SizeLarge
)

type ColonyAgentKind uint8

const (
	AgentWorker ColonyAgentKind = iota
	AgentMilitia

	// Tier2
	AgentFreighter
	AgentRedminer
	AgentCrippler
	AgentFighter
	AgentPrism
	AgentServo
	AgentRepeller
	AgentRepair
	AgentRecharger
	AgentGenerator
	AgentMortar
	AgentAntiAir

	// Tier3
	AgentRefresher
	AgentFlamer
	AgentDestroyer

	AgentKindNum

	// Buildings (not real agents/drones)
	AgentGunpoint
)
