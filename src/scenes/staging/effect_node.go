package staging

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/gesignal"
	"github.com/quasilyte/gmath"
	"github.com/quasilyte/roboden-game/viewport"
)

type effectNode struct {
	camera *viewport.Camera

	pos   gmath.Vec
	image resource.ImageID
	anim  *ge.Animation
	above bool

	EventCompleted gesignal.Event[gesignal.Void]
}

func newEffectNode(camera *viewport.Camera, pos gmath.Vec, above bool, image resource.ImageID) *effectNode {
	return &effectNode{
		camera: camera,
		pos:    pos,
		image:  image,
		above:  above,
	}
}

func (e *effectNode) Init(scene *ge.Scene) {
	s := scene.NewSprite(e.image)
	s.Pos.Base = &e.pos
	if e.above {
		e.camera.AddGraphicsAbove(s)
	} else {
		e.camera.AddGraphics(s)
	}

	e.anim = ge.NewAnimation(s, -1)
	e.anim.SetSecondsPerFrame(0.05)
}

func (e *effectNode) IsDisposed() bool {
	return e.anim.IsDisposed()
}

func (e *effectNode) Dispose() {
	e.anim.Sprite().Dispose()
}

func (e *effectNode) Update(delta float64) {
	if e.anim.Tick(delta) {
		e.EventCompleted.Emit(gesignal.Void{})
		e.Dispose()
		return
	}
}
