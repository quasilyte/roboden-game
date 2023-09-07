package staging

import (
	"fmt"
	"time"

	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/serverapi"
)

type replayPlayer struct {
	world *worldState

	choiceGen *choiceGenerator

	state *playerState
}

func newReplayPlayer(world *worldState, state *playerState, choiceGen *choiceGenerator) *replayPlayer {
	return &replayPlayer{
		world:     world,
		state:     state,
		choiceGen: choiceGen,
	}
}

func (p *replayPlayer) Init() {
	if p.choiceGen.creepsState == nil {
		p.state.selectedColony = p.state.colonies[0]
	}
}

func (p *replayPlayer) Update(computedDelta, delta float64) {
	for len(p.state.replay) > 0 {
		a := p.state.replay[0]
		if p.world.nodeRunner.ticks > a.Tick {
			panic(errIllegalAction)
		}
		if a.Tick != p.world.nodeRunner.ticks {
			return
		}
		p.state.replay = p.state.replay[1:]

		if p.choiceGen.creepsState == nil {
			if a.SelectedColony < 0 || a.SelectedColony >= len(p.state.colonies) {
				panic(errInvalidColonyIndex)
			}
			if p.world.GetColonyIndex(p.state.selectedColony) != a.SelectedColony {
				p.state.selectedColony = p.state.colonies[a.SelectedColony]
			}
		}

		ok := false
		if a.Kind == serverapi.ActionMove {
			ok = p.choiceGen.TryExecute(p.state.selectedColony, -1, gmath.Vec{X: a.Pos[0], Y: a.Pos[1]})
		} else {
			ok = p.choiceGen.TryExecute(p.state.selectedColony, int(a.Kind)-1, gmath.Vec{})
		}
		if !ok {
			fmt.Println("fail at", a.Tick, time.Second*time.Duration(p.world.nodeRunner.timePlayed), "player=", p.state.id)
			panic(errIllegalAction)
		}
	}
}

func (p *replayPlayer) GetState() *playerState { return p.state }
