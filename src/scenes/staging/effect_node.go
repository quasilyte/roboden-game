package staging

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/gesignal"
	"github.com/quasilyte/gmath"
)

type effectNode struct {
	world *worldState

	pos     gmath.Vec
	image   resource.ImageID
	anim    *ge.Animation
	layer   effectLayer
	rotates bool
	noFlip  bool

	rotation gmath.Rad
	scale    float64

	EventCompleted gesignal.Event[gesignal.Void]
}

func newEffectNodeFromSprite(world *worldState, layer effectLayer, sprite *ge.Sprite) *effectNode {
	e := &effectNode{
		world: world,
		layer: layer,
		anim:  ge.NewAnimation(sprite, -1),
		scale: 1,
	}
	e.anim.SetSecondsPerFrame(0.05)
	return e
}

func newEffectNode(world *worldState, pos gmath.Vec, layer effectLayer, image resource.ImageID) *effectNode {
	return &effectNode{
		world: world,
		pos:   pos,
		image: image,
		layer: layer,
		scale: 1,
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
	sprite.Rotation = &e.rotation
	sprite.SetScale(e.scale, e.scale)
	if !e.noFlip {
		sprite.FlipHorizontal = e.world.localRand.Bool()
	}
	switch e.layer {
	case aboveEffectLayer:
		e.world.stage.AddSpriteAbove(sprite)
	case slightlyAboveEffectLayer:
		e.world.stage.AddSpriteSlightlyAbove(sprite)
	default:
		e.world.stage.AddSprite(sprite)
	}
	if e.anim == nil {
		e.anim = ge.NewAnimation(sprite, -1)
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
	if e.rotates {
		e.rotation += gmath.Rad(delta * 2)
	}
}
