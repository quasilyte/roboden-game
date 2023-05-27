package staging

import (
	"github.com/quasilyte/ge/xslices"
	"github.com/quasilyte/roboden-game/gamedata"
	"github.com/quasilyte/roboden-game/serverapi"
)

type player interface {
	Init()
	Update(delta float64)
	HandleInput()
	GetState() *playerState
}

type playerState struct {
	id int

	colonies []*colonyCoreNode

	selectedColony *colonyCoreNode

	replay []serverapi.PlayerAction

	camera *cameraManager

	hasRoombas bool
}

func newPlayerState(camera *cameraManager) *playerState {
	pstate := &playerState{
		colonies: make([]*colonyCoreNode, 0, 1),
		camera:   camera,
	}

	return pstate
}

func (pstate *playerState) Init(world *worldState) {
	pstate.hasRoombas = xslices.Contains(world.tier2recipes, gamedata.FindRecipe(gamedata.RoombaAgentStats))
}

func getUnitColony(u any) *colonyCoreNode {
	switch u := u.(type) {
	case *colonyAgentNode:
		return u.colonyCore
	case *colonyCoreNode:
		return u
	default:
		return nil
	}
}
