package staging

import (
	"math"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
)

type servantSpawnerNode struct {
	rotateClockwise bool
	target          *colonyCoreNode
	sprite          *ge.Sprite
	rotation        gmath.Rad
	pos             gmath.Vec
	dir             gmath.Vec
	world           *worldState
	super           bool
}

func newServantSpawnerNode(world *worldState, pos, dir gmath.Vec, target *colonyCoreNode) *servantSpawnerNode {
	return &servantSpawnerNode{
		world:  world,
		pos:    pos,
		dir:    dir,
		target: target,
	}
}

func (n *servantSpawnerNode) IsDisposed() bool {
	return n.sprite.IsDisposed()
}

func (n *servantSpawnerNode) Init(scene *ge.Scene) {
	n.rotateClockwise = scene.Rand().Bool()

	n.sprite = scene.NewSprite(assets.ImageServantCreep)
	if n.super {
		n.sprite.FrameOffset.Y = n.sprite.FrameHeight
	}
	n.sprite.Rotation = &n.rotation
	n.sprite.Pos.Base = &n.pos
	n.world.stage.AddGraphics(n.sprite)
}

func (n *servantSpawnerNode) spawn() {
	creep := n.world.NewCreepNode(n.pos, servantCreepStats)
	creep.super = n.super
	creep.specialTarget = n.target
	n.world.nodeRunner.AddObject(creep)
	n.sprite.Dispose()
}

func (n *servantSpawnerNode) Update(delta float64) {
	n.pos = n.dir.Mulf(80 * delta).Add(n.pos)

	if n.rotateClockwise {
		n.rotation += gmath.Rad(delta * 8)
		if n.rotation >= 2*math.Pi {
			n.spawn()
		}
	} else {
		n.rotation -= gmath.Rad(delta * 8)
		if n.rotation <= -2*math.Pi {
			n.spawn()
		}
	}
}
