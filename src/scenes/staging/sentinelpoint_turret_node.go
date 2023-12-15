package staging

import (
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
)

type sentinelpointTurretNode struct {
	sprite *ge.Sprite
	world  *worldState
	pos    gmath.Vec

	worker *colonyAgentNode
}

func newSentinelpointTurretNode(world *worldState, pos gmath.Vec) *sentinelpointTurretNode {
	return &sentinelpointTurretNode{
		world: world,
		pos:   pos,
	}
}

func (t *sentinelpointTurretNode) Init(scene *ge.Scene) {
	turret := scene.NewSprite(assets.ImageSentinelpointAgentTurret)
	turret.Pos.Base = &t.pos
	turret.Pos.Offset = gmath.Vec{Y: -7}
	t.world.stage.AddSprite(turret)
	t.sprite = turret
}

func (t *sentinelpointTurretNode) Dispose() {
	t.sprite.Dispose()
}

func (t *sentinelpointTurretNode) IsDisposed() bool {
	return t.sprite.IsDisposed()
}

func (t *sentinelpointTurretNode) SetRotation(angle gmath.Rad) gmath.Vec {
	frame, fireOffset := findTurretFrame(angle)
	t.sprite.FrameOffset.X = t.sprite.FrameWidth * frame
	return fireOffset
}
