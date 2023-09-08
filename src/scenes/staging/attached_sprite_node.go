package staging

import (
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge"
	"github.com/quasilyte/gmath"
)

type attachedSpriteNode struct {
	world    *worldState
	owner    ge.SceneObject
	rotation gmath.Rad
	rotates  bool
	pos      ge.Pos
	image    resource.ImageID
	lifetime float64
	sprite   *ge.Sprite
}

func newAttachedSpriteNode(world *worldState, owner ge.SceneObject, lifetime float64, pos ge.Pos, rotates bool, image resource.ImageID) *attachedSpriteNode {
	return &attachedSpriteNode{
		world:    world,
		owner:    owner,
		pos:      pos,
		lifetime: lifetime,
		rotates:  rotates,
		image:    image,
	}
}

func (s *attachedSpriteNode) Init(scene *ge.Scene) {
	s.sprite = scene.NewSprite(s.image)
	s.sprite.Pos = s.pos
	s.sprite.Rotation = &s.rotation
	s.world.stage.AddSpriteAbove(s.sprite)

	if s.rotates {
		s.rotation = s.world.localRand.Rad()
	}
}

func (s *attachedSpriteNode) IsDisposed() bool {
	return s.sprite.IsDisposed()
}

func (s *attachedSpriteNode) dispose() {
	s.sprite.Dispose()
}

func (s *attachedSpriteNode) Update(delta float64) {
	if s.owner.IsDisposed() {
		s.dispose()
		return
	}

	s.lifetime -= delta
	if s.lifetime <= 0 {
		s.dispose()
		return
	}

	if s.rotates {
		s.rotation += gmath.Rad(delta * 8)
	}
}
