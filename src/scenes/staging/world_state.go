package staging

import (
	"github.com/quasilyte/colony-game/viewport"
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/gmath"
)

type worldState struct {
	rand *gmath.Rand

	camera *viewport.Camera

	essenceSources    []*essenceSourceNode
	creeps            []*creepNode
	colonies          []*colonyCoreNode
	coreConstructions []*colonyCoreConstructionNode

	width  float64
	height float64
	rect   gmath.Rect

	tmpAgentSlice []*colonyAgentNode
}

func (w *worldState) NewColonyCoreNode(config colonyConfig) *colonyCoreNode {
	n := newColonyCoreNode(config)
	w.colonies = append(w.colonies, n)
	return n
}

func (w *worldState) NewColonyCoreConstructionNode(pos gmath.Vec) *colonyCoreConstructionNode {
	n := newColonyCoreConstructionNode(w, pos)
	w.coreConstructions = append(w.coreConstructions, n)
	return n
}

func (w *worldState) NewCreepNode(pos gmath.Vec, stats *creepStats) *creepNode {
	n := newCreepNode(w, stats, pos)
	n.EventDestroyed.Connect(nil, func(x *creepNode) {
		w.creeps = xslices.Remove(w.creeps, x)
	})
	w.creeps = append(w.creeps, n)
	return n
}

func (w *worldState) NewEssenceSourceNode(stats *essenceSourceStats, pos gmath.Vec) *essenceSourceNode {
	n := newEssenceSourceNode(w.camera, stats, pos)
	n.EventDestroyed.Connect(nil, func(x *essenceSourceNode) {
		w.essenceSources = xslices.Remove(w.essenceSources, x)
	})
	w.essenceSources = append(w.essenceSources, n)
	return n
}

func (w *worldState) findColonyAgent(agents []*colonyAgentNode, pos gmath.Vec, r float64, skipIdling bool, f func(a *colonyAgentNode) bool) *colonyAgentNode {
	if len(agents) == 0 {
		return nil
	}
	var slider gmath.Slider
	slider.SetBounds(0, len(agents)-1)
	slider.TrySetValue(w.rand.IntRange(0, len(agents)-1))
	for i := 0; i < len(agents); i++ {
		slider.Inc()
		a := agents[slider.Value()]
		if skipIdling && (a.mode == agentModeCharging || a.mode == agentModeStandby) {
			continue
		}
		dist := a.pos.DistanceTo(pos)
		if dist > r {
			continue
		}
		if f(a) {
			return a
		}
	}
	return nil
}

func (w *worldState) FindColonyAgent(pos gmath.Vec, r float64, f func(a *colonyAgentNode) bool) *colonyAgentNode {
	for _, c := range w.colonies {
		skipIdling := c.body.Pos.DistanceTo(pos)*0.75 > r
		if a := w.findColonyAgent(c.combatAgents, pos, r, skipIdling, f); a != nil {
			return a
		}
		if a := w.findColonyAgent(c.agents, pos, r, skipIdling, f); a != nil {
			return a
		}
	}
	return nil
}
