package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
)

const siegeTurretAmmo = 6

type siegeTurretNode struct {
	sprite *ge.Sprite
	world  *worldState
	pos    gmath.Vec

	ammo   int
	target *creepNode
}

func newSiegeTurretNode(world *worldState, pos gmath.Vec) *siegeTurretNode {
	return &siegeTurretNode{
		world: world,
		pos:   pos,
		ammo:  siegeTurretAmmo,
	}
}

func (t *siegeTurretNode) Init(scene *ge.Scene) {
	turret := scene.NewSprite(assets.ImageSiegeAgentTurret)
	turret.Pos.Base = &t.pos
	turret.Pos.Offset = gmath.Vec{Y: -7}
	t.world.stage.AddSprite(turret)
	t.sprite = turret
}

func (t *siegeTurretNode) Dispose() {
	t.sprite.Dispose()
}

func (t *siegeTurretNode) IsDisposed() bool {
	return t.sprite.IsDisposed()
}

func (t *siegeTurretNode) SetRotation(angle gmath.Rad) gmath.Vec {
	frame, fireOffset := findTurretFrame(angle)
	t.sprite.FrameOffset.X = t.sprite.FrameWidth * frame
	return fireOffset
}
