package staging

import (
	"math"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/input"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/gsignal"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/session"
)

// This tutorial system is not very elegant.
// Instead of having an event-based system it can subscribe to,
// it has to query the game state and compare it with its expectations.
// To avoid too much redundant computations, we only do that once in a while
// with a randomized jitter.
// Also, the tutorial objects can't describe the interactive hints
// in a declarative way, so we'll have to hardcode every one of
// them here in the most adhoc way possible.

type tutorialRunner struct {
	updateFunc func() bool
}

type tutorialManager struct {
	input *input.Handler

	scene *ge.Scene

	inputMode string

	choice selectedChoice

	world        *worldState
	config       *session.LevelConfig
	tutorialStep int
	drone        *colonyAgentNode
	creep        *creepNode
	resource     *essenceSourceNode
	stepTicks    int

	explainedResourcePool  bool
	explainedWorker        bool
	explainedMilitia       bool
	explainedServoBots     bool
	explainedRepairBots    bool
	explainedFreighterBots bool
	explainedSuperElites   bool

	hint      *messageNode
	timedHint *messageNode

	updateDelay float64

	runner *tutorialRunner

	EventRequestPanelUpdate gsignal.Event[gsignal.Void]
	EventTriggerVictory     gsignal.Event[gsignal.Void]
}

func newTutorialManager(h *input.Handler, world *worldState) *tutorialManager {
	return &tutorialManager{
		input:       h,
		world:       world,
		config:      world.config,
		updateDelay: 2,
	}
}

func (m *tutorialManager) Init(scene *ge.Scene) {
	m.scene = scene

	m.inputMode = "keyboard"

	runners := [...]tutorialRunner{
		{
			updateFunc: m.updateTutorial1,
		},

		{
			updateFunc: m.updateTutorial2,
		},

		{
			updateFunc: m.updateTutorial3,
		},

		{
			updateFunc: m.updateTutorial4,
		},
	}

	m.runner = &runners[m.config.Tutorial.ID]
}

func (m *tutorialManager) IsDisposed() bool {
	return false
}

func (m *tutorialManager) Update(delta float64) {
	if m.drone != nil && m.drone.IsDisposed() {
		m.drone = nil
	}
	if m.creep != nil && m.creep.IsDisposed() {
		m.creep = nil
	}
	if m.timedHint != nil && m.timedHint.IsDisposed() {
		m.timedHint = nil
	}

	m.updateDelay = gmath.ClampMin(m.updateDelay-delta, 0)
	if m.updateDelay != 0 {
		return
	}
	m.updateDelay = m.scene.Rand().FloatRange(1.5, 4.5)

	m.stepTicks = gmath.ClampMin(m.stepTicks-1, 0)

	m.choice = selectedChoice{}
	m.runUpdateFunc()
}

func (m *tutorialManager) runUpdateFunc() {
	if len(m.world.colonies) == 0 {
		return
	}
	hintOpen := m.hint != nil
	if m.runner.updateFunc() {
		m.tutorialStep++
		if hintOpen && m.hint != nil {
			m.hint.Dispose()
			m.hint = nil
		}
	}
}

func (m *tutorialManager) OnChoice(choice selectedChoice) {
	m.choice = choice
	m.runUpdateFunc()
}

func (m *tutorialManager) updateTutorial1() bool {
	if !m.explainedResourcePool && m.world.colonies[0].resources > 100 && m.timedHint == nil {
		m.explainedResourcePool = true
		s := m.scene.Dict().Get("tutorial1.resource_bar")
		targetPos := ge.Pos{Base: &m.world.colonies[0].spritePos, Offset: gmath.Vec{X: -3, Y: 18}}
		m.timedHint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 148}, targetPos, s)
		m.timedHint.trackedObject = m.world.colonies[0]
		m.timedHint.timed = true
		m.timedHint.time = 18
		m.scene.AddObject(m.timedHint)
	}
	if !m.explainedWorker && m.timedHint == nil && m.tutorialStep >= 19 {
		worker := m.findDrone(searchWorkers, func(a *colonyAgentNode) bool {
			return a.stats.Kind == gamedata.AgentWorker
		})
		if worker != nil {
			m.explainDrone(worker, "tutorial1.hint_worker")
			m.explainedWorker = true
		}
	}
	if !m.explainedMilitia && m.timedHint == nil && m.tutorialStep >= 19 {
		militia := m.findDrone(searchFighters, func(a *colonyAgentNode) bool {
			return a.stats.Kind == gamedata.AgentMilitia
		})
		if militia != nil {
			m.explainDrone(militia, "tutorial1.hint_militia")
			m.explainedMilitia = true
		}
	}

	switch m.tutorialStep {
	case 0:
		s := m.scene.Dict().Get("tutorial1.your_colony")
		targetPos := ge.Pos{Base: &m.world.colonies[0].spritePos}
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.scene.AddObject(m.hint)
		m.stepTicks = 4
		return true
	case 1:
		return m.stepTicks == 0
	case 2:
		// Explain the action cards.
		s := m.scene.Dict().Get("tutorial1.action_cards", m.inputMode)
		targetPos := gmath.Vec{X: 812, Y: 50}
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.scene.AddObject(m.hint)
		return true
	case 3:
		ok := (m.choice.Option.special != specialChoiceNone && m.choice.Option.special != specialChoiceMoveColony) ||
			(m.choice.Option.special == specialChoiceNone && len(m.choice.Option.effects) != 0)
		if ok {
			return true
		}
	case 4:
		// Explain the resources priority.
		s := m.scene.Dict().Get("tutorial1.resources_priority")
		targetPos := gmath.Vec{X: 812 + (36 * 0), Y: 516}
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.scene.AddObject(m.hint)
		return true
	case 5:
		for _, effect := range m.choice.Option.effects {
			if effect.priority == priorityResources {
				return true
			}
		}
	case 6:
		// Explain the growth priority.
		s := m.scene.Dict().Get("tutorial1.growth_priority")
		targetPos := gmath.Vec{X: 812 + (36 * 1), Y: 516}
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.scene.AddObject(m.hint)
		return true
	case 7:
		for _, effect := range m.choice.Option.effects {
			if effect.priority == priorityGrowth {
				return true
			}
		}
	case 8:
		// Explain the evolution priority.
		s := m.scene.Dict().Get("tutorial1.evolution_priority")
		targetPos := gmath.Vec{X: 812 + (36 * 2), Y: 516}
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.scene.AddObject(m.hint)
		return true
	case 9:
		for _, effect := range m.choice.Option.effects {
			if effect.priority == priorityEvolution {
				return true
			}
		}
	case 10:
		// Explain the security priority.
		s := m.scene.Dict().Get("tutorial1.security_priority")
		targetPos := gmath.Vec{X: 812 + (36 * 3), Y: 516}
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.scene.AddObject(m.hint)
		return true
	case 11:
		for _, effect := range m.choice.Option.effects {
			if effect.priority == prioritySecurity {
				return true
			}
		}
	case 12:
		s := m.scene.Dict().Get("tutorial1.camera", m.inputMode)
		targetPos := ge.Pos{Offset: gmath.Vec{X: 1160, Y: 530}}
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.scene.AddObject(m.hint)
		return true
	case 13:
		targetPos := ge.Pos{Offset: gmath.Vec{X: 340, Y: 780}}
		cameraCenter := m.world.camera.Offset.Add(m.world.camera.Rect.Center())
		if targetPos.Resolve().DistanceSquaredTo(cameraCenter) <= (280 * 280) {
			return true
		}
	case 14:
		s := m.scene.Dict().Get("tutorial1.move", m.inputMode)
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, gmath.Vec{}, s)
		m.scene.AddObject(m.hint)
		return true
	case 15:
		if m.choice.Option.special == specialChoiceMoveColony {
			return true
		}
	case 16:
		return !m.world.colonies[0].IsFlying()
	case 17:
		s := m.scene.Dict().Get("tutorial1.fill_resources")
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, gmath.Vec{}, s)
		m.scene.AddObject(m.hint)
		return true
	case 18:
		if m.world.colonies[0].resources > (maxVisualResources * 0.5) {
			return true
		}
	case 19:
		s := m.scene.Dict().Get("tutorial1.build_action")
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, gmath.Vec{}, s)
		m.scene.AddObject(m.hint)
		return true
	case 20:
		if len(m.world.constructions) != 0 {
			return true
		}
	case 21:
		s := m.scene.Dict().Get("tutorial1.base_construction")
		targetPos := ge.Pos{Base: &m.world.constructions[0].pos, Offset: gmath.Vec{Y: 14}}
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.scene.AddObject(m.hint)
		return true
	case 22:
		if len(m.world.constructions) == 0 || m.world.constructions[0].progress >= 0.2 {
			return true
		}
	case 23:
		s := m.scene.Dict().Get("tutorial1.finish_construction")
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, gmath.Vec{}, s)
		m.scene.AddObject(m.hint)
		return true
	}

	return false
}

func (m *tutorialManager) updateTutorial2() bool {
	switch m.tutorialStep {
	case 0:
		m.world.evolutionEnabled = false
		s := m.scene.Dict().Get("tutorial2.factions_first_choice")
		targetPos := gmath.Vec{X: 812, Y: 224}
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.scene.AddObject(m.hint)
		return true
	case 1:
		if m.choice.Option.text != "" && m.choice.Faction != gamedata.GreenFactionTag {
			m.hint.HideLines()
		}
		return m.choice.Faction == gamedata.GreenFactionTag
	case 2:
		s := m.scene.Dict().Get("tutorial2.factions_second_choice")
		targetPos := gmath.Vec{X: 956, Y: 340}
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.scene.AddObject(m.hint)
		return true
	case 3:
		if m.choice.Option.special == specialChoiceNone && len(m.choice.Option.effects) != 0 {
			return true
		}
	case 4:
		for _, colony := range m.world.colonies {
			drone := colony.agents.Find(searchWorkers|searchFighters, func(a *colonyAgentNode) bool {
				return a.faction != gamedata.NeutralFactionTag && a.mode == agentModeStandby
			})
			if drone != nil {
				m.drone = drone
				drone.AssignMode(agentModePosing, gmath.Vec{}, nil)
				return true
			}
		}
	case 5:
		s := m.scene.Dict().Get("tutorial2.faction_drone")
		targetPos := ge.Pos{Base: &m.drone.spritePos}
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.hint.trackedObject = m.drone
		m.scene.AddObject(m.hint)
		m.stepTicks = 5
		return true
	case 6:
		return m.stepTicks == 0
	case 7:
		s := m.scene.Dict().Get("tutorial2.yellow_faction")
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, gmath.Vec{}, s)
		m.scene.AddObject(m.hint)
		m.stepTicks = 3
		return true
	case 8:
		if m.stepTicks > 0 {
			return false
		}
		for _, colony := range m.world.colonies {
			drone := colony.agents.Find(searchWorkers|searchFighters, func(a *colonyAgentNode) bool {
				return a.faction == gamedata.YellowFactionTag && a.mode == agentModeStandby
			})
			if drone != nil {
				m.drone = drone
				drone.AssignMode(agentModePosing, gmath.Vec{}, nil)
				return true
			}
		}
	case 9:
		s := m.scene.Dict().Get("tutorial2.yellow_faction_done")
		targetPos := ge.Pos{Base: &m.drone.spritePos}
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.hint.trackedObject = m.drone
		m.scene.AddObject(m.hint)
		m.stepTicks = 7
		return true
	case 10:
		if m.stepTicks > 0 {
			return false
		}
		for _, colony := range m.world.colonies {
			drone := colony.agents.Find(searchWorkers|searchFighters, func(a *colonyAgentNode) bool {
				return a.faction == gamedata.RedFactionTag && a.mode == agentModeStandby
			})
			if drone != nil {
				m.drone = drone
				drone.AssignMode(agentModePosing, gmath.Vec{}, nil)
				return true
			}
		}
	case 11:
		s := m.scene.Dict().Get("tutorial2.red_faction_done")
		targetPos := ge.Pos{Base: &m.drone.spritePos}
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.hint.trackedObject = m.drone
		m.scene.AddObject(m.hint)
		m.stepTicks = 7
		return true
	case 12:
		if m.stepTicks > 0 {
			return false
		}
		for _, colony := range m.world.colonies {
			drone := colony.agents.Find(searchWorkers|searchFighters, func(a *colonyAgentNode) bool {
				return a.faction == gamedata.GreenFactionTag && a.mode == agentModeStandby
			})
			if drone != nil {
				m.drone = drone
				drone.AssignMode(agentModePosing, gmath.Vec{}, nil)
				return true
			}
		}
	case 13:
		s := m.scene.Dict().Get("tutorial2.green_faction_done")
		targetPos := ge.Pos{Base: &m.drone.spritePos}
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.hint.trackedObject = m.drone
		m.scene.AddObject(m.hint)
		m.stepTicks = 7
		return true
	case 14:
		if m.stepTicks > 0 {
			return false
		}
		for _, colony := range m.world.colonies {
			drone := colony.agents.Find(searchWorkers|searchFighters, func(a *colonyAgentNode) bool {
				return a.faction == gamedata.BlueFactionTag && a.mode == agentModeStandby
			})
			if drone != nil {
				m.drone = drone
				drone.AssignMode(agentModePosing, gmath.Vec{}, nil)
				return true
			}
		}
	case 15:
		m.world.movementEnabled = false
		s := m.scene.Dict().Get("tutorial2.blue_faction_done")
		targetPos := ge.Pos{Base: &m.drone.spritePos}
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.hint.trackedObject = m.drone
		m.scene.AddObject(m.hint)
		m.stepTicks = 7
		return true
	case 16:
		return m.stepTicks == 0
	case 17:
		s := m.scene.Dict().Get("tutorial2.recycle_drones")
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, gmath.Vec{}, s)
		m.scene.AddObject(m.hint)
		m.stepTicks = 5
		for _, colony := range m.world.colonies {
			colony.factionWeights.SetWeight(gamedata.YellowFactionTag, 0)
			colony.factionWeights.SetWeight(gamedata.RedFactionTag, 0)
			colony.factionWeights.SetWeight(gamedata.GreenFactionTag, 0)
			colony.factionWeights.SetWeight(gamedata.BlueFactionTag, 0)
			colony.factionWeights.SetWeight(gamedata.NeutralFactionTag, 1)
			colony.agents.Each(func(a *colonyAgentNode) {
				if a.stats.Tier != 1 {
					return
				}
				a.AssignMode(agentModeRecycleReturn, gmath.Vec{}, nil)
			})
		}
		m.EventRequestPanelUpdate.Emit(gsignal.Void{})
		return true
	case 18:
		return m.stepTicks == 0
	case 19:
		m.world.movementEnabled = true
		s := m.scene.Dict().Get("tutorial2.request_cloner")
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, gmath.Vec{}, s)
		m.scene.AddObject(m.hint)
		m.stepTicks = 6
		return true
	case 20:
		if m.stepTicks > 0 {
			return false
		}
		m.world.evolutionEnabled = true
		for _, colony := range m.world.colonies {
			cloner := colony.agents.Find(searchWorkers, func(a *colonyAgentNode) bool {
				return a.stats.Kind == gamedata.AgentCloner
			})
			if cloner != nil {
				m.drone = cloner
				return true
			}
		}
	case 21:
		s := m.scene.Dict().Get("tutorial2.cloner_ability")
		targetPos := ge.Pos{Base: &m.drone.spritePos}
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.hint.trackedObject = m.drone
		m.scene.AddObject(m.hint)
		m.stepTicks = 5
		return true
	case 22:
		return m.stepTicks == 0
	case 23:
		s := m.scene.Dict().Get("tutorial2.tier3_intro")
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, gmath.Vec{}, s)
		m.scene.AddObject(m.hint)
		m.stepTicks = 5
		return true
	case 24:
		return m.stepTicks == 0
	case 25:
		s := m.scene.Dict().Get("tutorial2.request_destroyer")
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, gmath.Vec{}, s)
		m.scene.AddObject(m.hint)
		m.stepTicks = 5
		return true
	case 26:
		if m.stepTicks > 0 {
			return false
		}
		numFighters := 0
		m.world.colonies[0].agents.Find(searchFighters, func(a *colonyAgentNode) bool {
			if a.stats.Kind == gamedata.AgentFighter {
				numFighters++
			}
			return false
		})
		return numFighters >= 2
	case 27:
		s := m.scene.Dict().Get("tutorial2.evo_points")
		targetPos := ge.Pos{Base: &m.world.colonies[0].spritePos, Offset: gmath.Vec{X: -19, Y: -30}}
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.scene.AddObject(m.hint)
		return true
	case 28:
		return m.world.colonies[0].evoPoints >= blueEvoThreshold
	case 29:
		for _, colony := range m.world.colonies {
			destroyer := colony.agents.Find(searchFighters, func(a *colonyAgentNode) bool {
				return a.stats.Kind == gamedata.AgentDestroyer
			})
			if destroyer != nil {
				m.drone = destroyer
				return true
			}
		}
	case 30:
		s := m.scene.Dict().Get("tutorial2.destroyer_ability")
		targetPos := ge.Pos{Base: &m.drone.spritePos}
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.hint.trackedObject = m.drone
		m.scene.AddObject(m.hint)
		m.stepTicks = 4
		return true
	case 31:
		return m.stepTicks == 0
	case 32:
		m.EventTriggerVictory.Emit(gsignal.Void{})
		return true
	}

	return false
}

func (m *tutorialManager) explainDrone(drone *colonyAgentNode, textKey string) {
	s := m.scene.Dict().Get(textKey)
	targetPos := ge.Pos{Base: &drone.spritePos, Offset: gmath.Vec{Y: -4}}
	m.timedHint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 148}, targetPos, s)
	m.timedHint.trackedObject = drone
	m.timedHint.timed = true
	m.timedHint.time = 20
	m.scene.AddObject(m.timedHint)
}

func (m *tutorialManager) updateTutorial3() bool {
	var freighter *colonyAgentNode
	var servoBot *colonyAgentNode
	var repairBot *colonyAgentNode
	for _, colony := range m.world.colonies {
		colony.agents.Find(searchWorkers, func(a *colonyAgentNode) bool {
			switch a.stats.Kind {
			case gamedata.AgentFreighter:
				freighter = a
			case gamedata.AgentServo:
				servoBot = a
			case gamedata.AgentRepair:
				repairBot = a
			}
			return false
		})
	}
	if freighter != nil && !m.explainedFreighterBots && m.timedHint == nil {
		m.explainDrone(freighter, "tutorial3.hint_freighter")
		m.explainedFreighterBots = true
	}
	if servoBot != nil && !m.explainedServoBots && m.timedHint == nil {
		m.explainDrone(servoBot, "tutorial3.hint_servobot")
		m.explainedServoBots = true
	}
	if repairBot != nil && !m.explainedRepairBots && m.timedHint == nil {
		m.explainDrone(repairBot, "tutorial3.hint_repairbot")
		m.explainedRepairBots = true
	}

	switch m.tutorialStep {
	case 0:
		s := m.scene.Dict().Get("tutorial3.base_select", m.inputMode)
		targetPos := ge.Pos{Base: &m.world.colonies[1].spritePos}
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.scene.AddObject(m.hint)
		return true
	case 1:
		return m.world.selectedColony != m.world.colonies[0]
	case 2:
		s := m.scene.Dict().Get("tutorial3.base_controls")
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, gmath.Vec{}, s)
		m.scene.AddObject(m.hint)
		return true
	case 3:
		if m.choice.Option.special == specialChoiceMoveColony {
			return true
		}
	case 4:
		s := m.scene.Dict().Get("tutorial3.shared_actions")
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, gmath.Vec{}, s)
		m.scene.AddObject(m.hint)
		m.stepTicks = 6
		return true
	case 5:
		return m.stepTicks == 0
	case 6:
		for _, creep := range m.world.creeps {
			switch creep.stats.kind {
			case creepBase, creepTurret:
				// Ignore
			default:
				m.creep = creep
				return true
			}
		}
	case 7:
		s := m.scene.Dict().Get("tutorial3.enemy_drone")
		targetPos := ge.Pos{Base: &m.creep.pos, Offset: gmath.Vec{Y: -4}}
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.hint.trackedObject = m.creep
		m.scene.AddObject(m.hint)
		return true
	case 8:
		return m.creep == nil
	case 9:
		for _, creep := range m.world.creeps {
			if creep.stats.kind == creepBase {
				m.creep = creep
				break
			}
		}
		if m.creep == nil {
			return true
		}
		s := m.scene.Dict().Get("tutorial3.locate_base")
		targetPos := ge.Pos{Base: &m.creep.pos, Offset: gmath.Vec{Y: 14}}
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.hint.trackedObject = m.creep
		m.scene.AddObject(m.hint)
		return true
	case 10:
		if m.creep == nil {
			return true
		}
		targetPos := ge.Pos{Base: &m.creep.pos, Offset: gmath.Vec{Y: 8}}
		cameraCenter := m.world.camera.Offset.Add(m.world.camera.Rect.Center())
		if targetPos.Resolve().DistanceSquaredTo(cameraCenter) <= (280 * 280) {
			return true
		}
	case 11:
		s := m.scene.Dict().Get("tutorial3.attack_action")
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, gmath.Vec{}, s)
		m.scene.AddObject(m.hint)
		m.stepTicks = 6
		return true
	case 12:
		return m.stepTicks == 0
	case 13:
		s := m.scene.Dict().Get("tutorial3.radius_action")
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, gmath.Vec{}, s)
		m.scene.AddObject(m.hint)
		m.stepTicks = 6
		return true
	case 14:
		return m.stepTicks == 0
	case 15:
		s := m.scene.Dict().Get("tutorial3.final_goal")
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, gmath.Vec{}, s)
		m.scene.AddObject(m.hint)
		m.stepTicks = 12
		return true
	case 16:
		return m.stepTicks == 0
	}

	return false
}

func (m *tutorialManager) updateTutorial4() bool {
	if !m.explainedSuperElites && m.timedHint == nil {
		superElite := m.findDrone(searchFighters|searchWorkers, func(a *colonyAgentNode) bool {
			return a.rank == 2
		})
		if superElite != nil {
			m.explainDrone(superElite, "tutorial4.hint_superelite")
			m.explainedSuperElites = true
		}
	}

	switch m.tutorialStep {
	case 0:
		s := m.scene.Dict().Get("tutorial4.locate_boss")
		targetPos := ge.Pos{Base: &m.world.boss.pos}
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.hint.trackedObject = m.world.boss
		m.scene.AddObject(m.hint)
		return true
	case 1:
		targetPos := ge.Pos{Base: &m.world.boss.pos}
		cameraCenter := m.world.camera.Offset.Add(m.world.camera.Rect.Center())
		if targetPos.Resolve().DistanceSquaredTo(cameraCenter) <= (280 * 280) {
			return true
		}
	case 2:
		s := m.scene.Dict().Get("tutorial4.radar")
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, gmath.Vec{}, s)
		m.scene.AddObject(m.hint)
		m.stepTicks = 6
		return true
	case 3:
		return m.stepTicks == 0
	case 4:
		s := m.scene.Dict().Get("tutorial4.boss_warning")
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, gmath.Vec{}, s)
		m.scene.AddObject(m.hint)
		m.stepTicks = 6
		return true
	case 5:
		return m.stepTicks == 0
	case 6:
		var redCrystal *essenceSourceNode
		closestDist := math.MaxFloat64
		for _, e := range m.world.essenceSources {
			if e.stats != redCrystalSource {
				continue

			}
			dist := e.pos.DistanceSquaredTo(m.world.colonies[0].pos)
			if dist < closestDist {
				closestDist = dist
				redCrystal = e
			}
		}
		s := m.scene.Dict().Get("tutorial4.red_crystals")
		targetPos := ge.Pos{Base: &redCrystal.pos, Offset: gmath.Vec{Y: -4}}
		m.resource = redCrystal
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.hint.trackedObject = redCrystal
		m.scene.AddObject(m.hint)
		return true
	case 7:
		return m.resource.IsDisposed()
	case 8:
		for _, colony := range m.world.colonies {
			elite := colony.agents.Find(searchFighters|searchWorkers, func(a *colonyAgentNode) bool {
				return a.rank > 0
			})
			if elite != nil {
				m.drone = elite
				return true
			}
		}
	case 9:
		s := m.scene.Dict().Get("tutorial4.elite_drone")
		targetPos := ge.Pos{Base: &m.drone.spritePos, Offset: gmath.Vec{Y: -6}}
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.hint.trackedObject = m.drone
		m.scene.AddObject(m.hint)
		m.stepTicks = 6
		return true
	case 10:
		return m.stepTicks == 0
	case 11:
		s := m.scene.Dict().Get("tutorial4.red_miner_request")
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, gmath.Vec{}, s)
		m.scene.AddObject(m.hint)
		return true
	case 12:
		for _, colony := range m.world.colonies {
			redminer := colony.agents.Find(searchFighters|searchWorkers, func(a *colonyAgentNode) bool {
				return a.stats.Kind == gamedata.AgentRedminer
			})
			if redminer != nil {
				m.drone = redminer
				return true
			}
		}
	case 13:
		s := m.scene.Dict().Get("tutorial4.red_miner_done")
		targetPos := ge.Pos{Base: &m.drone.spritePos}
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.hint.trackedObject = m.drone
		m.scene.AddObject(m.hint)
		m.stepTicks = 6
		return true
	case 14:
		return m.stepTicks == 0
	case 15:
		var redOil *essenceSourceNode
		closestDist := math.MaxFloat64
		for _, e := range m.world.essenceSources {
			if e.stats != redOilSource {
				continue

			}
			dist := e.pos.DistanceSquaredTo(m.world.colonies[0].pos)
			if dist < closestDist {
				closestDist = dist
				redOil = e
			}
		}
		s := m.scene.Dict().Get("tutorial4.red_oil")
		targetPos := ge.Pos{Base: &redOil.pos}
		m.resource = redOil
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, targetPos, s)
		m.hint.trackedObject = redOil
		m.scene.AddObject(m.hint)
		m.stepTicks = 14
		return true
	case 16:
		return m.stepTicks == 0 ||
			(m.stepTicks < 8 && m.world.colonies[0].pos.DistanceSquaredTo(m.resource.pos) <= (280*280))
	case 17:
		s := m.scene.Dict().Get("tutorial4.final_goal")
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 16, Y: 16}, gmath.Vec{}, s)
		m.scene.AddObject(m.hint)
		m.stepTicks = 12
		return true
	case 18:
		return m.stepTicks == 0
	}

	return false
}

func (m *tutorialManager) findDrone(flags agentSearchFlags, f func(a *colonyAgentNode) bool) *colonyAgentNode {
	for _, colony := range m.world.colonies {
		drone := colony.agents.Find(flags, f)
		if drone != nil {
			return drone
		}
	}
	return nil
}
