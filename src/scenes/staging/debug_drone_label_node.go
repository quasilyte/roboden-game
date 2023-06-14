package staging

import (
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/assets"
)

type debugDroneLabelNode struct {
	drone  *colonyAgentNode
	label  *ge.Label
	pstate *playerState

	hpRendered float64
}

func newDebugDroneLabelNode(pstate *playerState, drone *colonyAgentNode) *debugDroneLabelNode {
	return &debugDroneLabelNode{
		pstate: pstate,
		drone:  drone,
	}
}

func (l *debugDroneLabelNode) IsDisposed() bool {
	return l.label.IsDisposed()
}

func (l *debugDroneLabelNode) dispose() {
	l.label.Dispose()
}

func (l *debugDroneLabelNode) Init(scene *ge.Scene) {
	l.drone.EventDestroyed.Connect(l, func(*colonyAgentNode) {
		l.dispose()
	})

	l.label = ge.NewLabel(assets.BitmapFont1)
	l.label.Pos.Base = &l.drone.spritePos
	l.label.Width = 32
	l.label.Height = 24
	l.label.Pos.Offset.X = -l.label.Width * 0.5
	l.label.Pos.Offset.Y = 6
	l.label.AlignHorizontal = ge.AlignHorizontalCenter
	l.label.AlignVertical = ge.AlignVerticalCenter
	l.label.ColorScale.SetRGBA(0x9d, 0xd7, 0x93, 200)
	l.pstate.camera.Private.AddGraphicsAbove(l)
}

func (l *debugDroneLabelNode) BoundsRect() gmath.Rect {
	return l.drone.sprite.BoundsRect()
}

func (l *debugDroneLabelNode) DrawWithOffset(dst *ebiten.Image, offset gmath.Vec) {
	l.label.DrawWithOffset(dst, offset)
}

func (l *debugDroneLabelNode) Update(delta float64) {
	if l.drone.health == l.hpRendered {
		return
	}
	l.hpRendered = l.drone.health
	l.label.Text = strconv.Itoa(int(l.hpRendered))
}
