package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/input"
	"github.com/quasilyte/gmath"
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

	hint *tutorialHintNode

	updateDelay float64

	runner *tutorialRunner
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
	m.updateDelay = gmath.ClampMin(m.updateDelay-delta, 0)
	if m.updateDelay != 0 {
		return
	}
	m.updateDelay = m.scene.Rand().FloatRange(1.5, 4.5)

	m.choice = selectedChoice{}
	m.runUpdateFunc()
}

func (m *tutorialManager) runUpdateFunc() {
	if m.runner.updateFunc() {
		m.tutorialStep++
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
		targetPos := ge.Pos{Base: &m.world.colonies[0].pos}
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, targetPos, s)
		m.scene.AddObject(m.hint)
		return true
	case 1:
		if m.choice.Option.special == specialChoiceMoveColony {
			m.hint.Dispose()
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
			m.hint.Dispose()
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
				m.hint.Dispose()
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
				m.hint.Dispose()
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
				m.hint.Dispose()
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
				m.hint.Dispose()
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
			m.hint.Dispose()
			return true
		}
	case 15:
		s := m.scene.Dict().Get("tutorial1.fill_resources")
		targetPos := ge.Pos{Base: &m.world.colonies[0].pos, Offset: gmath.Vec{X: -3, Y: 10}}
		m.hint = newWorldTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, targetPos, s)
		m.scene.AddObject(m.hint)
		return true
	case 16:
		if m.world.colonies[0].resources > (maxVisualResources * 0.5) {
			m.hint.Dispose()
			return true
		}
	case 17:
		s := m.scene.Dict().Get("tutorial1.build_action")
		m.hint = newScreenTutorialHintNode(m.world.camera, gmath.Vec{X: 14, Y: 160}, gmath.Vec{}, s)
		m.scene.AddObject(m.hint)
		return true
	case 18:
		if len(m.world.constructions) != 0 {
			m.hint.Dispose()
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
			m.hint.Dispose()
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

func (m *tutorialManager) updateTutorial2() bool { return false }
func (m *tutorialManager) updateTutorial3() bool { return false }
func (m *tutorialManager) updateTutorial4() bool { return false }
func (m *tutorialManager) updateTutorial5() bool { return false }
