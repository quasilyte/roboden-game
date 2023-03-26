package staging

import (
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
	stepTicks    int

	hint *tutorialHintNode

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

		{
			updateFunc: m.updateTutorial5,
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
		if m.hint != nil {
			m.hint.HideLines()
		}
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
	switch m.tutorialStep {
	case 0:
		s := m.scene.Dict().Get("tutorial1.your_colony")
		targetPos := ge.Pos{Base: &m.world.colonies[0].spritePos}
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, targetPos, s)
		m.scene.AddObject(m.hint)
		return true
	case 1:
		if m.choice.Option.special == specialChoiceMoveColony {
			return true
		}
	case 2:
		return !m.world.colonies[0].IsFlying()
	case 3:
		// Explain the action cards.
		s := m.scene.Dict().Get("tutorial1.action_cards", m.inputMode)
		targetPos := gmath.Vec{X: 812, Y: 50}
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, targetPos, s)
		m.scene.AddObject(m.hint)
		return true
	case 4:
		ok := (m.choice.Option.special != specialChoiceNone && m.choice.Option.special != specialChoiceMoveColony) ||
			(m.choice.Option.special == specialChoiceNone && len(m.choice.Option.effects) != 0)
		if ok {
			return true
		}
	case 5:
		// Explain the resources priority.
		s := m.scene.Dict().Get("tutorial1.resources_priority")
		targetPos := gmath.Vec{X: 812 + (36 * 0), Y: 516}
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, targetPos, s)
		m.scene.AddObject(m.hint)
		return true
	case 6:
		for _, effect := range m.choice.Option.effects {
			if effect.priority == priorityResources {
				return true
			}
		}
	case 7:
		// Explain the growth priority.
		s := m.scene.Dict().Get("tutorial1.growth_priority")
		targetPos := gmath.Vec{X: 812 + (36 * 1), Y: 516}
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, targetPos, s)
		m.scene.AddObject(m.hint)
		return true
	case 8:
		for _, effect := range m.choice.Option.effects {
			if effect.priority == priorityGrowth {
				return true
			}
		}
	case 9:
		// Explain the evolution priority.
		s := m.scene.Dict().Get("tutorial1.evolution_priority")
		targetPos := gmath.Vec{X: 812 + (36 * 2), Y: 516}
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, targetPos, s)
		m.scene.AddObject(m.hint)
		return true
	case 10:
		for _, effect := range m.choice.Option.effects {
			if effect.priority == priorityEvolution {
				return true
			}
		}
	case 11:
		// Explain the security priority.
		s := m.scene.Dict().Get("tutorial1.security_priority")
		targetPos := gmath.Vec{X: 812 + (36 * 3), Y: 516}
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, targetPos, s)
		m.scene.AddObject(m.hint)
		return true
	case 12:
		for _, effect := range m.choice.Option.effects {
			if effect.priority == prioritySecurity {
				return true
			}
		}
	case 13:
		s := m.scene.Dict().Get("tutorial1.camera", m.inputMode)
		targetPos := ge.Pos{Offset: gmath.Vec{X: 1540, Y: 420}}
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, targetPos, s)
		m.scene.AddObject(m.hint)
		return true
	case 14:
		targetPos := ge.Pos{Offset: gmath.Vec{X: 1540, Y: 420}}
		cameraCenter := m.world.camera.Offset.Add(m.world.camera.Rect.Center())
		if targetPos.Offset.DistanceSquaredTo(cameraCenter) <= (280 * 280) {
			return true
		}
	case 15:
		s := m.scene.Dict().Get("tutorial1.fill_resources")
		targetPos := ge.Pos{Base: &m.world.colonies[0].spritePos, Offset: gmath.Vec{X: -3, Y: 10}}
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, targetPos, s)
		m.scene.AddObject(m.hint)
		return true
	case 16:
		if m.world.colonies[0].resources > (maxVisualResources * 0.5) {
			return true
		}
	case 17:
		s := m.scene.Dict().Get("tutorial1.build_action")
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, gmath.Vec{}, s)
		m.scene.AddObject(m.hint)
		return true
	case 18:
		if len(m.world.constructions) != 0 {
			return true
		}
	case 19:
		s := m.scene.Dict().Get("tutorial1.base_construction")
		targetPos := ge.Pos{Base: &m.world.constructions[0].pos, Offset: gmath.Vec{Y: 14}}
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, targetPos, s)
		m.scene.AddObject(m.hint)
		return true
	case 20:
		if len(m.world.constructions) == 0 || m.world.constructions[0].progress >= 0.2 {
			return true
		}
	case 21:
		s := m.scene.Dict().Get("tutorial1.finish_construction")
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, gmath.Vec{}, s)
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
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, targetPos, s)
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
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, targetPos, s)
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
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, targetPos, s)
		m.scene.AddObject(m.hint)
		m.stepTicks = 5
		return true
	case 6:
		return m.stepTicks == 0
	case 7:
		s := m.scene.Dict().Get("tutorial2.yellow_faction")
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, gmath.Vec{}, s)
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
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, targetPos, s)
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
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, targetPos, s)
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
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, targetPos, s)
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
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, targetPos, s)
		m.scene.AddObject(m.hint)
		m.stepTicks = 7
		return true
	case 16:
		return m.stepTicks == 0
	case 17:
		s := m.scene.Dict().Get("tutorial2.recycle_drones")
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, gmath.Vec{}, s)
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
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, gmath.Vec{}, s)
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
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, targetPos, s)
		m.scene.AddObject(m.hint)
		m.stepTicks = 5
		return true
	case 22:
		return m.stepTicks == 0
	case 23:
		s := m.scene.Dict().Get("tutorial2.tier3_intro")
		// targetPos := ge.Pos{Base: &m.world.colonies[0].spritePos, Offset: gmath.Vec{X: -19, Y: -30}}
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, gmath.Vec{}, s)
		m.scene.AddObject(m.hint)
		m.stepTicks = 5
		return true
	case 24:
		return m.stepTicks == 0
	case 25:
		s := m.scene.Dict().Get("tutorial2.request_destroyer")
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, gmath.Vec{}, s)
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
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, targetPos, s)
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
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, targetPos, s)
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

func (m *tutorialManager) updateTutorial3() bool { return false }
func (m *tutorialManager) updateTutorial4() bool { return false }
func (m *tutorialManager) updateTutorial5() bool { return false }
