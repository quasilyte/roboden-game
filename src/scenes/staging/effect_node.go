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

	scale float64

	EventCompleted gesignal.Event[gesignal.Void]
}

func newEffectNodeFromSprite(camera *viewport.Camera, above bool, sprite *ge.Sprite) *effectNode {
	e := &effectNode{
		camera: camera,
		above:  above,
		anim:   ge.NewAnimation(sprite, -1),
		scale:  1,
	}
	e.anim.SetSecondsPerFrame(0.05)
	return e
}

func newEffectNode(camera *viewport.Camera, pos gmath.Vec, above bool, image resource.ImageID) *effectNode {
	return &effectNode{
		camera: camera,
		pos:    pos,
		image:  image,
		above:  above,
		scale:  1,
	}
}

func (e *effectNode) Init(scene *ge.Scene) {
	var sprite *ge.Sprite
	if e.anim == nil {
		sprite = scene.NewSprite(e.image)
		sprite.Pos.Base = &e.pos
	} else {
		sprite = e.anim.Sprite()
	}
	sprite.Scale = e.scale
	if e.above {
		e.camera.AddSpriteAbove(sprite)
	} else {
		e.camera.AddSprite(sprite)
	}
	if e.anim == nil {
		e.anim = ge.NewAnimation(sprite, -1)
		e.anim.SetSecondsPerFrame(0.05)
	}
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
