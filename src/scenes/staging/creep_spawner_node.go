package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/gamedata"
)

type creepSpawnerNode struct {
	scene      *ge.Scene
	world      *worldState
	delay      float64
	pos        gmath.Vec
	creepDest  gmath.Vec
	creepStats *gamedata.CreepStats
	fragScore  int
	super      bool
	disposed   bool
}

func newCreepSpawnerNode(world *worldState, delay float64, pos, dest gmath.Vec, stats *gamedata.CreepStats) *creepSpawnerNode {
	return &creepSpawnerNode{
		world:      world,
		delay:      delay,
		pos:        pos,
		creepDest:  dest,
		creepStats: stats,
	}
}

func (spawner *creepSpawnerNode) Init(scene *ge.Scene) {
	spawner.scene = scene
}

func (spawner *creepSpawnerNode) IsDisposed() bool {
	return spawner.disposed
}

func (spawner *creepSpawnerNode) Update(delta float64) {
	if spawner.disposed {
		return
	}

	spawner.delay -= delta
	if spawner.delay <= 0 {
		spawner.disposed = true
		creep := spawner.world.NewCreepNode(spawner.pos, spawner.creepStats)
		creep.super = spawner.super
		spawner.world.nodeRunner.AddObject(creep)
		creep.SendTo(spawner.creepDest)
		creep.fragScore = spawner.fragScore
		return
	}
}
