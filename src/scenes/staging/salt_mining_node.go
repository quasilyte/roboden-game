package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
)

type sulfurMiningNode struct {
	disposed    bool
	miners      []*colonyAgentNode
	minersArray [3]*colonyAgentNode
	source      *essenceSourceNode
}

func newSulfurMiningNode(source *essenceSourceNode) *sulfurMiningNode {
	n := &sulfurMiningNode{source: source}
	n.miners = n.minersArray[:0]
	return n
}

func (n *sulfurMiningNode) Init(scene *ge.Scene) {
	n.source.beingHarvested = true
}

func (n *sulfurMiningNode) IsDisposed() bool {
	return n.disposed
}

func (n *sulfurMiningNode) Dispose() {
	for _, a := range n.miners {
		a.AssignMode(agentModeStandby, gmath.Vec{}, nil)
	}
	n.source.beingHarvested = false
	n.disposed = true
}

func (n *sulfurMiningNode) Update(delta float64) {
	miners := n.miners[:0]
	for _, a := range n.miners {
		if a.IsDisposed() || a.mode != agentModeMineSulfurEssence {
			continue
		}
		miners = append(miners, a)
	}
	n.miners = miners

	if len(n.miners) == 0 {
		n.Dispose()
		return
	}
}
